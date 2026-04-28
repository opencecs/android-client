package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ========== 安卓容器列表后台轮询服务 ==========

const (
	androidPollInterval  = 5 * time.Second // 主循环扫描间隔
	androidQueryTimeout  = 8 * time.Second // 单台设备查询超时
	androidMaxConcurrent = 50             // 最大并发查询数
	androidMaxFailCount  = 4               // 连续失败超过此值认为设备离线
)

// AndroidDeviceCache 单台设备的安卓容器缓存
type AndroidDeviceCache struct {
	List        interface{} // 完整的 /android 接口响应（原始 JSON 解析结果）
	Version     int64       // 最后更新时间戳（UnixMilli），用于前端增量比对
	LastAttempt time.Time   // 最后一次查询时间
	Error       string      // 最后一次错误信息（空=正常）
	FailCount   int         // 连续失败次数
	Status      string      // "ok" | "auth_fail" | "error" | "offline"
}

// ScreenshotEntry 截图缓存条目
type ScreenshotEntry struct {
	Data    string    // base64 DataURL，如 "data:image/jpeg;base64,..."
	Version int64     // UnixMilli 时间戳
	Updated time.Time
}

const (
	screenshotPollInterval  = 1000 * time.Millisecond // 后端截图轮询间隔
	screenshotHTTPTimeout   = 2 * time.Second        // 单次截图 HTTP 超时
	screenshotMaxConcurrent = 20                     // 最大并发抓图 goroutine 数
)

// StartAndroidPoll 启动安卓容器列表后台轮询（幂等，重复调用无副作用）
func (a *App) StartAndroidPoll() {
	a.androidPollMutex.Lock()
	if a.androidPollRunning {
		a.androidPollMutex.Unlock()
		return
	}
	a.androidPollRunning = true
	a.androidPollMutex.Unlock()

	go func() {
		ticker := time.NewTicker(androidPollInterval)
		defer ticker.Stop()
		for range ticker.C {
			a.dispatchAndroidPoll()
		}
	}()
}

// dispatchAndroidPoll 对当前所有在线设备并发轮询一次
func (a *App) dispatchAndroidPoll() {
	// 获取当前需要监控的在线设备 IP 列表
	a.deviceIPsMutex.RLock()
	ips := make([]string, len(a.deviceIPs))
	copy(ips, a.deviceIPs)
	a.deviceIPsMutex.RUnlock()

	if len(ips) == 0 {
		return
	}

	// 只轮询在线设备
	onlineIPs := make([]string, 0, len(ips))
	a.deviceStatusMutex.RLock()
	for _, ip := range ips {
		if s, ok := a.deviceStatusMap[ip]; ok && s != nil && s.Status == "online" {
			onlineIPs = append(onlineIPs, ip)
		}
	}
	a.deviceStatusMutex.RUnlock()

	if len(onlineIPs) == 0 {
		return
	}

	// 信号量控制并发数
	sem := make(chan struct{}, androidMaxConcurrent)
	var wg sync.WaitGroup
	for _, ip := range onlineIPs {
		wg.Add(1)
		sem <- struct{}{}
		go func(deviceIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			a.pollSingleDevice(deviceIP)
		}(ip)
	}
	wg.Wait()
}

// pollSingleDevice 查询单台设备的 /android 接口并写入缓存
func (a *App) pollSingleDevice(ip string) {
	password := a.getDevicePasswordInternal(ip)

	url := fmt.Sprintf("http://%s/android", deviceAddr(ip))
	ctx, cancel := context.WithTimeout(context.Background(), androidQueryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		a.updateAndroidCacheError(ip, "build_request_error", err.Error(), false)
		return
	}
	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.updateAndroidCacheError(ip, "network_error", err.Error(), true)
		return
	}
	defer resp.Body.Close()

	// 认证失败
	if resp.StatusCode == http.StatusUnauthorized {
		a.updateAndroidCacheError(ip, "auth_fail", "Authentication Failed", false)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.updateAndroidCacheError(ip, "read_error", err.Error(), true)
		return
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		a.updateAndroidCacheError(ip, "parse_error", err.Error(), true)
		return
	}

	// 成功：写入缓存
	a.androidCacheMutex.Lock()
	a.androidCache[ip] = &AndroidDeviceCache{
		List:        parsed,
		Version:     time.Now().UnixMilli(),
		LastAttempt: time.Now(),
		Error:       "",
		FailCount:   0,
		Status:      "ok",
	}
	a.androidCacheMutex.Unlock()
}

// updateAndroidCacheError 更新缓存的错误状态
// countFail=true 时累加失败次数；auth_fail 不累加
func (a *App) updateAndroidCacheError(ip, status, errMsg string, countFail bool) {
	a.androidCacheMutex.Lock()
	defer a.androidCacheMutex.Unlock()

	existing := a.androidCache[ip]
	failCount := 0
	if existing != nil {
		failCount = existing.FailCount
	}
	if countFail {
		failCount++
	}

	finalStatus := status
	if countFail && failCount >= androidMaxFailCount {
		finalStatus = "offline"
	}

	if existing == nil {
		a.androidCache[ip] = &AndroidDeviceCache{}
		existing = a.androidCache[ip]
	}
	existing.LastAttempt = time.Now()
	existing.Error = errMsg
	existing.FailCount = failCount
	existing.Status = finalStatus
	// 认证失败时清空缓存列表，防止前端读到过期数据
	if status == "auth_fail" {
		existing.List = nil
		existing.Version = time.Now().UnixMilli()
	}
}

// getDevicePasswordInternal 内部获取设备密码（不加锁外部调用）
func (a *App) getDevicePasswordInternal(ip string) string {
	a.devicePasswordsMutex.RLock()
	defer a.devicePasswordsMutex.RUnlock()
	if a.devicePasswords == nil {
		return ""
	}
	return a.devicePasswords[ip]
}

// ========== IPC 接口（供前端调用）==========

// GetAndroidCacheVersions 返回所有设备的缓存版本号（极轻量，2秒轮询用）
// 返回: map[ip]versionUnixMilli
func (a *App) GetAndroidCacheVersions() map[string]int64 {
	a.androidCacheMutex.RLock()
	defer a.androidCacheMutex.RUnlock()

	result := make(map[string]int64, len(a.androidCache))
	for ip, cache := range a.androidCache {
		result[ip] = cache.Version
	}
	return result
}

// GetAndroidContainersList 返回指定设备的完整安卓容器缓存
// ips 为空时返回所有设备
func (a *App) GetAndroidContainersList(ips []string) map[string]interface{} {
	a.androidCacheMutex.RLock()
	defer a.androidCacheMutex.RUnlock()

	result := make(map[string]interface{})

	if len(ips) == 0 {
		// 返回所有
		for ip, cache := range a.androidCache {
			result[ip] = map[string]interface{}{
				"list":    cache.List,
				"version": cache.Version,
				"status":  cache.Status,
				"error":   cache.Error,
			}
		}
		return result
	}

	// 返回指定设备
	for _, ip := range ips {
		if cache, ok := a.androidCache[ip]; ok {
			result[ip] = map[string]interface{}{
				"list":    cache.List,
				"version": cache.Version,
				"status":  cache.Status,
				"error":   cache.Error,
			}
		} else {
			result[ip] = nil
		}
	}
	return result
}

// TriggerAndroidRefresh 立即触发指定设备重新查询（操作后调用）
// ips 为空时触发所有在线设备
func (a *App) TriggerAndroidRefresh(ips []string) {
	go func() {
		if len(ips) == 0 {
			a.dispatchAndroidPoll()
			return
		}
		sem := make(chan struct{}, androidMaxConcurrent)
		var wg sync.WaitGroup
		for _, ip := range ips {
			wg.Add(1)
			sem <- struct{}{}
			go func(deviceIP string) {
				defer wg.Done()
				defer func() { <-sem }()
				a.pollSingleDevice(deviceIP)
			}(ip)
		}
		wg.Wait()
	}()
}

// StartScreenshotPoll 启动截图后台轮询（幂等，重复调用无副作用）
func (a *App) StartScreenshotPoll() {
	a.screenshotPollMutex.Lock()
	if a.screenshotPollRunning {
		a.screenshotPollMutex.Unlock()
		return
	}
	a.screenshotPollRunning = true
	a.screenshotPollMutex.Unlock()

	go func() {
		ticker := time.NewTicker(screenshotPollInterval)
		defer ticker.Stop()
		for range ticker.C {
			// 非阻塞：每轮 tick 独立 dispatch，不等上一轮全部完成
			// 避免慢容器拖慢整体刷新节奏
			go a.dispatchScreenshotPoll()
		}
	}()
}

// dispatchScreenshotPoll 对所有在线设备的运行中容器并发抓取截图
func (a *App) dispatchScreenshotPoll() {
	// 收集所有在线设备及其容器快照
	type containerTask struct {
		deviceIP      string
		containerName string
		screenshotURL string
	}

	a.androidCacheMutex.RLock()
	var tasks []containerTask
	for ip, cache := range a.androidCache {
		if cache == nil || cache.Status != "ok" || cache.List == nil {
			continue
		}
		// cache.List 是 /android 接口的原始 JSON 解析结果
		// 支持多种结构：
		// 1. []interface{} — 直接是容器数组
		// 2. map{code, data:{list:[...]}} — V3 标准响应
		// 3. map{list:[...]} — 简化响应
		var containers []interface{}
		switch v := cache.List.(type) {
		case []interface{}:
			containers = v
		case map[string]interface{}:
			// 优先找 data.list（V3 标准：{code, data:{list:[...]}}）
			if data, ok := v["data"].(map[string]interface{}); ok {
				if lst, ok := data["list"].([]interface{}); ok {
					containers = lst
				}
			}
			// 次找顶层 list
			if containers == nil {
				if lst, ok := v["list"].([]interface{}); ok {
					containers = lst
				}
			}
		}
		for _, c := range containers {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			// 只处理运行中的容器
			status, _ := cm["status"].(string)
			if status != "running" {
				continue
			}
			name, _ := cm["name"].(string)
			if name == "" {
				continue
			}
			url := a.buildScreenshotURL(ip, cm)
			if url == "" {
				continue
			}
			tasks = append(tasks, containerTask{
				deviceIP:      ip,
				containerName: name,
				screenshotURL: url,
			})
		}
	}
	a.androidCacheMutex.RUnlock()

	if len(tasks) == 0 {
		return
	}

	sem := make(chan struct{}, screenshotMaxConcurrent)
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(t containerTask) {
			defer wg.Done()
			defer func() { <-sem }()
			a.fetchAndCacheScreenshot(t.deviceIP, t.containerName, t.screenshotURL)
		}(task)
	}
	wg.Wait()
}

// buildScreenshotURL 根据容器信息构造截图 URL（与前端逻辑保持一致）
func (a *App) buildScreenshotURL(deviceIP string, cm map[string]interface{}) string {
	networkName, _ := cm["networkName"].(string)

	if networkName == "myt" {
		// myt 网络：直接使用容器 IP + 9082
		containerIP, _ := cm["ip"].(string)
		if containerIP == "" {
			return ""
		}
		return fmt.Sprintf("http://%s:9082/task=snap&level=1", containerIP)
	}

	// 非 myt 网络：从端口映射中找 9082 的宿主机端口
	mappedPort := extractPort9082FromContainer(cm)
	if mappedPort == 0 {
		// fallback：indexNum * 1 + 10000
		if idx, ok := toInt(cm["indexNum"]); ok && idx > 0 {
			mappedPort = 10000 + idx
		}
	}
	if mappedPort == 0 {
		return ""
	}

	// OpenCecs 公网设备：deviceIP 包含 ":" 时提取纯 IP，并将 HostPort 转换为公网端口
	if strings.Contains(deviceIP, ":") {
		pureIP := deviceIP[:strings.Index(deviceIP, ":")]
		// 查找端口映射（先精确匹配，再按 IP 前缀模糊匹配）
		a.cecsPortMapMutex.RLock()
		portMap := a.cecsPortMap[deviceIP]
		if portMap == nil {
			for key, pm := range a.cecsPortMap {
				if strings.HasPrefix(key, pureIP+":") {
					portMap = pm
					break
				}
			}
		}
		a.cecsPortMapMutex.RUnlock()
		if portMap != nil {
			if pub, ok := portMap[mappedPort]; ok {
				mappedPort = pub
			}
		}
		return fmt.Sprintf("http://%s:%d/task=snap&level=1", pureIP, mappedPort)
	}

	return fmt.Sprintf("http://%s:%d/task=snap&level=1", deviceIP, mappedPort)
}

// extractPort9082FromContainer 从容器的端口映射中提取 9082 的宿主机端口
func extractPort9082FromContainer(cm map[string]interface{}) int {
	const portKey = "9082/tcp"

	// V3 格式: portBindings["9082/tcp"][0].HostPort
	if pb, ok := cm["portBindings"].(map[string]interface{}); ok {
		if bindings, ok := pb[portKey].([]interface{}); ok && len(bindings) > 0 {
			if b, ok := bindings[0].(map[string]interface{}); ok {
				if p, ok := toInt(b["HostPort"]); ok {
					return p
				}
			}
		}
	}

	// Docker 原生格式: Ports 数组
	if ports, ok := cm["Ports"].([]interface{}); ok {
		for _, p := range ports {
			pm, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			privatePort, _ := toInt(pm["PrivatePort"])
			typ, _ := pm["Type"].(string)
			if privatePort == 9082 && typ == "tcp" {
				if pub, ok := toInt(pm["PublicPort"]); ok {
					return pub
				}
			}
		}
	}

	// 兼容旧格式: NetworkSettings.Ports
	if ns, ok := cm["NetworkSettings"].(map[string]interface{}); ok {
		if portMap, ok := ns["Ports"].(map[string]interface{}); ok {
			if bindings, ok := portMap[portKey].([]interface{}); ok && len(bindings) > 0 {
				if b, ok := bindings[0].(map[string]interface{}); ok {
					if p, ok := toInt(b["HostPort"]); ok {
						return p
					}
				}
			}
		}
	}

	// 兼容 PortBindings（大写）
	if pb, ok := cm["PortBindings"].(map[string]interface{}); ok {
		if bindings, ok := pb[portKey].([]interface{}); ok && len(bindings) > 0 {
			if b, ok := bindings[0].(map[string]interface{}); ok {
				if p, ok := toInt(b["HostPort"]); ok {
					return p
				}
			}
		}
	}

	return 0
}

// toInt 将 interface{} 转为 int，支持多种底层类型
func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		n := 0
		_, err := fmt.Sscanf(val, "%d", &n)
		return n, err == nil
	}
	return 0, false
}

// fetchAndCacheScreenshot 对单个容器发起截图 HTTP 请求并写入缓存
func (a *App) fetchAndCacheScreenshot(deviceIP, containerName, screenshotURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), screenshotHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", screenshotURL, nil)
	if err != nil {
		return
	}

	resp, err := a.screenshotHTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return
	}

	// 判断 Content-Type，构造 DataURL
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	// 只取 mime 部分（去掉参数）
	if idx := len(contentType); idx > 0 {
		for i, c := range contentType {
			if c == ';' || c == ' ' {
				contentType = contentType[:i]
				break
			}
		}
	}

	dataURL := "data:" + contentType + ";base64," + b64Encode(body)

	cacheKey := deviceIP + "_" + containerName
	now := time.Now()

	a.screenshotCacheMutex.Lock()
	a.screenshotCache[cacheKey] = &ScreenshotEntry{
		Data:    dataURL,
		Version: now.UnixMilli(),
		Updated: now,
	}
	// 更新该设备的版本号（取所有容器中最新的时间戳）
	if cur, ok := a.screenshotVersions[deviceIP]; !ok || now.UnixMilli() > cur {
		a.screenshotVersions[deviceIP] = now.UnixMilli()
	}
	a.screenshotCacheMutex.Unlock()
}

// ========== 截图缓存 IPC 接口（供前端调用）==========

// GetScreenshotVersions 返回所有设备的截图版本号（极轻量，轮询用）
// 返回: map[ip]versionUnixMilli
func (a *App) GetScreenshotVersions() map[string]int64 {
	a.screenshotCacheMutex.RLock()
	defer a.screenshotCacheMutex.RUnlock()

	result := make(map[string]int64, len(a.screenshotVersions))
	for ip, v := range a.screenshotVersions {
		result[ip] = v
	}
	return result
}

// GetScreenshots 返回指定设备下所有容器的截图数据
// ips 为空时返回所有设备；返回 map["ip_containerName"]dataURL
func (a *App) GetScreenshots(ips []string) map[string]string {
	a.screenshotCacheMutex.RLock()
	defer a.screenshotCacheMutex.RUnlock()

	result := make(map[string]string)

	if len(ips) == 0 {
		// 返回全部
		for key, entry := range a.screenshotCache {
			if entry != nil && entry.Data != "" {
				result[key] = entry.Data
			}
		}
		return result
	}

	// 只返回指定设备的容器截图
	ipSet := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		ipSet[ip] = struct{}{}
	}
	for key, entry := range a.screenshotCache {
		if entry == nil || entry.Data == "" {
			continue
		}
		// key 格式为 "ip_containerName"
		for ip := range ipSet {
			if len(key) > len(ip)+1 && key[:len(ip)] == ip && key[len(ip)] == '_' {
				result[key] = entry.Data
				break
			}
		}
	}
	return result
}

// ClearScreenshotCache 清除指定设备的截图缓存
// ips 为空时清除所有设备缓存
func (a *App) ClearScreenshotCache(ips []string) {
	a.screenshotCacheMutex.Lock()
	defer a.screenshotCacheMutex.Unlock()

	if len(ips) == 0 {
		a.screenshotCache = make(map[string]*ScreenshotEntry)
		a.screenshotVersions = make(map[string]int64)
		return
	}
	for _, ip := range ips {
		// screenshotCache key 格式为 "ip_containerName"，按前缀批量删除
		prefix := ip + "_"
		for key := range a.screenshotCache {
			if len(key) > len(prefix) && key[:len(prefix)] == prefix {
				delete(a.screenshotCache, key)
			}
		}
		delete(a.screenshotVersions, ip)
	}
}

// b64Encode 将字节切片编码为 base64 字符串
func b64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// SetOpenCecsPortMap 前端调用：推送 OpenCecs 端口映射数据到 Go 后端
// 参数格式：map[deviceIp]map[privatePortStr]publicPort
func (a *App) SetOpenCecsPortMap(data map[string]map[string]int) {
	a.cecsPortMapMutex.Lock()
	defer a.cecsPortMapMutex.Unlock()

	a.cecsPortMap = make(map[string]map[int]int, len(data))
	for deviceIp, ports := range data {
		pm := make(map[int]int, len(ports))
		for privStr, pub := range ports {
			if priv, ok := toInt(privStr); ok {
				pm[priv] = pub
			}
		}
		a.cecsPortMap[deviceIp] = pm
	}
}

// resolveOpenCecsAddr 解析 OpenCecs 公网设备的正确 host 和 port。
// 当 deviceIP 含端口（如 "219.139.239.165:12831"）时，提取纯 IP 并从 cecsPortMap
// 查找 privatePort 对应的 publicPort。局域网设备原样返回 (deviceIP, privatePort)。
func (a *App) resolveOpenCecsAddr(deviceIP string, privatePort int) (string, int) {
	host, _, err := net.SplitHostPort(deviceIP)
	if err != nil {
		// 不含端口，局域网设备
		return deviceIP, privatePort
	}
	// 含端口 → OpenCecs 公网设备，查表
	a.cecsPortMapMutex.RLock()
	defer a.cecsPortMapMutex.RUnlock()

	if pm, ok := a.cecsPortMap[deviceIP]; ok {
		if pub, ok := pm[privatePort]; ok {
			return host, pub
		}
	}
	// 模糊匹配：按 IP 前缀查找
	prefix := host + ":"
	for key, pm := range a.cecsPortMap {
		if strings.HasPrefix(key, prefix) {
			if pub, ok := pm[privatePort]; ok {
				return host, pub
			}
		}
	}
	// 未找到映射，返回纯 IP + 原始端口
	return host, privatePort
}

// ========== RPA Agent 专用 API ==========

// extractPort9083FromContainer 从容器的端口映射中提取 9083 的宿主机端口
func extractPort9083FromContainer(cm map[string]interface{}) int {
	const portKey = "9083/tcp"

	// V3 格式: portBindings["9083/tcp"][0].HostPort
	if pb, ok := cm["portBindings"].(map[string]interface{}); ok {
		if bindings, ok := pb[portKey].([]interface{}); ok && len(bindings) > 0 {
			if b, ok := bindings[0].(map[string]interface{}); ok {
				if p, ok := toInt(b["HostPort"]); ok {
					return p
				}
			}
		}
	}

	// Docker 原生格式: Ports 数组
	if ports, ok := cm["Ports"].([]interface{}); ok {
		for _, p := range ports {
			pm, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			privatePort, _ := toInt(pm["PrivatePort"])
			typ, _ := pm["Type"].(string)
			if privatePort == 9083 && typ == "tcp" {
				if pub, ok := toInt(pm["PublicPort"]); ok {
					return pub
				}
			}
		}
	}

	// 兼容旧格式: NetworkSettings.Ports
	if ns, ok := cm["NetworkSettings"].(map[string]interface{}); ok {
		if portMap, ok := ns["Ports"].(map[string]interface{}); ok {
			if bindings, ok := portMap[portKey].([]interface{}); ok && len(bindings) > 0 {
				if b, ok := bindings[0].(map[string]interface{}); ok {
					if p, ok := toInt(b["HostPort"]); ok {
						return p
					}
				}
			}
		}
	}

	// 兼容 PortBindings（大写）
	if pb, ok := cm["PortBindings"].(map[string]interface{}); ok {
		if bindings, ok := pb[portKey].([]interface{}); ok && len(bindings) > 0 {
			if b, ok := bindings[0].(map[string]interface{}); ok {
				if p, ok := toInt(b["HostPort"]); ok {
					return p
				}
			}
		}
	}

	return 0
}

// calcRpaPortFromContainer 根据容器信息计算 RPA 连接地址
// 返回 (rpaIp, rpaPort)：
//   - myt 桥接模式：直接用容器 ip:9083
//   - bridge 非桥接模式：优先按坑位计算 30000+(indexNum-1)*100+2；portBindings["9083/tcp"] 有值时覆盖
func calcRpaPortFromContainer(cm map[string]interface{}, deviceIP string) (string, int) {
	networkName, _ := cm["networkName"].(string)

	if networkName == "myt" {
		// 桥接模式：直接用容器内网 IP:9083
		containerIP, _ := cm["ip"].(string)
		if containerIP == "" {
			return deviceIP, 9083
		}
		return containerIP, 9083
	}

	// 非桥接模式：优先按坑位计算
	rpaPort := 0
	if idx, ok := toInt(cm["indexNum"]); ok && idx > 0 {
		rpaPort = 30000 + (idx-1)*100 + 2
	}

	// portBindings["9083/tcp"] 有值时覆盖坑位计算结果
	if mappedPort := extractPort9083FromContainer(cm); mappedPort > 0 {
		rpaPort = mappedPort
	}

	if rpaPort == 0 {
		rpaPort = 9083 // 最终兜底
	}

	return deviceIP, rpaPort
}

// extractContainersFromCache 从缓存的原始响应中提取容器数组（复用 dispatchScreenshotPoll 的解析逻辑）
func extractContainersFromCache(list interface{}) []interface{} {
	switch v := list.(type) {
	case []interface{}:
		return v
	case map[string]interface{}:
		// V3 标准响应: {code, data:{list:[...]}}
		if data, ok := v["data"].(map[string]interface{}); ok {
			if lst, ok := data["list"].([]interface{}); ok {
				return lst
			}
		}
		// 简化响应: {list:[...]}
		if lst, ok := v["list"].([]interface{}); ok {
			return lst
		}
	}
	return nil
}

// RpaGetAndroidContainers IPC 方法，供前端 RPA Agent 调用
// 读取 androidCache，提取所有 running 容器，计算 RPA 端口，返回结构化列表
// 返回格式: map[deviceIP]ContainerInfo[]，其中 ContainerInfo 包含 rpaIp、rpaPort、networkMode 等
func (a *App) RpaGetAndroidContainers(ips []string) map[string]interface{} {
	a.androidCacheMutex.RLock()
	defer a.androidCacheMutex.RUnlock()

	result := make(map[string]interface{})

	// 构建待查询的 IP 集合
	targetIPs := make(map[string]struct{})
	if len(ips) == 0 {
		// 返回所有设备
		for ip := range a.androidCache {
			targetIPs[ip] = struct{}{}
		}
	} else {
		for _, ip := range ips {
			targetIPs[ip] = struct{}{}
		}
	}

	for ip := range targetIPs {
		cache, ok := a.androidCache[ip]
		if !ok || cache == nil {
			result[ip] = map[string]interface{}{
				"status":     "no_cache",
				"containers": []interface{}{},
				"error":      "设备缓存不存在，请确认设备在线",
			}
			continue
		}

		if cache.Status != "ok" || cache.List == nil {
			result[ip] = map[string]interface{}{
				"status":     cache.Status,
				"containers": []interface{}{},
				"error":      cache.Error,
			}
			continue
		}

		allContainers := extractContainersFromCache(cache.List)
		runningList := make([]interface{}, 0)

		for _, c := range allContainers {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			status, _ := cm["status"].(string)
			name, _ := cm["name"].(string)
			if status != "running" || name == "" {
				continue
			}

			networkName, _ := cm["networkName"].(string)
			indexNum := 0
			if idx, ok := toInt(cm["indexNum"]); ok {
				indexNum = idx
			}
			containerIP, _ := cm["ip"].(string)

			rpaIp, rpaPort := calcRpaPortFromContainer(cm, ip)

			containerInfo := map[string]interface{}{
				"name":        name,
				"status":      status,
				"networkMode": networkName,
				"indexNum":    indexNum,
				"containerIP": containerIP,
				"rpaIp":       rpaIp,
				"rpaPort":     rpaPort,
				"deviceIP":    ip,
			}

			// 附带原始信息（调试用）
			if imgVersion, ok := cm["imageVersion"].(string); ok {
				containerInfo["imageVersion"] = imgVersion
			}

			runningList = append(runningList, containerInfo)
		}

		result[ip] = map[string]interface{}{
			"status":     "ok",
			"containers": runningList,
			"total":      len(allContainers),
			"running":    len(runningList),
		}
	}

	fmt.Printf("[RpaGetAndroidContainers] ips=%v result_keys=%d\n", ips, len(result))
	return result
}
