package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// InstallTunnel 安装公网穿透(frpc)到指定设备
// 流程：解压frpc → 生成配置（含SSH代理）→ SSH上传 → 注册系统服务 → 启动
func (a *App) InstallTunnel(deviceIP string, serverAddr string, serverPort int, token string) map[string]interface{} {
	log.Printf("[公网穿透] 开始安装到设备: %s, 服务器: %s:%d", deviceIP, serverAddr, serverPort)

	ip := extractPureIP(deviceIP)

	// 1. 从远程下载frpc zip包到临时目录
	localDir := filepath.Join(os.TempDir(), "frpc-install")
	os.RemoveAll(localDir)
	os.MkdirAll(localDir, 0755)
	defer os.RemoveAll(localDir)

	frpcDownloadURL := "http://47.107.33.172:10011/api/v1/moyu/download/frpc"
	localZipPath := filepath.Join(localDir, "frpc.zip")
	log.Printf("[公网穿透] 从远程下载frpc: %s", frpcDownloadURL)

	if err := downloadFile(frpcDownloadURL, localZipPath); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("下载frpc失败: %v", err)}
	}

	if err := extractFRPCFromZip(localZipPath, localDir); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("解压frpc失败: %v", err)}
	}

	// 2. 生成 frpc.toml 配置（含SSH代理规则，暴露设备22端口）
	configContent := generateFrpcConfig(serverAddr, serverPort, token, ip)
	configPath := filepath.Join(localDir, "frpc.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("生成配置失败: %v", err)}
	}

	// 3. 生成 systemd/openrc 服务文件
	deployDir := filepath.Join(localDir, "deploy")
	os.MkdirAll(deployDir, 0755)
	generateTunnelAlpineOpenRC(deployDir)
	generateTunnelDebianSystemd(deployDir)

	// 4. SSH连接
	sshClient, err := dialSSH(ip)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接失败: %v", err)}
	}
	defer sshClient.Close()

	// 停止已有服务
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S sh -c 'rc-service frpc stop 2>/dev/null; systemctl stop frpc 2>/dev/null; killall frpc 2>/dev/null'", extensionSSHPassword))
	time.Sleep(1 * time.Second)

	// 清理旧文件
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -f /home/user/frpc /home/user/frpc.toml", extensionSSHPassword))
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -rf /home/user/deploy-frpc", extensionSSHPassword))

	// 5. SFTP上传
	sftpClient, err := sftpNewClient(sshClient)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SFTP连接失败: %v", err)}
	}
	defer sftpClient.Close()

	sftpClient.MkdirAll("/home/user/deploy-frpc")

	if err := sftpUploadFile(sftpClient, filepath.Join(localDir, "frpc"), "/home/user/frpc", 0755); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传frpc失败: %v", err)}
	}
	if err := sftpUploadFile(sftpClient, configPath, "/home/user/frpc.toml", 0644); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传配置文件失败: %v", err)}
	}
	if err := sftpUploadDir(sftpClient, deployDir, "/home/user/deploy-frpc"); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传部署脚本失败: %v", err)}
	}

	// 6. 注册并启动服务
	osType, _ := detectDeviceOS(sshClient)
	log.Printf("[公网穿透] 设备 %s OS类型: %s", ip, osType)

	if osType == "alpine" {
		err = installTunnelAlpineService(sshClient)
	} else {
		err = installTunnelDebianService(sshClient)
	}
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	// 等待frpc启动并查看日志
	time.Sleep(3 * time.Second)
	frpcLog, _ := runSSHCmd(sshClient, "tail -20 /home/user/logs/frpc.log 2>/dev/null || echo no-log")
		log.Printf("[公网穿透] frpc日志:\n%s", frpcLog)

	// 检查frpc进程是否真正运行（重试3次）
	frpcRunning := false
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)
		psOutput, _ := runSSHCmd(sshClient, "ps aux | grep /home/user/frpc | grep -v grep")
		if strings.TrimSpace(psOutput) != "" {
			frpcRunning = true
			break
		}
	}
	if !frpcRunning {
		log.Printf("[公网穿透] frpc未运行，可能架构不匹配")
		return map[string]interface{}{"success": false, "message": "frpc启动失败，可能架构不匹配，请确认frpc为ARM64版本"}
	}
	// 查询frps获取实际分配的远程端口
	remoteAddress := ""
	webAddress := ""
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		dashboardURL := fmt.Sprintf("http://%s:7500/api/proxy/tcp", serverAddr)
		req, _ := http.NewRequest("GET", dashboardURL, nil)
		req.SetBasicAuth("admin", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var result struct {
			Proxies []struct {
				Name string `json:"name"`
				Conf struct {
					RemotePort int `json:"remotePort"`
				} `json:"conf"`
			} `json:"proxies"`
		}
		if json.Unmarshal(body, &result) == nil {
			for _, p := range result.Proxies {
				if p.Name == "ssh" {
					remoteAddress = fmt.Sprintf("%s:%d", serverAddr, p.Conf.RemotePort)
				}
				if p.Name == "frpc-web" {
					webAddress = fmt.Sprintf("http://%s:%d", serverAddr, p.Conf.RemotePort)
				}
			}
		}
		if remoteAddress != "" || webAddress != "" {
			break
		}
	}

	log.Printf("[公网穿透] 安装成功: %s -> %s:%d, SSH远程地址: %s, Web管理地址: %s", deviceIP, serverAddr, serverPort, remoteAddress, webAddress)
	msg := fmt.Sprintf("公网穿透安装成功，已连接到 %s:%d", serverAddr, serverPort)
	if remoteAddress != "" || webAddress != "" {
		var parts []string
		if remoteAddress != "" {
			parts = append(parts, fmt.Sprintf("SSH远程地址: %s", remoteAddress))
		}
		if webAddress != "" {
			parts = append(parts, fmt.Sprintf("Web管理地址: %s", webAddress))
		}
		msg = fmt.Sprintf("公网穿透安装成功，%s", strings.Join(parts, "，"))
	}
	return map[string]interface{}{
		"success":       true,
		"message":       msg,
		"remoteAddress": remoteAddress,
		"webAddress":    webAddress,
	}
	}

// UninstallTunnel 卸载公网穿透(frpc)
func (a *App) UninstallTunnel(deviceIP string) map[string]interface{} {
	log.Printf("[公网穿透] 开始卸载: %s", deviceIP)
	ip := extractPureIP(deviceIP)

	sshClient, err := dialSSH(ip)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接失败: %v", err)}
	}
	defer sshClient.Close()

	osType, _ := detectDeviceOS(sshClient)

	if osType == "alpine" {
		steps := []struct {
			desc string
			cmd  string
		}{
			{"停止服务", fmt.Sprintf("echo '%s' | sudo -S rc-service frpc stop 2>/dev/null || true", extensionSSHPassword)},
			{"强杀进程", fmt.Sprintf("echo '%s' | sudo -S killall frpc 2>/dev/null || true", extensionSSHPassword)},
			{"移除开机启动", fmt.Sprintf("echo '%s' | sudo -S rc-update del frpc default 2>/dev/null || true", extensionSSHPassword)},
			{"删除服务文件", fmt.Sprintf("echo '%s' | sudo -S rm -f /etc/init.d/frpc", extensionSSHPassword)},
		}
		for _, step := range steps {
			output, err := sshExecCommandWithOutput(sshClient, step.cmd)
			if err != nil && !strings.Contains(output, "not found") && !strings.Contains(output, "not installed") {
				log.Printf("[公网穿透] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
			}
		}
	} else {
		steps := []struct {
			desc string
			cmd  string
		}{
			{"停止服务", fmt.Sprintf("echo '%s' | sudo -S systemctl stop frpc 2>/dev/null || true", extensionSSHPassword)},
			{"强杀进程", fmt.Sprintf("echo '%s' | sudo -S killall frpc 2>/dev/null || true", extensionSSHPassword)},
			{"禁用服务", fmt.Sprintf("echo '%s' | sudo -S systemctl disable frpc 2>/dev/null || true", extensionSSHPassword)},
			{"删除服务文件", fmt.Sprintf("echo '%s' | sudo -S rm -f /etc/systemd/system/frpc.service", extensionSSHPassword)},
			{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", extensionSSHPassword)},
		}
		for _, step := range steps {
			output, err := sshExecCommandWithOutput(sshClient, step.cmd)
			if err != nil && !strings.Contains(output, "not found") {
				log.Printf("[公网穿透] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
			}
		}
	}

	// 清理文件
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -f /home/user/frpc /home/user/frpc.toml /home/user/frpc_store.json", extensionSSHPassword))
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -rf /home/user/deploy-frpc", extensionSSHPassword))

	log.Printf("[公网穿透] 卸载成功")
	return map[string]interface{}{
		"success": true,
		"message": "公网穿透卸载成功",
	}
}

// generateFrpcConfig 生成 frpc.toml 配置内容（含SSH代理规则）
// 不指定remotePort，由frps服务端自动分配端口
func generateFrpcConfig(serverAddr string, serverPort int, token string, deviceIP string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("serverAddr = \"%s\"\n", serverAddr))
	sb.WriteString(fmt.Sprintf("serverPort = %d\n", serverPort))

	if token != "" {
		sb.WriteString("\n[auth]\n")
		sb.WriteString("method = \"token\"\n")
		sb.WriteString(fmt.Sprintf("token = \"%s\"\n", token))
	}

	// Web管理界面
	sb.WriteString("\n[webServer]\n")
	sb.WriteString("addr = \"0.0.0.0\"\n")
	sb.WriteString("port = 7400\n")
	sb.WriteString("user = \"admin\"\n")
	sb.WriteString("password = \"admin\"\n")

	// 持久化存储，保持代理状态（避免重启后frps重新分配端口）
	sb.WriteString("\n[store]\n")
	sb.WriteString("path = \"./frpc_store.json\"\n")

	// SSH代理规则：将设备22端口映射到frps服务器（端口由服务端自动分配）
	sb.WriteString(fmt.Sprintf("\n[[proxies]]\n"))
	sb.WriteString("name = \"ssh\"\n")
	sb.WriteString("type = \"tcp\"\n")
	sb.WriteString("localIP = \"127.0.0.1\"\n")
	sb.WriteString("localPort = 22\n")

	// Web管理代理规则：将设备7400端口映射到frps服务器（端口由服务端自动分配）
	sb.WriteString(fmt.Sprintf("\n[[proxies]]\n"))
	sb.WriteString("name = \"frpc-web\"\n")
	sb.WriteString("type = \"tcp\"\n")
	sb.WriteString("localIP = \"127.0.0.1\"\n")
	sb.WriteString("localPort = 7400\n")

	return sb.String()
}

// generateTunnelAlpineOpenRC 生成Alpine OpenRC服务文件
func generateTunnelAlpineOpenRC(dir string) {
	content := `#!/sbin/openrc-run
name="frpc"
description="FRP Client - Public Network Tunnel"

command="/home/user/frpc"
command_args="-c /home/user/frpc.toml"
command_user="root"
command_background=true
pidfile="/run/${RC_SVCNAME}.pid"
directory="/home/user"

output_log="/home/user/logs/frpc.log"
error_log="/home/user/logs/frpc.log"

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

// generateTunnelDebianSystemd 生成Debian systemd服务文件
func generateTunnelDebianSystemd(dir string) {
	content := `[Unit]
Description=FRP Client - Public Network Tunnel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/user
ExecStart=/home/user/frpc -c /home/user/frpc.toml

Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
	os.WriteFile(filepath.Join(dir, "debian-systemd.service"), []byte(content), 0644)
}

// installTunnelAlpineService 注册Alpine系统服务
func installTunnelAlpineService(sshClient *ssh.Client) error {
	steps := []struct {
		desc string
		cmd  string
	}{
		{"复制服务文件", fmt.Sprintf("echo '%s' | sudo -S sh -c 'cp /home/user/deploy-frpc/alpine-openrc /etc/init.d/frpc && sed -i \"s/\\r$//\" /etc/init.d/frpc && chmod +x /etc/init.d/frpc'", extensionSSHPassword)},
		{"创建日志目录", fmt.Sprintf("echo '%s' | sudo -S mkdir -p /home/user/logs", extensionSSHPassword)},
		{"注册开机启动", fmt.Sprintf("echo '%s' | sudo -S rc-update add frpc default 2>/dev/null || true", extensionSSHPassword)},
		{"启动服务", fmt.Sprintf("echo '%s' | sudo -S rc-service frpc start", extensionSSHPassword)},
	}
	return runSteps(sshClient, steps)
}

// installTunnelDebianService 注册Debian系统服务
func installTunnelDebianService(sshClient *ssh.Client) error {
	steps := []struct {
		desc string
		cmd  string
	}{
		{"复制服务文件", fmt.Sprintf("echo '%s' | sudo -S sh -c 'cp /home/user/deploy-frpc/debian-systemd.service /etc/systemd/system/frpc.service && sed -i \"s/\\r$//\" /etc/systemd/system/frpc.service'", extensionSSHPassword)},
		{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", extensionSSHPassword)},
		{"注册开机启动", fmt.Sprintf("echo '%s' | sudo -S systemctl enable frpc", extensionSSHPassword)},
		{"创建日志目录", fmt.Sprintf("echo '%s' | sudo -S mkdir -p /home/user/logs", extensionSSHPassword)},
		{"启动服务", fmt.Sprintf("echo '%s' | sudo -S systemctl start frpc", extensionSSHPassword)},
	}
	return runSteps(sshClient, steps)
}

// extractFRPCFromZip 从本地zip包中解压frpc二进制和配置文件
func extractFRPCFromZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// 只解压frpc相关文件
		baseName := filepath.Base(f.Name)
		if baseName != "frpc" && baseName != "frpc.toml" {
			continue
		}
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("打开zip内文件 %s 失败: %w", f.Name, err)
		}

		dstPath := filepath.Join(targetDir, baseName)
		dst, err := os.Create(dstPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("创建文件 %s 失败: %w", dstPath, err)
		}

		if _, err := io.Copy(dst, rc); err != nil {
			dst.Close()
			rc.Close()
			return fmt.Errorf("写入文件 %s 失败: %w", dstPath, err)
		}
		dst.Close()
		rc.Close()

		if baseName == "frpc" {
			os.Chmod(dstPath, 0755)
		}
		log.Printf("[公网穿透] 解压文件: %s -> %s", f.Name, baseName)
	}
	return nil
}

// sftpNewClient 创建SFTP客户端
func sftpNewClient(sshClient *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(sshClient)
}

// InstallTunnelServer 安装公网穿透服务端(frpcs)到用户指定的服务器
func (a *App) InstallTunnelServer(serverIP string, sshUser string, sshPassword string, sshPort int, bindPort int, dashboardPort int, dashboardUser string, dashboardPassword string) map[string]interface{} {
	// 清理服务器地址中的协议前缀
	serverIP = strings.TrimPrefix(serverIP, "http://")
	serverIP = strings.TrimPrefix(serverIP, "https://")
	serverIP = strings.TrimRight(serverIP, "/")

	log.Printf("[公网穿透服务端] 开始安装: %s", serverIP)

	if serverIP == "" {
		return map[string]interface{}{"success": false, "message": "服务器地址不能为空"}
	}
	if sshUser == "" {
		sshUser = "root"
	}
	if sshPassword == "" {
		return map[string]interface{}{"success": false, "message": "SSH密码不能为空"}
	}
	if sshPort == 0 {
		sshPort = 22
	}
	if bindPort == 0 {
		bindPort = 7000
	}
	if dashboardPort == 0 {
		dashboardPort = 7500
	}
	if dashboardUser == "" {
		dashboardUser = "admin"
	}
	if dashboardPassword == "" {
		dashboardPassword = "admin"
	}

	// 1. 下载frps
	localDir := filepath.Join(os.TempDir(), "frps-install")
	os.RemoveAll(localDir)
	os.MkdirAll(localDir, 0755)
	defer os.RemoveAll(localDir)

	// 从本地下载frps tar.gz，依次尝试多个镜像
	localTarPath := filepath.Join(localDir, "frps.tar.gz")
	mirrors := []string{
		"https://ghfast.top/https://github.com/fatedier/frp/releases/download/v0.68.1/frp_0.68.1_linux_amd64.tar.gz",
		"https://mirror.ghproxy.com/https://github.com/fatedier/frp/releases/download/v0.68.1/frp_0.68.1_linux_amd64.tar.gz",
		"https://github.com/fatedier/frp/releases/download/v0.68.1/frp_0.68.1_linux_amd64.tar.gz",
	}
	var downloadErr error
	for _, mirror := range mirrors {
		log.Printf("[公网穿透服务端] 尝试下载frps: %s", mirror)
		downloadErr = downloadFile(mirror, localTarPath)
		if downloadErr == nil {
			break
		}
		log.Printf("[公网穿透服务端] 下载失败(%s): %v，尝试下一个镜像", mirror, downloadErr)
		os.Remove(localTarPath)
	}
	if downloadErr != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("下载frps失败(已尝试所有镜像): %v", downloadErr)}
	}
	if err := extractFRPSFromTarGz(localTarPath, localDir); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("解压frps失败: %v", err)}
	}

	// 2. 生成 frps.toml 配置
	configContent := generateFrpsConfig(bindPort, dashboardPort, dashboardUser, dashboardPassword)
	configPath := filepath.Join(localDir, "frps.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("生成配置失败: %v", err)}
	}

	// 3. 生成 systemd 服务文件
	deployDir := filepath.Join(localDir, "deploy")
	os.MkdirAll(deployDir, 0755)
	generateFrpsSystemdService(deployDir)

	// 4. SSH连接到用户服务器
	sshConfig := &ssh.ClientConfig{
		User:            sshUser,
		Auth: []ssh.AuthMethod{
				ssh.Password(sshPassword),
				ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
					answers := make([]string, len(questions))
					for i := range answers {
						answers[i] = sshPassword
					}
					return answers, nil
				}),
			},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", serverIP, sshPort)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接 %s 失败: %v", addr, err)}
	}
	defer sshClient.Close()

	// 停止已有frps服务
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S systemctl stop frps 2>/dev/null; echo '%s' | sudo -S killall frps 2>/dev/null; true", sshPassword, sshPassword))
	time.Sleep(2 * time.Second)
	// 确保端口已释放，强制杀掉占用端口的进程
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S fuser -k %d/tcp 2>/dev/null; echo '%s' | sudo -S fuser -k %d/tcp 2>/dev/null; true", sshPassword, bindPort, sshPassword, dashboardPort))
	time.Sleep(1 * time.Second)

	// 清理旧文件
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -f /usr/local/bin/frps /etc/frps/frps.toml", sshPassword))
	runSSHCmd(sshClient, fmt.Sprintf("echo '%s' | sudo -S rm -rf /tmp/deploy-frps", sshPassword))

	// 5. SFTP上传所有文件
	sftpClient, err := sftpNewClient(sshClient)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SFTP连接失败: %v", err)}
	}
	defer sftpClient.Close()

	sftpClient.MkdirAll("/tmp/deploy-frps")
	if err := sftpUploadFile(sftpClient, filepath.Join(localDir, "frps"), "/tmp/deploy-frps/frps", 0755); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传frps失败: %v", err)}
	}
	if err := sftpUploadFile(sftpClient, configPath, "/tmp/deploy-frps/frps.toml", 0644); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传配置文件失败: %v", err)}
	}
	if err := sftpUploadDir(sftpClient, deployDir, "/tmp/deploy-frps"); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("上传服务文件失败: %v", err)}
	}

	// 6. 注册并启动服务
	steps := []struct {
		desc string
		cmd  string
	}{
		{"移动二进制", fmt.Sprintf("echo '%s' | sudo -S cp /tmp/deploy-frps/frps /usr/local/bin/frps && echo '%s' | sudo -S chmod +x /usr/local/bin/frps", sshPassword, sshPassword)},
		{"创建配置目录", fmt.Sprintf("echo '%s' | sudo -S mkdir -p /etc/frps", sshPassword)},
		{"移动配置文件", fmt.Sprintf("echo '%s' | sudo -S cp /tmp/deploy-frps/frps.toml /etc/frps/frps.toml", sshPassword)},
		{"复制服务文件", fmt.Sprintf("echo '%s' | sudo -S sh -c 'cp /tmp/deploy-frps/frps.service /etc/systemd/system/frps.service && sed -i \"s/\r$//\" /etc/systemd/system/frps.service'", sshPassword)},
		{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", sshPassword)},
		{"注册开机启动", fmt.Sprintf("echo '%s' | sudo -S systemctl enable frps", sshPassword)},
		{"启动服务", fmt.Sprintf("echo '%s' | sudo -S systemctl start frps", sshPassword)},
	}
	for _, step := range steps {
		output, err := sshExecCommandWithOutput(sshClient, step.cmd)
		if err != nil {
			log.Printf("[公网穿透服务端] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
		}
	}

// 7. 等待启动并验证
	time.Sleep(3 * time.Second)
	frspRunning := false
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)
		psOutput, _ := sshExecCommandWithOutput(sshClient, "ps aux | grep /usr/local/bin/frps | grep -v grep")
		if strings.TrimSpace(psOutput) != "" {
			frspRunning = true
			break
		}
	}
	if !frspRunning {
			// 输出诊断信息
			runOutput, _ := sshExecCommandWithOutput(sshClient, fmt.Sprintf("echo '%s' | sudo -S timeout 5 /usr/local/bin/frps -c /etc/frps/frps.toml 2>&1 || true", sshPassword))
			log.Printf("[公网穿透服务端] 运行frps输出: %s", runOutput)
			configOutput, _ := sshExecCommandWithOutput(sshClient, "cat /etc/frps/frps.toml")
			log.Printf("[公网穿透服务端] 配置文件: %s", configOutput)
			return map[string]interface{}{"success": false, "message": fmt.Sprintf("frps启动失败\n运行输出: %s", strings.TrimSpace(runOutput))}
		}

		// 检查端口是否监听
	portOutput, _ := sshExecCommandWithOutput(sshClient, fmt.Sprintf("ss -tlnp | grep ':%d '", bindPort))
	if strings.TrimSpace(portOutput) == "" {
		log.Printf("[公网穿透服务端] 端口 %d 未监听", bindPort)
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("frps端口 %d 未监听，可能启动失败", bindPort)}
	}

	dashboardURL := fmt.Sprintf("http://%s:%d", serverIP, dashboardPort)
	log.Printf("[公网穿透服务端] 安装成功: %s, 绑定端口: %d, 管理面板: %s", serverIP, bindPort, dashboardURL)

	return map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("公网穿透服务端安装成功，服务器: %s，绑定端口: %d，管理面板: %s", serverIP, bindPort, dashboardURL),
		"serverAddr":    serverIP,
		"bindPort":      bindPort,
		"dashboardURL":  dashboardURL,
		"dashboardUser": dashboardUser,
	}
}

// UninstallTunnelServer 卸载公网穿透服务端(frps)
func (a *App) UninstallTunnelServer(serverIP string, sshUser string, sshPassword string, sshPort int) map[string]interface{} {
	log.Printf("[公网穿透服务端] 开始卸载: %s", serverIP)

	if sshUser == "" {
		sshUser = "root"
	}
	if sshPort == 0 {
		sshPort = 22
	}

	sshConfig := &ssh.ClientConfig{
		User:            sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = sshPassword
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", serverIP, sshPort)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("SSH连接 %s 失败: %v", addr, err)}
	}
	defer sshClient.Close()

	steps := []struct {
		desc string
		cmd  string
	}{
		{"停止服务", fmt.Sprintf("echo '%s' | sudo -S systemctl stop frps 2>/dev/null || true", sshPassword)},
		{"禁用服务", fmt.Sprintf("echo '%s' | sudo -S systemctl disable frps 2>/dev/null || true", sshPassword)},
		{"删除服务文件", fmt.Sprintf("echo '%s' | sudo -S rm -f /etc/systemd/system/frps.service", sshPassword)},
		{"重载systemd", fmt.Sprintf("echo '%s' | sudo -S systemctl daemon-reload", sshPassword)},
		{"删除二进制", fmt.Sprintf("echo '%s' | sudo -S rm -f /usr/local/bin/frps", sshPassword)},
		{"删除配置", fmt.Sprintf("echo '%s' | sudo -S rm -rf /etc/frps", sshPassword)},
		{"清理临时文件", fmt.Sprintf("echo '%s' | sudo -S rm -rf /tmp/deploy-frps", sshPassword)},
	}

	for _, step := range steps {
		output, err := sshExecCommandWithOutput(sshClient, step.cmd)
		if err != nil && !strings.Contains(output, "not found") {
			log.Printf("[公网穿透服务端] %s: %v (%s)", step.desc, err, strings.TrimSpace(output))
		}
	}

	log.Printf("[公网穿透服务端] 卸载成功")
	return map[string]interface{}{"success": true, "message": "公网穿透服务端卸载成功"}
}

// generateFrpsConfig 生成 frps.toml 配置内容
func generateFrpsConfig(bindPort int, dashboardPort int, dashboardUser string, dashboardPassword string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("bindPort = %d\n", bindPort))

	// 允许自动分配的端口范围，避免与低端口冲突
	sb.WriteString("\nallowPorts = [\n")
	sb.WriteString("  { start = 10000, end = 60000 }\n")
	sb.WriteString("]\n")

	// Dashboard
	sb.WriteString("\n[webServer]\n")
	sb.WriteString(fmt.Sprintf("addr = \"0.0.0.0\"\n"))
	sb.WriteString(fmt.Sprintf("port = %d\n", dashboardPort))
	sb.WriteString(fmt.Sprintf("user = \"%s\"\n", dashboardUser))
	sb.WriteString(fmt.Sprintf("password = \"%s\"\n", dashboardPassword))

	return sb.String()
}

// generateFrpsSystemdService 生成 frps systemd 服务文件
func generateFrpsSystemdServiceContent() string {
	return `[Unit]
Description=FRP Server - Public Network Tunnel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/frps
ExecStart=/usr/local/bin/frps -c /etc/frps/frps.toml

Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
}

func generateFrpsSystemdService(dir string) {
	content := `[Unit]
Description=FRP Server - Public Network Tunnel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/frps
ExecStart=/usr/local/bin/frps -c /etc/frps/frps.toml

Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
	os.WriteFile(filepath.Join(dir, "frps.service"), []byte(content), 0644)
}

// extractFRPSFromTarGz 从tar.gz包中解压frps二进制
func extractFRPSFromTarGz(tarPath, targetDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("打开tar.gz文件失败: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("解压gzip失败: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	found := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取tar失败: %w", err)
		}
		baseName := filepath.Base(hdr.Name)
		if baseName != "frps" && baseName != "frps.toml" {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		dstPath := filepath.Join(targetDir, baseName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("创建文件 %s 失败: %w", dstPath, err)
		}
		if _, err := io.Copy(dst, tr); err != nil {
			dst.Close()
			return fmt.Errorf("写入文件 %s 失败: %w", dstPath, err)
		}
		dst.Close()
		if baseName == "frps" {
			os.Chmod(dstPath, 0755)
		}
		log.Printf("[公网穿透服务端] 解压文件: %s -> %s", hdr.Name, baseName)
		found = true
	}
	if !found {
		return fmt.Errorf("tar.gz包中未找到frps文件")
	}
	return nil
}

// extractFRPSFromZip 从zip包中解压frps二进制
func extractFRPSFromZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %w", err)
	}
	defer r.Close()

	// 列出zip内所有文件用于调试
	var fileNames []string
	for _, f := range r.File {
		fileNames = append(fileNames, f.Name)
	}
	log.Printf("[公网穿透服务端] zip内文件列表: %v", fileNames)

	found := false
	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName != "frps" && baseName != "frps.toml" {
			continue
		}
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("打开zip内文件 %s 失败: %w", f.Name, err)
		}

		dstPath := filepath.Join(targetDir, baseName)
		dst, err := os.Create(dstPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("创建文件 %s 失败: %w", dstPath, err)
		}

		if _, err := io.Copy(dst, rc); err != nil {
			dst.Close()
			rc.Close()
			return fmt.Errorf("写入文件 %s 失败: %w", dstPath, err)
		}
		dst.Close()
		rc.Close()

		if baseName == "frps" {
			os.Chmod(dstPath, 0755)
		}
		log.Printf("[公网穿透服务端] 解压文件: %s -> %s", f.Name, baseName)
			found = true
		}
		if !found {
			return fmt.Errorf("zip包中未找到frps文件，包内文件: %v", fileNames)
		}
		return nil
	}
