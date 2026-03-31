package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/q191201771/lal/pkg/logic"
)

type RtmpService struct {
	ctx          context.Context
	lalServer    logic.ILalServer
	app          *App
	localIP      string
	savedStreams []string // 保存用户创建的房间号
	useSrs       bool
	srsBinary    string
	srsWorkDir   string
	srsCmd       *exec.Cmd
}

func NewRtmpService() *RtmpService {
	return &RtmpService{
		savedStreams: []string{},
		srsBinary:    "srs",
	}
}

func (r *RtmpService) SetApp(app *App) {
	r.app = app
}

// StartRtmpServer 启动流媒体服务器
func (r *RtmpService) StartRtmpServer() {
	r.localIP = GetLocalIP()
	startedSrs, err := r.tryStartLocalSrsServer()
	if err != nil {
		fmt.Printf("SRS start failed: %v\n", err)
	}
	if startedSrs {
		r.useSrs = true
		fmt.Println("SRS Server started on :1935, WebRTC/HTTP on :1985, RTC UDP/TCP on :8000")
		return
	}

	r.useSrs = false

	// 构造配置 JSON
	// LAL v0.37.4+ 的正确配置格式
	// 关键：default_http 配置启动 HTTP 服务器
	confWithWebRTC := `{
		"rtmp": {
			"enable": true,
			"addr": ":1935",
			"gop_num": 1
		},
		"default_http": {
			"http_listen_addr": ":8083"
		},
		"httpflv": {
			"enable": true,
			"enable_gop_cache": false
		},
		"http_api": {
			"enable": true,
			"addr": ":8084"
		},
		"rtsp": {
			"enable": true,
			"addr": ":554"
		},
		"webrtc": {
			"enable": true
		},
		"hls": {
			"enable": false
		}
	}`

	// 使用 ModOption 传入配置
	modOption := func(option *logic.Option) {
		option.ConfRawContent = []byte(confWithWebRTC)
	}

	l := logic.NewLalServer(modOption)
	r.lalServer = l

	go func() {
		err := l.RunLoop()
		if err != nil {
			fmt.Printf("LAL Server error: %v\n", err)
		}
	}()
	fmt.Println("RTMP Server started on :1935, WebRTC/HTTP on :8083, RTSP on :554, API on :8084")

	startProxy1985()
}

func (r *RtmpService) tryStartLocalSrsServer() (bool, error) {
	if r.isSrsApiAvailable() {
		return true, nil
	}

	bin := r.getSrsBinary()
	path, err := exec.LookPath(bin)
	if err != nil {
		return false, nil
	}

	cmd := exec.Command(path, "-c", "conf/rtc.conf")
	if workDir := r.getSrsWorkDir(); workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return false, err
	}
	r.srsCmd = cmd

	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if r.isSrsApiAvailable() {
			return true, nil
		}
	}
	return false, fmt.Errorf("SRS API not available")
}

func (r *RtmpService) getSrsBinary() string {
	if value := os.Getenv("MYT_SRS_BIN"); value != "" {
		return value
	}
	if value := os.Getenv("SRS_BIN"); value != "" {
		return value
	}
	return r.srsBinary
}

func (r *RtmpService) getSrsWorkDir() string {
	if value := os.Getenv("MYT_SRS_HOME"); value != "" {
		return value
	}
	if value := os.Getenv("SRS_HOME"); value != "" {
		return value
	}
	return r.srsWorkDir
}

func (r *RtmpService) isSrsApiAvailable() bool {
	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:1985/api/v1/versions")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func startProxy1985() {
	targetBase := "http://127.0.0.1:8083"
	client := &http.Client{Timeout: 5 * time.Second}
	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			// log.Printf("[1985->8083] read request body error: %v", readErr)
			_ = readErr
		}
		_ = r.Body.Close()
		// log.Printf("[1985->8083] -> %s %s?%s from %s", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr)
		// if len(requestBody) > 0 {
		// 	log.Printf("[1985->8083] -> body %d %s", len(requestBody), trimLogBody(requestBody, 512))
		// }

		resp, usedPath, err := forwardToLal(client, targetBase, r, requestBody, r.URL.Path)
		if err == nil && resp.StatusCode == http.StatusNotFound {
			altPath := mapRtcPath(r.URL.Path)
			if altPath != "" && altPath != r.URL.Path {
				// log.Printf("[1985->8083] retry %s -> %s", r.URL.Path, altPath)
				_ = resp.Body.Close()
				resp, usedPath, err = forwardToLal(client, targetBase, r, requestBody, altPath)
			}
		}
		if err != nil {
			// log.Printf("[1985->8083] forward error: %v", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		respBody, readRespErr := io.ReadAll(resp.Body)
		if readRespErr != nil {
			// log.Printf("[1985->8083] read response body error: %v", readRespErr)
			_ = readRespErr
		}
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		_ = usedPath
		// log.Printf("[1985->8083] <- %d %s body=%s", resp.StatusCode, usedPath, trimLogBody(respBody, 512))
	})

	// log.Printf("[1985->8083] listen :1985")
		_ = http.ListenAndServe(":1985", handler)
	}()
}

func forwardToLal(client *http.Client, targetBase string, r *http.Request, body []byte, path string) (*http.Response, string, error) {
	targetURL := targetBase + path
	if r.URL.RawQuery != "" {
		targetURL = targetURL + "?" + r.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, targetURL, err
	}
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Host = strings.TrimPrefix(targetBase, "http://")
	resp, err := client.Do(req)
	if err != nil {
		return nil, targetURL, err
	}
	return resp, targetURL, nil
}

func mapRtcPath(path string) string {
	// LAL WebRTC API 使用不同的路径格式
	// 客户端发送 JSON: {"api":"...", "sdp":"..."}
	if strings.HasPrefix(path, "/rtc/v1/whep") || strings.HasPrefix(path, "/rtc/v1/play") {
		// LAL 使用 /webrtc/v1/play (注意有 v1)
		return strings.Replace(path, "/rtc/v1/play", "/webrtc/v1/play", 1)
	}
	if strings.HasPrefix(path, "/rtc/v1/whip") {
		return strings.Replace(path, "/rtc/v1/whip", "/webrtc/v1/publish", 1)
	}
	return ""
}

func trimLogBody(body []byte, limit int) string {
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "...(truncated)"
}

// GetLocalIP 获取本机局域网IP
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: 遍历网卡
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return ""
		}
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// StreamInfo 流信息结构
type StreamInfo struct {
	StreamName   string `json:"streamName"`
	AppName      string `json:"appName"`
	PublisherIP  string `json:"publisherIP"`
	StartTime    string `json:"startTime"`
	VideoBitrate int    `json:"videoBitrate"`
	AudioBitrate int    `json:"audioBitrate"`
}

// LalStatResponse LAL 统计接口响应
type LalStatResponse struct {
	ErrorCode int    `json:"error_code"`
	Desp      string `json:"desp"`
	Data      struct {
		Groups []struct {
			AppName    string `json:"app_name"`
			StreamName string `json:"stream_name"`
			PubSession *struct {
				SessionId     string `json:"session_id"`
				Protocol      string `json:"protocol"`
				BaseType      string `json:"base_type"`
				RemoteAddr    string `json:"remote_addr"`
				StartTime     string `json:"start_time"`
				ReadBytesSum  int    `json:"read_bytes_sum"`
				WroteBytesSum int    `json:"wrote_bytes_sum"`
				BitrateKbits  int    `json:"bitrate_kbits"`
			} `json:"pub"`
			SubSessions []struct {
				StreamName string `json:"stream_name"`
				RemoteAddr string `json:"remote_addr"`
				Type       string `json:"type"` // "RTMP", "FLV", "TS", "HLS", "WebRTC"
			} `json:"subs"`
		} `json:"groups"`
	} `json:"data"`
}

// AddStreamName 添加房间号
func (r *RtmpService) AddStreamName(name string) {
	for _, s := range r.savedStreams {
		if s == name {
			return
		}
	}
	r.savedStreams = append(r.savedStreams, name)
}

// DeleteStreamName 删除房间号
func (r *RtmpService) DeleteStreamName(name string) {
	var newStreams []string
	for _, s := range r.savedStreams {
		if s != name {
			newStreams = append(newStreams, s)
		}
	}
	r.savedStreams = newStreams
}

// GetSavedStreams 获取保存的房间号列表 (包含状态)
func (r *RtmpService) GetSavedStreams() []StreamInfo {
	// 获取当前活跃流
	active := r.GetActiveStreams()
	activeMap := make(map[string]StreamInfo)
	for _, s := range active {
		activeMap[s.StreamName] = s
	}

	var result []StreamInfo
	for _, name := range r.savedStreams {
		if info, ok := activeMap[name]; ok {
			result = append(result, info)
		} else {
			// 如果不活跃，返回仅包含名称的空状态
			result = append(result, StreamInfo{
				StreamName:  name,
				AppName:     "live", // 默认 live
				PublisherIP: "",     // 空表示不活跃
			})
		}
	}
	// 同时，如果有活跃流不在保存列表中，也应该显示出来 (可选，或者自动添加到保存列表)
	// 这里为了简单，先只显示保存列表中的。
	// 但如果用户推了一个不在列表里的流，应该也能看到。
	for _, s := range active {
		found := false
		for _, name := range r.savedStreams {
			if name == s.StreamName {
				found = true
				break
			}
		}
		if !found {
			result = append(result, s)
		}
	}

	return result
}

// GetActiveStreams 获取当前活跃流列表 (原始方法保持不变，供内部使用)
func (r *RtmpService) GetActiveStreams() []StreamInfo {
	if r.useSrs {
		return r.getActiveStreamsFromSrs()
	}
	
	// log.Println("[推流] ⏰ 查询活跃推流列表...")
	resp, err := http.Get("http://127.0.0.1:8084/api/stat/all_group")
	if err != nil {
		// log.Printf("[推流] ❌ 查询失败: %v\n", err)
		return []StreamInfo{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("[推流] ❌ 读取响应失败: %v\n", err)
		return []StreamInfo{}
	}

	var stat LalStatResponse
	if err := json.Unmarshal(body, &stat); err != nil {
		// log.Printf("[推流] ❌ 解析响应失败: %v\n", err)
		return []StreamInfo{}
	}

	var streams []StreamInfo
	for _, group := range stat.Data.Groups {
		if group.PubSession != nil {
			streams = append(streams, StreamInfo{
				StreamName:   group.StreamName,
				AppName:      group.AppName,
				PublisherIP:  strings.Split(group.PubSession.RemoteAddr, ":")[0],
				StartTime:    group.PubSession.StartTime,
				VideoBitrate: group.PubSession.BitrateKbits * 1024, // 转换为 bps 以匹配前端显示
				AudioBitrate: 0,                                    // API 中没有区分音视频码率，统一用总码率
			})
		}
	}
	
	// if len(streams) == 0 {
	// 	log.Println("[推流] ✓ 查询完成: 当前无活跃推流")
	// } else {
	// 	log.Printf("[推流] ✓ 查询完成: 发现 %d 个活跃推流\n", len(streams))
	// 	for _, s := range streams {
	// 		log.Printf("[推流]   ├─ 房间号: %s, 来源IP: %s, 码率: %d Kbps",
	// 			s.StreamName, s.PublisherIP, s.VideoBitrate/1024)
	// 	}
	// }
	
	return streams
}

func (r *RtmpService) getActiveStreamsFromSrs() []StreamInfo {
	// log.Println("[推流-SRS] ⏰ 查询活跃推流列表...")
	resp, err := http.Get("http://127.0.0.1:1985/api/v1/streams")
	if err != nil {
		// log.Printf("[推流-SRS] ❌ 查询失败: %v\n", err)
		return []StreamInfo{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("[推流-SRS] ❌ 读取响应失败: %v\n", err)
		return []StreamInfo{}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// log.Printf("[推流-SRS] ❌ 解析响应失败: %v\n", err)
		return []StreamInfo{}
	}

	streamsRaw, ok := payload["streams"].([]interface{})
	if !ok {
		// log.Println("[推流-SRS] ✓ 查询完成: 当前无活跃推流")
		return []StreamInfo{}
	}

	var streams []StreamInfo
	for _, item := range streamsRaw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		streamName := readString(m, "stream")
		if streamName == "" {
			streamName = readString(m, "name")
		}
		appName := readString(m, "app")
		publisherIP := ""
		startTime := ""
		videoBitrate := 0
		audioBitrate := 0

		if publish := readMap(m, "publish"); publish != nil {
			publisherIP = readString(publish, "ip")
			startTime = readString(publish, "start_time")
		}
		if kbps := readMap(m, "kbps"); kbps != nil {
			videoBitrate = int(readFloat(kbps, "video") * 1024)
			audioBitrate = int(readFloat(kbps, "audio") * 1024)
		}

		if streamName != "" {
			streams = append(streams, StreamInfo{
				StreamName:   streamName,
				AppName:      appName,
				PublisherIP:  publisherIP,
				StartTime:    startTime,
				VideoBitrate: videoBitrate,
				AudioBitrate: audioBitrate,
			})
		}
	}

	// if len(streams) == 0 {
	// 	log.Println("[推流-SRS] ✓ 查询完成: 当前无活跃推流")
	// } else {
	// 	log.Printf("[推流-SRS] ✓ 查询完成: 发现 %d 个活跃推流\n", len(streams))
	// 	for _, s := range streams {
	// 		log.Printf("[推流-SRS]   ├─ 房间号: %s, 来源IP: %s, 视频: %d Kbps, 音频: %d Kbps",
	// 			s.StreamName, s.PublisherIP, s.VideoBitrate/1024, s.AudioBitrate/1024)
	// 	}
	// }

	return streams
}

// GetStreamPullers 获取指定流的拉流客户端信息
type PullerInfo struct {
	DeviceIP   string `json:"deviceIP"`
	Type       string `json:"type"`
	StreamName string `json:"streamName"`
}

func (r *RtmpService) GetStreamPullers(streamName string) []PullerInfo {
	if r.useSrs {
		return r.getStreamPullersFromSrs(streamName)
	}
	resp, err := http.Get("http://127.0.0.1:8084/api/stat/all_group")
	if err != nil {
		return []PullerInfo{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []PullerInfo{}
	}

	var stat LalStatResponse
	if err := json.Unmarshal(body, &stat); err != nil {
		return []PullerInfo{}
	}

	var pullers []PullerInfo
	for _, group := range stat.Data.Groups {
		if group.StreamName == streamName {
			for _, sub := range group.SubSessions {
				ip := strings.Split(sub.RemoteAddr, ":")[0]
				pullers = append(pullers, PullerInfo{
					DeviceIP:   ip,
					Type:       sub.Type,
					StreamName: sub.StreamName,
				})
			}
		}
	}
	return pullers
}

func (r *RtmpService) getStreamPullersFromSrs(streamName string) []PullerInfo {
	resp, err := http.Get("http://127.0.0.1:1985/api/v1/streams")
	if err != nil {
		return []PullerInfo{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []PullerInfo{}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return []PullerInfo{}
	}

	streamsRaw, ok := payload["streams"].([]interface{})
	if !ok {
		return []PullerInfo{}
	}

	streamID := ""
	for _, item := range streamsRaw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := readString(m, "stream")
		if name == "" {
			name = readString(m, "name")
		}
		if name == streamName {
			streamID = readString(m, "id")
			if streamID == "" {
				streamID = readString(m, "stream_id")
			}
			break
		}
	}

	if streamID == "" {
		return []PullerInfo{}
	}

	clientsResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:1985/api/v1/streams/%s/clients", streamID))
	if err != nil {
		return []PullerInfo{}
	}
	defer clientsResp.Body.Close()

	clientsBody, err := io.ReadAll(clientsResp.Body)
	if err != nil {
		return []PullerInfo{}
	}

	var clientsPayload map[string]interface{}
	if err := json.Unmarshal(clientsBody, &clientsPayload); err != nil {
		return []PullerInfo{}
	}

	clientsRaw, ok := clientsPayload["clients"].([]interface{})
	if !ok {
		return []PullerInfo{}
	}

	var pullers []PullerInfo
	for _, item := range clientsRaw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		clientType := readString(m, "type")
		if clientType == "publish" {
			continue
		}
		ip := readString(m, "ip")
		if ip == "" {
			ip = readString(m, "client_ip")
		}
		pullers = append(pullers, PullerInfo{
			DeviceIP:   ip,
			Type:       clientType,
			StreamName: streamName,
		})
	}

	return pullers
}

func readString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func readFloat(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func readMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	value, ok := m[key]
	if !ok || value == nil {
		return nil
	}
	if child, ok := value.(map[string]interface{}); ok {
		return child
	}
	return nil
}

// PushStreamToDevices 推送流到设备 (WebRTC)
func (r *RtmpService) PushStreamToDevices(streamName string, deviceIPs []string) map[string]interface{} {
	localIP := GetLocalIP()
	if localIP == "" {
		return map[string]interface{}{"success": false, "message": "无法获取本机IP"}
	}

	httpflvURL := fmt.Sprintf("http://%s:8083/live/%s.flv", localIP, streamName)

	successCount := 0

	for _, ip := range deviceIPs {
		go func(deviceIP string) {
			// 1. 设置源
			// GET /modifydev?cmd=4&type=webrtc&path={url}&resolution=1
			setUrl := fmt.Sprintf("http://%s/modifydev?cmd=4&type=webrtc&path=%s&resolution=1", deviceAddr(deviceIP), httpflvURL)

			r.app.HttpRequest(HTTPRequestParams{
				URL:    setUrl,
				Method: "GET",
			})

			// 2. 热启动
			// GET /camera?cmd=start
			startUrl := fmt.Sprintf("http://%s/camera?cmd=start", deviceAddr(deviceIP))
			r.app.HttpRequest(HTTPRequestParams{
				URL:    startUrl,
				Method: "GET",
			})
		}(ip)
		successCount++ // 异步执行，暂且认为发送成功
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("已向 %d 台设备发送推流指令", successCount),
	}
}

// GetWebRTCPlayURL 生成 WebRTC 播放地址（兼容 webrtc:// 协议）
func (r *RtmpService) GetWebRTCPlayURL(streamName string, useSDP bool) string {
	localIP := GetLocalIP()
	if localIP == "" {
		localIP = "localhost"
	}

	// 如果需要 SDP 协商地址
	if useSDP {
		// 返回 HTTP 信令地址
		return fmt.Sprintf("http://%s:1985/rtc/v1/whep/?app=live&stream=%s", localIP, streamName)
	}

	// 返回 webrtc:// URL (兼容格式)
	return fmt.Sprintf("webrtc://%s:1985/live/%s", localIP, streamName)
}

// ParseWebRTCURL 解析 webrtc:// URL 并返回 HTTP 信令地址
func (r *RtmpService) ParseWebRTCURL(webrtcURL string) (string, error) {
	// 支持格式：
	// webrtc://host/app/stream
	// webrtc://host:port/app/stream
	// webrtc://host/live/stream (默认 app=live)

	if !strings.HasPrefix(webrtcURL, "webrtc://") {
		return "", fmt.Errorf("invalid webrtc URL: must start with webrtc://")
	}

	// 去掉 webrtc:// 前缀
	remainder := strings.TrimPrefix(webrtcURL, "webrtc://")

	// 分离 host:port 和路径
	parts := strings.SplitN(remainder, "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid webrtc URL format: missing path")
	}

	hostPort := parts[0]
	path := parts[1]

	// 默认端口 1985
	if !strings.Contains(hostPort, ":") {
		hostPort = hostPort + ":1985"
	}

	// 解析路径: app/stream
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 {
		return "", fmt.Errorf("invalid webrtc URL path: must be app/stream")
	}

	app := pathParts[0]
	stream := strings.Join(pathParts[1:], "/")

	// 构建 HTTP 信令 URL (WHEP 协议)
	httpURL := fmt.Sprintf("http://%s/rtc/v1/whep/?app=%s&stream=%s", hostPort, app, stream)

	return httpURL, nil
}

// GetWebRTCPushURL 生成 WebRTC 推流地址（WHIP 协议）
func (r *RtmpService) GetWebRTCPushURL(streamName string) string {
	localIP := GetLocalIP()
	if localIP == "" {
		localIP = "localhost"
	}

	// 返回 WHIP 信令地址
	return fmt.Sprintf("http://%s:1985/rtc/v1/whip/?app=live&stream=%s", localIP, streamName)
}

// StopDevicePlay 停止设备播放
func (r *RtmpService) StopDevicePlay(deviceIPs []string) map[string]interface{} {
	for _, ip := range deviceIPs {
		go func(deviceIP string) {
			// GET /camera?cmd=stop
			stopUrl := fmt.Sprintf("http://%s/camera?cmd=stop", deviceAddr(deviceIP))
			r.app.HttpRequest(HTTPRequestParams{
				URL:    stopUrl,
				Method: "GET",
			})
		}(ip)
	}
	return map[string]interface{}{
		"success": true,
		"message": "已发送停止指令",
	}
}

// StopPushSession 踢掉推流
func (r *RtmpService) StopPushSession(streamName string) map[string]interface{} {
	if r.useSrs {
		return map[string]interface{}{"success": false, "message": "SRS 暂不支持踢流"}
	}
	// 调用 lal kick 接口
	// http://127.0.0.1:8084/api/ctrl/kick_session?stream_name=xxx
	url := fmt.Sprintf("http://127.0.0.1:8084/api/ctrl/kick_session?stream_name=%s", streamName)
	resp, err := http.Get(url)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	return map[string]interface{}{"success": true, "message": "已断开推流"}
}
