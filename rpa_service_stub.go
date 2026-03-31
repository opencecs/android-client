//go:build !cgorpa

package main

// rpa_service_stub.go — RPA SDK 桩实现（无 CGO 模式）
// 当不使用 -tags cgorpa 编译时，提供相同方法签名，返回"SDK 未集成"提示。
// 等 libmytrpc.dll 到位后，使用 -tags cgorpa 编译即可切换到真实 CGO 版本。

// RpaRequest 前端传入的设备定位信息
type RpaRequest struct {
	DeviceIP    string `json:"deviceIp"`
	RpaPort     int    `json:"rpaPort"`
	ContainerID string `json:"containerId"`
	Password    string `json:"password"`
}

func rpaNotAvailable() map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": "RPA SDK 未集成（需使用 -tags cgorpa 编译并提供 libmytrpc.dll）",
	}
}

// RpaGetScreenContext 获取屏幕上下文
func (a *App) RpaGetScreenContext(req RpaRequest) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaGetDeviceEnv 获取设备环境信息
func (a *App) RpaGetDeviceEnv(req RpaRequest) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaClick 点击指定坐标
func (a *App) RpaClick(req RpaRequest, x, y int) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaSwipe 滑动操作
func (a *App) RpaSwipe(req RpaRequest, x0, y0, x1, y1, durationMs int) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaSendText 向当前输入框发送文字
func (a *App) RpaSendText(req RpaRequest, text string) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaKeyPress 按下指定按键码
func (a *App) RpaKeyPress(req RpaRequest, keyCode int) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaOpenApp 启动指定包名的 App
func (a *App) RpaOpenApp(req RpaRequest, pkg string) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaStopApp 停止指定包名的 App
func (a *App) RpaStopApp(req RpaRequest, pkg string) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaExecCmd 在设备内执行 shell 命令（层2 RPA 通道）
func (a *App) RpaExecCmd(req RpaRequest, cmd string, waitForExit bool) map[string]interface{} {
	return rpaNotAvailable()
}

// RpaCloseConnection 关闭指定设备的 RPA 连接
func (a *App) RpaCloseConnection(req RpaRequest) map[string]interface{} {
	return map[string]interface{}{"success": true}
}

// RpaGetVersion 获取 SDK 版本号
func (a *App) RpaGetVersion() map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"version": 0,
		"message": "RPA SDK 未集成",
	}
}
