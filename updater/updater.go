package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Updater struct {
	config          *UpdateConfig
	versionService  *VersionService
	downloadService *DownloadService
	verifier        *Verifier
	state           *UpdateState
	progressCh      chan DownloadProgress
	doneCh          chan struct{}
	wailsApp        *application.App
}

func NewUpdater(config *UpdateConfig) *Updater {
	updater := &Updater{
		config:          config,
		versionService:  NewVersionService(config),
		downloadService: NewDownloadService(config),
		verifier:        NewVerifier(config.PublicKeyPath),
		state: &UpdateState{
			State:          UpdateStateIdle,
			CurrentVersion: AppVersion,
		},
		progressCh: make(chan DownloadProgress, 100),
		doneCh:     make(chan struct{}),
	}

	updater.downloadService.SetProgressChannel(updater.progressCh)

	go updater.cleanupOldFiles()

	return updater
}

func (u *Updater) SetWailsApp(app *application.App) {
	u.wailsApp = app
}

func (u *Updater) cleanupOldFiles() {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[Updater] 无法获取可执行文件路径: %v", err)
		return
	}

	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)
	oldPath := filepath.Join(exeDir, exeName+".old")

	if _, err := os.Stat(oldPath); err == nil {
		log.Printf("[Updater] 清理旧文件: %s", oldPath)

		for i := 1; i <= 10; i++ {
			if err := os.Remove(oldPath); err == nil {
				log.Printf("[Updater] 已清理旧文件")
				return
			}
			log.Printf("[Updater] 等待清理旧文件 (%d/10): %v", i, err)
			time.Sleep(500 * time.Millisecond)
		}

		log.Printf("[Updater] 无法清理旧文件，将在下下次启动时重试")
	}
}

func (u *UpdateState) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"state":            u.State,
		"currentVersion":   u.CurrentVersion,
		"latestVersion":    u.LatestVersion,
		"downloadProgress": u.DownloadProgress,
		"errorMessage":     u.ErrorMessage,
		"updateLog":        u.UpdateLog,
	}
}

func (u *Updater) StartAutoCheck() {
	if !u.config.AutoCheck {
		log.Printf("[Updater] 自动检查已禁用")
		return
	}

	log.Printf("[Updater] 启动自动检查定时器, 间隔: %d秒", u.config.CheckInterval)

	go func() {
		ticker := time.NewTicker(time.Duration(u.config.CheckInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-u.doneCh:
				log.Printf("[Updater] 停止自动检查")
				return
			case <-ticker.C:
				u.checkForUpdate()
			}
		}
	}()
}

func (u *Updater) Stop() {
	select {
	case <-u.doneCh:
		return
	default:
		close(u.doneCh)
	}
}

func (u *Updater) GetState() *UpdateState {
	return u.state
}

func (u *Updater) GetConfig() *UpdateConfig {
	return u.config
}

func (u *Updater) UpdateConfig(newConfig *UpdateConfig) error {
	u.config.AutoCheck = newConfig.AutoCheck
	u.config.AutoUpdate = newConfig.AutoUpdate
	u.config.CheckInterval = newConfig.CheckInterval
	u.config.ProxyURL = newConfig.ProxyURL
	u.config.Channel = newConfig.Channel

	return u.config.Save()
}

func (u *Updater) CheckForUpdate() (*ReleaseAsset, error) {
	u.state.State = UpdateStateChecking
	u.state.ErrorMessage = ""

	releaseAsset, err := u.versionService.CheckForUpdate(AppVersion, u.config.Channel)
	if err != nil {
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = err.Error()
		return nil, err
	}

	if releaseAsset != nil {
		u.state.LatestVersion = releaseAsset.Version
		u.state.State = UpdateStateIdle
		u.config.LastCheckTime = time.Now()
		u.config.LastCheckedVersion = releaseAsset.Version
		u.config.Save()
		return releaseAsset, nil
	}

	u.state.State = UpdateStateIdle
	u.config.LastCheckTime = time.Now()
	u.config.Save()
	return nil, nil
}

func (u *Updater) checkForUpdate() {
	asset, err := u.CheckForUpdate()
	if err != nil {
		log.Printf("[Updater] 自动检查失败: %v", err)
		return
	}

	if asset != nil && u.config.AutoUpdate {
		log.Printf("[Updater] 自动下载更新包")
		u.StartUpdate()
	}
}

func (u *Updater) StartUpdate() error {
	asset, err := u.CheckForUpdate()
	if err != nil {
		return err
	}

	if asset == nil {
		return fmt.Errorf("当前已是最新版本")
	}

	go u.monitorDownloadProgress()
	return u.performUpdate(asset)
}

func (u *Updater) monitorDownloadProgress() {
	for {
		select {
		case <-u.doneCh:
			return
		case progress, ok := <-u.progressCh:
			if !ok {
				return
			}
			u.state.DownloadProgress = progress.Progress
		}
	}
}

func (u *Updater) performUpdate(asset *ReleaseAsset) error {
	log.Printf("[Updater] 开始执行更新流程, 版本: %s", asset.Version)

	tmpDir := u.downloadService.GetTempDir()
	defer u.downloadService.CleanupTempDir()

	updateDir := filepath.Join(tmpDir, "update")
	os.MkdirAll(updateDir, 0755)

	zipPath := filepath.Join(updateDir, "update.zip")

	u.state.State = UpdateStateDownloading
	u.state.DownloadProgress = 0

	log.Printf("[Updater] 下载更新包: %s", asset.DownloadURL)
	log.Printf("[Updater] 临时目录: %s", zipPath)
	if err := u.downloadService.DownloadFile(asset.DownloadURL, zipPath); err != nil {
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = fmt.Sprintf("下载失败: %v", err)
		log.Printf("[Updater] 下载失败: %v", err)
		return err
	}
	log.Printf("[Updater] 下载完成，大小: %d bytes", getFileSize(zipPath))

	log.Printf("[Updater] 下载完成, 开始校验")
	u.state.State = UpdateStateVerifying

	if asset.Checksum != "" {
		valid, err := u.verifier.VerifyFile(zipPath, asset.Checksum)
		if err != nil || !valid {
			u.state.State = UpdateStateFailed
			u.state.ErrorMessage = "校验失败: 更新包可能已损坏"
			log.Printf("[Updater] 校验失败: %v", err)
			return fmt.Errorf("校验失败")
		}
		log.Printf("[Updater] 校验通过")
	} else {
		log.Printf("[Updater] 跳过校验（无校验信息）")
	}

	u.state.State = UpdateStateInstalling

	log.Printf("[Updater] 开始安装更新")
	if err := u.extractAndInstall(zipPath, updateDir, asset); err != nil {
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = fmt.Sprintf("安装失败: %v", err)
		return err
	}

	u.state.State = UpdateStateComplete
	u.state.UpdateLog = fmt.Sprintf("更新完成\n版本: %s\n时间: %s", asset.Version, time.Now().Format(time.RFC3339))

	log.Printf("[Updater] 更新流程完成，准备重启应用")

	u.config.LastCheckTime = time.Now()
	u.config.Save()

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return info.Size()
}

func (u *Updater) extractAndInstall(zipPath, updateDir string, asset *ReleaseAsset) error {
	log.Printf("[Updater] 解压更新包: %s", zipPath)

	zipFile, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %w", err)
	}
	defer zipFile.Close()

	for _, file := range zipFile.File {
		path := filepath.Join(updateDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(path), 0755)

		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("解压文件失败: %s: %w", file.Name, err)
		}

		dst, err := os.Create(path)
		if err != nil {
			src.Close()
			return fmt.Errorf("创建文件失败: %s: %w", file.Name, err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return fmt.Errorf("写入文件失败: %s: %w", file.Name, err)
		}

		src.Close()
		dst.Close()

		log.Printf("[Updater] 解压文件: %s", file.Name)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	log.Printf("[Updater] 复制文件到安装目录: %s", exeDir)

	if err := u.copyUpdateFilesWithRename(updateDir, exeDir, filepath.Base(exePath)); err != nil {
		return err
	}

	// Ensure the main executable has correct permissions, especially on macOS
	log.Printf("[Updater] 设置主程序可执行权限: %s", exePath)
	if err := os.Chmod(exePath, 0755); err != nil {
		return fmt.Errorf("设置可执行权限失败: %w", err)
	}

	return nil
}

func (u *Updater) copyUpdateFilesWithRename(srcDir, dstDir, exeName string) error {
	// Determine the fallback executable name based on OS
	fallbackExeName := "MYT"
	if runtime.GOOS == "windows" {
		fallbackExeName = "MYT.exe"
	}

	// Track if we found and updated the executable
	//exeUpdated := false

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(srcDir, path)
		if relPath == "." || relPath == "" {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Check if this is the executable we need to update
		isExe := relPath == exeName || relPath == fallbackExeName
		if isExe {
			// Use the actual destination executable name, not the source name
			finalDstPath := filepath.Join(dstDir, exeName)
			if err := u.renameAndCopyExe(path, finalDstPath); err != nil {
				return err
			}
			//exeUpdated = true
			return nil
		}

		return u.copyFile(path, dstPath)
	})
}

func (u *Updater) renameAndCopyExe(src, dst string) error {
	oldPath := dst + ".old"

	log.Printf("[Updater] 重命名旧文件: %s -> %s", dst, oldPath)

	if _, err := os.Stat(dst); err == nil {
		if err := os.Rename(dst, oldPath); err != nil {
			log.Printf("[Updater] 重命名失败，尝试删除: %v", err)
			if delErr := os.Remove(dst); delErr != nil {
				return fmt.Errorf("无法删除旧文件: %w", delErr)
			}
			log.Printf("[Updater] 已删除旧文件: %s", dst)
		} else {
			log.Printf("[Updater] 已重命名旧文件到: %s", oldPath)
		}
	}

	log.Printf("[Updater] 复制新文件: %s -> %s", src, dst)
	if err := u.copyFile(src, dst); err != nil {
		return fmt.Errorf("复制新文件失败: %w", err)
	}

	log.Printf("[Updater] 文件更新完成: %s", dst)
	return nil
}

func (u *Updater) copyFile(src, dst string) error {
	// Get source file info to preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Preserve file permissions from source
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}

func (u *Updater) RestartApp() error {
	log.Printf("[Updater] 准备重启应用")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	log.Printf("[Updater] 使用可执行文件路径: %s", exePath)

	args := os.Args[1:]

	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(exePath)

	// 注意：不要在 Windows 上使用 HideWindow=true，否则会导致新进程窗口被隐藏
	// 如果需要隐藏窗口，应该在程序启动时通过命令行参数控制，而不是在这里设置

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动新进程失败: %w", err)
	}

	log.Printf("[Updater] 新进程已启动，退出当前应用")

	if u.wailsApp != nil {
		u.wailsApp.Quit()
	}

	return nil
}

func (u *Updater) GetUpdateLog() string {
	return u.state.UpdateLog
}

func (u *Updater) SubscribeProgress() <-chan DownloadProgress {
	return u.progressCh
}

func (u *Updater) WriteUpdateLog(message string) {
	u.state.UpdateLog += message + "\n"
	log.Printf("[Updater] %s", message)

	if u.config.UpdateLogPath != "" {
		logFile, err := os.OpenFile(u.config.UpdateLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			defer logFile.Close()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
		}
	}
}

func (u *Updater) ClearUpdateLog() {
	u.state.UpdateLog = ""

	if u.config.UpdateLogPath != "" {
		os.Remove(u.config.UpdateLogPath)
	}
}

func (u *Updater) GetInstallDir() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[Updater] 获取安装目录失败: %v", err)
		return ""
	}
	return filepath.Dir(exePath)
}

func (u *Updater) NeedsElevation() bool {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[Updater] 获取可执行文件路径失败: %v", err)
		return false
	}

	exeDir := filepath.Dir(exePath)

	programFiles := []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("ProgramW6432"),
	}

	for _, pf := range programFiles {
		if pf != "" && strings.HasPrefix(exeDir, pf) {
			log.Printf("[Updater] 需要UAC提权: 程序安装在 %s", exeDir)
			return true
		}
	}

	log.Printf("[Updater] 不需要UAC提权: 程序安装在 %s", exeDir)
	return false
}

func (u *Updater) StartElevatedUpdate(asset *ReleaseAsset) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	tmpDir := u.downloadService.GetTempDir()
	zipPath := filepath.Join(tmpDir, "update.zip")

	u.state.State = UpdateStateDownloading
	u.state.DownloadProgress = 0

	go u.monitorDownloadProgress()

	log.Printf("[Updater] 下载更新包: %s", asset.DownloadURL)
	log.Printf("[Updater] 临时目录: %s", zipPath)
	if err := u.downloadService.DownloadFile(asset.DownloadURL, zipPath); err != nil {
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = fmt.Sprintf("下载失败: %v", err)
		return err
	}

	log.Printf("[Updater] 下载完成, 开始校验")
	u.state.State = UpdateStateVerifying

	if asset.Checksum != "" {
		valid, err := u.verifier.VerifyFile(zipPath, asset.Checksum)
		if err != nil || !valid {
			u.state.State = UpdateStateFailed
			u.state.ErrorMessage = "校验失败: 更新包可能已损坏"
			return fmt.Errorf("校验失败")
		}
		log.Printf("[Updater] 校验通过")
	}

	log.Printf("[Updater] 准备启动更新...")
	log.Printf("[Updater] 可执行文件: %s", exePath)
	log.Printf("[Updater] ZIP路径: %s", zipPath)
	log.Printf("[Updater] 工作目录: %s", exeDir)
	log.Printf("[Updater] ZIP文件存在: %v", fileExists(zipPath))
	log.Printf("[Updater] ZIP文件大小: %d bytes", getFileSize(zipPath))

	args := []string{
		"--update",
		"--zip", zipPath,
		"--restart", exePath,
		"--version", asset.Version,
	}

	// 使用平台特定的更新启动函数
	return runUpdateWithElevation(u, exePath, zipPath, exeDir, args, asset)
}
