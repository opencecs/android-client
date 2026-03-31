package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// -------------------------------------------------------
// Camera pusher — 独立 embed Release_cam/（静态编译，无 DLL）
// P2P 用 Release_native/（原始轻量版，含 libsrt），各自独立。
// -------------------------------------------------------

//go:embed camdemo/windows_pusher/Release_cam/*
var camPusherAssets embed.FS

const camPusherAssetDir = "camdemo/windows_pusher/Release_cam"

var (
	camExeOnce sync.Once
	camExePath string
	camExeErr  error
)

func ensureCamExe() (string, error) {
	camExeOnce.Do(func() {
		dir := filepath.Join(os.TempDir(), "edgeclient-cam")
		if err := os.MkdirAll(dir, 0755); err != nil {
			camExeErr = fmt.Errorf("cannot create cam temp dir: %w", err)
			return
		}
		if err := extractCamAssets(dir); err != nil {
			camExeErr = err
			return
		}
		camExePath = filepath.Join(dir, "WindowsPusher.exe")
		fmt.Printf("[CamPusher] extracted to %s\n", camExePath)
	})
	return camExePath, camExeErr
}

func extractCamAssets(destDir string) error {
	return fs.WalkDir(camPusherAssets, camPusherAssetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(camPusherAssetDir, filepath.FromSlash(path))
		if rel == ".gitkeep" {
			return nil
		}
		data, err := camPusherAssets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embed %s: %w", path, err)
		}
		if len(data) == 0 {
			return fmt.Errorf("embedded file %s is empty", path)
		}
		want := sha256.Sum256(data)
		dest := filepath.Join(destDir, rel)
		if existing, err := os.ReadFile(dest); err == nil {
			if sha256.Sum256(existing) == want {
				return nil
			}
		}
		return os.WriteFile(dest, data, 0755)
	})
}

// -------------------------------------------------------
// Process management
// -------------------------------------------------------

type windowsPusherProcess struct {
	cmd *exec.Cmd
	pid int
}

var (
	pusherProcesses   = map[int]*windowsPusherProcess{}
	pusherProcessesMu sync.Mutex
	pusherNextID      = 1
)

// StartWindowsPusher launches the embedded camera WindowsPusher.exe.
func (a *App) StartWindowsPusher(args []string) map[string]interface{} {
	exePath, err := ensureCamExe()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "摄像头推流工具初始化失败: " + err.Error(),
		}
	}

	var stderr bytes.Buffer
	cmd := exec.Command(exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "启动失败: " + err.Error(),
		}
	}

	pusherProcessesMu.Lock()
	id := pusherNextID
	pusherNextID++
	exitCh := make(chan error, 1)
	pusherProcesses[id] = &windowsPusherProcess{cmd: cmd, pid: id}
	pusherProcessesMu.Unlock()

	go func() {
		err := cmd.Wait()
		exitCh <- err
		pusherProcessesMu.Lock()
		delete(pusherProcesses, id)
		pusherProcessesMu.Unlock()
	}()

	// 等待最多 3 秒，检测进程是否立即崩溃退出
	select {
	case waitErr := <-exitCh:
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" && waitErr != nil {
			errMsg = waitErr.Error()
		}
		if errMsg == "" {
			errMsg = "进程异常退出，请检查设备IP/端口是否可达"
		}
		fmt.Printf("[CamPusher] Process exited quickly id=%d err=%v stderr=%s\n", id, waitErr, errMsg)
		return map[string]interface{}{
			"success": false,
			"message": errMsg,
		}
	case <-time.After(3 * time.Second):
		// 进程存活超过3秒，认为启动成功
	}

	fmt.Printf("[CamPusher] Started id=%d pid=%d args=%v\n", id, cmd.Process.Pid, args)
	return map[string]interface{}{
		"success": true,
		"pid":     id,
		"message": "摄像头推流已启动",
	}
}

// StopWindowsPusher stops the camera pusher process by internal id.
func (a *App) StopWindowsPusher(id int) map[string]interface{} {
	pusherProcessesMu.Lock()
	proc, ok := pusherProcesses[id]
	pusherProcessesMu.Unlock()

	if !ok {
		return map[string]interface{}{"success": true, "message": "进程已停止"}
	}

	if proc.cmd != nil && proc.cmd.Process != nil {
		if err := proc.cmd.Process.Kill(); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": "停止失败: " + err.Error(),
			}
		}
	}

	pusherProcessesMu.Lock()
	delete(pusherProcesses, id)
	pusherProcessesMu.Unlock()

	fmt.Printf("[CamPusher] Stopped id=%d\n", id)
	return map[string]interface{}{"success": true, "message": "已停止"}
}

// StopAllWindowsPushers stops every running camera pusher (called on app exit).
func StopAllWindowsPushers() {
	pusherProcessesMu.Lock()
	defer pusherProcessesMu.Unlock()
	for id, proc := range pusherProcesses {
		if proc.cmd != nil && proc.cmd.Process != nil {
			_ = proc.cmd.Process.Kill()
		}
		delete(pusherProcesses, id)
	}
}

// CapturePreviewFrame invokes WindowsPusher.exe --capture-preview and returns
// a JPEG DataURL.  The C++ side first checks a shared temp file written by
// the running push process, so the camera is never double-opened.
func (a *App) CapturePreviewFrame(camName string, width, height int) map[string]interface{} {
	exePath, err := ensureCamExe()
	if err != nil {
		return map[string]interface{}{"success": false, "message": "摄像头推流工具初始化失败: " + err.Error()}
	}

	args := []string{"--capture-preview",
		"--width", fmt.Sprintf("%d", width),
		"--height", fmt.Sprintf("%d", height),
	}
	if camName != "" {
		args = append(args, "--cam-name", camName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[CamPusher] CapturePreviewFrame error: %v  stderr: %s\n", err, stderr.String())
	}

	scanner := bufio.NewScanner(&stdout)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "PREVIEW_JPEG:") {
			b64 := strings.TrimPrefix(line, "PREVIEW_JPEG:")
			return map[string]interface{}{
				"success": true,
				"dataURL": "data:image/jpeg;base64," + b64,
			}
		}
	}

	errMsg := strings.TrimSpace(stderr.String())
	return map[string]interface{}{"success": false, "message": "未能获取预览帧: " + errMsg}
}

// ListCameraDevices calls WindowsPusher.exe --list-cameras.
func (a *App) ListCameraDevices() map[string]interface{} {
	exePath, err := ensureCamExe()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "摄像头推流工具初始化失败: " + err.Error(),
			"cameras": []string{},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exePath, "--list-cameras")
	cmd.Dir = filepath.Dir(exePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[CamPusher] ListCameraDevices error: %v  stderr: %s\n", err, stderr.String())
	}

	cameras := []string{}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "CAMERA:") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "CAMERA:"))
			if name != "" {
				cameras = append(cameras, name)
			}
		}
	}

	fmt.Printf("[CamPusher] ListCameraDevices found %d cameras: %v\n", len(cameras), cameras)
	return map[string]interface{}{
		"success": true,
		"cameras": cameras,
	}
}

// resetCamExe allows re-extraction (useful after an update).
// Not exposed to the frontend, can be called internally if needed.
func resetCamExe() {
	camExeOnce = sync.Once{}
	camExePath = ""
	camExeErr = nil
}
