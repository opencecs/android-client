package main

import (
	"archive/zip"
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/webplayer
var webplayerAssets embed.FS

//go:embed all:player_dist
var playerAssets embed.FS

func main() {
	debugMode := flag.Bool("debug", false, "Enable debug mode with context menu")

	updateMode := flag.Bool("update", false, "Run in elevated update mode")
	restartFromUpdate := flag.Bool("restart-from-update", false, "Restarted from elevated update")
	zipPath := flag.String("zip", "", "Update zip path for elevated update")
	restartPath := flag.String("restart", "", "Path to restart after elevated update")
	updateVersion := flag.String("version", "", "New version for elevated update")

	flag.Parse()

	// ========== 日志输出到文件（dev/debug 模式）==========
	// Windows 下 wails3 dev 没有控制台，log 写到 logs/app.log 方便实时查看
	if err := os.MkdirAll("logs", 0755); err == nil {
		logFilePath := filepath.Join("logs", "app.log")
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			multi := io.MultiWriter(os.Stderr, logFile)
			log.SetOutput(multi)
			log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
			log.Printf("[Main] 日志同步写入: %s", logFilePath)
		}
	}

	envDebug := os.Getenv("APP_DEBUG")
	enableDebug := *debugMode || strings.ToLower(envDebug) == "true" || envDebug == "1" || IsDevBuild()

	if enableDebug {
		log.Printf("[Main] Debug mode enabled")
	}

	exePath, _ := os.Executable()

	if *updateMode && *zipPath != "" && *restartPath != "" {
		log.Printf("[Main] === 提权更新模式 ===")
		log.Printf("[Main] 接收参数: zip=%s, restart=%s, version=%s", *zipPath, *restartPath, *updateVersion)

		executeElevatedUpdate(*zipPath, *restartPath, *updateVersion)

		log.Printf("[Main] 更新完成，准备重启主程序...")
		time.Sleep(1 * time.Second)

		args := []string{"--restart-from-update", "--version", *updateVersion}
		cmd := exec.Command(exePath, args...)
		cmd.Dir = filepath.Dir(exePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Printf("[Main] 重启失败: %v", err)
			fmt.Println("重启失败，请手动启动程序")
		} else {
			log.Printf("[Main] 重启命令已发送")
		}

		os.Exit(0)
		return
	}

	if *restartFromUpdate {
		log.Printf("[Main] === 从提权更新重启 ===")
		log.Printf("[Main] 新版本: %s", *updateVersion)
		fmt.Printf("更新完成！当前版本: %s\n", *updateVersion)
	}

	appService := NewApp()
	appService.startup()
	updaterService := NewUpdaterService()

	mainHandler := application.AssetFileServerFS(assets)

	webplayerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var contentType string
		var err error

		path := strings.TrimPrefix(r.URL.Path, "/webplayer")
		if path == "" {
			path = "frontend/webplayer/play.html"
		} else {
			path = "frontend/webplayer" + path
		}
		data, err = webplayerAssets.ReadFile(path)
		if strings.HasSuffix(path, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(path, ".svg") {
			contentType = "image/svg+xml"
		}

		if err != nil {
			log.Printf("[Main] 文件未找到: %s, error: %v", r.URL.Path, err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		w.Write(data)
	})

	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/webplayer") {
			webplayerHandler.ServeHTTP(w, r)
		} else {
			mainHandler.ServeHTTP(w, r)
		}
	})

	wailsApp := application.New(application.Options{
		Name:        "魔云腾-V3-客户端",
		Description: "ARM边缘计算设备管理客户端",
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(updaterService),
		},
		Assets: application.AssetOptions{
			Handler: combinedHandler,
		},
	})

	appService.SetWailsApp(wailsApp)
	updaterService.SetWailsApp(wailsApp)

	mainWindow := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:                      "魔云腾-V3-客户端",
		Width:                      1460,
		Height:                     945,
		BackgroundColour:           application.NewRGB(27, 38, 54),
		URL:                        "/",
		DevToolsEnabled:            enableDebug,
		OpenInspectorOnStartup:     false,
		DefaultContextMenuDisabled: !enableDebug,
	})

	mainWindow.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		log.Println("[Main] 主窗口关闭，强制退出应用...")
		appService.CleanupProjectionWindows()
		appService.CleanupProjectionProcesses()
		appService.StopAllP2P()
		wailsApp.Quit()
	})

	err := wailsApp.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func executeElevatedUpdate(zipPath, restartPath, version string) {
	log.Printf("[ElevatedUpdate] ========== 开始执行提权更新 ==========")
	log.Printf("[ElevatedUpdate] ZIP路径: %s", zipPath)
	log.Printf("[ElevatedUpdate] 重启路径: %s", restartPath)
	log.Printf("[ElevatedUpdate] 新版本: %s", version)
	log.Printf("[ElevatedUpdate] ZIP文件存在: %v", fileExists(zipPath))

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[ElevatedUpdate] 获取可执行文件路径失败: %v", err)
		return
	}
	exeName := filepath.Base(exePath)
	log.Printf("[ElevatedUpdate] 当前程序: %s", exePath)
	log.Printf("[ElevatedUpdate] 程序名称: %s", exeName)

	updateDir := filepath.Join(os.TempDir(), "edgeclient-elevated-update")
	cleanupDir := filepath.Dir(zipPath)
	defer os.RemoveAll(updateDir)
	defer os.RemoveAll(cleanupDir)

	log.Printf("[ElevatedUpdate] 更新临时目录: %s", updateDir)
	log.Printf("[ElevatedUpdate] 清理临时目录: %s", cleanupDir)
	os.MkdirAll(updateDir, 0755)

	log.Printf("[ElevatedUpdate] 打开压缩包...")
	zipFile, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Printf("[ElevatedUpdate] 打开压缩包失败: %v", err)
		return
	}
	defer zipFile.Close()
	log.Printf("[ElevatedUpdate] 压缩包包含 %d 个文件", len(zipFile.File))

	for _, file := range zipFile.File {
		dst := filepath.Join(updateDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(dst, 0755)
			log.Printf("[ElevatedUpdate] 创建目录: %s", file.Name)
			continue
		}

		os.MkdirAll(filepath.Dir(dst), 0755)

		src, err := file.Open()
		if err != nil {
			log.Printf("[ElevatedUpdate] 打开文件失败 %s: %v", file.Name, err)
			continue
		}

		dstFile, err := os.Create(dst)
		if err != nil {
			src.Close()
			log.Printf("[ElevatedUpdate] 创建文件失败 %s: %v", file.Name, err)
			continue
		}

		if _, err := io.Copy(dstFile, src); err != nil {
			log.Printf("[ElevatedUpdate] 写入文件失败 %s: %v", file.Name, err)
		} else {
			log.Printf("[ElevatedUpdate] 解压文件: %s", file.Name)
		}

		src.Close()
		dstFile.Close()

		os.Chmod(dst, 0755)
	}

	newExePath := filepath.Join(updateDir, exeName)
	if _, err := os.Stat(newExePath); err != nil {
		log.Printf("[ElevatedUpdate] 未找到主程序文件，查找可执行文件: %v", err)
		for _, f := range zipFile.File {
			if strings.HasSuffix(f.Name, ".exe") {
				newExePath = filepath.Join(updateDir, f.Name)
				break
			}
		}
	}

	log.Printf("[ElevatedUpdate] 目标程序路径: %s", newExePath)

	if _, err := os.Stat(newExePath); err != nil {
		log.Printf("[ElevatedUpdate] 目标程序不存在: %v", err)
		return
	}

	oldPath := exePath + ".old"
	log.Printf("[ElevatedUpdate] 重命名旧文件: %s -> %s", exePath, oldPath)

	if _, err := os.Stat(exePath); err == nil {
		if err := os.Rename(exePath, oldPath); err != nil {
			log.Printf("[ElevatedUpdate] 重命名失败，尝试删除: %v", err)
			if delErr := os.Remove(exePath); delErr != nil {
				log.Printf("[ElevatedUpdate] 删除旧文件失败: %v", delErr)
				return
			}
			log.Printf("[ElevatedUpdate] 已删除旧文件")
		} else {
			log.Printf("[ElevatedUpdate] 已重命名旧文件")
		}
	}

	log.Printf("[ElevatedUpdate] 复制新文件: %s -> %s", newExePath, exePath)
	if err := copyFile(newExePath, exePath); err != nil {
		log.Printf("[ElevatedUpdate] 复制新文件失败: %v", err)
		return
	}

	os.Chmod(exePath, 0755)
	log.Printf("[ElevatedUpdate] 设置权限完成")

	log.Printf("[ElevatedUpdate] 提权更新完成")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
