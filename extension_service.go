package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// 扩展服务 - SSH连接配置
const (
	extensionSSHUser     = "user"
	extensionSSHPassword = "myt"
	extensionSSHPort     = 22
)

// mytPanelUpdateResponse 远程更新检查接口返回结构
type mytPanelUpdateResponse struct {
	CodeID int `json:"code_id"`
	Msg    string `json:"msg"`
	Data   struct {
		Changelog     string `json:"changelog"`
		DownloadURL   string `json:"download_url"`
		FileSize      int64  `json:"file_size"`
		ForceUpdate   bool   `json:"force_update"`
		LatestVersion string `json:"latest_version"`
		SHA256        string `json:"sha256"`
	} `json:"data"`
}

// InstallMytPanel 安装魔云互联(myt-panel)到指定设备
// 流程：查询最新版本 → 下载到本地 → SSH上传 → 注册系统服务 → 启动
func (a *App) InstallMytPanel(deviceIP string) map[string]interface{} {
	log.Printf("[扩展服务] 开始安装魔云互联到设备: %s", deviceIP)

	ip := extractPureIP(deviceIP)

	// 1. 查询远程最新版本（device参数固定为r1s）
	downloadURL, latestVersion, sha256Hash, err := checkMytPanelUpdate()
	if err != nil {
		log.Printf("[扩展服务] 查询最新版本失败: %v", err)
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("查询最新版本失败: %v", err)}
	}
	log.Printf("[扩展服务] 最新版本: %s, 下载地址: %s", latestVersion, downloadURL)

	// 3. 下载到本地临时目录
	localDir := filepath.Join(os.TempDir(), "myt-panel-install")
	os.RemoveAll(localDir)
	os.MkdirAll(localDir, 0755)
	defer os.RemoveAll(localDir) // 安装完成后清理

	binaryPath := filepath.Join(localDir, "myt-panel")
	log.Printf("[扩展服务] 开始下载到: %s", binaryPath)
	if err := downloadFile(downloadURL, binaryPath); err != nil {
		log.Printf("[扩展服务] 下载失败: %v", err)
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("下载安装包失败: %v", err)}
	}

	// 写入sha256文件
	if sha256Hash != "" {
		os.WriteFile(filepath.Join(localDir, "myt-panel.sha256"), []byte(sha256Hash), 0644)
	}

	// 4. 生成服务配置文件到本地 deploy 目录
	deployDir := filepath.Join(localDir, "deploy")
	os.MkdirAll(deployDir, 0755)
	generateAlpineOpenRC(deployDir)
	generateDebianSystemd(deployDir)

	// 5. 建立SSH连接
	sshClient, err := dialSSH(ip)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接失败: %v", err)}
	}
	defer sshClient.Close()

	// 停止已运行的服务
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S sh -c 'rc-service myt-panel stop 2>/dev/null; systemctl stop myt-panel 2>/dev/null; killall myt-panel 2>/dev/null'", extensionSSHPassword))
	time.Sleep(1 * time.Second)

	// 清理旧文件（可能属于root，需要sudo删除）
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -f /home/user/myt-panel /home/user/myt-panel.sha256", extensionSSHPassword))
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -rf /home/user/deploy", extensionSSHPassword))

	// 6. SFTP上传文件
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SFTP连接失败: %v", err)}
	}
	defer sftpClient.Close()

	sftpClient.MkdirAll("/home/user/deploy")

	if err := sftpUploadFile(sftpClient, binaryPath, "/home/user/myt-panel", 0755); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传myt-panel失败: %v", err)}
	}
	sha256Path := filepath.Join(localDir, "myt-panel.sha256")
	if _, err := os.Stat(sha256Path); err == nil {
		sftpUploadFile(sftpClient, sha256Path, "/home/user/myt-panel.sha256", 0644)
	}
	if err := sftpUploadDir(sftpClient, deployDir, "/home/user/deploy"); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传部署脚本失败: %v", err)}
	}

	// 7. 检测OS并注册系统服务
	osType, _ := detectDeviceOS(sshClient)
	log.Printf("[扩展服务] 设备 %s OS类型: %s", ip, osType)

	if osType == "alpine" {
		err = installAlpineService(sshClient)
	} else {
		err = installDebianService(sshClient)
	}
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	// 8. 健康检查
	panelURL := fmt.Sprintf("http://%s:8081", ip)
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		output, _ := runSSHCmd(sshClient, "netstat -tlnp 2>/dev/null | grep 8081 || ss -tlnp | grep 8081")
		if strings.Contains(output, "8081") {
			log.Printf("[扩展服务] 魔云互联安装成功，访问地址: %s", panelURL)
			return map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("魔云互联 v%s 安装成功", latestVersion),
				"url":     panelURL,
			}
		}
		// 重试启动
		if osType == "alpine" {
			runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rc-service myt-panel start 2>/dev/null", extensionSSHPassword))
		} else {
			runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S systemctl start myt-panel 2>/dev/null", extensionSSHPassword))
		}
	}

	log.Printf("[扩展服务] 服务启动超时")
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("魔云互联 v%s 安装完成，服务启动中，请稍后访问 %s", latestVersion, panelURL),
		"url":     panelURL,
	}
}

// checkMytPanelUpdate 查询远程最新版本（device固定为r1s）
func checkMytPanelUpdate() (downloadURL, latestVersion, sha256 string, err error) {
	apiURL := "https://newapi.moyunteng.com/api/v1/update/check?device=r1s&arch=arm64&version=0"
	log.Printf("[扩展服务] 查询最新版本: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", "", fmt.Errorf("请求更新接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result mytPanelUpdateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.CodeID != 200 || result.Data.DownloadURL == "" {
		return "", "", "", fmt.Errorf("未找到可用更新 (code_id=%d, msg=%s)", result.CodeID, result.Msg)
	}

	return result.Data.DownloadURL, result.Data.LatestVersion, result.Data.SHA256, nil
}

// UninstallMytPanel 卸载魔云互联(myt-panel)
func (a *App) UninstallMytPanel(deviceIP string) map[string]interface{} {
	log.Printf("[扩展服务] 开始卸载魔云互联: %s", deviceIP)
	ip := extractPureIP(deviceIP)

	sshClient, err := dialSSH(ip)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接失败: %v", err)}
	}
	defer sshClient.Close()

	osType, _ := detectDeviceOS(sshClient)

	if osType == "alpine" {
		// 停止服务 → 移除开机启动 → 删除服务文件
		steps := []struct {
			desc string
			cmd  string
		}{
			{"停止服务", fmt.Sprintf("echo '%s' | sudo -S rc-service myt-panel stop 2>/dev/null || true", extensionSSHPassword)},
			{"移除开机启动", fmt.Sprintf("echo '%s' | sudo -S rc-update del myt-panel default 2>/dev/null || true", extensionSSHPassword)},
			{"删除服务文件", fmt.Sprintf("echo '%s' | sudo -S rm -f /etc/init.d/myt-panel", extensionSSHPassword)},
		}
		for _, step := range steps {
			output, err := sshExecCommandWithOutput(sshClient, step.cmd)
			if err != nil && !strings.Contains(output, "not found") && !strings.Contains(output, "not installed") {
				log.Printf("[扩展服务] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
			}
		}
	} else {
		steps := []struct {
			desc string
			cmd  string
		}{
			{"停止服务", fmt.Sprintf("echo '%s' | sudo -S systemctl stop myt-panel 2>/dev/null || true", extensionSSHPassword)},
			{"禁用服务", fmt.Sprintf("echo '%s' | sudo -S systemctl disable myt-panel 2>/dev/null || true", extensionSSHPassword)},
			{"删除服务文件", fmt.Sprintf("echo '%s' | sudo -S rm -f /etc/systemd/system/myt-panel.service", extensionSSHPassword)},
			{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", extensionSSHPassword)},
		}
		for _, step := range steps {
			output, err := sshExecCommandWithOutput(sshClient, step.cmd)
			if err != nil && !strings.Contains(output, "not found") {
				log.Printf("[扩展服务] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
			}
		}
	}

	// 清理文件
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -f /home/user/myt-panel /home/user/myt-panel.sha256", extensionSSHPassword))
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -rf /home/user/deploy /home/user/logs", extensionSSHPassword))

	log.Printf("[扩展服务] 魔云互联卸载成功")
	return map[string]interface{}{
		"success": true,
		"message": "魔云互联卸载成功",
	}
}

// downloadFile 下载文件到本地
func downloadFile(url, filePath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	log.Printf("[扩展服务] 下载完成，大小: %d bytes", written)
	return nil
}

// generateAlpineOpenRC 生成Alpine OpenRC服务文件
func generateAlpineOpenRC(dir string) {
	content := `#!/sbin/openrc-run
name="myt-panel"
description="MYT Cloud Phone Panel"

command="/home/user/myt-panel"
command_args="-port 8081"
command_user="root"
command_background=true
pidfile="/run/${RC_SVCNAME}.pid"
directory="/home/user"

output_log="/home/user/logs/myt-panel.log"
error_log="/home/user/logs/myt-panel.log"

respawn_delay=3
respawn_max=0

depend() {
    need net
    after firewall
}

start_pre() {
    mkdir -p /home/user/logs
}
`
	os.WriteFile(filepath.Join(dir, "alpine-openrc"), []byte(content), 0755)
}

// generateDebianSystemd 生成Debian systemd服务文件
func generateDebianSystemd(dir string) {
	content := `[Unit]
Description=MYT Cloud Phone Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/user
ExecStart=/home/user/myt-panel -port 8081

Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
	os.WriteFile(filepath.Join(dir, "debian-systemd.service"), []byte(content), 0644)
}

// installAlpineService 注册Alpine系统服务
func installAlpineService(sshClient *ssh.Client) error {
	steps := []struct {
		desc string
		cmd  string
	}{
		{"复制服务文件", fmt.Sprintf("echo '%s' | sudo -S sh -c 'cp /home/user/deploy/alpine-openrc /etc/init.d/myt-panel && sed -i \"s/\\r$//\" /etc/init.d/myt-panel && chmod +x /etc/init.d/myt-panel'", extensionSSHPassword)},
		{"创建日志目录", fmt.Sprintf("echo '%s' | sudo -S mkdir -p /home/user/logs", extensionSSHPassword)},
		{"注册开机启动", fmt.Sprintf("echo '%s' | sudo -S rc-update add myt-panel default 2>/dev/null || true", extensionSSHPassword)},
		{"启动服务", fmt.Sprintf("echo '%s' | sudo -S rc-service myt-panel start", extensionSSHPassword)},
	}
	return runSteps(sshClient, steps)
}

// installDebianService 注册Debian系统服务
func installDebianService(sshClient *ssh.Client) error {
	steps := []struct {
		desc string
		cmd  string
	}{
		{"复制服务文件", fmt.Sprintf("echo '%s' | sudo -S sh -c 'cp /home/user/deploy/debian-systemd.service /etc/systemd/system/myt-panel.service && sed -i \"s/\\r$//\" /etc/systemd/system/myt-panel.service'", extensionSSHPassword)},
		{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", extensionSSHPassword)},
		{"注册开机启动", fmt.Sprintf("echo '%s' | sudo -S systemctl enable myt-panel", extensionSSHPassword)},
		{"创建日志目录", fmt.Sprintf("echo '%s' | sudo -S mkdir -p /home/user/logs", extensionSSHPassword)},
		{"启动服务", fmt.Sprintf("echo '%s' | sudo -S systemctl start myt-panel", extensionSSHPassword)},
	}
	return runSteps(sshClient, steps)
}

// runSteps 逐步执行命令
func runSteps(sshClient *ssh.Client, steps []struct {
	desc string
	cmd  string
},
) error {
	for _, step := range steps {
		output, err := sshExecCommandWithOutput(sshClient, step.cmd)
		if err != nil {
			errMsg := fmt.Sprintf("%s失败", step.desc)
			if output != "" {
				cleanOutput := strings.TrimSpace(strings.ReplaceAll(output, "[sudo] password for user:", ""))
				if cleanOutput != "" {
					errMsg = fmt.Sprintf("%s失败: %s", step.desc, cleanOutput)
				}
			}
			log.Printf("[扩展服务] %s", errMsg)
			return fmt.Errorf("%s", errMsg)
		}
		log.Printf("[扩展服务] %s成功", step.desc)
	}
	return nil
}

func runSSHCmd(client *ssh.Client, cmd string) (string, error) {
	return sshExecCommandWithOutput(client, cmd)
}

// ---- 通用工具函数 ----

func extractPureIP(ip string) string {
	if ip == "" {
		return ip
	}
	lastColon := strings.LastIndex(ip, ":")
	if lastColon == -1 {
		return ip
	}
	afterColon := ip[lastColon+1:]
	if isAllDigits(afterColon) {
		return ip[:lastColon]
	}
	return ip
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// findLatestVersionDir 保留用于本地fallback查找
func findLatestVersionDir(releaseBaseDir string) (string, error) {
	var versionDirs []string
	err := filepath.WalkDir(releaseBaseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		binaryPath := filepath.Join(path, "myt-panel")
		if info, err := os.Stat(binaryPath); err == nil && !info.IsDir() {
			versionDirs = append(versionDirs, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("遍历release目录失败: %w", err)
	}
	if len(versionDirs) == 0 {
		return "", fmt.Errorf("未找到myt-panel安装文件，请确认 %s 目录下有版本文件", releaseBaseDir)
	}
	sort.Strings(versionDirs)
	return versionDirs[len(versionDirs)-1], nil
}

func dialSSH(deviceIP string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            extensionSSHUser,
		Auth:            []ssh.AuthMethod{ssh.Password(extensionSSHPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", deviceIP, extensionSSHPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接 %s 失败: %w", addr, err)
	}
	return client, nil
}

func sshExecCommandWithOutput(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH session失败: %w", err)
	}
	defer session.Close()

	// 设置命令超时，避免长时间阻塞
	done := make(chan struct{})
	var output []byte
	var execErr error
	go func() {
		output, execErr = session.CombinedOutput(cmd)
		close(done)
	}()

	select {
	case <-done:
		if execErr != nil {
			return string(output), fmt.Errorf("执行命令失败 [%s]: %w", cmd, execErr)
		}
		return string(output), nil
	case <-time.After(5 * time.Minute):
		session.Close()
		return "", fmt.Errorf("执行命令超时 [%s]", cmd)
	}
}

func detectDeviceOS(client *ssh.Client) (string, error) {
	output, err := sshExecCommandWithOutput(client, "cat /etc/os-release 2>/dev/null | grep -i alpine || echo 'NOT_ALPINE'")
	if err != nil {
		output2, err2 := sshExecCommandWithOutput(client, "which apk 2>/dev/null || echo 'NO_APK'")
		if err2 != nil {
			return "debian", nil
		}
		if strings.Contains(output2, "apk") && !strings.Contains(output2, "NO_APK") {
			return "alpine", nil
		}
		return "debian", nil
	}
	if strings.Contains(strings.ToLower(output), "alpine") {
		return "alpine", nil
	}
	return "debian", nil
}

func sftpUploadFile(client *sftp.Client, localPath string, remotePath string, perm fs.FileMode) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败 %s: %w", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败 %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	if _, err := remoteFile.ReadFrom(localFile); err != nil {
		return fmt.Errorf("写入远程文件失败 %s: %w", remotePath, err)
	}
	if err := client.Chmod(remotePath, perm); err != nil {
		log.Printf("[扩展服务] 设置文件权限失败 %s: %v (非致命)", remotePath, err)
	}
	log.Printf("[扩展服务] SFTP上传成功: %s -> %s (权限: %04o)", localPath, remotePath, perm)
	return nil
}

func sftpUploadDir(client *sftp.Client, localDir string, remoteDir string) error {
	if err := client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败 %s: %w", remoteDir, err)
	}
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return fmt.Errorf("读取本地目录失败 %s: %w", localDir, err)
	}
	for _, entry := range entries {
		localPath := filepath.Join(localDir, entry.Name())
		remotePath := fmt.Sprintf("%s/%s", remoteDir, entry.Name())
		if entry.IsDir() {
			if err := sftpUploadDir(client, localPath, remotePath); err != nil {
				return err
			}
		} else {
			perm := fs.FileMode(0644)
			if strings.HasSuffix(entry.Name(), ".sh") {
				perm = 0755
			}
			if err := sftpUploadFile(client, localPath, remotePath, perm); err != nil {
				return err
			}
		}
	}
	return nil
}
