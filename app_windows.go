//go:build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// openFileDialog 通用文件选择对话框（Windows特定实现）
func openFileDialog(title, filterName, filterPattern string) map[string]interface{} {
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Text.Encoding
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = "%s"
$dialog.Filter = "%s (%s)|%s"
$dialog.FilterIndex = 1
$dialog.RestoreDirectory = $true
$dialog.ShowDialog() | Out-Null
if ($dialog.FileName) {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($dialog.FileName)
    [Convert]::ToBase64String($bytes)
}
`, title, filterName, filterPattern, filterPattern)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		log.Printf("打开文件选择对话框失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件选择对话框失败: %v", err),
		}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		log.Printf("用户取消选择文件")
		return map[string]interface{}{
			"success": false,
			"message": "用户取消选择",
		}
	}

	// Base64 解码获取 UTF-8 编码的文件路径
	fileBytes, err := ConvertBase64ToBytes(outputStr)
	if err != nil {
		log.Printf("文件路径解码失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件路径解码失败: %v", err),
		}
	}

	filePath := string(fileBytes)
	log.Printf("选择文件成功: %s", filePath)
	return map[string]interface{}{
		"success": true,
		"message": "文件选择成功",
		"path":    filePath,
	}
}

// setSocketBroadcast sets the SO_BROADCAST option on a socket
func setSocketBroadcast(fd uintptr) error {
	return syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
}

func configureHiddenProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
}

func bringProcessToFront(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}

	hwnd, err := findMainWindow(uint32(pid))
	if err != nil {
		return err
	}

	return bringWindowToFront(hwnd)
}

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows            = user32.NewProc("EnumWindows")
	procGetWindowThreadProcess = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible        = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procShowWindow             = user32.NewProc("ShowWindow")
	procGetWindowTextLengthW   = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW         = user32.NewProc("GetWindowTextW")
	
	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	procOpenProcess            = kernel32.NewProc("OpenProcess")
	procCloseHandle            = kernel32.NewProc("CloseHandle")
	procGetExitCodeProcess     = kernel32.NewProc("GetExitCodeProcess")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// isWindowsProcessRunning 检查Windows进程是否正在运行
func isWindowsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// 尝试打开进程句柄
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)
	
	if handle == 0 {
		return false
	}
	defer procCloseHandle.Call(handle)
	
	// 检查进程退出码：STILL_ACTIVE (259) 表示进程仍在运行
	var exitCode uint32
	ret, _, _ := procGetExitCodeProcess.Call(handle, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false // GetExitCodeProcess 调用失败
	}
	
	return exitCode == 259 // STILL_ACTIVE
}

func findMainWindow(pid uint32) (WindowHandle, error) {
	var hwnd WindowHandle

	cb := syscall.NewCallback(func(handle uintptr, lparam uintptr) uintptr {
		var windowPid uint32
		procGetWindowThreadProcess.Call(handle, uintptr(unsafe.Pointer(&windowPid)))
		if windowPid != pid {
			return 1
		}

		visible, _, _ := procIsWindowVisible.Call(handle)
		if visible == 0 {
			return 1
		}

		hwnd = WindowHandle(handle)
		return 0
	})

	procEnumWindows.Call(cb, 0)

	if hwnd == 0 {
		return 0, fmt.Errorf("window not found for pid %d", pid)
	}

	return hwnd, nil
}

func bringWindowToFront(hwnd WindowHandle) error {
	if hwnd == 0 {
		return fmt.Errorf("invalid window handle")
	}
	const swRestore = 9
	procShowWindow.Call(uintptr(hwnd), swRestore)
	r, _, callErr := procSetForegroundWindow.Call(uintptr(hwnd))
	if r == 0 {
		return fmt.Errorf("SetForegroundWindow failed: %v", callErr)
	}
	return nil
}

func findWindowByTitle(title string) (WindowHandle, error) {
	if title == "" {
		return 0, fmt.Errorf("empty title")
	}

	var hwnd WindowHandle
	lowerTitle := strings.ToLower(title)

	cb := syscall.NewCallback(func(handle uintptr, lparam uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(handle)
		if visible == 0 {
			return 1
		}

		length, _, _ := procGetWindowTextLengthW.Call(handle)
		if length == 0 {
			return 1
		}

		buf := make([]uint16, length+1)
		procGetWindowTextW.Call(handle, uintptr(unsafe.Pointer(&buf[0])), length+1)
		text := windows.UTF16ToString(buf)
		if text == "" {
			return 1
		}

		if strings.Contains(strings.ToLower(text), lowerTitle) {
			hwnd = WindowHandle(handle)
			return 0
		}

		return 1
	})

	procEnumWindows.Call(cb, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("window not found for title %s", title)
	}
	return hwnd, nil
}
