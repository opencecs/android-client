package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// ========== 设备心跳检测服务 ==========

// 存储空间查询定时器
var storageQueryTicker *time.Ticker
var storageQueryStop chan struct{}

// ========== TCP Ping 核心函数 ==========

// tcpPing 对指定设备执行轻量级TCP连接测试
// 参数:
//   - ip: 设备IP地址
//   - port: TCP端口号(通常为8000)
//   - timeout: 连接超时时间
// 返回:
//   - latency: TCP连接建立耗时(毫秒)
//   - err: 连接错误(nil表示成功)
// 特点:
//   - 仅建立TCP连接,无数据传输,网络开销最小(约60字节)
//   - 立即关闭连接,不占用设备资源
//   - 适用于大规模设备场景的高频心跳检测
func tcpPing(ip string, port int, timeout time.Duration) (latency int64, err error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	start := time.Now()
	
	// 使用 net.DialTimeout 建立TCP连接
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		// 连接失败: 可能是设备离线、端口未监听或网络不通
		return 0, err
	}
	defer conn.Close()
	
	// 计算连接建立耗时
	latency = time.Since(start).Milliseconds()
	return latency, nil
}

// deviceAddr 返回用于 TCP/HTTP 访问的完整 host:port 地址。
// 若 deviceIP 已含端口（如 "1.2.3.4:8187"），直接返回；否则追加默认端口 8000。
func deviceAddr(deviceIP string) string {
	_, _, err := net.SplitHostPort(deviceIP)
	if err == nil {
		// 已含端口，直接使用
		return deviceIP
	}
	return deviceIP + ":8000"
}

// ========== HTTP 接口数据结构 ==========

// DeviceInfoResponse /info 接口返回的数据结构
type DeviceInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		CurrentVersion int `json:"currentVersion"` // 当前API版本
		LatestVersion  int `json:"latestVersion"`  // 最新API版本
	} `json:"data"`
}

// DeviceDetailResponse /info/device 接口返回的数据结构
type DeviceDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"` // 错误消息
	Data struct {
		// 存储信息
		MMCTotal string `json:"mmctotal"` // 总存储空间（MB，字符串格式）
		MMCUse   string `json:"mmcuse"`   // 已用存储空间（MB，字符串格式）
		MMCFree  string `json:"mmcfree"`  // 可用存储空间（MB，字符串格式）
		
		// 固件信息
		Version string `json:"version"` // SDK版本
		Model   string `json:"model"`   // 设备型号
		
		// 系统运行信息
		IP          string `json:"ip"`           // 设备IP
		IP1         string `json:"ip_1"`         // 第二IP
		HWAddr      string `json:"hwaddr"`       // MAC地址
		HWAddr1     string `json:"hwaddr_1"`     // MAC地址2
		CPUTemp     int    `json:"cputemp"`      // CPU温度
		CPULoad     string `json:"cpuload"`      // CPU负载
		MemTotal    string `json:"memtotal"`     // 内存总量（MB）
		MemUse      string `json:"memuse"`       // 内存使用量（MB）
		Speed       string `json:"speed"`        // 网络速度
		MMCRead     string `json:"mmcread"`      // 磁盘读取量
		MMCWrite    string `json:"mmcwrite"`     // 磁盘写入量
		SysUptime   string `json:"sysuptime"`    // 系统运行时间（秒）
		MMCModel    string `json:"mmcmodel"`     // 磁盘型号
		MMCTemp     string `json:"mmctemp"`      // 磁盘温度
		Network4G   string `json:"network4g"`    // 4G网络状态
		NetworkEth0 string `json:"netWork_eth0"` // 以太网状态
		DeviceID    string `json:"deviceId"`     // 设备ID
	} `json:"data"`
}

// ResetAllDevicesOffline 重置所有设备为离线状态(应用启动/重启时调用)
// 这个方法用于应用启动时强制重置所有设备状态,确保从离线开始重新验证
func (a *App) ResetAllDevicesOffline() {
	a.deviceStatusMutex.Lock()
	defer a.deviceStatusMutex.Unlock()
	
	resetCount := 0
	for ip, status := range a.deviceStatusMap {
		if status.Status != "offline" {
			status.Status = "offline"
			status.ConsecutiveSuccesses = 0
			status.ConsecutiveFailures = 0
			log.Printf("[心跳] 🔄 应用重启: 设备 %s 重置为离线状态", ip)
			resetCount++
		}
	}
	
	log.Printf("[心跳] 应用重启: 共重置 %d 个设备为离线状态", resetCount)
}

// getDeviceName 根据IP获取设备名称
func (a *App) getDeviceName(ip string) string {
	a.deviceNamesMutex.RLock()
	defer a.deviceNamesMutex.RUnlock()
	if name, ok := a.deviceNames[ip]; ok && name != "" {
		return name
	}
	return ip
}

// UpdateMonitoredDevices 更新需要监控的设备列表，names 为 IP→名称映射（可传 nil）
func (a *App) UpdateMonitoredDevices(ips []string, names map[string]string) {
	log.Printf("[心跳] ⚠️⚠️⚠️ UpdateMonitoredDevices 被调用! 设备数量: %d", len(ips))

	// 保存设备名称映射
	a.deviceNamesMutex.Lock()
	a.deviceNames = make(map[string]string, len(ips))
	for _, ip := range ips {
		if names != nil {
			if n, ok := names[ip]; ok {
				a.deviceNames[ip] = n
			}
		}
	}
	a.deviceNamesMutex.Unlock()

	a.deviceIPsMutex.Lock()
	a.deviceIPs = make([]string, len(ips))
	copy(a.deviceIPs, ips)
	a.deviceIPsMutex.Unlock()

	// ========== 初始化设备状态 ==========
	// 策略: 只初始化新设备,已存在的设备不重置状态
	// 这样可以避免 watch 触发时导致的设备闪烁
	a.deviceStatusMutex.Lock()
	for _, ip := range ips {
		if _, exists := a.deviceStatusMap[ip]; !exists {
			// 新设备: 初始化为离线状态
			a.deviceStatusMap[ip] = &DeviceStatus{
				IP:     ip,
				Status: "offline",
			}
			name := ip
			if names != nil {
				if n, ok := names[ip]; ok && n != "" {
					name = n
				}
			}
			log.Printf("[心跳] 🆕 新设备 %s (%s) 初始化为离线状态", ip, name)
		}
		// 已存在的设备: 不做任何修改,保持当前状态
	}
	a.deviceStatusMutex.Unlock()
	
	// ========== 同步清理不在新设备列表中的设备状态 ==========
	// 创建新设备列表的快速查找map
	newIPSet := make(map[string]bool, len(ips))
	for _, ip := range ips {
		newIPSet[ip] = true
	}
	
	// 清理deviceStatusMap中不在新列表的设备
	a.deviceStatusMutex.Lock()
	var removedIPs []string
	for ip := range a.deviceStatusMap {
		if !newIPSet[ip] {
			delete(a.deviceStatusMap, ip)
			removedIPs = append(removedIPs, ip)
		}
	}
	a.deviceStatusMutex.Unlock()
	
	if len(removedIPs) > 0 {
		log.Printf("[心跳] 清理缓存: 移除 %d 个不在监控列表的设备状态: %v", len(removedIPs), removedIPs)
	}
	
	log.Printf("[心跳] 更新监控设备列表，共 %d 个设备", len(ips))
}

// UpdateDevicePasswords 更新设备密码映射（前端调用）
func (a *App) UpdateDevicePasswords(passwords map[string]string) {
	a.devicePasswordsMutex.Lock()
	// 记录哪些设备的密码发生了变化
	var updatedIPs []string
	for ip, newPassword := range passwords {
		oldPassword, exists := a.devicePasswords[ip]
		if !exists || oldPassword != newPassword {
			updatedIPs = append(updatedIPs, ip)
		}
	}
	a.devicePasswords = passwords
	a.devicePasswordsMutex.Unlock()
	
	log.Printf("[心跳] 更新设备密码映射，共 %d 个设备", len(passwords))
	
	// 为密码已更新的设备触发认证检查
	for _, ip := range updatedIPs {
		go a.OnDevicePasswordUpdated(ip)
	}
}

// StartDeviceHeartbeat 启动设备心跳检测服务
func (a *App) StartDeviceHeartbeat() {
	// 防止重复启动
	a.heartbeatMutex.Lock()
	if a.heartbeatRunning {
		log.Printf("[心跳] ⚠️ 心跳服务已在运行，忽略重复启动请求")
		a.heartbeatMutex.Unlock()
		return
	}
	// 重建 stop channel（防止上次 Stop 后 channel 已关闭导致新 goroutine 立即退出）
	a.heartbeatStop = make(chan struct{})
	a.heartbeatRunning = true
	a.heartbeatMutex.Unlock()
	
	// log.Printf("[心跳] 启动设备心跳检测服务 (TCP Ping模式)")
	
	// ========== 定时器1: TCP Ping心跳检测 (1秒间隔) ==========
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		defer func() {
			a.heartbeatMutex.Lock()
			a.heartbeatRunning = false
			a.heartbeatMutex.Unlock()
			// log.Printf("[TCP Ping] 心跳检测服务已停止")
		}()
		
		// log.Printf("[TCP Ping] 定时器已启动，间隔: 1秒")
		
		// 立即执行一次TCP Ping检测
		a.tcpPingAllDevices()
		
		for {
			select {
			case <-ticker.C:
				a.tcpPingAllDevices()
			case <-a.heartbeatStop:
				// log.Printf("[TCP Ping] 收到停止信号，退出心跳检测")
				return
			}
		}
	}()
	
	// ========== 定时器2: API版本检查和存储查询 (60秒间隔，按需触发) ==========
	// 注意: API版本检查和存储查询在tcpPingSingleDevice中根据状态变化触发
	// 这里启动一个定时器，为所有在线设备定期检查API版本和存储信息
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		
		// log.Printf("[定时检查] 定时器已启动，间隔: 60秒 (API版本+存储信息)")
		
		// 延迟10秒后首次执行，确保设备已上线
		time.Sleep(10 * time.Second)
		
		for {
			select {
			case <-ticker.C:
				// 为所有在线设备触发API检查和存储查询
				a.deviceStatusMutex.RLock()
				var onlineIPs []string
				for ip, status := range a.deviceStatusMap {
					if status.Status == "online" {
						onlineIPs = append(onlineIPs, ip)
					}
				}
				a.deviceStatusMutex.RUnlock()
				
				// 异步检查所有在线设备的API版本和存储信息
				for _, ip := range onlineIPs {
					go func(deviceIP string) {
						a.checkDeviceAPIVersion(deviceIP)
						a.checkDeviceStorage(deviceIP)
					}(ip)
				}
			case <-a.heartbeatStop:
				log.Printf("[定时检查] 收到停止信号，退出定时检查")
				return
			}
		}
	}()

	// ========== 安卓容器列表轮询服务 ==========
	a.StartAndroidPoll()

	// ========== 安卓容器截图轮询服务 ==========
	a.StartScreenshotPoll()
}

// StopDeviceHeartbeat 停止设备心跳检测服务
func (a *App) StopDeviceHeartbeat() {
	// log.Printf("[心跳] 停止设备心跳检测服务")
	close(a.heartbeatStop)
	// 注意：heartbeatStop channel 在 StartDeviceHeartbeat 启动时会重建，此处无需重建
}

// checkAllDevicesStatus 并发检测所有设备状态（使用并发控制）
func (a *App) checkAllDevicesStatus() {
	a.deviceIPsMutex.RLock()
	deviceIPs := make([]string, len(a.deviceIPs))
	copy(deviceIPs, a.deviceIPs)
	a.deviceIPsMutex.RUnlock()
	
	if len(deviceIPs) == 0 {
		return
	}
	
	// log.Printf("[心跳] 开始检测 %d 个设备状态", len(deviceIPs))
	
	// ========== 大规模设备并发策略 ==========
	// 根据设备数量动态调整并发数
	// 设计目标: 5000台设备在3秒内完成检测
	
	deviceCount := len(deviceIPs)
	var maxWorkers int
	
	switch {
	case deviceCount <= 50:
		// 小规模: 使用 CPU×2
		maxWorkers = runtime.NumCPU() * 2
	case deviceCount <= 500:
		// 中规模: 使用 CPU×4
		maxWorkers = runtime.NumCPU() * 4
	case deviceCount <= 2000:
		// 大规模: 固定200个并发
		maxWorkers = 200
	default:
		// 超大规模: 固定500个并发 (5000台÷500并发≈10秒完成)
		maxWorkers = 500
	}
	
	// 创建信号量（buffered channel）控制并发数
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	
	// 并发检测所有设备，但限制同时执行的goroutine数量
	for _, ip := range deviceIPs {
		wg.Add(1)
		
		// 获取信号量
		semaphore <- struct{}{}
		
		go func(deviceIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			a.checkSingleDeviceStatus(deviceIP)
		}(ip)
	}
	
	// 等待所有检测完成
	wg.Wait()
	
	// log.Printf("[心跳] 完成 %d 个设备检测，耗时: %v，并发数: %d", len(deviceIPs), time.Since(startTime), maxWorkers)
}

// checkSingleDeviceStatus 检测单个设备状态并获取API版本（仅支持 V3 设备）
// 只负责判断在线/离线和获取API版本，不查询存储信息
func (a *App) checkSingleDeviceStatus(deviceIP string) {
	checkStart := time.Now()
	
	// 使用 /info 端点检查在线状态和API版本
	infoURL := fmt.Sprintf("http://%s/info", deviceAddr(deviceIP))
	
	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		responseTime := time.Since(checkStart).Milliseconds()
		a.updateDeviceStatus(deviceIP, "offline", responseTime, nil)
		log.Printf("[心跳] ❌ 设备 %s 创建请求失败: %v", deviceIP, err)
		return
	}
	
	// 添加认证头（如果有密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()
	
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[心跳] 🔑 设备 %s 使用认证", deviceIP)
	}
	
	// 使用共享的 HTTP Client（复用连接）
	resp, err := a.httpClient.Do(req)
	
	if err != nil {
		// 请求失败，设备离线
		responseTime := time.Since(checkStart).Milliseconds()
		a.updateDeviceStatus(deviceIP, "offline", responseTime, nil)
		log.Printf("[心跳] ❌ 设备 %s 离线 (错误: %v)", deviceIP, err)
		return
	}
	defer resp.Body.Close()
	
	// 401 表示设备端口可达但认证失败，不能视为在线
	if resp.StatusCode == 401 {
		responseTime := time.Since(checkStart).Milliseconds()
		// 无论是否有密码，401 都意味着未通过认证，标记离线并通知前端输入密码
		a.updateDeviceStatus(deviceIP, "offline", responseTime, nil)
		if password != "" {
			log.Printf("[心跳] 🔒 设备 %s 密码认证失败 (状态码: 401)，密码可能错误", deviceIP)
		} else {
			log.Printf("[心跳] 🔑 设备 %s 需要认证密码 (状态码: 401)", deviceIP)
		}
		go a.addToAuthQueue(deviceIP)
		return
	}
	
	// 只有 2xx 状态码才认为设备在线
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// 读取并解析 /info 接口的 JSON
		body, err := io.ReadAll(resp.Body)
		responseTime := time.Since(checkStart).Milliseconds() // 读取完响应体后再计算时间
		
		if err != nil {
			log.Printf("[心跳] ⚠️ 设备 %s 读取 /info 响应失败: %v", deviceIP, err)
			a.updateDeviceStatus(deviceIP, "online", responseTime, nil)
			return
		}
		
		var infoResp DeviceInfoResponse
		if err := json.Unmarshal(body, &infoResp); err != nil {
			log.Printf("[心跳] ⚠️ 设备 %s 解析 /info JSON失败: %v, 原始数据: %s", deviceIP, err, string(body))
			a.updateDeviceStatus(deviceIP, "online", responseTime, nil)
			return
		}
		
		log.Printf("[心跳] 🔍 设备 %s 解析 /info 成功: code=%d, currentVersion=%d, latestVersion=%d", 
			deviceIP, infoResp.Code, infoResp.Data.CurrentVersion, infoResp.Data.LatestVersion)
		
		// 更新设备状态和API版本
		a.updateDeviceStatus(deviceIP, "online", responseTime, &infoResp)
		
		log.Printf("[心跳] ✅ 设备 %s 在线 (响应: %dms, API: %d/%d)", 
			deviceIP, 
			responseTime,
			infoResp.Data.CurrentVersion,
			infoResp.Data.LatestVersion,
		)
	} else {
		// 其他状态码（3xx、4xx、5xx）都认为设备异常
		responseTime := time.Since(checkStart).Milliseconds()
		a.updateDeviceStatus(deviceIP, "offline", responseTime, nil)
		log.Printf("[心跳] ⚠️ 设备 %s 异常 (状态码: %d)", deviceIP, resp.StatusCode)
	}
}

// ========== TCP Ping 状态机实现 ==========

// tcpPingSingleDevice 对单个设备执行TCP Ping+HTTP验证并更新状态
// 实现状态机逻辑:
//   - 先进行TCP Ping检测端口连通性和延迟
//   - TCP成功后进行HTTP /info验证确认是我们的设备服务
//   - 连续4次失败(TCP失败或HTTP验证失败) → 标记为离线
//   - 连续2次成功(TCP<50ms且HTTP验证通过) → 标记为在线
//   - 状态变化(离线→在线)时检查401认证,通过后触发API版本检查和存储查询
func (a *App) tcpPingSingleDevice(deviceIP string) {
	// ========== 1. 执行TCP Ping ==========
	// TCP连接超时设置为2000ms，容忍局域网抖动和公网延迟
	addr := deviceAddr(deviceIP)
	start := time.Now()
	conn, tcpErr := net.DialTimeout("tcp", addr, 2000*time.Millisecond)
	var latency int64
	var err error
	if tcpErr != nil {
		err = tcpErr
	} else {
		latency = time.Since(start).Milliseconds()
		conn.Close()
	}
	_ = latency
	
	// ========== 2. 并发安全 - 获取并更新设备状态 ==========
	a.deviceStatusMutex.Lock()
	defer a.deviceStatusMutex.Unlock()
	
	// 获取或创建设备状态
	status := a.deviceStatusMap[deviceIP]
	if status == nil {
		status = &DeviceStatus{
			IP:     deviceIP,
			Status: "offline", // 新设备默认离线状态,需要连续2次验证通过才上线
		}
		a.deviceStatusMap[deviceIP] = status
		// log.Printf("[TCP Ping] 🆕 新设备 %s 初始化为离线状态", deviceIP)
	}
	
	// 记录当前状态(用于判断状态变化)
	oldStatus := status.Status
	now := time.Now()
	
	// ========== 3. 处理Ping结果 ==========
	// 延迟阈值500ms，局域网正常时远不会触发，容忍网络抖动
	if err != nil || latency > 500 {
		// ========== TCP失败或延迟>150ms处理 ==========
		status.ConsecutiveSuccesses = 0      // 清零成功计数
		status.ConsecutiveFailures++         // 增加失败计数
		status.LastCheckAt = now
		
		if err != nil {
			log.Printf("[TCP Ping] ❌ 设备 %s (%s) TCP连接失败 (连续失败%d次): %v", deviceIP, a.getDeviceName(deviceIP), status.ConsecutiveFailures, err)
		} else {
			log.Printf("[TCP Ping] ⚠️ 设备 %s (%s) 延迟过高: %dms > 150ms (连续失败%d次)", deviceIP, a.getDeviceName(deviceIP), latency, status.ConsecutiveFailures)
			status.ResponseTime = latency
		}
		
		// 连续6次失败 → 标记离线（容忍短暂抖动，约6秒）
		if status.ConsecutiveFailures >= 6 {
			if status.Status != "offline" {
				status.Status = "offline"
				log.Printf("[TCP Ping] ❌ 设备 %s (%s) 离线 (连续%d次失败, 原因: %v)", deviceIP, a.getDeviceName(deviceIP), status.ConsecutiveFailures, err)
			}
		}
		
		// 失败时显示上次成功的延迟(如果有)
		if err != nil && status.LastSuccessLatency > 0 {
			status.ResponseTime = status.LastSuccessLatency
		}
	} else {
		// ========== TCP Ping成功且延迟≤150ms,进行HTTP验证 ==========
		// 释放锁,避免HTTP请求阻塞其他设备的状态更新
		a.deviceStatusMutex.Unlock()

		// 进行HTTP /info验证,确认这是我们的设备服务,并获取详细结果
		httpResult, _ := a.quickHttpCheckResult(deviceIP)

		// 重新获取锁
		a.deviceStatusMutex.Lock()

		// 再次获取status(可能在HTTP请求期间被修改)
		status = a.deviceStatusMap[deviceIP]
		if status == nil {
			return // 设备已被移除
		}

		if httpResult == httpCheckFailed {
			// HTTP验证失败,端口被其他服务占用或设备API异常
			status.ConsecutiveSuccesses = 0
			status.ConsecutiveFailures++
			status.LastCheckAt = now

			log.Printf("[TCP Ping] ❌ 设备 %s (%s) HTTP验证失败 (端口可能被占用, 连续失败%d次)", deviceIP, a.getDeviceName(deviceIP), status.ConsecutiveFailures)

			// 连续6次失败 → 标记离线
			if status.ConsecutiveFailures >= 6 {
				if status.Status != "offline" {
					status.Status = "offline"
					log.Printf("[TCP Ping] ❌ 设备 %s (%s) 离线 (HTTP验证失败, 连续%d次)", deviceIP, a.getDeviceName(deviceIP), status.ConsecutiveFailures)
				}
			}
			return
		}

		if httpResult == httpCheckAuthRequired {
			// 设备端口可达但需要认证,不标记为 online,通知前端输入密码
			status.ConsecutiveSuccesses = 0
			status.ConsecutiveFailures = 0
			status.LastCheckAt = now
			// 保持或设置为 offline,直到认证成功后的 /info 返回 2xx
			if status.Status != "offline" && status.Status != "" {
				// 已经在线的设备突然变成401(密码变了),标为离线
				status.Status = "offline"
				log.Printf("[TCP Ping] 🔒 设备 %s (%s) 认证失败，标记离线等待重新认证", deviceIP, a.getDeviceName(deviceIP))
			} else {
				log.Printf("[TCP Ping] 🔒 设备 %s (%s) 需要认证，等待前端输入密码", deviceIP, a.getDeviceName(deviceIP))
			}
			go a.addToAuthQueue(deviceIP)
			return
		}

		// ========== httpCheckOK: TCP和HTTP(2xx)都成功 ==========
		status.ConsecutiveFailures = 0      // 清零失败计数
		status.ConsecutiveSuccesses++       // 增加成功计数
		status.ResponseTime = latency       // 更新当前延迟
		status.LastSuccessLatency = latency // 缓存成功延迟
		status.LastCheckAt = now

		// 输出每个设备的独立延迟(用于调试和监控)
		// log.Printf("[TCP Ping] 📡 设备 %s (%s) 延迟: %dms (连续成功%d次)", deviceIP, a.getDeviceName(deviceIP), latency, status.ConsecutiveSuccesses)

		// 连续2次成功 → 标记在线
		if status.ConsecutiveSuccesses >= 2 {
			if status.Status != "online" {
				status.Status = "online"
				log.Printf("[TCP Ping] ✅ 设备 %s (%s) 上线 (连续%d次验证通过, 延迟%dms)",
					deviceIP, a.getDeviceName(deviceIP), status.ConsecutiveSuccesses, latency)

				// 离线→在线状态变化: 触发API版本检查和存储查询
				if oldStatus == "offline" || oldStatus == "" {
					// 释放锁后异步执行API查询(此时已经过认证,无需再checkDeviceAuth)
					go func() {
						a.checkDeviceAPIVersion(deviceIP)
						// 离线→在线时立即查询存储(不受60秒限制)
						a.deviceStatusMutex.Lock()
						if st := a.deviceStatusMap[deviceIP]; st != nil {
							st.LastStorageCheckTime = time.Time{} // 重置为零值,强制立即查询
						}
						a.deviceStatusMutex.Unlock()
						a.checkDeviceStorage(deviceIP)
					}()
				}
			}
		}
	}
}

// quickHttpCheck 快速HTTP检查,验证8000端口上运行的是我们的设备API服务
// 返回: true=认证通过(2xx), false=认证失败/网络异常/非我们的服务
// 注意: 401 表示认证失败,不视为在线,调用方需通过 quickHttpCheckResult 获取完整结果
func (a *App) quickHttpCheck(deviceIP string) bool {
	result, _ := a.quickHttpCheckResult(deviceIP)
	return result == httpCheckOK
}

// httpCheckResult 表示 HTTP 检查的详细结果
type httpCheckResult int

const (
	httpCheckOK          httpCheckResult = iota // 2xx: 认证通过,设备在线
	httpCheckAuthRequired                        // 401: 设备需要认证(端口可达但未授权)
	httpCheckFailed                              // 其他: 网络错误/非我们的服务
)

// quickHttpCheckResult 返回详细的HTTP检查结果
func (a *App) quickHttpCheckResult(deviceIP string) (httpCheckResult, int) {
	infoURL := fmt.Sprintf("http://%s/info", deviceAddr(deviceIP))

	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		return httpCheckFailed, 0
	}

	// 添加认证(如果有)
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return httpCheckFailed, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// 认证通过(有密码时密码正确,无密码时设备无需认证)
		return httpCheckOK, resp.StatusCode
	}

	if resp.StatusCode == 401 {
		// 端口可达但认证失败: 设备存在但未授权,需要密码
		log.Printf("[HTTP验证] 🔒 设备 %s 返回 401,需要认证", deviceIP)
		return httpCheckAuthRequired, 401
	}

	// 其他状态码,可能不是我们的服务
	log.Printf("[HTTP验证] ⚠️ 设备 %s 返回状态码 %d, 可能不是我们的设备服务", deviceIP, resp.StatusCode)
	return httpCheckFailed, resp.StatusCode
}

// ========== API版本检查函数 ==========

// checkDeviceAPIVersion 检查设备API版本(复用HTTP /info接口)
// 触发条件:
//   1. 设备从离线→在线时立即触发
//   2. 设备持续在线每60秒触发一次
// 功能:
//   - 调用HTTP /info接口获取API版本
//   - 更新DeviceStatus的APIVersion和LatestVersion字段
//   - 记录LastAPICheckTime防止频繁调用
func (a *App) checkDeviceAPIVersion(deviceIP string) {
	// ========== 1. 检查是否需要查询 ==========
	a.deviceStatusMutex.RLock()
	status := a.deviceStatusMap[deviceIP]
	shouldCheck := false
	
	if status != nil && status.Status == "online" {
		// 检查距离上次API检查是否超过60秒
		if time.Since(status.LastAPICheckTime) >= 60*time.Second {
			shouldCheck = true
		}
	}
	a.deviceStatusMutex.RUnlock()
	
	if !shouldCheck {
		return
	}
	
	log.Printf("[API检查] 🔍 开始检查设备 %s 的API版本", deviceIP)
	
	// ========== 2. 调用HTTP /info接口 ==========
	infoURL := fmt.Sprintf("http://%s/info", deviceAddr(deviceIP))
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		log.Printf("[API检查] ❌ 设备 %s 创建请求失败: %v", deviceIP, err)
		return
	}
	
	// 添加认证(如果有)
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()
	
	if password != "" {
		req.SetBasicAuth("admin", password)
	}
	
	// 发送请求
	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Printf("[API检查] ❌ 设备 %s 请求失败: %v", deviceIP, err)
		return
	}
	defer resp.Body.Close()
	
	// ========== 3. 解析响应 ==========
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[API检查] ⚠️ 设备 %s 读取响应失败: %v", deviceIP, err)
			return
		}
		
		var infoResp DeviceInfoResponse
		if err := json.Unmarshal(body, &infoResp); err != nil {
			log.Printf("[API检查] ⚠️ 设备 %s 解析JSON失败: %v", deviceIP, err)
			return
		}
		
		// ========== 4. 更新API版本信息 ==========
		a.deviceStatusMutex.Lock()
		if status := a.deviceStatusMap[deviceIP]; status != nil {
			status.APIVersion = fmt.Sprintf("%d", infoResp.Data.CurrentVersion)
			status.LatestVersion = fmt.Sprintf("%d", infoResp.Data.LatestVersion)
			status.LastAPICheckTime = time.Now()
			
			log.Printf("[API检查] ✅ 设备 %s API版本: current=%s, latest=%s", 
				deviceIP, status.APIVersion, status.LatestVersion)
		}
		a.deviceStatusMutex.Unlock()
	} else {
		log.Printf("[API检查] ⚠️ 设备 %s 状态码异常: %d", deviceIP, resp.StatusCode)
	}
}

// ========== 存储查询函数 ==========

// checkDeviceStorage 检查设备存储信息(复用HTTP /info/device接口)
// 触发条件:
//   1. 设备从离线→在线时立即触发
//   2. 设备持续在线每60秒触发一次
// 功能:
//   - 调用HTTP /info/device接口获取存储和系统信息
//   - 更新DeviceStatus的存储、固件、系统运行信息字段
//   - 记录LastStorageCheckTime防止频繁调用
func (a *App) checkDeviceStorage(deviceIP string) {
	// ========== 1. 检查是否需要查询 ==========
	a.deviceStatusMutex.RLock()
	status := a.deviceStatusMap[deviceIP]
	shouldCheck := false
	
	if status != nil && status.Status == "online" {
		// 检查距离上次存储查询是否超过60秒
		if time.Since(status.LastStorageCheckTime) >= 60*time.Second {
			shouldCheck = true
		}
	}
	a.deviceStatusMutex.RUnlock()
	
	if !shouldCheck {
		return
	}
	
	// ========== 2. 调用querySingleDeviceStorage ==========
	// 复用现有的存储查询逻辑
	success := a.querySingleDeviceStorage(deviceIP)
	
	// ========== 3. 更新LastStorageCheckTime ==========
	if success {
		a.deviceStatusMutex.Lock()
		if status := a.deviceStatusMap[deviceIP]; status != nil {
			status.LastStorageCheckTime = time.Now()
		}
		a.deviceStatusMutex.Unlock()
	}
}

// ========== 并发TCP Ping函数 ==========

// tcpPingAllDevices 并发检测所有设备(使用动态Worker Pool)
// 并发策略:
//   - ≤50台设备: CPU×4个worker
//   - ≤500台设备: CPU×8个worker
//   - ≤2000台设备: 500个worker
//   - 5000台设备: 1000个worker
// 特点:
//   - 使用信号量控制并发数,避免goroutine爆炸
//   - TCP Ping轻量级,可承受比HTTP更高的并发
//   - 记录批量执行的总耗时和成功率
func (a *App) tcpPingAllDevices() {
	// ========== 1. 复制设备列表(减少锁持有时间) ==========
	a.deviceIPsMutex.RLock()
	deviceIPs := make([]string, len(a.deviceIPs))
	copy(deviceIPs, a.deviceIPs)
	a.deviceIPsMutex.RUnlock()
	
	if len(deviceIPs) == 0 {
		return
	}
	
	// ========== 2. 根据设备数量动态调整并发数 ==========
	deviceCount := len(deviceIPs)
	var maxWorkers int
	
	switch {
	case deviceCount <= 50:
		// 小规模: 使用 CPU×4
		maxWorkers = runtime.NumCPU() * 4
	case deviceCount <= 500:
		// 中规模: 使用 CPU×8
		maxWorkers = runtime.NumCPU() * 8
	case deviceCount <= 2000:
		// 大规模: 固定500个并发
		maxWorkers = 500
	default:
		// 超大规模: 固定1000个并发 (5000台÷1000≈5秒完成)
		maxWorkers = 1000
	}
	
	// ========== 3. 创建信号量控制并发数 ==========
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	
	// ========== 4. 并发执行TCP Ping ==========
	for _, ip := range deviceIPs {
		wg.Add(1)
		
		// 获取信号量
		semaphore <- struct{}{}
		
		go func(deviceIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			a.tcpPingSingleDevice(deviceIP)
		}(ip)
	}
	
	// ========== 5. 等待所有检测完成 ==========
	wg.Wait()
	
	// ========== 6. 统计成功率 ==========
	// log.Printf("[TCP Ping] 完成批量检测: %d个设备, 耗时%v, 并发%d", 
	// 	deviceCount, time.Since(startTime), maxWorkers)
}







// updateDeviceStatus 更新设备状态和信息（并发安全）
// 兼容TCP Ping和HTTP API检查两种数据来源
// 只处理在线/离线状态和API版本，存储信息由 querySingleDeviceStorage 单独更新
func (a *App) updateDeviceStatus(ip, status string, responseTime int64, infoResp *DeviceInfoResponse) {
	a.deviceStatusMutex.Lock()
	defer a.deviceStatusMutex.Unlock()
	
	// 创建或更新设备状态
	statusData := &DeviceStatus{
		IP:           ip,
		Status:       status,
		LastCheckAt:  time.Now(),
		ResponseTime: responseTime,
	}
	
	// 如果有 /info 接口返回的信息（API版本）
	if infoResp != nil && infoResp.Code == 0 {
		statusData.APIVersion = fmt.Sprintf("%d", infoResp.Data.CurrentVersion)
		statusData.LatestVersion = fmt.Sprintf("%d", infoResp.Data.LatestVersion)
		log.Printf("[心跳] 💾 设备 %s 保存API版本: current=%s, latest=%s", ip, statusData.APIVersion, statusData.LatestVersion)
	} else {
		// 保留旧的API版本信息
		if oldStatus, exists := a.deviceStatusMap[ip]; exists {
			statusData.APIVersion = oldStatus.APIVersion
			statusData.LatestVersion = oldStatus.LatestVersion
			log.Printf("[心跳] 📦 设备 %s 保留旧API版本: current=%s, latest=%s", ip, statusData.APIVersion, statusData.LatestVersion)
		} else {
			log.Printf("[心跳] ⚠️ 设备 %s 无API版本信息", ip)
		}
	}
	
	// 保留旧的存储信息和TCP Ping状态（由其他函数单独更新）
	if oldStatus, exists := a.deviceStatusMap[ip]; exists {
		// 存储信息
		statusData.StorageTotal = oldStatus.StorageTotal
		statusData.StorageFree = oldStatus.StorageFree
		statusData.StorageUsed = oldStatus.StorageUsed
		// 固件信息
		statusData.SDKVersion = oldStatus.SDKVersion
		statusData.DeviceModel = oldStatus.DeviceModel
		// 系统运行信息
		statusData.CPUTemp = oldStatus.CPUTemp
		statusData.CPULoad = oldStatus.CPULoad
		statusData.MemoryTotal = oldStatus.MemoryTotal
		statusData.MemoryUsed = oldStatus.MemoryUsed
		statusData.MMCRead = oldStatus.MMCRead
		statusData.MMCWrite = oldStatus.MMCWrite
		statusData.MMCModel = oldStatus.MMCModel
		statusData.MMCTemp = oldStatus.MMCTemp
		statusData.SysUptime = oldStatus.SysUptime
		statusData.Speed = oldStatus.Speed
		statusData.Network4G = oldStatus.Network4G
		statusData.NetworkEth0 = oldStatus.NetworkEth0
		statusData.HWAddr = oldStatus.HWAddr
		statusData.HWAddr1 = oldStatus.HWAddr1
		statusData.IP1 = oldStatus.IP1
		statusData.DeviceID = oldStatus.DeviceID
		
		// ========== 保留TCP Ping状态跟踪字段 ==========
		statusData.ConsecutiveFailures = oldStatus.ConsecutiveFailures
		statusData.ConsecutiveSuccesses = oldStatus.ConsecutiveSuccesses
		statusData.LastSuccessLatency = oldStatus.LastSuccessLatency
		statusData.LastAPICheckTime = oldStatus.LastAPICheckTime
	}
	
	a.deviceStatusMap[ip] = statusData
	
	// 输出最终存储的数据用于调试
	log.Printf("[心跳] 📊 设备 %s 最终状态: status=%s, api=%s/%s, storage=%dMB, sdk=%s, model=%s", 
		ip, statusData.Status, statusData.APIVersion, statusData.LatestVersion, 
		statusData.StorageTotal, statusData.SDKVersion, statusData.DeviceModel)
}

// GetDevicesStatus 获取所有设备状态（供前端调用）
func (a *App) GetDevicesStatus() map[string]*DeviceStatus {
	a.deviceStatusMutex.RLock()
	defer a.deviceStatusMutex.RUnlock()
	
	// 复制一份返回，避免并发问题
	result := make(map[string]*DeviceStatus, len(a.deviceStatusMap))
	for ip, status := range a.deviceStatusMap {
		statusCopy := *status
		result[ip] = &statusCopy
	}
	
	return result
}

// GetDeviceStatus 获取单个设备状态（供前端调用）
func (a *App) GetDeviceStatus(ip string) *DeviceStatus {
	a.deviceStatusMutex.RLock()
	defer a.deviceStatusMutex.RUnlock()
	
	if status, exists := a.deviceStatusMap[ip]; exists {
		statusCopy := *status
		return &statusCopy
	}
	
	return nil
}

// ForceRefreshDeviceInfo 强制刷新设备信息(API版本和存储)
// 供前端"刷新"按钮调用,不管设备在线离线都会查询
func (a *App) ForceRefreshDeviceInfo(deviceIPs []string) map[string]interface{} {
	log.Printf("[强制刷新] 收到强制刷新请求，设备数: %d", len(deviceIPs))
	
	if len(deviceIPs) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "设备列表为空",
		}
	}
	
	// 🔧 重置所有设备的LastAPICheckTime和LastStorageCheckTime
	a.deviceStatusMutex.Lock()
	for _, ip := range deviceIPs {
		if status := a.deviceStatusMap[ip]; status != nil {
			status.LastAPICheckTime = time.Time{}      // 重置为零值,强制刷新
			status.LastStorageCheckTime = time.Time{} // 重置存储查询时间
		}
	}
	a.deviceStatusMutex.Unlock()
	
	log.Printf("[强制刷新] ✅ 已重置 %d 个设备的查询时间", len(deviceIPs))
	
	// 🔧 立即触发API版本检查和存储查询
	go func() {
		log.Printf("[强制刷新] 开始立即查询设备API版本和存储信息...")
		for _, ip := range deviceIPs {
			a.checkDeviceAPIVersion(ip)
			a.checkDeviceStorage(ip)
		}
		log.Printf("[强制刷新] ✅ 已触发 %d 个设备的API版本和存储查询", len(deviceIPs))
	}()
	
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("已触发 %d 个设备的刷新", len(deviceIPs)),
	}
}

// ========== 存储空间查询服务 ==========

// queryAllOnlineDevicesStorage 并发查询所有在线设备的存储信息（使用并发控制）
func (a *App) queryAllOnlineDevicesStorage() {
	// 获取所有在线设备
	a.deviceStatusMutex.RLock()
	var onlineIPs []string
	for ip, status := range a.deviceStatusMap {
		if status.Status == "online" {
			onlineIPs = append(onlineIPs, ip)
		}
	}
	a.deviceStatusMutex.RUnlock()
	
	if len(onlineIPs) == 0 {
		log.Printf("[存储] 没有在线设备需要查询")
		return
	}
	
	log.Printf("[存储] 开始查询 %d 个在线设备的存储信息", len(onlineIPs))
	startTime := time.Now()
	
	// ========== 大规模设备并发策略 ==========
	// 存储查询比心跳慢(15秒超时),需要更多并发
	deviceCount := len(onlineIPs)
	var maxWorkers int
	
	switch {
	case deviceCount <= 50:
		// 小规模: 使用 CPU×2
		maxWorkers = runtime.NumCPU() * 2
	case deviceCount <= 500:
		// 中规模: 使用 CPU×4
		maxWorkers = runtime.NumCPU() * 4
	case deviceCount <= 2000:
		// 大规模: 固定300个并发
		maxWorkers = 300
	default:
		// 超大规模: 固定800个并发 (5000台÷800并发≈6秒完成)
		maxWorkers = 800
	}
	
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	successCount := 0
	var successMutex sync.Mutex
	
	// 并发查询所有在线设备，但限制同时执行的数量
	for _, ip := range onlineIPs {
		wg.Add(1)
		
		// 获取信号量
		semaphore <- struct{}{}
		
		go func(deviceIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			if a.querySingleDeviceStorage(deviceIP) {
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}
		}(ip)
	}
	
	// 等待所有查询完成
	wg.Wait()
	
	elapsed := time.Since(startTime)
	log.Printf("[存储] 完成查询，成功: %d/%d，耗时: %v，并发数: %d", successCount, len(onlineIPs), elapsed, maxWorkers)
}

// querySingleDeviceStorage 查询单个设备的存储信息
func (a *App) querySingleDeviceStorage(deviceIP string) bool {
	// 使用 /info/device 接口获取存储信息
	detailURL := fmt.Sprintf("http://%s/info/device", deviceAddr(deviceIP))
	
	// 创建带超时的 context (设备负载高时命令执行可能需要更长时间)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", detailURL, nil)
	if err != nil {
		log.Printf("[存储] ❌ 设备 %s 创建请求失败: %v", deviceIP, err)
		return false
	}
	
	// 添加认证头（如果有密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()
	
	if password != "" {
		req.SetBasicAuth("admin", password)
	}
	
	// 使用共享的 HTTP Client（复用连接）
	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Printf("[存储] ❌ 设备 %s 查询失败: %v", deviceIP, err)
		return false
	}
	defer resp.Body.Close()
	
	// 401表示需要认证
	if resp.StatusCode == 401 {
		log.Printf("[存储] 🔒 设备 %s 需要认证 (状态码: 401)", deviceIP)
		return false
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[存储] ⚠️ 设备 %s 返回状态码: %d", deviceIP, resp.StatusCode)
		return false
	}
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[存储] ❌ 设备 %s 读取响应失败: %v", deviceIP, err)
		return false
	}
	
	// 输出原始响应用于调试
	log.Printf("[存储] 🔍 设备 %s 原始响应: %s", deviceIP, string(body))
	
	// 解析 JSON
	var detailResp DeviceDetailResponse
	if err := json.Unmarshal(body, &detailResp); err != nil {
		log.Printf("[存储] ❌ 设备 %s 解析JSON失败: %v", deviceIP, err)
		return false
	}
	
	if detailResp.Code != 0 {
		// Code 50: 设备内部命令执行超时（设备负载高或系统繁忙）
		if detailResp.Code == 50 {
			log.Printf("[存储] ⚠️ 设备 %s 内部命令超时 (code=50, message=%s)，保留旧数据，等待下次查询", 
				deviceIP, detailResp.Message)
			// 返回 false 表示本次查询失败，但不清空已有数据
			return false
		}
		log.Printf("[存储] ⚠️ 设备 %s 返回错误码: %d, 消息: %s", deviceIP, detailResp.Code, detailResp.Message)
		return false
	}
	
	// 辅助函数：安全转换字符串为 int64
	safeParseInt := func(s string, fieldName string) int64 {
		if s == "" {
			log.Printf("[存储] ⚠️ 设备 %s %s为空字符串，使用默认值0", deviceIP, fieldName)
			return 0
		}
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Printf("[存储] ⚠️ 设备 %s %s转换失败: %v (原始值: %s)，使用默认值0", deviceIP, fieldName, err, s)
			return 0
		}
		return val
	}
	
	// 转换字符串为 int64
	storageTotal := safeParseInt(detailResp.Data.MMCTotal, "MMCTotal")
	storageUsed := safeParseInt(detailResp.Data.MMCUse, "MMCUse")
	memTotal := safeParseInt(detailResp.Data.MemTotal, "MemTotal")
	memUse := safeParseInt(detailResp.Data.MemUse, "MemUse")
	
	// 计算可用空间：总空间 - 已用空间
	storageFree := int64(0)
	if storageTotal > 0 && storageUsed > 0 {
		storageFree = storageTotal - storageUsed
	}
	
	log.Printf("[存储] 📊 设备 %s 存储计算: 总计=%dMB, 已用=%dMB, 可用=%dMB", 
		deviceIP, storageTotal, storageUsed, storageFree)
	
	// 获取SDK版本和设备型号
	sdkVersion := detailResp.Data.Version
	deviceModel := detailResp.Data.Model
	log.Printf("[存储] 📱 设备 %s 固件信息: SDK版本=%s, 设备型号=%s", 
		deviceIP, sdkVersion, deviceModel)
	
	// 更新所有设备信息
	a.deviceStatusMutex.Lock()
	if status, exists := a.deviceStatusMap[deviceIP]; exists {
		// 存储信息
		status.StorageTotal = storageTotal
		status.StorageFree = storageFree
		status.StorageUsed = storageUsed
		
		// 固件信息
		status.SDKVersion = sdkVersion
		status.DeviceModel = deviceModel
		
		// 系统运行信息
		status.CPUTemp = detailResp.Data.CPUTemp
		status.CPULoad = detailResp.Data.CPULoad
		status.MemoryTotal = memTotal
		status.MemoryUsed = memUse
		status.MMCRead = detailResp.Data.MMCRead
		status.MMCWrite = detailResp.Data.MMCWrite
		status.MMCModel = detailResp.Data.MMCModel
		status.MMCTemp = detailResp.Data.MMCTemp
		status.SysUptime = detailResp.Data.SysUptime
		status.Speed = detailResp.Data.Speed
		status.Network4G = detailResp.Data.Network4G
		status.NetworkEth0 = detailResp.Data.NetworkEth0
		status.HWAddr = detailResp.Data.HWAddr
		status.HWAddr1 = detailResp.Data.HWAddr1
		status.IP1 = detailResp.Data.IP1
		status.DeviceID = detailResp.Data.DeviceID
		
		log.Printf("[存储] ✅ 设备 %s 完整信息更新: 存储=%dMB/%dMB, CPU=%d°C/%s, 内存=%dMB/%dMB, 网速=%s, SDK=%s", 
			deviceIP, storageFree, storageTotal, status.CPUTemp, status.CPULoad, 
			status.MemoryUsed, status.MemoryTotal, status.Speed, sdkVersion)
	}
	a.deviceStatusMutex.Unlock()
	
	return true
}

// ========== 401认证检查和队列管理 ==========

// checkDeviceAuth 检查设备是否需要401认证
// 返回: true=认证通过或不需要认证, false=需要认证但未通过
func (a *App) checkDeviceAuth(deviceIP string) bool {
	// 检查是否已有密码
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()
	
	// 调用 /info 接口测试认证
	infoURL := fmt.Sprintf("http://%s/info", deviceAddr(deviceIP))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		log.Printf("[认证检查] ❌ 设备 %s 创建请求失败: %v", deviceIP, err)
		return false
	}
	
	if password != "" {
		req.SetBasicAuth("admin", password)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Printf("[认证检查] ❌ 设备 %s 请求失败: %v", deviceIP, err)
		return false
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 401 {
		log.Printf("[认证检查] 🔒 设备 %s 需要401认证", deviceIP)
		return false
	}
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[认证检查] ✅ 设备 %s 认证通过", deviceIP)
		return true
	}
	
	log.Printf("[认证检查] ⚠️ 设备 %s 返回状态码: %d", deviceIP, resp.StatusCode)
	return false
}

// addToAuthQueue 将设备加入401认证队列
// 通过发送事件通知前端,由用户输入密码
func (a *App) addToAuthQueue(deviceIP string) {
	// 检查设备是否已在队列中(通过检查是否已有密码正在等待输入)
	a.devicePasswordsMutex.RLock()
	_, hasPassword := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()
	
	if hasPassword {
		// 已有密码但认证失败,可能是密码错误
		log.Printf("[认证队列] ⚠️ 设备 %s 已有密码但认证失败,通知前端重新输入", deviceIP)
	}
	
	// 发送事件到前端,请求用户输入密码
	a.emitEvent("device:auth:required", map[string]interface{}{
		"deviceIP": deviceIP,
		"message":  "设备需要认证密码",
	})
	
	log.Printf("[认证队列] 📮 已发送认证请求事件: device:auth:required, deviceIP=%s", deviceIP)
}

// OnDevicePasswordUpdated 当设备密码更新后触发(前端调用UpdateDevicePasswords后调用此方法)
// 检查等待认证的设备,通过认证后立即执行API查询
func (a *App) OnDevicePasswordUpdated(deviceIP string) {
	log.Printf("[认证队列] 🔑 设备 %s 密码已更新,检查认证状态", deviceIP)
	
	// 检查设备当前状态
	a.deviceStatusMutex.RLock()
	status := a.deviceStatusMap[deviceIP]
	isOnline := status != nil && status.Status == "online"
	a.deviceStatusMutex.RUnlock()
	
	if !isOnline {
		log.Printf("[认证队列] ⚠️ 设备 %s 不在线,跳过认证检查", deviceIP)
		return
	}
	
	// 检查认证是否通过
	if a.checkDeviceAuth(deviceIP) {
		log.Printf("[认证队列] ✅ 设备 %s 认证通过,立即执行API查询", deviceIP)
		
		// 认证通过,立即执行API检查和存储查询
		go func() {
			a.checkDeviceAPIVersion(deviceIP)
			
			// 重置存储查询时间,强制立即查询
			a.deviceStatusMutex.Lock()
			if status := a.deviceStatusMap[deviceIP]; status != nil {
				status.LastStorageCheckTime = time.Time{}
			}
			a.deviceStatusMutex.Unlock()
			
			a.checkDeviceStorage(deviceIP)
		}()
	} else {
		log.Printf("[认证队列] ❌ 设备 %s 认证失败,密码可能不正确", deviceIP)
		// 继续等待用户输入正确密码
	}
}
