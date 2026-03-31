//go:build cgorpa

package main

/*
#cgo windows,amd64 CFLAGS: -DLIBMYTRPC_EXPORTS
#cgo windows,amd64 LDFLAGS: -LC:/cgo_libs/mytrpc -lmytrpc -lws2_32 -Wl,-Bstatic -lstdc++ -lgcc -lwinpthread -Wl,-Bdynamic
#include <windows.h>
#include <stdbool.h>
#include "C:/cgo_libs/mytrpc/libmytrpc.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// ============================================================
// rpa_service.go — RPA SDK cgo 绑定层
// 连接池管理 + 所有 Wails RPA 方法（层2工具）
// ============================================================

// ---------- 连接池 ----------

type rpaHandleEntry struct {
	handle    C.long
	createdAt time.Time
}

var (
	rpaPool   sync.Map    // key: "ip:port" → *rpaHandleEntry
	rpaPoolMu sync.Mutex  // 保护同一key的并发 openDevice 调用
)

const rpaHandleMaxAge = 10 * time.Minute // 超过此时间的连接重建

// rpaGetHandle 获取连接池中的 handle，不存在或连接断开时自动重建
func rpaGetHandle(ip string, port int) (C.long, error) {
	key := fmt.Sprintf("%s:%d", ip, port)

	// 先尝试从池中取
	if v, ok := rpaPool.Load(key); ok {
		entry := v.(*rpaHandleEntry)
		// 检查存活状态
		if C.checkLive(entry.handle) == 1 && time.Since(entry.createdAt) < rpaHandleMaxAge {
			return entry.handle, nil
		}
		// 连接已断开或超时，关闭旧连接
		C.closeDevice(entry.handle)
		rpaPool.Delete(key)
	}

	// 串行建立新连接，防止并发时重复 openDevice
	rpaPoolMu.Lock()
	defer rpaPoolMu.Unlock()

	// 双重检查
	if v, ok := rpaPool.Load(key); ok {
		entry := v.(*rpaHandleEntry)
		if C.checkLive(entry.handle) == 1 {
			return entry.handle, nil
		}
		C.closeDevice(entry.handle)
		rpaPool.Delete(key)
	}

	cIP := C.CString(ip)
	defer C.free(unsafe.Pointer(cIP))
	handle := C.openDevice(cIP, C.int(port), C.long(5))
	if handle == 0 {
		return 0, fmt.Errorf("RPA 连接失败: %s:%d", ip, port)
	}

	rpaPool.Store(key, &rpaHandleEntry{handle: handle, createdAt: time.Now()})
	log.Printf("[RPA] 建立连接 %s handle=%d", key, handle)
	return handle, nil
}

// rpaReleaseHandle 主动释放连接（任务结束后可选调用）
func rpaReleaseHandle(ip string, port int) {
	key := fmt.Sprintf("%s:%d", ip, port)
	if v, ok := rpaPool.Load(key); ok {
		entry := v.(*rpaHandleEntry)
		C.closeDevice(entry.handle)
		rpaPool.Delete(key)
		log.Printf("[RPA] 释放连接 %s", key)
	}
}

// ---------- 请求/响应数据结构 ----------

// RpaRequest 前端传入的设备定位信息
type RpaRequest struct {
	DeviceIP    string `json:"deviceIp"`
	RpaPort     int    `json:"rpaPort"`
	ContainerID string `json:"containerId"`
	Password    string `json:"password"`
}

// ---------- Wails RPA API — 层2工具 ----------

// RpaGetScreenContext 获取屏幕上下文（优先 XML 精简节点摘要，超过阈值时截图）
// 返回：{ type: "nodes"|"screenshot", content: "...", width: int, height: int }
func (a *App) RpaGetScreenContext(req RpaRequest) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	// 优先尝试获取 XML
	xmlPtr := C.dumpNodeXml(handle, C.int(0))
	if xmlPtr != nil {
		defer C.freeRpcPtr(unsafe.Pointer(xmlPtr))
		xmlStr := C.GoString(xmlPtr)
		if len(xmlStr) > 50 { // 有效 XML
			summary := parseXMLToNodeSummary(xmlStr)
			return map[string]interface{}{
				"success": true,
				"type":    "nodes",
				"content": summary,
			}
		}
	}

	// XML 不可用，降级截图（jpg 压缩，质量 60）
	var imgLen C.int
	imgPtr := C.takeCaptrueCompress(handle, C.int(1), C.int(60), &imgLen)
	if imgPtr == nil || imgLen == 0 {
		return map[string]interface{}{"success": false, "message": "截图失败"}
	}
	defer C.freeRpcPtr(unsafe.Pointer(imgPtr))
	imgBytes := C.GoBytes(unsafe.Pointer(imgPtr), imgLen)
	b64 := base64.StdEncoding.EncodeToString(imgBytes)

	return map[string]interface{}{
		"success": true,
		"type":    "screenshot",
		"content": b64,
	}
}

// xmlNode 用于解析 Android UI XML
type xmlNode struct {
	XMLName     xml.Name  `xml:"node"`
	Text        string    `xml:"text,attr"`
	ContentDesc string    `xml:"content-desc,attr"`
	Clickable   string    `xml:"clickable,attr"`
	Focusable   string    `xml:"focusable,attr"`
	Bounds      string    `xml:"bounds,attr"`
	ClassName   string    `xml:"class,attr"`
	Children    []xmlNode `xml:"node"`
}

// parseXMLToNodeSummary 解析 Android UI XML，提取所有可交互节点，返回精简摘要
// 格式：可点击节点：\n- [文本] center=(x,y) bounds=[x0,y0][x1,y1]\n...
func parseXMLToNodeSummary(xmlStr string) string {
	// Android XML 根节点通常是 <hierarchy>，内嵌 <node>
	type hierarchy struct {
		Nodes []xmlNode `xml:"node"`
	}
	var root hierarchy
	if err := xml.Unmarshal([]byte(xmlStr), &root); err != nil {
		// 解析失败：尝试截短原始 XML 返回
		if len(xmlStr) > 500 {
			return xmlStr[:500] + "...(truncated)"
		}
		return xmlStr
	}

	var lines []string
	var walk func(n xmlNode)
	walk = func(n xmlNode) {
		label := strings.TrimSpace(n.Text)
		if label == "" {
			label = strings.TrimSpace(n.ContentDesc)
		}
		// 只收录有文字标签且可点击/可聚焦的节点
		if label != "" && (n.Clickable == "true" || n.Focusable == "true") && n.Bounds != "" {
			cx, cy := boundsCenter(n.Bounds)
			if cx >= 0 {
				lines = append(lines, fmt.Sprintf("- [%s] center=(%d,%d) bounds=%s", label, cx, cy, n.Bounds))
			}
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, n := range root.Nodes {
		walk(n)
	}

	if len(lines) == 0 {
		// 没有可交互节点时降级返回截短 XML
		if len(xmlStr) > 500 {
			return xmlStr[:500] + "...(truncated)"
		}
		return xmlStr
	}
	return "可点击节点：\n" + strings.Join(lines, "\n")
}

// boundsCenter 解析 bounds 字符串，如 "[45,58][180,98]"，返回中心坐标
func boundsCenter(bounds string) (int, int) {
	var x0, y0, x1, y1 int
	n, _ := fmt.Sscanf(bounds, "[%d,%d][%d,%d]", &x0, &y0, &x1, &y1)
	if n != 4 {
		return -1, -1
	}
	return (x0 + x1) / 2, (y0 + y1) / 2
}

// RpaGetDeviceEnv 获取设备环境信息（已安装 App 列表 + 屏幕分辨率）
func (a *App) RpaGetDeviceEnv(req RpaRequest) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	cmd := C.CString("pm list packages -3 && wm size")
	defer C.free(unsafe.Pointer(cmd))
	resPtr := C.execCmd(handle, C.int(1), cmd)
	if resPtr == nil {
		return map[string]interface{}{"success": false, "message": "execCmd 返回空"}
	}
	defer C.freeRpcPtr(unsafe.Pointer(resPtr))
	output := C.GoString(resPtr)

	return map[string]interface{}{
		"success": true,
		"output":  output,
	}
}

// RpaClick 点击指定坐标（含随机偏移，模拟真人点击）
func (a *App) RpaClick(req RpaRequest, x, y int) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	// 随机偏移 ±25px，模拟真人手指不精准的点击区域
	rx := x + rand.Intn(51) - 25
	ry := y + rand.Intn(51) - 25

	ret := C.touchClick(handle, C.int(0), C.int(rx), C.int(ry))
	if ret == 0 {
		return map[string]interface{}{"success": false, "message": "点击失败"}
	}
	// 点击后随机等待 1200~2000ms，模拟真人反应节奏
	waitMs := 1200 + rand.Intn(800)
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return map[string]interface{}{"success": true}
}

// RpaSwipe 滑动操作（慢速 + 随机起止点偏移，模拟真人滑动）
func (a *App) RpaSwipe(req RpaRequest, x0, y0, x1, y1, durationMs int) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	// 默认滑动时长 1000~1400ms（缓慢自然的手指滑动）
	if durationMs <= 0 {
		durationMs = 1000 + rand.Intn(400)
	} else {
		// 即使模型指定了时长，也拉长到至少 900ms，避免过快
		if durationMs < 900 {
			durationMs = 900 + rand.Intn(300)
		}
	}

	// X轴起点/终点随机偏移 ±30px
	rx0 := x0 + rand.Intn(61) - 30
	rx1 := x1 + rand.Intn(61) - 30
	// Y轴起点/终点随机偏移 ±50px
	ry0 := y0 + rand.Intn(101) - 50
	ry1 := y1 + rand.Intn(101) - 50

	C.swipe(handle, C.int(0), C.int(rx0), C.int(ry0), C.int(rx1), C.int(ry1), C.long(durationMs), C.bool(false))

	// 滑动结束后等待内容加载：滑动时长 + 随机 2000~3500ms
	waitMs := durationMs + 2000 + rand.Intn(1500)
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return map[string]interface{}{"success": true}
}

// RpaSendText 向当前输入框发送文字
func (a *App) RpaSendText(req RpaRequest, text string) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	ret := C.sendText(handle, cText)
	if ret == 0 {
		return map[string]interface{}{"success": false, "message": "sendText 失败"}
	}
	time.Sleep(600 * time.Millisecond)
	return map[string]interface{}{"success": true}
}

// RpaKeyPress 按下指定按键码（Android KeyEvent 编码）
func (a *App) RpaKeyPress(req RpaRequest, keyCode int) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	ret := C.keyPress(handle, C.int(keyCode))
	if ret == 0 {
		return map[string]interface{}{"success": false, "message": "keyPress 失败"}
	}
	time.Sleep(800 * time.Millisecond)
	return map[string]interface{}{"success": true}
}

// RpaOpenApp 启动指定包名的 App，等待加载后自动获取界面节点摘要并返回
func (a *App) RpaOpenApp(req RpaRequest, pkg string) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	cPkg := C.CString(pkg)
	defer C.free(unsafe.Pointer(cPkg))
	ret := C.openApp(handle, cPkg)
	if ret == 0 {
		log.Printf("[RPA] openApp ret=0 pkg=%s，等待后继续", pkg)
	}
	// 等待 App 冷启动加载完成（通常需要 2-3s）
	time.Sleep(3000 * time.Millisecond)

	// 自动获取当前界面节点，确认是否已进入目标 App
	xmlPtr := C.dumpNodeXml(handle, C.int(0))
	if xmlPtr != nil {
		defer C.freeRpcPtr(unsafe.Pointer(xmlPtr))
		xmlStr := C.GoString(xmlPtr)
		if len(xmlStr) > 50 {
			summary := parseXMLToNodeSummary(xmlStr)
			return map[string]interface{}{
				"success": true,
				"nodes":   summary,
			}
		}
	}

	// XML 获取失败，界面可能仍在加载，再等 2s 重试一次
	time.Sleep(2000 * time.Millisecond)
	xmlPtr2 := C.dumpNodeXml(handle, C.int(0))
	if xmlPtr2 != nil {
		defer C.freeRpcPtr(unsafe.Pointer(xmlPtr2))
		xmlStr2 := C.GoString(xmlPtr2)
		if len(xmlStr2) > 50 {
			summary := parseXMLToNodeSummary(xmlStr2)
			return map[string]interface{}{
				"success": true,
				"nodes":   summary,
			}
		}
	}

	// 两次都拿不到 XML，返回成功但提示节点为空（让模型自行 get_screen_context 重试）
	return map[string]interface{}{
		"success": true,
		"nodes":   "界面节点暂未就绪，请稍候",
	}
}

// RpaStopApp 停止指定包名的 App
func (a *App) RpaStopApp(req RpaRequest, pkg string) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	cPkg := C.CString(pkg)
	defer C.free(unsafe.Pointer(cPkg))
	ret := C.stopApp(handle, cPkg)
	if ret == 0 {
		return map[string]interface{}{"success": false, "message": "stopApp 失败"}
	}
	return map[string]interface{}{"success": true}
}

// RpaExecCmd 在设备内执行 shell 命令（层2 RPA 通道）
// waitForExit: true 等待命令执行完成再返回，false 异步执行
func (a *App) RpaExecCmd(req RpaRequest, cmd string, waitForExit bool) map[string]interface{} {
	handle, err := rpaGetHandle(req.DeviceIP, req.RpaPort)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	cCmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(cCmd))
	wait := C.int(0)
	if waitForExit {
		wait = C.int(1)
	}
	resPtr := C.execCmd(handle, wait, cCmd)
	if resPtr == nil {
		return map[string]interface{}{"success": true, "output": ""}
	}
	defer C.freeRpcPtr(unsafe.Pointer(resPtr))
	output := C.GoString(resPtr)

	return map[string]interface{}{
		"success": true,
		"output":  output,
	}
}

// RpaCloseConnection 关闭指定设备的 RPA 连接（前端任务结束后调用）
func (a *App) RpaCloseConnection(req RpaRequest) map[string]interface{} {
	rpaReleaseHandle(req.DeviceIP, req.RpaPort)
	return map[string]interface{}{"success": true}
}

// RpaGetVersion 获取 SDK 版本号（用于连接测试）
func (a *App) RpaGetVersion() map[string]interface{} {
	ver := C.getVersion()
	return map[string]interface{}{
		"success": true,
		"version": int(ver),
	}
}
