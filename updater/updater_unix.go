//go:build !windows

package updater

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// runUpdateWithElevation runs the update process on non-Windows platforms
// (no UAC elevation needed on macOS/Linux)
func runUpdateWithElevation(u *Updater, exePath, zipPath, exeDir string, args []string, asset *ReleaseAsset) error {
	// On non-Windows platforms, we can run the update directly without elevation
	log.Printf("[Updater] 在非Windows平台运行更新，不需要UAC提权")
	log.Printf("[Updater] 可执行文件: %s", exePath)
	log.Printf("[Updater] 参数: %v", args)

	// Create the command
	cmd := exec.Command(exePath, args...)
	cmd.Dir = exeDir

	// Start the command
	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("启动更新进程失败: %v", err)
		log.Printf("[Updater] %s", errMsg)
		u.state.State = UpdateStateFailed
		u.state.ErrorMessage = errMsg
		return err
	}

	log.Printf("[Updater] 更新进程已启动，PID: %d", cmd.Process.Pid)
	u.state.State = UpdateStateInstalling
	u.state.UpdateLog = fmt.Sprintf("正在更新...\n版本: %s\n时间: %s", asset.Version, time.Now().Format(time.RFC3339))

	// Don't wait for the command to complete - let it run in the background
	// since this process will exit
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("[Updater] 更新进程退出，错误: %v", err)
		} else {
			log.Printf("[Updater] 更新进程成功完成")
		}
	}()

	return nil
}
