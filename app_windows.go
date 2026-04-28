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

// ---- 现代 IFileOpenDialog 文件夹选择器 (Vista+) ----

var (
	ole32  = windows.NewLazySystemDLL("ole32.dll")
	shell  = windows.NewLazySystemDLL("shell32.dll")

	procCoInitializeEx          = ole32.NewProc("CoInitializeEx")
	procCoUninitialize          = ole32.NewProc("CoUninitialize")
	procCoCreateInstance        = ole32.NewProc("CoCreateInstance")
	procSHCreateItemFromParsing = shell.NewProc("SHCreateItemFromParsingName")
	procCoTaskMemFree           = ole32.NewProc("CoTaskMemFree")
)

// COM IIDs
var (
	clsidFileOpenDialog = windows.GUID{Data1: 0xDC1C5A9C, Data2: 0xE88A, Data3: 0x4dde, Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	iidIFileOpenDialog  = windows.GUID{Data1: 0xD57C7288, Data2: 0xD4AD, Data3: 0x4768, Data4: [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
	iidIShellItem       = windows.GUID{Data1: 0x43826D1E, Data2: 0xE718, Data3: 0x42EE, Data4: [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE}}
)

const (
	FOS_PICKFOLDERS     = 0x00000020
	FOS_FORCEFILESYSTEM = 0x00000040
	SIGDN_FILESYSPATH   = 0x80058000
	COINIT_APARTMENTTHREADED = 0x2
	ERROR_CANCELLED     = 0x800704C7
)

// IFileOpenDialog vtable offsets
// IUnknown(3): QueryInterface,AddRef,Release
// IModalWindow(1): Show
// IFileDialog: SetFileTypes,SetFileTypeIndex,GetFileTypeIndex,Advise,Unadvise,SetOptions,GetOptions,SetDefaultFolder,SetFolder,GetFolder,GetCurrentSelection,SetFileName,GetFileName,SetTitle,SetOkButtonLabel,SetFileNameLabel,GetResult,AddPlace,SetDefaultExtension,Close,SetClientGuid,ClearClientData,SetFilter
const (
	vtblShow             = 3  // IModalWindow::Show
	vtblSetOptions       = 9  // IFileDialog::SetOptions
	vtblSetDefaultFolder = 11 // IFileDialog::SetDefaultFolder
	vtblSetTitle         = 17 // IFileDialog::SetTitle
	vtblGetResult        = 20 // IFileDialog::GetResult
)

// IShellItem vtable offsets
// IUnknown(3): QueryInterface,AddRef,Release
// IShellItem: BindToHandler,GetParent,GetDisplayName,GetAttributes,Compare
const (
	siVtblGetDisplayName = 5 // IShellItem::GetDisplayName
)

// IUnknown::Release vtable offset
const vtblRelease = 2

// comRelease 调用IUnknown::Release释放COM对象
func comRelease(obj uintptr) {
	if obj == 0 {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	release := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblRelease)*unsafe.Sizeof(uintptr(0))))
	syscall.Syscall(release, 1, obj, 0, 0)
}

func pickFolderWindows(startPath string) (string, error) {
	// 初始化COM
	procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	defer procCoUninitialize.Call()

	// 创建 IFileOpenDialog 实例
	var dialog uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		1|4, // CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER
		uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
		uintptr(unsafe.Pointer(&dialog)),
	)
	if hr != 0 || dialog == 0 {
		return "", fmt.Errorf("CoCreateInstance failed: 0x%08X", hr)
	}
	defer comRelease(dialog)

	vtbl := *(*uintptr)(unsafe.Pointer(dialog))

	// SetOptions: FOS_PICKFOLDERS | FOS_FORCEFILESYSTEM
	setOptions := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblSetOptions)*unsafe.Sizeof(uintptr(0))))
	_, _, _ = syscall.Syscall(setOptions, 2, dialog, FOS_PICKFOLDERS|FOS_FORCEFILESYSTEM, 0)

	// SetTitle
	setTitle := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblSetTitle)*unsafe.Sizeof(uintptr(0))))
	titlePtr, _ := windows.UTF16PtrFromString("请选择文件夹")
	syscall.Syscall(setTitle, 2, dialog, uintptr(unsafe.Pointer(titlePtr)), 0)

	// SetDefaultFolder (if startPath provided)
	if startPath != "" {
		var shellItem uintptr
		pathPtr, _ := windows.UTF16PtrFromString(startPath)
		hr2, _, _ := procSHCreateItemFromParsing.Call(
			uintptr(unsafe.Pointer(pathPtr)),
			0,
			uintptr(unsafe.Pointer(&iidIShellItem)),
			uintptr(unsafe.Pointer(&shellItem)),
		)
		if hr2 == 0 && shellItem != 0 {
			setDefaultFolder := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblSetDefaultFolder)*unsafe.Sizeof(uintptr(0))))
			syscall.Syscall(setDefaultFolder, 2, dialog, shellItem, 0)
			comRelease(shellItem)
		}
	}

	// Show
	show := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblShow)*unsafe.Sizeof(uintptr(0))))
	hr3, _, _ := syscall.Syscall(show, 2, dialog, 0, 0)
	if hr3 != 0 {
		if hr3 == ERROR_CANCELLED {
			return "", fmt.Errorf("用户取消选择")
		}
		return "", fmt.Errorf("对话框显示失败: 0x%08X", hr3)
	}

	// GetResult
	var resultItem uintptr
	getResult := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(vtblGetResult)*unsafe.Sizeof(uintptr(0))))
	hr4, _, _ := syscall.Syscall(getResult, 2, dialog, uintptr(unsafe.Pointer(&resultItem)), 0)
	if hr4 != 0 || resultItem == 0 {
		return "", fmt.Errorf("获取选择结果失败")
	}
	defer comRelease(resultItem)

	// IShellItem::GetDisplayName
	siVtbl := *(*uintptr)(unsafe.Pointer(resultItem))
	getDisplayName := *(*uintptr)(unsafe.Pointer(siVtbl + uintptr(siVtblGetDisplayName)*unsafe.Sizeof(uintptr(0))))
	var namePtr uintptr
	syscall.Syscall(getDisplayName, 3, resultItem, SIGDN_FILESYSPATH, uintptr(unsafe.Pointer(&namePtr)))
	if namePtr == 0 {
		return "", fmt.Errorf("获取路径名失败")
	}
	defer procCoTaskMemFree.Call(namePtr)

	selectedPath := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(namePtr)))
	return selectedPath, nil
}

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
