//go:build windows

package updater

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// runUpdateWithElevation runs the update process with UAC elevation on Windows
func runUpdateWithElevation(u *Updater, exePath, zipPath, exeDir string, args []string, asset *ReleaseAsset) error {
	argStr := strings.Join(args, " ")
	log.Printf("[Updater] 命令行: %s %s", exePath, argStr)

	showCmd := int32(1)

	log.Printf("[Updater] 加载 shell32.dll...")
	mod := windows.NewLazyDLL("shell32.dll")
	log.Printf("[Updater] 获取 ShellExecuteW 过程地址...")
	proc := mod.NewProc("ShellExecuteW")

	log.Printf("[Updater] 调用 ShellExecuteW...")
	result, _, err := proc.Call(
		0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("runas"))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(exePath))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(argStr))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(exeDir))),
		uintptr(showCmd))

	log.Printf("[Updater] ShellExecuteW 返回: result=%d, err=%v", result, err)

	if result == 0 || (result > 0 && result <= 32) {
		errMsg := fmt.Sprintf("ShellExecute 失败，返回值: %d, 错误: %v", result, err)
		log.Printf("[Updater] %s", errMsg)
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = errMsg
		return errors.New(errMsg)
	}

	log.Printf("[Updater] UAC提权请求已发送成功")
	u.state.State = UpdateStateInstalling
	u.state.UpdateLog = fmt.Sprintf("正在提权更新...\n版本: %s\n时间: %s", asset.Version, time.Now().Format(time.RFC3339))

	return nil
}
