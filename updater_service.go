package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"edgeclient/updater"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// UpdaterService 更新服务 - V3 Service
type UpdaterService struct {
	updater  *updater.Updater
	wailsApp *application.App
}

// NewUpdaterService 创建更新服务实例
func NewUpdaterService() *UpdaterService {
	config := updater.DefaultUpdateConfig()

	service := &UpdaterService{
		updater:  updater.NewUpdater(config),
		wailsApp: nil,
	}

	// 启动自动检查（如果启用）
	service.updater.StartAutoCheck()

	return service
}

// SetWailsApp 设置Wails应用实例
func (s *UpdaterService) SetWailsApp(app *application.App) {
	s.wailsApp = app
	if s.updater != nil {
		s.updater.SetWailsApp(app)
	}
}

// CheckForUpdate 检查更新
func (s *UpdaterService) CheckForUpdate() map[string]interface{} {
	log.Printf("[UpdaterService] 收到检查更新请求")

	if s.updater == nil {
		return map[string]interface{}{
			"success": false,
			"message": "更新服务未初始化",
			"state":   "error",
		}
	}

	// 检查更新
	result, err := s.updater.CheckForUpdate()
	if err != nil {
		log.Printf("[UpdaterService] 检查更新失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"state":   "error",
		}
	}

	if result != nil {
		// 发现新版本
		return map[string]interface{}{
			"success":      true,
			"hasUpdate":    true,
			"version":      result.Version,
			"downloadUrl":  result.DownloadURL,
			"checksum":     result.Checksum,
			"fileSize":     result.FileSize,
			"releaseNotes": result.ReleaseNotes,
			"mandatory":    result.Mandatory,
			"state":        "available",
		}
	}

	// 当前已是最新版本
	return map[string]interface{}{
		"success":   true,
		"hasUpdate": false,
		"message":   "当前已是最新版本",
		"state":     "uptodate",
	}
}

// StartUpdate 开始更新
func (s *UpdaterService) StartUpdate() map[string]interface{} {
	log.Printf("[UpdaterService] 收到开始更新请求")

	if s.updater == nil {
		return map[string]interface{}{
			"success": false,
			"message": "更新服务未初始化",
			"state":   "error",
		}
	}

	asset, err := s.updater.CheckForUpdate()
	if err != nil {
		log.Printf("[UpdaterService] 检查更新失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"state":   "failed",
		}
	}

	if asset == nil {
		return map[string]interface{}{
			"success": false,
			"message": "当前已是最新版本",
			"state":   "uptodate",
		}
	}

	if s.updater.NeedsElevation() {
		log.Printf("[UpdaterService] 检测到需要UAC提权")
		log.Printf("[UpdaterService] 目录: %s", s.updater.GetInstallDir())

		log.Printf("[UpdaterService] 准备下载更新包...")
		err := s.updater.StartElevatedUpdate(asset)
		if err != nil {
			log.Printf("[UpdaterService] 提权更新启动失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("启动提权更新失败: %v", err),
				"state":   "failed",
			}
		}

		log.Printf("[UpdaterService] ShellExecute成功，准备退出主程序")

		go func() {
			time.Sleep(500 * time.Millisecond)
			log.Printf("[UpdaterService] 自动退出主程序")
			if s.wailsApp != nil {
				s.wailsApp.Quit()
			} else {
				log.Printf("[UpdaterService] wailsApp 为 nil，直接退出")
				os.Exit(0)
			}
		}()

		return map[string]interface{}{
			"success": true,
			"message": "已请求管理员权限，请在新窗口中完成更新",
			"state":   "elevating",
			"version": asset.Version,
		}
	}

	err = s.updater.StartUpdate()
	if err != nil {
		log.Printf("[UpdaterService] 更新失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"state":   "failed",
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "更新完成",
		"state":   "complete",
	}
}

// ConfigureUpdate 配置更新参数
func (s *UpdaterService) ConfigureUpdate(autoCheck bool, autoUpdate bool, checkInterval int, channel string, proxyURL string) map[string]interface{} {
	log.Printf("[UpdaterService] 收到配置更新请求: autoCheck=%v, autoUpdate=%v, checkInterval=%d, channel=%s",
		autoCheck, autoUpdate, checkInterval, channel)

	if s.updater == nil {
		return map[string]interface{}{
			"success": false,
			"message": "更新服务未初始化",
		}
	}

	// 创建新配置
	newConfig := &updater.UpdateConfig{
		AutoCheck:     autoCheck,
		AutoUpdate:    autoUpdate,
		CheckInterval: checkInterval,
		Channel:       channel,
		ProxyURL:      proxyURL,
	}

	// 更新配置并保存
	err := s.updater.UpdateConfig(newConfig)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "配置已保存",
	}
}

// GetUpdateConfig 获取更新配置
func (s *UpdaterService) GetUpdateConfig() map[string]interface{} {
	log.Printf("[UpdaterService] 收到获取配置请求")

	if s.updater == nil {
		config := updater.DefaultUpdateConfig()
		return map[string]interface{}{
			"checkUrl":           config.CheckURL,
			"channel":            config.Channel,
			"autoCheck":          config.AutoCheck,
			"autoUpdate":         config.AutoUpdate,
			"checkInterval":      config.CheckInterval,
			"lastCheckTime":      "",
			"lastCheckedVersion": "",
		}
	}

	config := s.updater.GetConfig()
	state := s.updater.GetState()

	return map[string]interface{}{
		"checkUrl":           config.CheckURL,
		"channel":            config.Channel,
		"autoCheck":          config.AutoCheck,
		"autoUpdate":         config.AutoUpdate,
		"checkInterval":      config.CheckInterval,
		"lastCheckTime":      config.LastCheckTime.Format("2006-01-02 15:04:05"),
		"lastCheckedVersion": state.LatestVersion,
	}
}

// GetVersionInfo 获取当前版本信息
func (s *UpdaterService) GetVersionInfo() map[string]interface{} {
	log.Printf("[UpdaterService] 收到获取版本信息请求")

	return map[string]interface{}{
		"version":   updater.AppVersion,
		"buildTime": updater.BuildTime,
	}
}

// CancelUpdate 取消更新
func (s *UpdaterService) CancelUpdate() map[string]interface{} {
	log.Printf("[UpdaterService] 收到取消更新请求")

	if s.updater == nil {
		return map[string]interface{}{
			"success": false,
			"message": "更新服务未初始化",
		}
	}

	s.updater.Stop()
	return map[string]interface{}{
		"success": true,
		"message": "已取消更新",
	}
}

// GetUpdateState 获取更新状态
func (s *UpdaterService) GetUpdateState() map[string]interface{} {
	log.Printf("[UpdaterService] 收到获取更新状态请求")

	if s.updater == nil {
		return map[string]interface{}{
			"state":            "idle",
			"currentVersion":   updater.AppVersion,
			"downloadProgress": 0,
			"errorMessage":     "",
		}
	}

	state := s.updater.GetState()
	return map[string]interface{}{
		"state":            state.State,
		"currentVersion":   state.CurrentVersion,
		"downloadProgress": state.DownloadProgress,
		"errorMessage":     state.ErrorMessage,
	}
}

// GetUpdateLog 获取更新日志
func (s *UpdaterService) GetUpdateLog() map[string]interface{} {
	log.Printf("[UpdaterService] 收到获取更新日志请求")

	if s.updater == nil {
		return map[string]interface{}{
			"updateLog": "",
		}
	}

	state := s.updater.GetState()
	return map[string]interface{}{
		"updateLog": state.UpdateLog,
	}
}

// RestartApp 重启应用
func (s *UpdaterService) RestartApp() map[string]interface{} {
	log.Printf("[UpdaterService] 收到重启应用请求")

	if s.updater == nil {
		return map[string]interface{}{
			"success": false,
			"message": "更新服务未初始化",
		}
	}

	if err := s.updater.RestartApp(); err != nil {
		log.Printf("[UpdaterService] 重启应用失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("重启失败: %v", err),
		}
	}

	log.Printf("[UpdaterService] 重启命令已发送")
	return map[string]interface{}{
		"success": true,
		"message": "正在重启应用",
	}
}
