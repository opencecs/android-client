package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode"

	"gitee.com/zoums/dget"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// DeviceStatus 设备状态信息（扩展版，包含设备详细信息）
type DeviceStatus struct {
	IP           string    `json:"ip"`
	Status       string    `json:"status"`        // online/offline
	LastCheckAt  time.Time `json:"lastCheckAt"`
	ResponseTime int64     `json:"responseTime"`  // 响应时间（毫秒）- TCP Ping延迟或HTTP响应时间
	
	// 设备详细信息（从 /info 接口获取）
	APIVersion      string `json:"apiVersion"`      // API当前版本
	LatestVersion   string `json:"latestVersion"`   // API最新版本
	
	// NVME存储信息（从 /info/device 接口获取，需要认证）
	StorageTotal int64  `json:"storageTotal"`  // 总存储空间（MB）
	StorageFree  int64  `json:"storageFree"`   // 可用存储空间（MB）
	StorageUsed  int64  `json:"storageUsed"`   // 已用存储空间（MB）
	
	// 主机固件信息（从 /info/device 接口获取）
	SDKVersion   string `json:"sdkVersion"`    // SDK版本/主机固件版本
	DeviceModel  string `json:"deviceModel"`   // 设备型号
	
	// 系统运行信息（从 /info/device 接口获取）
	CPUTemp      int    `json:"cpuTemp"`       // CPU温度
	CPULoad      string `json:"cpuLoad"`       // CPU负载
	MemoryTotal  int64  `json:"memoryTotal"`   // 内存总量（MB）
	MemoryUsed   int64  `json:"memoryUsed"`    // 内存使用量（MB）
	MMCRead      string `json:"mmcRead"`       // 磁盘读取量
	MMCWrite     string `json:"mmcWrite"`      // 磁盘写入量
	MMCModel     string `json:"mmcModel"`      // 磁盘型号
	MMCTemp      string `json:"mmcTemp"`       // 磁盘温度
	SysUptime    string `json:"sysUptime"`     // 系统运行时间（秒）
	Speed        string `json:"speed"`         // 网络速度
	Network4G    string `json:"network4g"`     // 4G网络状态
	NetworkEth0  string `json:"networkEth0"`   // 以太网状态
	HWAddr       string `json:"hwaddr"`        // MAC地址
	HWAddr1      string `json:"hwaddr_1"`      // MAC地址2
	IP1          string `json:"ip_1"`          // 第二IP
	DeviceID     string `json:"deviceId"`      // 设备ID
	
	// ========== TCP Ping 状态跟踪字段（新增） ==========
	ConsecutiveFailures  int       `json:"-"`                  // 连续失败次数（不暴露给前端）
	ConsecutiveSuccesses int       `json:"-"`                  // 连续成功次数（不暴露给前端）
	LastSuccessLatency   int64     `json:"lastSuccessLatency"` // 最近一次成功的延迟（毫秒）- 失败时显示此值
	LastAPICheckTime     time.Time `json:"-"`                  // 上次API版本检查时间（用于60秒间隔控制）
	LastStorageCheckTime time.Time `json:"-"`                  // 上次存储查询时间（用于60秒间隔控制）
}

// App struct - V3 Service
type App struct {
	wailsApp                *application.App                      // V3应用实例
	imagePullProgress       float64                               // 当前镜像拉取进度 (0-100)
	downloadProgress        float64                               // 当前镜像下载进度 (0-100)
	progressMutex           sync.Mutex                            // 进度更新互斥锁
	downloadCancel          context.CancelFunc                    // 下载取消函数
	downloadCtx             context.Context                       // 下载上下文
	downloadMutex           sync.Mutex                            // 下载控制互斥锁
	uploadCancel            context.CancelFunc                    // 上传取消函数
	uploadCtx               context.Context                       // 上传上下文
	uploadMutex             sync.Mutex                            // 上传控制互斥锁
	projectionWindows       map[string]*application.WebviewWindow // 投屏窗口管理 {deviceIP -> window}
	windowAlwaysOnTop       map[string]bool                       // 投屏窗口置顶状态 {deviceIP -> alwaysOnTop}
	RtmpService             *RtmpService                          // 流媒体服务
	P2PManager              *P2PManager                           // P2P推流管理
	BatchTaskService        *BatchTaskService                     // 批量任务服务
	currentBatchImportTask  *BatchImportTask                      // 当前批量导入任务
	log                     *Logger                               // 日志记录器
	// 设备状态管理
	deviceStatusMap      map[string]*DeviceStatus           // 设备状态缓存 {deviceIP -> status}
	deviceStatusMutex    sync.RWMutex                       // 设备状态读写锁
	deviceIPs            []string                           // 需要监控的设备IP列表
	deviceIPsMutex       sync.RWMutex                       // 设备IP列表读写锁
	devicePasswords      map[string]string                  // 设备密码映射 {deviceIP -> password}
	devicePasswordsMutex sync.RWMutex                       // 设备密码读写锁
	deviceNames          map[string]string                  // 设备名称映射 {deviceIP -> name}
	deviceNamesMutex     sync.RWMutex                       // 设备名称读写锁
	heartbeatStop        chan struct{}                      // 心跳检测停止信号
	heartbeatRunning     bool                               // 心跳服务运行状态标志
	heartbeatMutex       sync.Mutex                         // 保护运行状态的互斥锁
	// HTTP客户端（复用连接）
	httpClient           *http.Client                       // 共享的HTTP客户端，用于设备通信
	// 安卓容器列表后台轮询
	androidCache         map[string]*AndroidDeviceCache     // 安卓容器列表缓存 {deviceIP -> cache}
	androidCacheMutex    sync.RWMutex                       // 安卓缓存读写锁
	androidPollRunning   bool                               // 安卓轮询服务运行状态
	androidPollMutex     sync.Mutex                         // 保护轮询运行状态的互斥锁
	// 安卓容器截图后台轮询
	screenshotCache       map[string]*ScreenshotEntry        // 截图缓存 {"ip_containerName" -> entry}
	screenshotVersions    map[string]int64                   // 每台设备的截图版本号 {ip -> UnixMilli}
	screenshotCacheMutex  sync.RWMutex                       // 截图缓存读写锁
	screenshotPollRunning bool                               // 截图轮询服务运行状态
	screenshotPollMutex   sync.Mutex                         // 保护截图轮询运行状态的互斥锁
	screenshotHTTPClient  *http.Client                       // 截图专用 HTTP Client（连接池复用）
	// OpenCecs 端口映射：deviceIp → map[privatePort]publicPort
	cecsPortMap      map[string]map[int]int
	cecsPortMapMutex sync.RWMutex
}

// WindowHandle 表示跨平台窗口句柄
type WindowHandle uintptr

// ProjectionConfig 投屏窗口配置
type ProjectionConfig struct {
	DeviceIP      string // 设备IP (作为窗口唯一标识)
	TCPPort       int    // TCP端口
	UDPPort       int    // UDP端口
	ControlPort   int    // 控制端口
	Orient        int    // 默认旋转方向 (0竖屏 1横屏)
	List          string // 批量控制列表
	Term          string // 窗口标题
	ContainerID   string // 容器ID
	ContainerName string // 容器名称
	Width         int    // 窗口宽度
	Height        int    // 窗口高度
}

// CheckPortOpen 检查本地端口是否开放
func (a *App) CheckPortOpen(port int) bool {
	timeout := time.Second
	// 尝试连接本地回环地址
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// SetWailsApp 设置Wails应用实例
func (a *App) SetWailsApp(app *application.App) {
	a.wailsApp = app
	a.projectionWindows = make(map[string]*application.WebviewWindow)
	a.windowAlwaysOnTop = make(map[string]bool)
	
	// 初始化日志记录器
	a.log = &Logger{prefix: "BatchTask"}
	
	// 初始化批量任务服务
	configDir, err := os.UserConfigDir()
	if err != nil {
		a.log.Error("获取配置目录失败: %v", err)
		return
	}
	batchTaskDataDir := filepath.Join(configDir, "edgeclient", "batch_tasks")
	
	a.BatchTaskService = &BatchTaskService{}
	if err := a.BatchTaskService.InitBatchTaskService(a, batchTaskDataDir); err != nil {
		a.log.Error("初始化批量任务服务失败: %v", err)
	}
}

// Container 容器结构体
type Container struct {
	ID          string
	Name        string
	Status      string
	IP          string
	NetworkName string
	NetworkMode string
	Ports       []Port
}

// Port 端口结构体
type Port struct {
	PrivatePort int
	PublicPort  int
	Type        string
}

// HTTPRequestParams HTTP请求参数
type HTTPRequestParams struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HttpRequest 通用HTTP请求处理函数
func (a *App) HttpRequest(params HTTPRequestParams) map[string]interface{} {
	shouldLog := strings.Contains(params.URL, "/modifydev") || strings.Contains(params.URL, "/camera")
	if shouldLog {
		log.Printf("[HttpRequest] %s %s", params.Method, params.URL)
	}
	// 创建HTTP请求
	req, err := http.NewRequest(params.Method, params.URL, bytes.NewBuffer([]byte(params.Body)))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加请求头
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	// 设置默认Content-Type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}
	if shouldLog {
		log.Printf("[HttpRequest] status=%d body=%s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果无法解析为JSON，直接返回原始响应
		return map[string]interface{}{
			"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
			"status":  resp.StatusCode,
			"body":    string(body),
			"raw":     string(body),
			"headers": resp.Header,
		}
	}

	// 返回解析后的JSON响应
	return map[string]interface{}{
		"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":  resp.StatusCode,
		"body":    result,
		"raw":     string(body),
		"headers": resp.Header,
	}
}

// FlexibleInt 灵活的整数类型，可以接受字符串或数字
type FlexibleInt int

// UnmarshalJSON 自定义JSON反序列化，支持字符串和数字
func (fi *FlexibleInt) UnmarshalJSON(data []byte) error {
	// 尝试作为数字解析
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fi = FlexibleInt(i)
		return nil
	}
	
	// 尝试作为字符串解析
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// 将字符串转换为整数
		i, err := strconv.Atoi(s)
		if err != nil {
			*fi = 0
			return nil
		}
		*fi = FlexibleInt(i)
		return nil
	}
	
	*fi = 0
	return nil
}

// AnnouncementData 公告数据结构
type AnnouncementData struct {
	AnnouncementID  int         `json:"announcementId"`
	Title           string      `json:"title"`
	Content         string      `json:"content"`
	AnnouncementType string     `json:"announcementType"`
	DisplayDuration FlexibleInt `json:"displayDuration"`
	ValidDuration   FlexibleInt `json:"validDuration"`
	IsPopup         bool        `json:"isPopup"`
	CreatedAt       string      `json:"createdAt"`
	ValidUntil      string      `json:"validUntil"`
}

// AnnouncementResponse 公告接口响应结构
type AnnouncementResponse struct {
	CodeID  int              `json:"code_id"`
	Msg     string           `json:"msg"`
	Data    AnnouncementData `json:"data"` // data为单个对象
}

// GetAnnouncement 获取系统公告
func (a *App) GetAnnouncement() map[string]interface{} {
	log.Printf("[GetAnnouncement] 开始获取系统公告...")
	
	apiURL := "https://newapi.moyunteng.com/api/v1/announcement"
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 创建GET请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("[GetAnnouncement] 创建请求失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"msg":     fmt.Sprintf("创建请求失败: %v", err),
			"data":    nil,
		}
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GetAnnouncement] 请求失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"msg":     fmt.Sprintf("请求失败: %v", err),
			"data":    nil,
		}
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GetAnnouncement] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"msg":     fmt.Sprintf("读取响应失败: %v", err),
			"data":    nil,
		}
	}
	
	log.Printf("[GetAnnouncement] 响应状态码: %d", resp.StatusCode)
	log.Printf("[GetAnnouncement] 响应内容: %s", string(body))
	
	// 解析JSON响应
	var result AnnouncementResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[GetAnnouncement] 解析JSON失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"msg":     fmt.Sprintf("解析响应失败: %v", err),
			"data":    nil,
		}
	}
	
	// 检查data是否为空（无公告时 announcementId 为 0）
	if result.Data.AnnouncementID == 0 {
		log.Printf("[GetAnnouncement] 当前无公告")
		return map[string]interface{}{
			"code_id": result.CodeID,
			"msg":     result.Msg,
			"data":    nil,
		}
	}

	// 取公告数据
	item := &result.Data
	log.Printf("[GetAnnouncement] 成功获取公告: title=%s, isPopup=%v", item.Title, item.IsPopup)
	
	// 返回标准格式
	return map[string]interface{}{
		"code_id": result.CodeID,
		"msg":     result.Msg,
		"data": map[string]interface{}{
			"announcementId":   item.AnnouncementID,
			"title":            item.Title,
			"content":          item.Content,
			"announcementType": item.AnnouncementType,
			"displayDuration":  int(item.DisplayDuration),
			"validDuration":    int(item.ValidDuration),
			"isPopup":          item.IsPopup,
			"createdAt":        item.CreatedAt,
			"validUntil":       item.ValidUntil,
		},
	}
}

// GetUserRabbetList 获取用户实例列表（无需 token，使用 term_info_nologn 接口）
func (a *App) GetUserRabbetList(rabbet string) map[string]interface{} {
	log.Printf("[GetUserRabbetList] rabbet=%s", rabbet)

	apiURL := "https://www.moyunteng.com/api/api.php"

	// host 为 JSON 数组序列化后的字符串，_ts 为秒级时间戳（整数）
	hostArr := []string{rabbet}
	hostJSON, _ := json.Marshal(hostArr)
	hostStr := string(hostJSON) // 形如 ["xxxx"]
	tsInt := time.Now().Unix()
	ts := fmt.Sprintf("%d", tsInt)

	// 签名：按 key 排序后取值拼接
	signParams := map[string]string{
		"host": hostStr,
		"_ts":  ts,
	}
	keys := []string{"_ts", "host"}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, signParams[k])
	}
	signStr := strings.Join(parts, "#") + "#@#1234A98413G=--..234"
	h := md5.New()
	h.Write([]byte(signStr))
	sign := hex.EncodeToString(h.Sum(nil))

	log.Printf("[GetUserRabbetList] signStr=%s sign=%s", signStr, sign)

	// data 为 JSON：host 是字符串（数组序列化后），_ts 是数字，_sign 是字符串
	dataMap := map[string]interface{}{
		"host":  hostStr,
		"_ts":   tsInt,
		"_sign": sign,
	}
	dataJSON, _ := json.Marshal(dataMap)
	log.Printf("[GetUserRabbetList] data=%s", string(dataJSON))

	// type 和 data 用标准 form 编码
	formData := url.Values{}
	formData.Set("type", "term_info_nologn")
	formData.Set("data", string(dataJSON))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[GetUserRabbetList] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.moyunteng.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://www.moyunteng.com")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GetUserRabbetList] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GetUserRabbetList] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[GetUserRabbetList] 响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{
			"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
			"status":  resp.StatusCode,
			"message": string(body),
		}
	}

	return map[string]interface{}{
		"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":  resp.StatusCode,
		"data":    result,
	}
}

// GetUserRabbetListWithToken 获取用户实例列表（授权页面，需要 token）
func (a *App) GetUserRabbetListWithToken(token string, page, limit int, rabbet string) map[string]interface{} {
	log.Printf("[GetUserRabbetListWithToken] page=%d, limit=%d, rabbet=%s", page, limit, rabbet)

	apiURL := fmt.Sprintf("https://www.moyunteng.com/pc/ajax.php?c=index&a=get_user_rabbet_list_v2")

	formData := url.Values{}
	// formData.Set("type", "user_host_oper")
	formData.Set("page", strconv.Itoa(page))
	formData.Set("limit", strconv.Itoa(limit))
	formData.Set("rabbet", rabbet)
	formData.Set("token", token)

	log.Printf("[GetUserRabbetList] API URL: %s", apiURL)
	log.Printf("[GetUserRabbetList] FormData: %s", formData.Encode())

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[GetUserRabbetList] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.moyunteng.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://www.moyunteng.com")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GetUserRabbetList] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GetUserRabbetList] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[GetUserRabbetList] 响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{
			"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
			"status":  resp.StatusCode,
			"message": string(body),
		}
	}

	return map[string]interface{}{
		"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":  resp.StatusCode,
		"data":    result,
	}

}

// GetPackage 获取购买套餐列表
func (a *App) GetPackage(token string) map[string]interface{} {
	log.Printf("[GetPackage] 开始获取套餐列表")

	apiURL := "https://www.moyunteng.com/api/api.php"

	formData := url.Values{}
	formData.Set("type", "get_package")

	dataMap := map[string]string{"token": token}
	dataJSON, _ := json.Marshal(dataMap)
	formData.Set("data", string(dataJSON))

	log.Printf("[GetPackage] API URL: %s", apiURL)
	log.Printf("[GetPackage] FormData: %s", formData.Encode())

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[GetPackage] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GetPackage] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GetPackage] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[GetPackage] 响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{
			"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
			"status":  resp.StatusCode,
			"message": string(body),
		}
	}

	return map[string]interface{}{
		"success": resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":  resp.StatusCode,
		"data":    result,
	}
}

// OrderParams 下单参数
type OrderParams struct {
	Rabbet  string `json:"rabbet"`
	Package string `json:"package"`
	Money   string `json:"money"`
	PayType string `json:"paytype"`
	Token   string `json:"token"`
}

// OrderResult 下单结果
type OrderResult struct {
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	QRCode string `json:"qrcode,omitempty"`
	Data   any    `json:"data,omitempty"`
}

// CreateOrder 创建订单
func (a *App) CreateOrder(params OrderParams) map[string]interface{} {
	log.Printf("[CreateOrder] 开始创建订单, params: %+v", params)

	apiURL := "https://www.moyunteng.com/api/api.php"

	formData := url.Values{}
	formData.Set("type", "get_order_v2")

	orderData := map[string]string{
		"rabbet":  params.Rabbet,
		"package": params.Package,
		"money":   params.Money,
		"paytype": params.PayType,
		"token":   params.Token,
	}
	orderDataJSON, _ := json.Marshal(orderData)
	formData.Set("data", string(orderDataJSON))

	log.Printf("[CreateOrder] API URL: %s", apiURL)
	log.Printf("[CreateOrder] FormData: %s", formData.Encode())

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[CreateOrder] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.moyunteng.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://www.moyunteng.com")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[CreateOrder] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[CreateOrder] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[CreateOrder] 响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	var result OrderResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[CreateOrder] 解析OrderResult失败，尝试解析通用响应: %v", err)

		var genericResult map[string]interface{}
		if jsonErr := json.Unmarshal(body, &genericResult); jsonErr == nil {
			log.Printf("[CreateOrder] 通用响应: %+v", genericResult)
			return map[string]interface{}{
				"success": false,
				"message": genericResult["message"],
				"raw":     string(body),
			}
		}

		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
			"raw":     string(body),
		}
	}

	log.Printf("[CreateOrder] 解析结果: Code=%s, Msg=%s", result.Code, result.Msg)

	if result.Code != "0" && result.Code != "200" {
		return map[string]interface{}{
			"success": false,
			"message": result.Msg,
		}
	}

	return map[string]interface{}{
		"success": true,
		"code":    result.Code,
		"msg":     result.Msg,
		"data":    result.Data,
	}
}

// QueryOrderParams 查询订单参数
type QueryOrderParams struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// OrderData 订单数据
type OrderData struct {
	ID      string `json:"id"`
	OID     string `json:"oid"`
	UID     string `json:"uid"`
	PID     string `json:"pid"`
	CTime   string `json:"ctime"`
	Price   string `json:"price"`
	Rabbet  string `json:"rabbet"`
	UTime   string `json:"utime"`
	PInfo   string `json:"pinfo"`
	PayType string `json:"paytype"`
	State   string `json:"state"`
}

// QueryOrderResult 查询订单结果
type QueryOrderResult struct {
	Code string    `json:"code"`
	Msg  string    `json:"msg"`
	Data OrderData `json:"data"`
}

// QueryOrderStatus 查询订单支付状态
func (a *App) QueryOrderStatus(params QueryOrderParams) map[string]interface{} {
	log.Printf("[QueryOrderStatus] 开始查询订单状态, params: %+v", params)

	apiURL := "https://www.moyunteng.com/api/api.php"

	formData := url.Values{}
	formData.Set("type", "query_order")

	orderData := map[string]string{
		"id":    params.ID,
		"token": params.Token,
	}
	orderDataJSON, _ := json.Marshal(orderData)
	formData.Set("data", string(orderDataJSON))

	log.Printf("[QueryOrderStatus] API URL: %s", apiURL)
	log.Printf("[QueryOrderStatus] FormData: %s", formData.Encode())

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[QueryOrderStatus] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://www.moyunteng.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://www.moyunteng.com")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[QueryOrderStatus] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[QueryOrderStatus] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[QueryOrderStatus] 响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	var result QueryOrderResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[QueryOrderStatus] 解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
			"raw":     string(body),
		}
	}

	log.Printf("[QueryOrderStatus] 解析结果: Code=%s, Msg=%s, Data.State=%s", result.Code, result.Msg, result.Data.State)

	orderState, _ := strconv.Atoi(result.Data.State)

	return map[string]interface{}{
		"success": true,
		"code":    result.Code,
		"msg":     result.Msg,
		"state":   orderState,
		"data":    result.Data,
	}
}

// ProgressReader 带进度回调的Reader包装器
type ProgressReader struct {
	Reader    io.Reader
	Size      int64
	ReadBytes int64
	Callback  func(float64)
}

// Read 实现io.Reader接口
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.ReadBytes += int64(n)
		progress := float64(pr.ReadBytes) / float64(pr.Size) * 100
		if pr.Callback != nil {
			pr.Callback(progress)
		}
	}
	return n, err
}

// DeviceInfo 设备信息结构体

type DeviceInfo struct {
	IP      string `json:"ip"`
	Type    string `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    int    `json:"port"` // RCore TCP 端口，默认 30105
}

// AndroidItem 安卓云机实例结构体
type AndroidItem struct {
	Name         string                   `json:"name"`
	Status       string                   `json:"status"`
	IndexNum     int                      `json:"indexNum"`
	DataPath     string                   `json:"dataPath"`
	Image        string                   `json:"image"`
	IP           string                   `json:"ip"`
	NetworkName  string                   `json:"networkName"`
	Subnet       string                   `json:"subnet"`
	Gateway      string                   `json:"gateway"`
	PortBindings map[string][]PortBinding `json:"portBindings"`
}

// PortBinding 端口绑定结构体
type PortBinding struct {
	HostIp   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// DockerContainer Docker容器结构体
type DockerContainer struct {
	ID              string                 `json:"ID"`
	Names           []string               `json:"Names"`
	Image           string                 `json:"Image"`
	Status          string                 `json:"Status"`
	Created         int64                  `json:"Created"`
	Ports           []DockerPort           `json:"Ports"`
	Config          *DockerConfig          `json:"Config"`
	HostConfig      *DockerHostConfig      `json:"HostConfig"`
	NetworkSettings *DockerNetworkSettings `json:"NetworkSettings"`
	Labels          map[string]string      `json:"Labels"`
}

// DockerPort Docker端口结构体
type DockerPort struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

// DockerConfig Docker配置结构体
type DockerConfig struct {
	Labels map[string]string `json:"Labels"`
}

// DockerHostConfig Docker主机配置结构体
type DockerHostConfig struct {
	Devices []DockerDevice `json:"Devices"`
}

// DockerDevice Docker设备结构体
type DockerDevice struct {
	PathOnHost        string `json:"PathOnHost"`
	PathInContainer   string `json:"PathInContainer"`
	CgroupPermissions string `json:"CgroupPermissions"`
}

// DockerNetworkSettings Docker网络设置结构体
type DockerNetworkSettings struct {
	Ports map[string][]DockerPortBinding `json:"Ports"`
}

// DockerPortBinding Docker端口绑定结构体
type DockerPortBinding struct {
	HostIp   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// 镜像相关结构体定义（用于纯Go Docker镜像拉取）
type Manifest struct {
	Config Descriptor   `json:"config"`
	Layers []Descriptor `json:"layers"`
}

type Descriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type ImageTag struct {
	Host     string
	Name     string
	Version  string
	Hash     string
	Authed   bool
	AuthInfo TokenInfo
}

type ManifestFile struct {
	Config   string
	RepoTags []string
	Layers   []string
}

type TokenInfo struct {
	AccessToken string `json:"access_token"`
	Token       string `json:"token"`
}

// VPCProxy VPC代理服务器结构体
type VPCProxy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// VPCHost VPC主机结构体
type VPCHost struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Type      string `json:"type"`
	VPCStatus string `json:"vpcStatus"`
}

// 全局变量
var (
	runningProjections = make(map[string]*os.Process) // key: ip_containerName 组合唯一标识
	pendingProjections = make(map[string]bool)         // 正在启动中的投屏（解压/查找阶段）
	projectionLock     sync.Mutex
)

// 模拟设备列表数据
var mockDevices = []DeviceInfo{
	{
		IP:      "192.168.2.100",
		Type:    "q1",
		ID:      "device-1",
		Name:    "q1_v3",
		Version: "v3",
	},
	{
		IP:      "192.168.1.101",
		Type:    "q1",
		ID:      "device-2",
		Name:    "q1_v2",
		Version: "v2",
	},
	{
		IP:      "192.168.1.102",
		Type:    "q1",
		ID:      "device-3",
		Name:    "q1_10",
		Version: "v1",
	},
	{
		IP:      "192.168.1.103",
		Type:    "q1",
		ID:      "device-4",
		Name:    "q1",
		Version: "v0",
	},
}

// 模拟容器列表数据
var mockInstances = []AndroidItem{
	{
		Name:         "云机-001",
		Status:       "running",
		IndexNum:     1,
		DataPath:     "/data/android-001",
		Image:        "android-11.0",
		IP:           "192.168.2.101",
		NetworkName:  "br0",
		Subnet:       "255.255.255.0",
		Gateway:      "192.168.2.1",
		PortBindings: map[string][]PortBinding{},
	},
	{
		Name:         "云机-002",
		Status:       "shutdown",
		IndexNum:     1,
		DataPath:     "/data/android-002",
		Image:        "android-10.0",
		IP:           "192.168.2.102",
		NetworkName:  "br0",
		Subnet:       "255.255.255.0",
		Gateway:      "192.168.2.1",
		PortBindings: map[string][]PortBinding{},
	},
	{
		Name:         "云机-003",
		Status:       "restarting",
		IndexNum:     2,
		DataPath:     "/data/android-003",
		Image:        "android-11.0",
		IP:           "192.168.2.103",
		NetworkName:  "br0",
		Subnet:       "255.255.255.0",
		Gateway:      "192.168.2.1",
		PortBindings: map[string][]PortBinding{},
	},
}

// 模拟VPC代理列表数据
var mockVPCProxies = []VPCProxy{
	{
		ID:   "proxy-1",
		Name: "代理服务器1",
		IP:   "10.0.0.1",
	},
	{
		ID:   "proxy-2",
		Name: "代理服务器2",
		IP:   "10.0.0.2",
	},
}

// 模拟VPC主机列表数据
var mockVPCHosts = []VPCHost{
	{
		ID:        "host-1",
		Name:      "云机-001",
		IP:        "192.168.2.101",
		Type:      "android",
		VPCStatus: "enabled",
	},
	{
		ID:        "host-2",
		Name:      "云机-002",
		IP:        "192.168.2.102",
		Type:      "android",
		VPCStatus: "disabled",
	},
	{
		ID:        "host-3",
		Name:      "主机-001",
		IP:        "192.168.2.100",
		Type:      "host",
		VPCStatus: "enabled",
	},
}

// discoverDevices 使用 UDP 广播 "lgcloud" 到所有网卡的广播地址，以及常见私有网段的广播地址
func discoverDevices() ([]DeviceInfo, error) {
	deviceMap := make(map[string]DeviceInfo)

	log.Printf("start discoverDevices")

	// 创建UDP socket，绑定到 0.0.0.0:0
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 允许广播
	rawConn, err := conn.(*net.UDPConn).SyscallConn()
	if err != nil {
		return nil, err
	}

	err = rawConn.Control(func(fd uintptr) {
		// 调用平台特定的函数，具体实现由build tags决定
		err = setSocketBroadcast(fd)
	})
	if err != nil {
		return nil, err
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Println("本地监听地址:", localAddr)

	message := []byte("lgcloud")

	// 收集所有要广播的目标地址
	broadcastTargets := make(map[string]bool)

	// 1. 先添加有限广播地址（本地网络）
	broadcastTargets["255.255.255.255"] = true

	// 2. 获取所有网卡的广播地址
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			// 跳过回环接口和未启用的接口
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok || ipNet.IP.IsLoopback() {
					continue
				}

				if ipNet.IP.To4() != nil {
					// 计算本机网卡的广播地址 (确保只操作4字节)
					ip4 := ipNet.IP.To4()
					mask4 := ipNet.Mask
					broadcastIP := net.IPv4(
						ip4[0]|^mask4[0],
						ip4[1]|^mask4[1],
						ip4[2]|^mask4[2],
						ip4[3]|^mask4[3],
					)
					if broadcastIP[0] != 127 {
						broadcastTargets[broadcastIP.String()] = true
					}
				}
			}
		}
	}

	// 3. 扫描常见私有网段的广播地址（即使路由器转发广播也能到达）
	// 10.x.x.x 网段
	for a := 1; a <= 10; a++ {
		broadcastTargets[fmt.Sprintf("10.%d.255.255", a)] = true
	}
	// 172.16.x.x - 172.31.x.x 网段
	for b := 16; b <= 31; b++ {
		broadcastTargets[fmt.Sprintf("172.%d.255.255", b)] = true
	}
	// 192.168.x.x 网段
	for c := 0; c <= 255; c++ {
		broadcastTargets[fmt.Sprintf("192.168.%d.255", c)] = true
	}

	// 发送广播到所有目标地址
	broadcastCount := 0
	for targetIP := range broadcastTargets {
		target := &net.UDPAddr{
			IP:   net.ParseIP(targetIP),
			Port: 7678,
		}
		_, err = conn.WriteTo(message, target)
		if err == nil {
			broadcastCount++
			log.Printf("已发送广播到 %s:7678", targetIP)
		}
	}
	log.Printf("共发送 %d 个广播", broadcastCount)

	// 设置读取超时
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return []DeviceInfo{}, err
		}

		response := string(buf[:n])
		log.Printf("收到来自 %s 的响应: %s\n", addr.String(), response)
		parts := strings.Split(response, ":")
		if len(parts) >= 3 {
			ip := strings.Split(addr.String(), ":")[0]
			deviceID := parts[1]
			deviceName := parts[2]

			// 解析版本信息
			version := ""
			if strings.Contains(deviceName, "_v") {
				// 提取版本号 (如 q1_v3 -> v3)
				versionParts := strings.Split(deviceName, "_v")
				if len(versionParts) > 1 {
					version = "v" + versionParts[len(versionParts)-1]
				}
			} else if strings.Contains(deviceName, "_") {
				// 处理 q1_10 -> v1 的情况
				versionParts := strings.Split(deviceName, "_")
				if len(versionParts) > 1 && versionParts[1] == "10" {
					version = "v1"
				}
			} else {
				// 不带后缀是 v0
				version = "v0"
			}

			if _, exists := deviceMap[deviceID]; !exists {
				deviceMap[deviceID] = DeviceInfo{
					IP:      ip,
					Type:    parts[0],
					ID:      deviceID,
					Name:    deviceName,
					Version: version,
					Port:    30105,
				}
			} else {
				existing := deviceMap[deviceID]
				existing.IP = ip
				existing.Version = version
				if existing.Port == 0 {
					existing.Port = 30105
				}
				deviceMap[deviceID] = existing
			}
		}
	}

	// 转换map为slice
	devices := make([]DeviceInfo, 0, len(deviceMap))
	for _, d := range deviceMap {
		devices = append(devices, d)
	}

	// 按IP排序
	sort.Slice(devices, func(i, j int) bool {
		return strings.Compare(devices[i].IP, devices[j].IP) < 0
	})
	return devices, nil
}

// DiscoverDevicesManually 手动发现指定IP的设备
// ips: 逗号或换行分隔的IP列表
// 返回: 成功发现的设备列表, 失败的IP列表
func (a *App) DiscoverDevicesManually(ips string) ([]DeviceInfo, []string, error) {
	log.Printf("DiscoverDevicesManually: %s", ips)

	// 解析IP列表
	var ipList []string
	for _, ip := range strings.FieldsFunc(ips, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ' ' || r == '\t'
	}) {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			// 验证IP格式
			if net.ParseIP(ip) == nil {
				log.Printf("无效的IP地址: %s", ip)
				continue
			}
			ipList = append(ipList, ip)
		}
	}

	if len(ipList) == 0 {
		return nil, nil, errors.New("未提供有效的IP地址")
	}

	// 创建UDP socket
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Printf("本地监听地址: %s", localAddr)

	message := []byte("lgcloud")

	// 设置读取超时
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	deviceMap := make(map[string]DeviceInfo)
	failedIPs := make([]string, 0)
	responseCh := make(chan struct {
		ip       string
		response string
	}, len(ipList))

	// 向每个IP发送查询
	for _, ip := range ipList {
		target := &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: 7678,
		}
		_, err = conn.WriteTo(message, target)
		if err != nil {
			log.Printf("发送到 %s 失败: %v", ip, err)
			failedIPs = append(failedIPs, ip)
		} else {
			log.Printf("已发送查询到 %s:7678", ip)
		}
	}

	// 收集响应
	buf := make([]byte, 1024)
	responsesReceived := make(map[string]bool)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}

		ip := strings.Split(addr.String(), ":")[0]
		if responsesReceived[ip] {
			continue
		}
		responsesReceived[ip] = true

		response := string(buf[:n])
		log.Printf("收到来自 %s 的响应: %s", ip, response)
		responseCh <- struct {
			ip       string
			response string
		}{ip: ip, response: response}
	}

	// 处理响应
	totalResponses := len(responseCh)
	for i := 0; i < totalResponses; i++ {
		select {
		case data := <-responseCh:
			parts := strings.Split(data.response, ":")
			if len(parts) >= 3 {
				deviceID := parts[1]
				deviceName := parts[2]

				version := ""
				if strings.Contains(deviceName, "_v") {
					versionParts := strings.Split(deviceName, "_v")
					if len(versionParts) > 1 {
						version = "v" + versionParts[len(versionParts)-1]
					}
				} else if strings.Contains(deviceName, "_") {
					versionParts := strings.Split(deviceName, "_")
					if len(versionParts) > 1 && versionParts[1] == "10" {
						version = "v1"
					}
				} else {
					version = "v0"
				}

				deviceMap[deviceID] = DeviceInfo{
					IP:      data.ip,
					Type:    parts[0],
					ID:      deviceID,
					Name:    deviceName,
					Version: version,
					Port:    30105,
				}
			}
		default:
			break
		}
	}

	// 找出没有响应的IP
	for _, ip := range ipList {
		if !responsesReceived[ip] {
			failedIPs = append(failedIPs, ip)
		}
	}

	// 转换map为slice
	devices := make([]DeviceInfo, 0, len(deviceMap))
	for _, d := range deviceMap {
		devices = append(devices, d)
	}

	// 按IP排序
	sort.Slice(devices, func(i, j int) bool {
		return strings.Compare(devices[i].IP, devices[j].IP) < 0
	})

	log.Printf("DiscoverDevicesManually 完成: 成功=%d, 失败=%d", len(devices), len(failedIPs))
	return devices, failedIPs, nil
}

// GetWebplayerHTML 返回webplayer的HTML内容
func (a *App) GetWebplayerHTML() string {
	// 读取play.html文件内容
	content, err := webplayerAssets.ReadFile("frontend/webplayer/play.html")
	if err != nil {
		return fmt.Sprintf("Error reading play.html: %v", err)
	}

	return string(content)
}

// GetWebplayerAsset 返回webplayer的资源文件内容
func (a *App) GetWebplayerAsset(path string) string {
	// 构建完整的资源路径
	fullPath := fmt.Sprintf("frontend/webplayer/%s", path)

	// 读取资源文件内容
	content, err := webplayerAssets.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading asset %s: %v", path, err)
	}

	return string(content)
}

// emitEvent 发送事件到前端 (V3事件系统)
func (a *App) emitEvent(eventName string, data interface{}) {
	if a.wailsApp != nil {
		a.log.Info("[Event] 发送事件: %s", eventName)
		a.wailsApp.Event.Emit(eventName, data)
	} else {
		a.log.Error("[Event] 无法发送事件 %s: wailsApp 未初始化", eventName)
	}
}

// NewApp creates a new App application struct
func NewApp() *App {
	// ========== 大规模设备优化 HTTP Client 配置 ==========
	// 设计目标: 支持 5000+ 设备同时在线
	// 核心策略: 连接池复用 + Keep-Alive + 并发控制
	
	transport := &http.Transport{
		// 连接池配置 - 支持大规模设备
		MaxIdleConns:        10000,            // 全局最大空闲连接数 (支持5000台设备×2个接口)
		MaxIdleConnsPerHost: 2,                // 每设备保持2个连接 (心跳+存储并发)
		MaxConnsPerHost:     5,                // 每设备最大并发连接数 (防止单设备占用过多)
		IdleConnTimeout:     120 * time.Second, // 空闲连接保持2分钟 (覆盖多轮查询)
		
		// Keep-Alive 配置 - 避免频繁握手
		DisableKeepAlives:   false,            // 启用 HTTP Keep-Alive (关键!)
		
		// TCP 层优化
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,        // 连接建立超时
			KeepAlive: 60 * time.Second,       // TCP Keep-Alive 探测间隔
		}).DialContext,
		
		// 超时配置
		ResponseHeaderTimeout: 10 * time.Second, // 等待响应头超时
		ExpectContinueTimeout: 1 * time.Second,  // 100-Continue 超时
		
		// 性能优化
		DisableCompression:    false,          // 启用压缩节省带宽
		ForceAttemptHTTP2:     false,          // 禁用HTTP/2 (设备只支持HTTP/1.1)
		
		// 资源限制
		MaxResponseHeaderBytes: 4096,          // 限制响应头大小 (4KB足够)
	}
	
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   0, // 不设置全局超时，在每个请求中单独使用 context.WithTimeout
	}

	// 截图专用 HTTP Client：高并发连接池，短超时
	screenshotTransport := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true, // 截图是二进制，无需压缩
		ForceAttemptHTTP2:   false,
	}
	screenshotHTTPClient := &http.Client{
		Transport: screenshotTransport,
		Timeout:   0,
	}

	app := &App{
		imagePullProgress: 0,
		downloadProgress:  0,
		RtmpService:       NewRtmpService(),
		P2PManager:        NewP2PManager(),
		deviceStatusMap:   make(map[string]*DeviceStatus),
		deviceNames:       make(map[string]string),
		deviceIPs:         make([]string, 0),
		heartbeatStop:     make(chan struct{}),
		httpClient:        httpClient,            // 初始化共享 HTTP Client
		screenshotHTTPClient: screenshotHTTPClient, // 截图专用 HTTP Client
		// 安卓轮询缓存初始化
		androidCache:       make(map[string]*AndroidDeviceCache),
		// 截图缓存初始化
		screenshotCache:    make(map[string]*ScreenshotEntry),
		screenshotVersions: make(map[string]int64),
	}
	app.RtmpService.SetApp(app)

	return app
}

// startup is called when the app starts.
func (a *App) startup() {
	// 初始化逻辑在NewApp构造函数中完成
	if a.RtmpService != nil {
		a.RtmpService.StartRtmpServer()
	}
	
	// 注意：心跳检测服务由前端通过 initDeviceHeartbeat() 启动
	// 不在这里启动，避免重复启动
	log.Printf("[启动] App 启动完成，等待前端初始化心跳检测服务")

	// 后台预解压投屏资源(player.exe及其DLL)，避免首次投屏时的长时间等待
	go func() {
		start := time.Now()
		log.Printf("[启动] 开始后台预解压投屏资源...")
		if exePath, err := ensureEmbeddedPlayerExtracted(); err != nil {
			log.Printf("[启动] 后台预解压投屏资源失败: %v", err)
		} else {
			log.Printf("[启动] 后台预解压投屏资源完成，耗时 %v，路径: %s", time.Since(start), exePath)
		}
	}()
}

// CleanupProjectionProcesses 清理所有投屏进程
func (a *App) CleanupProjectionProcesses() {
	log.Printf("开始清理所有投屏进程")
	projectionLock.Lock()
	defer projectionLock.Unlock()

	// 遍历所有运行中的投屏进程并终止
	for containerID, proc := range runningProjections {
		if proc != nil {
			log.Printf("终止投屏进程: containerID=%s, PID=%d", containerID, proc.Pid)
			if err := proc.Kill(); err != nil {
				log.Printf("终止投屏进程失败: %v", err)
			} else {
				log.Printf("成功终止投屏进程: containerID=%s, PID=%d", containerID, proc.Pid)
			}
		}
	}

	// 清空runningProjections map
	runningProjections = make(map[string]*os.Process)
	log.Printf("投屏进程清理完成")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// chatTicketSecret 客服系统 JWT 签名密钥（与 10.10.0.31:8080 后端 CHAT_TICKET_SECRET 保持一致）
const chatTicketSecret = "Hk8BpN3qR7wE2yJ5uX9zA4cW6vF1nM8sL5tG"

// GenerateChatTicket 生成客服系统 JWT ticket（有效期 60 秒）
func (a *App) GenerateChatTicket(userID, userName string) map[string]interface{} {
	log.Printf("[GenerateChatTicket] ===== 开始生成 chat ticket =====")
	log.Printf("[GenerateChatTicket] 入参: userID=%q, userName=%q", userID, userName)

	now := time.Now().Unix()

	// 生成 jti（随机16字节hex）
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		log.Printf("[GenerateChatTicket] 生成 jti 失败: %v", err)
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	jti := hex.EncodeToString(jtiBytes)
	log.Printf("[GenerateChatTicket] jti=%s", jti)

	// JWT Header
	headerJSON, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	log.Printf("[GenerateChatTicket] header=%s", string(headerJSON))

	// JWT Payload
	payloadMap := map[string]interface{}{
		"sub":  userID,
		"name": userName,
		"iat":  now,
		"exp":  now + 300,
		"jti":  jti,
	}
	payloadJSON, _ := json.Marshal(payloadMap)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	log.Printf("[GenerateChatTicket] payload=%s", string(payloadJSON))
	log.Printf("[GenerateChatTicket] iat=%d, exp=%d (有效期300s, 到期时间: %s)",
		now, now+300, time.Unix(now+300, 0).Format("2006-01-02 15:04:05"))

	// HMAC-SHA256 签名
	signingInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(chatTicketSecret))
	mac.Write([]byte(signingInput))
	signatureB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	ticket := signingInput + "." + signatureB64
	log.Printf("[GenerateChatTicket] ticket 生成成功, 长度=%d", len(ticket))
	log.Printf("[GenerateChatTicket] ===== ticket 生成完成 =====")
	return map[string]interface{}{"success": true, "ticket": ticket}
}

// DiscoverDevices 获取设备列表
func (a *App) DiscoverDevices() []DeviceInfo {
	log.Printf("[IPC] 收到 DiscoverDevices 调用")
	log.Printf("[IPC] 调用时间: %s", time.Now().Format(time.RFC3339))

	// 调用真实的UDP设备发现功能
	devices, err := discoverDevices()
	if err != nil {
		log.Printf("[IPC] 设备发现失败: %v", err)
		// 失败时返回空列表
		return []DeviceInfo{}
	}

	log.Printf("[IPC] 返回设备数量: %d", len(devices))
	log.Printf("[IPC] 返回设备列表: %+v", devices)

	return devices
}

// GetContainers 获取容器列表
func (a *App) GetContainers(deviceIP string, version string, password string) interface{} {
	log.Printf("[IPC] 收到 GetContainers 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s", deviceIP, version)

	// 调用真实的容器获取逻辑
	containers, err := getContainers(deviceIP, version, password)
	if err != nil {
		log.Printf("[IPC] 获取容器列表失败: %v", err)
		// 失败时返回空结果
		if version == "v3" {
			// 返回错误信息，包含网络连接错误的情况
			return map[string]interface{}{
				"code":    500,
				"message": "网络连接错误",
			}
		} else {
			return []DockerContainer{}
		}
	}

	// log.Printf("[IPC] GetContainers 返回结果: %+v", containers)
	return containers
}

// getContainers 调用设备API获取容器列表
func getContainers(deviceIP string, version string, password string) (interface{}, error) {
	if version == "v3" {
		// V3 API: 调用设备的API接口
		return getV3Containers(deviceIP, password)
	} else {
		// V0-V2 Docker API: 调用Docker API
		return getDockerContainers(deviceIP, version, password)

	}
}

// getContainerByID 根据容器ID获取容器信息
func (a *App) getContainerByID(deviceIP string, containerID string, password string) (Container, error) {
	log.Printf("获取容器信息: deviceIP=%s, containerID=%s", deviceIP, containerID)

	// 尝试从完整ID中提取索引号和容器名称
	// 前端生成的完整ID格式: deviceIP_deviceId_indexNum_containerName
	containerIndexNum := 0
	containerName := containerID
	parts := strings.Split(containerID, "_")
	if len(parts) >= 3 {
		// 尝试从倒数第二个部分提取索引号
		if idx, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
			containerIndexNum = idx
			log.Printf("从完整ID中提取索引号: %s -> indexNum=%d", containerID, containerIndexNum)
		}
		// 容器名称是最后一部分
		if len(parts) >= 4 {
			containerName = parts[len(parts)-1]
			log.Printf("从完整ID中提取容器名称: %s -> %s", containerID, containerName)
		}
	}
	if containerIndexNum == 0 {
		if idx, err := strconv.Atoi(containerID); err == nil {
			containerIndexNum = idx
			log.Printf("从纯数字ID中提取索引号: %s -> indexNum=%d", containerID, containerIndexNum)
		}
	}

	// 首先尝试获取所有容器
	containersData := a.GetContainers(deviceIP, "v3", password)

	// 查找指定ID的容器
	if containersMap, ok := containersData.(map[string]interface{}); ok {
		if codeVal, ok := containersMap["code"].(float64); ok {
			if int(codeVal) == 61 {
				if msg, ok := containersMap["message"].(string); ok && msg == "Authentication Failed" {
					return Container{}, fmt.Errorf("Authentication Failed")
				}
			}
		}
		var dataList []interface{}
		var respCode interface{}
		var respMessage string
		if codeVal, ok := containersMap["code"]; ok {
			respCode = codeVal
		}
		if msgVal, ok := containersMap["message"].(string); ok {
			respMessage = msgVal
		}
		if dataMap, ok := containersMap["data"].(map[string]interface{}); ok {
			if listVal, ok := dataMap["list"].([]interface{}); ok {
				dataList = listVal
			} else if listVal, ok := dataMap["List"].([]interface{}); ok {
				dataList = listVal
			}
		}
		if dataList == nil {
			if listVal, ok := containersMap["list"].([]interface{}); ok {
				dataList = listVal
			} else if listVal, ok := containersMap["List"].([]interface{}); ok {
				dataList = listVal
			}
		}
		if dataList != nil {
			for _, item := range dataList {
				if containerMap, ok := item.(map[string]interface{}); ok {
						id := ""
						if idVal, ok := containerMap["id"].(string); ok {
							id = idVal
						} else if idVal, ok := containerMap["Id"].(string); ok {
							id = idVal
						} else if idVal, ok := containerMap["ID"].(string); ok {
							id = idVal
						} else if idVal, ok := containerMap["containerId"].(string); ok {
							id = idVal
						} else if idVal, ok := containerMap["containerID"].(string); ok {
							id = idVal
						}

						// 优先使用indexNum匹配（最可靠）
						if indexNumVal, ok := containerMap["indexNum"].(float64); ok {
							if int(indexNumVal) == containerIndexNum && containerIndexNum > 0 {
								log.Printf("通过indexNum找到容器: indexNum=%d, id=%s", int(indexNumVal), id)
								var container Container
								container.ID = id
								if ipVal, ok := containerMap["ip"].(string); ok {
									container.IP = ipVal
								}
								if networkNameVal, ok := containerMap["networkName"].(string); ok {
									container.NetworkName = networkNameVal
								}
								if networkModeVal, ok := containerMap["networkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["NetworkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["network"].(string); ok {
									container.NetworkMode = networkModeVal
								}

								// 提取端口信息
								if portBindings, ok := containerMap["portBindings"].(map[string]interface{}); ok {
									for portKey, bindings := range portBindings {
										if bindingsList, ok := bindings.([]interface{}); ok {
											for _, binding := range bindingsList {
												if bindingMap, ok := binding.(map[string]interface{}); ok {
													var port Port
													if strings.HasSuffix(portKey, "/tcp") {
														privatePortStr := strings.TrimSuffix(portKey, "/tcp")
														if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
															port.PrivatePort = privatePort
														}
														port.Type = "tcp"
													} else if strings.HasSuffix(portKey, "/udp") {
														privatePortStr := strings.TrimSuffix(portKey, "/udp")
														if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
															port.PrivatePort = privatePort
														}
														port.Type = "udp"
													}
													if hostPort, ok := bindingMap["HostPort"].(string); ok {
														if publicPort, err := strconv.Atoi(hostPort); err == nil {
															port.PublicPort = publicPort
														}
													}
													container.Ports = append(container.Ports, port)
												}
											}
										}
									}
								}

								return container, nil
							}
						}

						// 备选：检查ID或名称是否匹配
						if id == containerID || id == containerName {
							log.Printf("通过ID或名称找到容器: id=%s", id)
							var container Container
							container.ID = id
								if ipVal, ok := containerMap["ip"].(string); ok {
									container.IP = ipVal
								}
								if networkNameVal, ok := containerMap["networkName"].(string); ok {
									container.NetworkName = networkNameVal
								}
								if networkModeVal, ok := containerMap["networkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["NetworkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["network"].(string); ok {
									container.NetworkMode = networkModeVal
								}

							// 提取端口信息
							if portBindings, ok := containerMap["portBindings"].(map[string]interface{}); ok {
								for portKey, bindings := range portBindings {
									if bindingsList, ok := bindings.([]interface{}); ok {
										for _, binding := range bindingsList {
											if bindingMap, ok := binding.(map[string]interface{}); ok {
												var port Port
												if strings.HasSuffix(portKey, "/tcp") {
													privatePortStr := strings.TrimSuffix(portKey, "/tcp")
													if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
														port.PrivatePort = privatePort
													}
													port.Type = "tcp"
												} else if strings.HasSuffix(portKey, "/udp") {
													privatePortStr := strings.TrimSuffix(portKey, "/udp")
													if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
														port.PrivatePort = privatePort
													}
													port.Type = "udp"
												}
												if hostPort, ok := bindingMap["HostPort"].(string); ok {
													if publicPort, err := strconv.Atoi(hostPort); err == nil {
														port.PublicPort = publicPort
													}
												}
												container.Ports = append(container.Ports, port)
											}
										}
									}
								}
							}

							return container, nil
						}

						if nameVal, ok := containerMap["name"].(string); ok {
							if nameVal == containerID || nameVal == containerName {
								log.Printf("通过name找到容器: name=%s, id=%s", nameVal, id)
								var container Container
								container.ID = id
								if ipVal, ok := containerMap["ip"].(string); ok {
									container.IP = ipVal
								}
								if networkNameVal, ok := containerMap["networkName"].(string); ok {
									container.NetworkName = networkNameVal
								}
								if networkModeVal, ok := containerMap["networkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["NetworkMode"].(string); ok {
									container.NetworkMode = networkModeVal
								} else if networkModeVal, ok := containerMap["network"].(string); ok {
									container.NetworkMode = networkModeVal
								}

								// 提取端口信息
								if portBindings, ok := containerMap["portBindings"].(map[string]interface{}); ok {
									for portKey, bindings := range portBindings {
										if bindingsList, ok := bindings.([]interface{}); ok {
											for _, binding := range bindingsList {
												if bindingMap, ok := binding.(map[string]interface{}); ok {
													var port Port
													if strings.HasSuffix(portKey, "/tcp") {
														privatePortStr := strings.TrimSuffix(portKey, "/tcp")
														if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
															port.PrivatePort = privatePort
														}
														port.Type = "tcp"
													} else if strings.HasSuffix(portKey, "/udp") {
														privatePortStr := strings.TrimSuffix(portKey, "/udp")
														if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
															port.PrivatePort = privatePort
														}
														port.Type = "udp"
													}
													if hostPort, ok := bindingMap["HostPort"].(string); ok {
														if publicPort, err := strconv.Atoi(hostPort); err == nil {
															port.PublicPort = publicPort
														}
													}
													container.Ports = append(container.Ports, port)
												}
											}
										}
									}
								}

								return container, nil
							}
						}
					}
			}
		} else {
			// 尝试识别认证失败
			if codeNum, ok := respCode.(float64); ok {
				if int(codeNum) == 61 {
					return Container{}, fmt.Errorf("Authentication Failed")
				}
			}
			if strings.Contains(strings.ToLower(respMessage), "auth") ||
				strings.Contains(strings.ToLower(respMessage), "unauthorized") ||
				strings.Contains(respMessage, "认证") {
				return Container{}, fmt.Errorf("Authentication Failed")
			}
			return Container{}, fmt.Errorf("GetContainers返回错误: code=%v message=%s", respCode, respMessage)
		}
	}

	return Container{}, fmt.Errorf("容器不存在: %s", containerID)
}

// ListSharedDirFiles 列出共享目录中的文件
func (a *App) ListSharedDirFiles() map[string]interface{} {
	log.Printf("[IPC] 收到 ListSharedDirFiles 调用")

	// 使用自定义或默认共享目录
	sharedPath := getCustomSharedDirPath()

	// 检查目录是否存在
	if _, err := os.Stat(sharedPath); os.IsNotExist(err) {
		log.Printf("共享目录不存在，创建目录: %v", err)
		if err := os.MkdirAll(sharedPath, 0755); err != nil {
			log.Printf("创建共享目录失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建共享目录失败: %v", err),
			}
		}
	}

	// 读取目录内容，构建目录树
	tree, err := buildDirectoryTree(sharedPath, sharedPath)
	if err != nil {
		log.Printf("读取共享目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取共享目录失败: %v", err),
		}
	}

	log.Printf("列出共享目录文件成功: %s", sharedPath)
	return map[string]interface{}{
		"success":  true,
		"message":  "列出文件成功",
		"tree":     tree,
		"rootPath": sharedPath,
	}
}

// DirectoryNode 目录树节点
type DirectoryNode struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	IsDir    bool            `json:"isDir"`
	Size     int64           `json:"size"`
	ModTime  int64           `json:"modTime"`
	Children []DirectoryNode `json:"children"`
}

// buildDirectoryTree 构建目录树
func buildDirectoryTree(rootPath, currentPath string) (DirectoryNode, error) {
	info, err := os.Stat(currentPath)
	if err != nil {
		return DirectoryNode{}, err
	}

	node := DirectoryNode{
		Name:     filepath.Base(currentPath),
		Path:     currentPath,
		IsDir:    info.IsDir(),
		Size:     info.Size(),
		ModTime:  info.ModTime().Unix(),
		Children: []DirectoryNode{},
	}

	if info.IsDir() {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return DirectoryNode{}, err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(currentPath, entry.Name())
			childNode, err := buildDirectoryTree(rootPath, entryPath)
			if err != nil {
				continue
			}
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// UploadFileToCloudMachine 上传文件到云机
func (a *App) UploadFileToCloudMachine(deviceIP string, version string, containerID string, filePath string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 UploadFileToCloudMachine 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerID=%s, filePath=%s", deviceIP, version, containerID, filePath)

	// 验证参数
	if deviceIP == "" || filePath == "" {
		return map[string]interface{}{
			"success": false,
			"message": "设备IP和文件路径不能为空",
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return map[string]interface{}{
			"success": false,
			"message": "文件不存在",
		}
	}

	// 获取容器信息以找到9082端口映射
	container, err := a.getContainerByID(deviceIP, containerID, password)
	if err != nil {
		log.Printf("获取容器信息失败: %v", err)
		if err.Error() == "Authentication Failed" {
			return map[string]interface{}{
				"success": false,
				"code":    61,
				"message": "Authentication Failed",
			}
		}
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取容器信息失败: %v", err),
		}
	}

	uploadHost := deviceIP
	uploadPort := 0
	if container.NetworkName == "myt" || container.NetworkMode == "myt" {
		if container.IP != "" {
			uploadHost = container.IP
			uploadPort = 9082
		}
	}

	// OpenCecs 公网设备：uploadHost 含端口时需提取纯 IP，uploadPort 需转为公网端口
	if uploadPort == 0 {
		for _, p := range container.Ports {
			if p.PrivatePort == 9082 {
				uploadPort = p.PublicPort
				break
			}
		}
	}

	if uploadPort == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "未找到9082端口映射",
		}
	}

	// OpenCecs 公网设备：将 uploadHost(含端口) + uploadPort(LAN端口) 转为 纯IP + 公网端口
	uploadHost, uploadPort = a.resolveOpenCecsAddr(uploadHost, uploadPort)

	// 上传文件到云机
	if err := uploadFileToContainer(uploadHost, uploadPort, filePath); err != nil {
		log.Printf("上传文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("上传文件失败: %v", err),
		}
	}

	log.Printf("文件上传成功: %s -> http://%s:%d/upload", filePath, uploadHost, uploadPort)
	return map[string]interface{}{
		"success":   true,
		"message":   "文件上传成功",
		"filePath":  filePath,
		"uploadUrl": fmt.Sprintf("http://%s:%d/upload", uploadHost, uploadPort),
	}
}

// UpgradeSDK 升级SDK
func (a *App) UpgradeSDK(deviceIP string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 UpgradeSDK 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, password=%s", deviceIP, password)

	// 构造SDK升级URL
	url := fmt.Sprintf("http://%s/server/upgrade", deviceAddr(deviceIP))

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头
	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 60 * time.Second, // 增加超时时间，因为SDK升级可能需要较长时间
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("调用SDK升级API失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("调用SDK升级API失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 读取响应内容以获取更多信息
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SDK升级API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
		return map[string]interface{}{
			"success":    false,
			"message":    fmt.Sprintf("SDK升级API返回错误: 状态码 %d", resp.StatusCode),
			"statusCode": resp.StatusCode,
			"response":   string(body),
		}
	}

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取SDK升级API响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取SDK升级API响应失败: %v", err),
		}
	}

	log.Printf("SDK升级API响应: %s", string(body))

	// 处理Server-Sent Events (SSE)格式的响应
	lines := strings.Split(string(body), "\n")
	var completeData string
	var event string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data != "" {
				if event == "complete" {
					completeData = data
					break
				}
			}
		}
	}

	// 解析完整事件的数据
	if completeData != "" {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(completeData), &result); err != nil {
			log.Printf("解析SDK升级完成事件失败: %v, 数据: %s", err, completeData)
			return map[string]interface{}{
				"success":  false,
				"message":  fmt.Sprintf("解析SDK升级完成事件失败: %v", err),
				"response": string(body),
			}
		}

		// 检查升级是否成功
		if msg, ok := result["msg"].(string); ok && msg == "success" {
			return map[string]interface{}{
				"success": true,
				"message": "SDK升级成功",
				"data":    result,
			}
		}

		return map[string]interface{}{
			"success": true,
			"message": "SDK升级请求成功，正在进行升级...",
			"data":    result,
		}
	}

	// 尝试直接解析响应体（兼容旧格式）
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("解析SDK升级API响应失败: %v, 响应内容: %s", err, string(body))
		// 即使解析失败，如果响应包含"success"，也认为升级成功
		if strings.Contains(string(body), "success") {
			return map[string]interface{}{
				"success": true,
				"message": "SDK升级成功",
			}
		}
		return map[string]interface{}{
			"success":  false,
			"message":  fmt.Sprintf("解析SDK升级API响应失败: %v", err),
			"response": string(body),
		}
	}

	// 检查V3 API响应格式
	if code, ok := result["code"].(float64); ok {
		if code != 0 {
			message := "未知错误"
			if msg, ok := result["message"].(string); ok {
				message = msg
			}
			log.Printf("SDK升级API返回错误: code=%f, message=%s", code, message)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("SDK升级失败: %s", message),
				"code":    code,
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "SDK升级请求成功，正在进行升级...",
		"data":    result,
	}
}

// UpgradeDevice 升级设备
func (a *App) UpgradeDevice(deviceIP string, version string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 UpgradeDevice 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, password=%s", deviceIP, version, password)

	// 直接调用现有的UpgradeSDK方法，忽略version参数（因为升级逻辑对所有版本相同）
	return a.UpgradeSDK(deviceIP, password)
}

// DockerExecResult docker exec执行结果
type DockerExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// sanitizeString 清理字符串中的无效UTF-8字符和控制字符
func sanitizeString(s string) string {
	return strings.ToValidUTF8(strings.Map(func(r rune) rune {
		// 保留换行符和制表符，移除其他控制字符
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s), "")
}

// dockerExec 执行docker容器内命令
func dockerExec(deviceIP string, dockerPort int, containerID string, cmd []string, password string, version string) (DockerExecResult, error) {
	// 构造docker exec API URL
	var url string
	// 使用 deviceAddr 处理已含端口的设备IP（如 OpenCecs 公网设备）
	baseAddr := deviceAddr(deviceIP)
	if version == "v3" && dockerPort == 8000 {
		// V3设备使用8000端口的docker端点
		url = fmt.Sprintf("http://%s/docker/containers/%s/exec", baseAddr, containerID)
	} else {
		// 标准Docker API
		url = fmt.Sprintf("http://%s:%d/containers/%s/exec", deviceIP, dockerPort, containerID)
	}

	// 准备exec创建请求
	execRequest := map[string]interface{}{
		"AttachStdout": true,
		"AttachStderr": true,
		"Tty":          true,
		"Cmd":          cmd,
	}

	// 序列化请求体
	execReqBody, err := json.Marshal(execRequest)
	if err != nil {
		return DockerExecResult{}, err
	}

	// 创建exec
	client := &http.Client{}
	execReq, err := http.NewRequest("POST", url, bytes.NewBuffer(execReqBody))
	if err != nil {
		return DockerExecResult{}, err
	}

	// 添加请求头
	execReq.Header.Set("Content-Type", "application/json")

	// 添加认证头（所有版本都需要）
	if password != "" {
		execReq.SetBasicAuth("admin", password)
	}

	resp, err := client.Do(execReq)
	if err != nil {
		return DockerExecResult{}, err
	}
	defer resp.Body.Close()

	// 读取exec创建响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return DockerExecResult{}, err
	}

	// 检查响应状态码
	if resp.StatusCode == http.StatusNotFound {
		return DockerExecResult{}, fmt.Errorf("容器未找到: %s", string(respBody))
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return DockerExecResult{}, fmt.Errorf("创建exec失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 检查是否是"Not Found"响应
	responseStr := string(respBody)
	if strings.Contains(strings.ToLower(responseStr), "not found") {
		return DockerExecResult{}, fmt.Errorf("容器未找到: %s", responseStr)
	}

	// 解析exec创建响应
	var execResp struct {
		ID string `json:"Id"`
	}

	// 尝试解析标准Docker API响应
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		// 尝试解析V3 API响应格式
		var v3Resp struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}

		if err := json.Unmarshal(respBody, &v3Resp); err != nil {
			// 直接返回错误，包含响应内容以便调试
			return DockerExecResult{}, fmt.Errorf("解析exec响应失败: %v, 响应内容: %s", err, string(respBody))
		}

		// 检查V3 API响应
		if v3Resp.Code != 0 {
			return DockerExecResult{}, fmt.Errorf("V3 API错误: %s", v3Resp.Message)
		}

		// 尝试从data中提取ID
		if dataMap, ok := v3Resp.Data.(map[string]interface{}); ok {
			if id, ok := dataMap["Id"].(string); ok {
				execResp.ID = id
			} else if id, ok := dataMap["id"].(string); ok {
				execResp.ID = id
			} else {
				return DockerExecResult{}, fmt.Errorf("V3 API响应中未找到Id字段")
			}
		} else {
			return DockerExecResult{}, fmt.Errorf("V3 API响应data格式错误")
		}
	}

	// 启动exec
	var startUrl string
	if version == "v3" && dockerPort == 8000 {
		// V3设备使用8000端口的docker端点
		startUrl = fmt.Sprintf("http://%s/docker/exec/%s/start", baseAddr, execResp.ID)
	} else {
		// 标准Docker API
		startUrl = fmt.Sprintf("http://%s:%d/exec/%s/start", deviceIP, dockerPort, execResp.ID)
	}
	startReq, err := http.NewRequest("POST", startUrl, nil)
	if err != nil {
		return DockerExecResult{}, err
	}

	startReq.Header.Set("Content-Type", "application/json")
	startReq.Body = io.NopCloser(strings.NewReader(`{"Detach":false,"Tty":false}`))

	// 添加认证头（所有版本都需要）
	if password != "" {
		startReq.SetBasicAuth("admin", password)
	}

	startResp, err := client.Do(startReq)
	if err != nil {
		return DockerExecResult{}, err
	}
	defer startResp.Body.Close()

	// 读取执行结果
	output, err := io.ReadAll(startResp.Body)
	if err != nil {
		return DockerExecResult{}, err
	}

	// Docker API 返回的是 multiplexed stream 格式
	// 每个消息块的格式：[stream_type(1字节), 0, 0, 0, size(4字节), payload]
	// 我们需要跳过这些头部，只提取实际内容
	cleanedOutput := stripDockerStreamHeaders(output)

	return DockerExecResult{
		ExitCode: 0,
		Stdout:   cleanedOutput,
		Stderr:   "",
	}, nil
}

// stripDockerStreamHeaders 移除 Docker multiplexed stream 的头部并清理输出
func stripDockerStreamHeaders(data []byte) string {
	var result []byte
	
	// Docker stream 格式: [stream_type(1), 0, 0, 0, size(4), payload...]
	i := 0
	for i < len(data) {
		// 检查是否还有足够的字节读取头部
		if i+8 > len(data) {
			// 剩余数据不足8字节，可能是末尾数据，直接添加
			result = append(result, data[i:]...)
			break
		}
		
		// 读取 stream type (第1字节)
		// streamType := data[i]
		
		// 读取 payload size (第5-8字节，大端序)
		payloadSize := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])
		
		// 跳过 8 字节头部
		i += 8
		
		// 检查是否有足够的数据
		if i+payloadSize > len(data) {
			// payload 大小超出剩余数据，可能格式不对，保留原始数据
			result = append(result, data[i:]...)
			break
		}
		
		// 提取 payload
		if payloadSize > 0 {
			result = append(result, data[i:i+payloadSize]...)
			i += payloadSize
		}
	}
	
	// 如果没有提取到任何数据，可能不是 Docker stream 格式，返回原始数据
	if len(result) == 0 {
		result = data
	}
	
	// 清理 UTF-8 字符
	return sanitizeString(string(result))
}

// InstallAPK 安装APK文件到云机
func (a *App) InstallAPK(deviceIP string, version string, containerID string, filePath string, password string, options ...APKInstallOptions) map[string]interface{} {
	log.Printf("[IPC] 收到 InstallAPK 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerID=%s, filePath=%s", deviceIP, version, containerID, filePath)

	// 解析安装选项（如果提供）
	var opts APKInstallOptions
	if len(options) > 0 {
		opts = options[0]
		log.Printf("[InstallAPK] 安装选项: replace=%v, test=%v, grant=%v, deleteAfterInstall=%v", 
			opts.Replace, opts.Test, opts.Grant, opts.DeleteAfterInstall)
	} else {
		// 默认选项：替换、测试、授权（保持向后兼容）
		opts = APKInstallOptions{
			Replace: true,
			Test:    true,
			Grant:   true,
			DeleteAfterInstall: false,
		}
		log.Printf("[InstallAPK] 使用默认安装选项")
	}

	// 验证参数
	if deviceIP == "" || filePath == "" {
		return map[string]interface{}{
			"success": false,
			"message": "设备IP和文件路径不能为空",
		}
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("[InstallAPK] 文件不存在: %s", filePath)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件不存在: %s", filePath),
		}
	}
	if err != nil {
		log.Printf("[InstallAPK] 无法访问文件: %s, 错误: %v", filePath, err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("无法访问文件: %v", err),
		}
	}
	log.Printf("[InstallAPK] 文件存在，大小: %d bytes, 文件名: %s", fileInfo.Size(), filepath.Base(filePath))

	// 获取容器信息以找到9082端口映射
	container, err := a.getContainerByID(deviceIP, containerID, password)
	if err != nil {
		log.Printf("获取容器信息失败: %v", err)
		if err.Error() == "Authentication Failed" {
			return map[string]interface{}{
				"success": false,
				"code":    61,
				"message": "Authentication Failed",
			}
		}
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取容器信息失败: %v", err),
		}
	}

	uploadHost := deviceIP
	uploadPort := 0
	if container.NetworkName == "myt" || container.NetworkMode == "myt" {
		if container.IP != "" {
			uploadHost = container.IP
			uploadPort = 9082
		}
	}

	if uploadPort == 0 {
		for _, p := range container.Ports {
			if p.PrivatePort == 9082 {
				uploadPort = p.PublicPort
				break
			}
		}
	}

	if uploadPort == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "未找到9082端口映射",
		}
	}

	// OpenCecs 公网设备：将 uploadHost(含端口) + uploadPort(LAN端口) 转为 纯IP + 公网端口
	uploadHost, uploadPort = a.resolveOpenCecsAddr(uploadHost, uploadPort)

	// 找到docker API端口
	dockerPort := 2375
	if version == "v3" {
		// 对于V3设备，优先使用8000端口的docker接口
		dockerPort = 8000
	}

	// 上传文件到云机
	if err := uploadFileToContainer(uploadHost, uploadPort, filePath); err != nil {
		log.Printf("上传文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("上传文件失败: %v", err),
		}
	}


	// 准备容器内路径
	fileName := filepath.Base(filePath)
	uploadPath := "/sdcard/upload/" + fileName
	installPath := "/data/local/tmp/" + fileName


	// 只复制文件到临时目录，不删除源文件
	commands := []struct {
		Cmd     string
		Message string
	}{
		{fmt.Sprintf("cp '%s' '%s'", uploadPath, installPath), "复制文件到临时目录..."},
	}

	for _, step := range commands {
		log.Printf("执行命令: %s %s", step.Cmd, container.ID)

		result, err := dockerExec(deviceIP, dockerPort, container.ID, []string{"sh", "-c", step.Cmd}, password, version)
		if err != nil {
			log.Printf("执行命令失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("执行命令失败: %v", err),
			}
		}

		if result.ExitCode != 0 {
			log.Printf("命令执行失败，退出码: %d, 输出: %s, 错误: %s", result.ExitCode, result.Stdout, result.Stderr)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("命令执行失败: %s, 退出码: %d", step.Cmd, result.ExitCode),
				"stdout":  result.Stdout,
				"stderr":  result.Stderr,
			}
		}

		log.Printf("命令执行成功: %s", step.Message)
		if result.Stdout != "" {
			log.Printf("命令输出: %s", result.Stdout)
		}
		if result.Stderr != "" {
			log.Printf("命令错误输出: %s", result.Stderr)
		}
	}

	// ========== 检测文件类型，如果是APK则自动安装 ==========
	isAPK := strings.HasSuffix(strings.ToLower(fileName), ".apk")
	
	if isAPK {
		log.Printf("[InstallAPK] 检测到APK文件，开始自动安装: %s", fileName)
		
		// 构建安装命令选项
		var installFlags []string
		if opts.Replace {
			installFlags = append(installFlags, "-r") // 替换已有应用
		}
		if opts.Test {
			installFlags = append(installFlags, "-t") // 允许测试应用
		}
		if opts.Grant {
			installFlags = append(installFlags, "-g") // 自动授权
		}
		
		installCmd := fmt.Sprintf("pm install %s %s", strings.Join(installFlags, " "), installPath)
		log.Printf("[InstallAPK] 执行安装命令: %s", installCmd)
		
		installResult, err := dockerExec(
			deviceIP, 
			dockerPort, 
			container.ID, 
			[]string{"sh", "-c", fmt.Sprintf("sd -c '%s'", installCmd)},
			password, 
			version,
		)
		
		if err != nil || installResult.ExitCode != 0 {
			log.Printf("[InstallAPK] 安装失败: exitCode=%d, stderr=%s", installResult.ExitCode, installResult.Stderr)
			
			// 清理临时文件
			dockerExec(deviceIP, dockerPort, container.ID, 
				[]string{"sh", "-c", fmt.Sprintf("sd -c 'rm -f %s'", installPath)}, 
				password, version)
			
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("APK安装失败: %s", installResult.Stderr),
				"uploadPath":  uploadPath,
				"installPath": installPath,
			}
		}
		
		log.Printf("[InstallAPK] APK安装成功: %s", fileName)
		
		// 清理临时安装文件（保留上传目录的文件）
		dockerExec(deviceIP, dockerPort, container.ID, 
			[]string{"sh", "-c", fmt.Sprintf("sd -c 'rm -f %s'", installPath)}, 
			password, version)
		
		// 如果设置了安装后删除，删除上传目录的文件
		if opts.DeleteAfterInstall {
			log.Printf("[InstallAPK] 删除安装成功的APK文件: %s", uploadPath)
			dockerExec(deviceIP, dockerPort, container.ID, 
				[]string{"sh", "-c", fmt.Sprintf("sd -c 'rm -f %s'", uploadPath)}, 
				password, version)
		}
		
		// 返回安装成功信息
		return map[string]interface{}{
			"success":     true,
			"message":     "APK上传并安装成功",
			"uploadPath":  uploadPath,  // 上传目录路径 /sdcard/upload/
			"fileName":    fileName,
			"localPath":   filePath,
			"installed":   true,  // 标记已安装
			"deleted":     opts.DeleteAfterInstall,  // 是否删除了文件
		}
	}

	log.Printf("文件上传并复制完成: %s (源文件保留在: %s)", installPath, uploadPath)
	return map[string]interface{}{
		"success":     true,
		"message":     "文件上传完成",
		"uploadPath":  uploadPath,  // 上传目录路径 /sdcard/upload/
		"installPath": installPath, // 临时安装路径 /data/local/tmp/
		"fileName":    fileName,
		"localPath":   filePath, // 本地文件路径
		"installed":   false,  // 标记未安装（非APK文件）
	}
}

// APKInstallTarget 批量安装APK的目标信息
type APKInstallTarget struct {
	DeviceIP      string   `json:"deviceIP"`
	DeviceVersion string   `json:"deviceVersion"`
	ContainerID   string   `json:"containerID"`
	ContainerName string   `json:"containerName"`
	Password      string   `json:"password"`
}

// APKInstallOptions 安装选项
type APKInstallOptions struct {
	Replace           bool `json:"replace"`            // -r 替换已存在的应用
	Test              bool `json:"test"`               // -t 允许测试包
	Grant             bool `json:"grant"`              // -g 授予所有运行时权限
	DeleteAfterInstall bool `json:"deleteAfterInstall"` // 安装后删除上传的文件
}

// APKInstallResult 单个设备的安装结果
type APKInstallResult struct {
	DeviceIP      string `json:"deviceIP"`
	ContainerName string `json:"containerName"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Duration      int64  `json:"duration"` // 毫秒
}

// BatchInstallAPK 批量安装APK（并发执行）
func (a *App) BatchInstallAPK(targets []APKInstallTarget, filePath string, options APKInstallOptions) map[string]interface{} {
	log.Printf("[BatchInstallAPK] 收到批量安装请求，目标数量: %d, 文件: %s", len(targets), filePath)
	
	// 验证文件
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件不存在: %s", filePath),
		}
	}
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("无法访问文件: %v", err),
		}
	}
	log.Printf("[BatchInstallAPK] 文件验证通过，大小: %d bytes", fileInfo.Size())
	
	// 并发安装（限制并发数）
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	results := make([]APKInstallResult, 0, len(targets))
	
	for i, target := range targets {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		
		go func(index int, t APKInstallTarget) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			startTime := time.Now()
			log.Printf("[BatchInstallAPK] [%d/%d] 开始处理: %s (%s)", index+1, len(targets), t.DeviceIP, t.ContainerName)
			
			result := APKInstallResult{
				DeviceIP:      t.DeviceIP,
				ContainerName: t.ContainerName,
				Success:       false,
			}
			
			// 调用 InstallAPK（已包含上传+安装逻辑）
			// 注意: InstallAPK 会自动检测 .apk 文件并安装
			uploadResult := a.InstallAPK(t.DeviceIP, t.DeviceVersion, t.ContainerID, filePath, t.Password)
			
			if !uploadResult["success"].(bool) {
				result.Message = fmt.Sprintf("安装失败: %v", uploadResult["message"])
				result.Duration = time.Since(startTime).Milliseconds()
				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()
				log.Printf("[BatchInstallAPK] [%d/%d] 安装失败: %s - %s", index+1, len(targets), t.DeviceIP, result.Message)
				return
			}
			
			// 检查是否已安装（InstallAPK 自动安装了）
			installed, _ := uploadResult["installed"].(bool)
			if !installed {
				// 不是 APK 文件，应该不会发生
				result.Message = "文件不是 APK 格式"
				result.Duration = time.Since(startTime).Milliseconds()
				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()
				log.Printf("[BatchInstallAPK] [%d/%d] 错误: %s 不是APK文件", index+1, len(targets), t.DeviceIP)
				return
			}
			
			log.Printf("[BatchInstallAPK] [%d/%d] 安装成功: %s", index+1, len(targets), t.DeviceIP)
			
			// 可选: 删除上传目录中的文件
			if options.DeleteAfterInstall {
				uploadPath, ok := uploadResult["uploadPath"].(string)
				if ok && uploadPath != "" {
					dockerPort := 2375
					if t.DeviceVersion == "v3" {
						dockerPort = 8000
					}
					
					container, err := a.getContainerByID(t.DeviceIP, t.ContainerID, t.Password)
					if err == nil {
						// 删除上传目录的文件
						dockerExec(t.DeviceIP, dockerPort, container.ID, 
							[]string{"sh", "-c", fmt.Sprintf("sd -c 'rm -f %s'", uploadPath)}, 
							t.Password, t.DeviceVersion)
						log.Printf("[BatchInstallAPK] [%d/%d] 已删除上传文件: %s", index+1, len(targets), uploadPath)
					}
				}
			}
			
			result.Success = true
			result.Message = "安装成功"
			result.Duration = time.Since(startTime).Milliseconds()
			
			resultsMutex.Lock()
			results = append(results, result)
			resultsMutex.Unlock()
		}(i, target)
	}
	
	// 等待所有任务完成
	wg.Wait()
	
	// 统计结果
	successCount := 0
	failedCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failedCount++
		}
	}
	
	log.Printf("[BatchInstallAPK] 批量安装完成，成功: %d, 失败: %d", successCount, failedCount)
	
	return map[string]interface{}{
		"success":      true,
		"results":      results,
		"successCount": successCount,
		"failedCount":  failedCount,
		"total":        len(targets),
	}
}

// uploadFileToContainer 上传文件到容器
func uploadFileToContainer(ip string, port int, filePath string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	// 拷贝文件内容
	if _, err = io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	// 创建请求
	url := fmt.Sprintf("http://%s:%d/upload", ip, port)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{Timeout: 7200 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务器错误: %s", string(respBody))
	}

	return nil
}

// OpenImageCacheDirectory 打开镜像缓存目录
func (a *App) OpenImageCacheDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenImageCacheDirectory 调用")

	// 确定镜像缓存目录
	cacheDir := getCacheDir()
	imageDir := filepath.Join(cacheDir, "images")

	// 创建目标目录
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建镜像目录失败: %v", err),
		}
	}

	// 打开目录
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", imageDir)
	case "darwin":
		cmd = exec.Command("open", imageDir)
	case "linux":
		cmd = exec.Command("xdg-open", imageDir)
	default:
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("不支持的操作系统: %s", runtime.GOOS),
		}
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
	}
}

// OpenLocalImageDirectory 打开本地镜像目录
func (a *App) OpenLocalImageDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenLocalImageDirectory 调用")

	imagePath := getStorageBaseDir()

	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建本地镜像目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", imagePath)
	case "darwin":
		cmd = exec.Command("open", imagePath)
	default:
		cmd = exec.Command("xdg-open", imagePath)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("打开本地镜像目录成功: %s", imagePath)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    imagePath,
	}
}

// SelectImageFile 选择图片文件（支持 jpg, jpeg, png）
func (a *App) SelectImageFile() map[string]interface{} {
	log.Printf("[IPC] 收到 SelectImageFile 调用")

	if a.wailsApp == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Wails app not initialized",
		}
	}

	selection, err := a.wailsApp.Dialog.OpenFile().
		SetTitle("图片文件").
		AddFilter("图片文件 (jpg,jpeg,png)", "*.jpg;*.jpeg;*.png").
		PromptForSingleSelection()

	if err != nil {
		log.Printf("打开文件选择对话框失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件选择对话框失败: %v", err),
		}
	}

	if selection == "" {
		log.Printf("用户取消选择文件")
		return map[string]interface{}{
			"success": false,
			"message": "用户取消选择",
		}
	}

	log.Printf("选择文件成功: %s", selection)
	return map[string]interface{}{
		"success": true,
		"message": "文件选择成功",
		"path":    selection,
	}
}

// SelectVideoFile 选择视频文件
func (a *App) SelectVideoFile() map[string]interface{} {
	log.Printf("[IPC] 收到 SelectVideoFile 调用")

	if a.wailsApp == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Wails app not initialized",
		}
	}

	selection, err := a.wailsApp.Dialog.OpenFile().
		SetTitle("视频文件").
		AddFilter("视频文件 (mp4,avi,mov,mkv)", "*.mp4;*.avi;*.mov;*.mkv").
		PromptForSingleSelection()

	if err != nil {
		log.Printf("打开文件选择对话框失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件选择对话框失败: %v", err),
		}
	}

	if selection == "" {
		log.Printf("用户取消选择文件")
		return map[string]interface{}{
			"success": false,
			"message": "用户取消选择",
		}
	}

	log.Printf("选择文件成功: %s", selection)
	return map[string]interface{}{
		"success": true,
		"message": "文件选择成功",
		"path":    selection,
	}
}

// SelectZipFile 选择ZIP文件
func (a *App) SelectZipFile() map[string]interface{} {
	log.Printf("[IPC] 收到 SelectZipFile 调用")

	if a.wailsApp == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Wails app not initialized",
		}
	}

	selection, err := a.wailsApp.Dialog.OpenFile().
		SetTitle("选择模型文件").
		AddFilter("ZIP文件 (*.zip)", "*.zip").
		PromptForSingleSelection()

	if err != nil {
		// 检查是否是用户取消操作
		if err.Error() == "cancelled by user" {
			log.Printf("用户取消选择文件")
			return map[string]interface{}{
				"success": false,
				"message": "用户取消选择",
			}
		}
		log.Printf("打开文件选择对话框失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件选择对话框失败: %v", err),
		}
	}

	if selection == "" {
		log.Printf("未选择文件")
		return map[string]interface{}{
			"success": false,
			"message": "未选择文件",
		}
	}

	log.Printf("选择文件成功: %s", selection)
	return map[string]interface{}{
		"success": true,
		"message": "文件选择成功",
		"path":    selection,
	}
}

// SelectApkFile 选择APK文件
func (a *App) SelectApkFile() map[string]interface{} {
	log.Printf("[IPC] 收到 SelectApkFile 调用（多文件选择）")

	if a.wailsApp == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Wails app not initialized",
		}
	}

	// 修改为支持多文件选择
	selections, err := a.wailsApp.Dialog.OpenFile().
		SetTitle("选择 APK 文件（可多选）").
		AddFilter("APK文件 (*.apk)", "*.apk").
		PromptForMultipleSelection()

	if err != nil {
		// 检查是否是用户取消操作
		if err.Error() == "cancelled by user" {
			log.Printf("用户取消选择文件")
			return map[string]interface{}{
				"success": false,
				"message": "用户取消选择",
			}
		}
		log.Printf("打开文件选择对话框失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件选择对话框失败: %v", err),
		}
	}

	if len(selections) == 0 {
		log.Printf("未选择文件")
		return map[string]interface{}{
			"success": false,
			"message": "未选择文件",
		}
	}

	// 返回多文件列表，包含文件信息
	type FileInfo struct {
		Path string `json:"path"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	
	fileList := make([]FileInfo, 0, len(selections))
	for _, filePath := range selections {
		stat, err := os.Stat(filePath)
		if err != nil {
			log.Printf("无法获取文件信息: %s, 错误: %v", filePath, err)
			continue
		}
		
		fileList = append(fileList, FileInfo{
			Path: filePath,
			Name: filepath.Base(filePath),
			Size: stat.Size(),
		})
	}

	log.Printf("选择APK文件成功，共 %d 个文件", len(fileList))
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("成功选择 %d 个文件", len(fileList)),
		"files":   fileList,
		"count":   len(fileList),
	}
}

// ConvertBase64ToBytes Base64 解码为字节数组
func ConvertBase64ToBytes(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// OpenSharedDirectory 打开共享目录
func (a *App) OpenSharedDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenSharedDirectory 调用")

	// 使用自定义或默认共享目录
	sharedPath := getCustomSharedDirPath()

	// 创建目标目录
	if err := os.MkdirAll(sharedPath, 0755); err != nil {
		log.Printf("创建共享目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建共享目录失败: %v", err),
		}
	}

	// 打开目录，根据不同平台使用不同的命令
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", sharedPath)
	case "darwin":
		cmd = exec.Command("open", sharedPath)
	default:
		cmd = exec.Command("xdg-open", sharedPath)
	}

	if err := cmd.Start(); err != nil {
		log.Printf("打开共享目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开共享目录失败: %v", err),
		}
	}

	log.Printf("打开共享目录成功: %s", sharedPath)
	return map[string]interface{}{
		"success": true,
		"message": "打开共享目录成功",
		"path":    sharedPath,
	}
}

// UploadFileToSharedDir 上传文件到共享目录
func (a *App) UploadFileToSharedDir(deviceIP string, version string, containerID string, filePath string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 UploadFileToSharedDir 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerID=%s, filePath=%s", deviceIP, version, containerID, filePath)

	// 验证参数
	if deviceIP == "" || filePath == "" {
		return map[string]interface{}{
			"success": false,
			"message": "设备IP和文件路径不能为空",
		}
	}

	// 检查源文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return map[string]interface{}{
			"success": false,
			"message": "源文件不存在",
		}
	}

	// 确定共享目录
	sharedDir := "edgeclient/shared"
	userDataDir, err := os.UserCacheDir()
	if err != nil {
		userDataDir = os.TempDir()
	}
	sharedPath := filepath.Join(userDataDir, sharedDir)

	// 创建目标目录
	if err := os.MkdirAll(sharedPath, 0755); err != nil {
		log.Printf("创建共享目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建共享目录失败: %v", err),
		}
	}

	// 获取文件名
	fileName := filepath.Base(filePath)
	destPath := filepath.Join(sharedPath, fileName)

	// 复制文件
	if err := copyFile(filePath, destPath); err != nil {
		log.Printf("复制文件到共享目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("复制文件失败: %v", err),
		}
	}

	log.Printf("文件上传到共享目录成功: %s -> %s", filePath, destPath)
	return map[string]interface{}{
		"success": true,
		"message": "文件上传成功",
		"path":    destPath,
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// getV3Containers 获取V3设备的容器列表
func getV3Containers(deviceIP string, password string) (interface{}, error) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/android", deviceAddr(deviceIP))

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加认证头
	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return map[string]interface{}{
			"code":    61,
			"message": "Authentication Failed",
		}, nil
	}

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取V3 API响应失败: %v", err)
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析V3 API响应失败: %v", err)
	}

	// log.Printf("V3 API响应: %s", string(body))
	return result, nil
}

// callDockerAPI 调用Docker API，优先使用v3设备的8000端口docker端点
func callDockerAPI(deviceIP string, version string, endpoint string, method string, data []byte, password string) (*http.Response, error) {
	var url string
	var req *http.Request
	var err error

	// 对于v3设备，优先使用8000端口的docker端点
	if version == "v3" {
		url = fmt.Sprintf("http://%s/docker%s", deviceAddr(deviceIP), endpoint)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		// 添加认证头
		if password != "" {
			req.SetBasicAuth("admin", password)
		}

		// 发送请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// 如果8000端口失败，回退到2375端口
		log.Printf("v3 docker端点调用失败，回退到2375端口: %v", err)
	}

	// 使用2375端口的标准Docker API
	url = fmt.Sprintf("http://%s:2375%s", deviceIP, endpoint)
	req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// callDockerAPIWithReader 调用Docker API并支持流式上传
func callDockerAPIWithReader(ctx context.Context, deviceIP string, version string, endpoint string, method string, reader io.Reader, contentType string, password string) (*http.Response, error) {
	var url string
	var req *http.Request
	var err error

	// 对于v3设备，使用8000端口的docker端点
	if version == "v3" {
		// 对于SDK版本>=21的V3设备，强制使用8000端口
		// 目前没有直接获取SDK版本的参数，根据用户需求，所有V3设备上传镜像时都优先使用8000端口
		url = fmt.Sprintf("http://%s/docker%s", deviceAddr(deviceIP), endpoint)
		req, err = http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return nil, err
		}

		// 添加认证头
		if password != "" {
			req.SetBasicAuth("admin", password)
		}

		// 添加内容类型
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		// 发送请求（无超时，因为文件上传可能需要很长时间）
		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil {
			// 对于8000端口，即使返回500错误也不回退到2375端口，因为SDK>=21的设备可能没有开放2375端口
			return resp, nil
		}

		// 如果8000端口连接失败且有密码，尝试无认证的8000端口
		if err != nil && password != "" {
			log.Printf("带认证的8000端口调用失败，尝试无认证: %v", err)
			req, err = http.NewRequestWithContext(ctx, method, url, reader)
			if err != nil {
				return nil, err
			}

			// 不添加认证头
			// 添加内容类型
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}

			// 发送请求
			resp, err = client.Do(req)
			if err == nil {
				return resp, nil
			}
		}

		// 如果所有8000端口的尝试都失败，回退到2375端口
		log.Printf("v3 docker端点调用失败，回退到2375端口: %v", err)
	}

	// 使用2375端口的标准Docker API
	url = fmt.Sprintf("http://%s:2375%s", deviceIP, endpoint)
	req, err = http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	// 添加内容类型
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 发送请求（无超时，因为文件上传可能需要很长时间）
	client := &http.Client{}
	return client.Do(req)
}

// getDockerContainers 获取Docker容器列表
func getDockerContainers(deviceIP string, version string, password string) ([]DockerContainer, error) {
	resp, err := callDockerAPI(deviceIP, version, "/containers/json?all=true", "GET", nil, password)
	if err != nil {
		return nil, fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	// 解析JSON响应
	var containers []DockerContainer
	if err := json.Unmarshal(body, &containers); err != nil {
		return nil, fmt.Errorf("解析Docker API响应失败: %v", err)
	}

	// log.Printf("Docker API响应: %s", string(body))
	return containers, nil
}

// StartContainer 启动容器
func (a *App) StartContainer(deviceIP string, version string, containerId string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 StartContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerId=%s", deviceIP, version, containerId)

	// 调用真实的容器启动逻辑
	success, message := startContainer(deviceIP, version, containerId, password)

	result := map[string]interface{}{
		"success": success,
		"message": message,
	}
	log.Printf("[IPC] StartContainer 返回结果: %+v", result)
	return result
}

// GetLLMModelList 获取设备上的LLM模型列表
func (a *App) GetLLMModelList(deviceIP string, token string) map[string]interface{} {
	log.Printf("[GetLLMModelList] 获取模型列表: deviceIP=%s", deviceIP)

	// 构建URL
	url := fmt.Sprintf("http://%s/lm/local", deviceAddr(deviceIP))
	log.Printf("[GetLLMModelList] 请求URL: %s", url)

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[GetLLMModelList] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", token)
		log.Printf("[GetLLMModelList] 已添加认证头")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GetLLMModelList] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GetLLMModelList] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[GetLLMModelList] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	// 检查code字段
	if code, ok := result["code"]; ok {
		if codeInt, ok := code.(float64); ok && codeInt == 0 {
			result["success"] = true
		} else {
			result["success"] = false
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
	}

	return result
}

// LoadLLMModel 加载LLM模型
func (a *App) LoadLLMModel(deviceIP string, modelPath string, weightPath string, token string) map[string]interface{} {
	log.Printf("[LoadLLMModel] 加载模型: deviceIP=%s, modelPath=%s, weightPath=%s", deviceIP, modelPath, weightPath)

	// 构建URL
	url := fmt.Sprintf("http://%s/llm/model/load", deviceAddr(deviceIP))
	log.Printf("[LoadLLMModel] 请求URL: %s", url)

	// 构建请求体
	requestBody := map[string]string{
		"model_path":  modelPath,
		"weight_path": weightPath,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[LoadLLMModel] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[LoadLLMModel] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", token)
		log.Printf("[LoadLLMModel] 已添加认证头")
	}

	// 发送请求（加载模型可能需要较长时间）
	client := &http.Client{
		Timeout: 120 * time.Second, // 2分钟超时
	}

	log.Printf("[LoadLLMModel] 开始发送请求...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[LoadLLMModel] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[LoadLLMModel] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[LoadLLMModel] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return map[string]interface{}{
				"success": true,
				"message": "模型加载成功",
				"code":    0,
			}
		}
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("服务器返回错误 (状态码: %d): %s", resp.StatusCode, string(respBody)),
			"code":    resp.StatusCode,
		}
	}

	// 检查code字段
	if code, ok := result["code"]; ok {
		if codeInt, ok := code.(float64); ok && codeInt == 0 {
			result["success"] = true
		} else {
			result["success"] = false
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
	}

	return result
}

// InitLLMSession 初始化LLM会话
func (a *App) InitLLMSession(deviceIP string, embeddingPath string, tokenizerPath string, systemPrompt string, token string) map[string]interface{} {
	log.Printf("[InitLLMSession] 初始化会话: deviceIP=%s, embeddingPath=%s, tokenizerPath=%s", deviceIP, embeddingPath, tokenizerPath)

	// 构建URL
	url := fmt.Sprintf("http://%s/llm/session/init", deviceAddr(deviceIP))
	log.Printf("[InitLLMSession] 请求URL: %s", url)

	// 构建请求体
	requestBody := map[string]string{
		"embedding_path":  embeddingPath,
		"tokenizer_path":  tokenizerPath,
		"system_prompt":   systemPrompt,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[InitLLMSession] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	log.Printf("[InitLLMSession] 请求体: %s", string(jsonData))

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[InitLLMSession] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", token)
		log.Printf("[InitLLMSession] 已添加认证头")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[InitLLMSession] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[InitLLMSession] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[InitLLMSession] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return map[string]interface{}{
				"success": true,
				"message": "会话初始化成功",
				"code":    0,
			}
		}
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	// 检查code字段
	if code, ok := result["code"]; ok {
		if codeInt, ok := code.(float64); ok && codeInt == 0 {
			result["success"] = true
		} else {
			result["success"] = false
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
	}

	return result
}

// ChatWithLLM 与LLM模型对话
func (a *App) ChatWithLLM(deviceIP string, message string, token string) map[string]interface{} {
	log.Printf("[ChatWithLLM] 发送消息: deviceIP=%s, message=%s", deviceIP, message)

	// 构建URL
	url := fmt.Sprintf("http://%s/llm/chat", deviceAddr(deviceIP))
	log.Printf("[ChatWithLLM] 请求URL: %s", url)

	// 构建请求体
	requestBody := map[string]interface{}{
		"message": message,
		"stream":  false, // 非流式响应
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[ChatWithLLM] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ChatWithLLM] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", token)
		log.Printf("[ChatWithLLM] 已添加认证头")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second, // 1分钟超时
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ChatWithLLM] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ChatWithLLM] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[ChatWithLLM] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	// 检查code字段
	if code, ok := result["code"]; ok {
		if codeInt, ok := code.(float64); ok && codeInt == 0 {
			result["success"] = true
		} else {
			result["success"] = false
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
	}

	return result
}

// ChatWithOpenAI 使用OpenAI API进行对话
func (a *App) ChatWithOpenAI(apiKey string, apiBase string, model string, messages string, temperature float64, maxTokens int) map[string]interface{} {
	log.Printf("[ChatWithOpenAI] 发送消息: model=%s, temperature=%.2f, maxTokens=%d", model, temperature, maxTokens)

	// 如果apiBase为空，使用默认值
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

	// 构建URL
	url := fmt.Sprintf("%s/chat/completions", apiBase)
	log.Printf("[ChatWithOpenAI] 请求URL: %s", url)

	// 解析消息JSON
	var messagesList []map[string]interface{}
	if err := json.Unmarshal([]byte(messages), &messagesList); err != nil {
		log.Printf("[ChatWithOpenAI] 消息JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("消息JSON解析失败: %v", err),
		}
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    messagesList,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[ChatWithOpenAI] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	log.Printf("[ChatWithOpenAI] 请求体: %s", string(jsonData))

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ChatWithOpenAI] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ChatWithOpenAI] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ChatWithOpenAI] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[ChatWithOpenAI] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	// 检查是否有错误
	if errorInfo, ok := result["error"]; ok {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("OpenAI API错误: %v", errorInfo),
			"data":    result,
		}
	}

	// OpenAI API成功响应
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
		
		// 提取回复内容
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						result["response"] = content
					}
				}
			}
		}
	} else {
		result["success"] = false
	}

	return result
}

// ChatWithLocalOpenAIStream 使用本地OpenAI兼容API进行流式对话
func (a *App) ChatWithLocalOpenAIStream(deviceIP string, model string, messages string, temperature float64, maxTokens int, token string) map[string]interface{} {
	log.Printf("[ChatWithLocalOpenAIStream] 发送消息: deviceIP=%s, model=%s, temperature=%.2f, maxTokens=%d", deviceIP, model, temperature, maxTokens)

	// 构建URL
	url := fmt.Sprintf("http://%s/v1/chat/completions", deviceAddr(deviceIP))
	log.Printf("[ChatWithLocalOpenAIStream] 请求URL: %s", url)

	// 解析消息JSON
	var messagesList []map[string]interface{}
	if err := json.Unmarshal([]byte(messages), &messagesList); err != nil {
		log.Printf("[ChatWithLocalOpenAIStream] 消息JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("消息JSON解析失败: %v", err),
		}
	}

	// 构建请求体，启用流式传输
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    messagesList,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      true, // 启用流式传输
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[ChatWithLocalOpenAIStream] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	log.Printf("[ChatWithLocalOpenAIStream] 请求体: %s", string(jsonData))

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ChatWithLocalOpenAIStream] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// 发送请求
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ChatWithLocalOpenAIStream] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[ChatWithLocalOpenAIStream] 错误响应: %d, %s", resp.StatusCode, string(respBody))
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败，状态码: %d", resp.StatusCode),
		}
	}

	// 实时读取和转发流式响应
	var fullContent strings.Builder
	var fullReasoning strings.Builder
	var usageInfo map[string]interface{}
	var timingsInfo map[string]interface{} // 添加 timings 信息
	reader := bufio.NewReader(resp.Body)
	
	// 用于跟踪 <think> 标签的状态
	inThinkTag := false
	var thinkBuffer strings.Builder
	var beforeThinkContent strings.Builder
	var afterThinkContent strings.Builder
	thinkTagClosed := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[ChatWithLocalOpenAIStream] 读取流失败: %v", err)
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		// 解析 SSE 数据
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				log.Printf("[ChatWithLocalOpenAIStream] 解析chunk失败: %v", err)
				continue
			}

			// 提取 usage 信息（通常在最后一个消息中）
			if usage, ok := chunk["usage"].(map[string]interface{}); ok {
				usageInfo = usage
				log.Printf("[ChatWithLocalOpenAIStream] 收到usage信息: %+v", usage)
			}

			// 提取 timings 信息（llama.cpp 提供的性能数据）
			if timings, ok := chunk["timings"].(map[string]interface{}); ok {
				timingsInfo = timings
				log.Printf("[ChatWithLocalOpenAIStream] 收到timings信息: %+v", timings)
			}

			// 提取内容
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						// 提取思考过程（如果API直接提供reasoning_content）
						if reasoningContent, ok := delta["reasoning_content"].(string); ok && reasoningContent != "" {
							fullReasoning.WriteString(reasoningContent)
							// 实时发送思考过程
							a.emitEvent("ai-reasoning", reasoningContent)
						}
						
						// 提取回复内容
						if content, ok := delta["content"].(string); ok && content != "" {
							fullContent.WriteString(content)
							
							// 实时解析 <think> 标签并分类发送
							for _, char := range content {
								charStr := string(char)
								
								if !inThinkTag && !thinkTagClosed {
									// 检查是否是 <think> 标签的开始
									thinkBuffer.WriteString(charStr)
									if strings.HasSuffix(thinkBuffer.String(), "<think>") {
										inThinkTag = true
										// 移除标签本身，只保留标签前的内容
										beforeStr := thinkBuffer.String()
										beforeStr = beforeStr[:len(beforeStr)-7]
										if beforeStr != "" {
											beforeThinkContent.WriteString(beforeStr)
											// 实时发送标签前的内容
											a.emitEvent("ai-stream-chunk", beforeStr)
										}
										thinkBuffer.Reset()
									} else if !strings.HasPrefix("<think>", thinkBuffer.String()) {
										// 不是标签的开始，发送内容
										toSend := thinkBuffer.String()
										beforeThinkContent.WriteString(toSend)
										a.emitEvent("ai-stream-chunk", toSend)
										thinkBuffer.Reset()
										thinkBuffer.WriteString(charStr)
									}
								} else if inThinkTag {
									// 在 think 标签内，收集思考内容
									thinkBuffer.WriteString(charStr)
									if strings.HasSuffix(thinkBuffer.String(), "</think>") {
										// 标签结束
										thinkStr := thinkBuffer.String()
										thinkStr = thinkStr[:len(thinkStr)-8]
										if thinkStr != "" {
											fullReasoning.WriteString(thinkStr)
											// 实时发送思考内容
											a.emitEvent("ai-reasoning", thinkStr)
										}
										inThinkTag = false
										thinkTagClosed = true
										thinkBuffer.Reset()
									}
								} else if thinkTagClosed {
									// think 标签已关闭，后续都是正常内容
									afterThinkContent.WriteString(charStr)
									a.emitEvent("ai-stream-chunk", charStr)
								}
							}
						}
					}
				}
			}
		}
	}
	
	// 处理剩余的缓冲内容
	if thinkBuffer.Len() > 0 {
		remaining := thinkBuffer.String()
		if inThinkTag {
			fullReasoning.WriteString(remaining)
			a.emitEvent("ai-reasoning", remaining)
		} else {
			beforeThinkContent.WriteString(remaining)
			a.emitEvent("ai-stream-chunk", remaining)
		}
	}

	// 构建最终的内容和思考文本
	reasoningText := fullReasoning.String()
	contentText := beforeThinkContent.String() + afterThinkContent.String()
	
	log.Printf("[ChatWithLocalOpenAIStream] 接收完成 - 思考: %d字符, 内容: %d字符", 
		len(reasoningText), len(contentText))

	// 发送完成事件
	doneData := map[string]interface{}{
		"content":   contentText,
		"reasoning": reasoningText,
	}
	
	// 如果有 usage 信息，添加到完成事件中
	if usageInfo != nil {
		doneData["usage"] = usageInfo
	}
	
	// 如果有 timings 信息，添加到完成事件中
	if timingsInfo != nil {
		doneData["timings"] = timingsInfo
		log.Printf("[ChatWithLocalOpenAIStream] 传递timings到前端: %+v", timingsInfo)
	}
	
	a.emitEvent("ai-stream-done", doneData)

	return map[string]interface{}{
		"success":   true,
		"response":  contentText,
		"reasoning": reasoningText,
	}
}

// ChatWithLocalOpenAI 使用本地OpenAI兼容API进行对话（非流式，保留用于兼容）
func (a *App) ChatWithLocalOpenAI(deviceIP string, model string, messages string, temperature float64, maxTokens int, token string) map[string]interface{} {
	log.Printf("[ChatWithLocalOpenAI] 发送消息: deviceIP=%s, model=%s, temperature=%.2f, maxTokens=%d", deviceIP, model, temperature, maxTokens)

	// 构建URL
	url := fmt.Sprintf("http://%s/v1/chat/completions", deviceAddr(deviceIP))
	log.Printf("[ChatWithLocalOpenAI] 请求URL: %s", url)

	// 解析消息JSON
	var messagesList []map[string]interface{}
	if err := json.Unmarshal([]byte(messages), &messagesList); err != nil {
		log.Printf("[ChatWithLocalOpenAI] 消息JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("消息JSON解析失败: %v", err),
		}
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    messagesList,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[ChatWithLocalOpenAI] JSON序列化失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("JSON序列化失败: %v", err),
		}
	}

	log.Printf("[ChatWithLocalOpenAI] 请求体: %s", string(jsonData))

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ChatWithLocalOpenAI] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ChatWithLocalOpenAI] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ChatWithLocalOpenAI] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[ChatWithLocalOpenAI] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	// 检查是否有错误
	if errorInfo, ok := result["error"]; ok {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("本地OpenAI API错误: %v", errorInfo),
			"data":    result,
		}
	}

	// OpenAI API成功响应
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
		
		// 提取回复内容
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						result["response"] = content
					}
				}
			}
		}
	} else {
		result["success"] = false
	}

	return result
}

// StopContainer 停止容器
func (a *App) StopContainer(deviceIP string, version string, containerId string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 StopContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerId=%s", deviceIP, version, containerId)

	// 调用真实的容器停止逻辑
	success, message := stopContainer(deviceIP, version, containerId, password)

	result := map[string]interface{}{
		"success": success,
		"message": message,
	}
	log.Printf("[IPC] StopContainer 返回结果: %+v", result)
	return result
}

// startContainer 调用设备API启动容器
func startContainer(deviceIP string, version string, containerId string, password string) (bool, string) {
	if version == "v3" {
		// V3 API: 调用设备的API接口启动容器
		return startV3Container(deviceIP, containerId, password)
	} else {
		// V0-V2 Docker API: 调用Docker API启动容器
		return startDockerContainer(deviceIP, containerId, version, password)
	}
}

// stopContainer 调用设备API停止容器
func stopContainer(deviceIP string, version string, containerId string, password string) (bool, string) {
	if version == "v3" {
		// V3 API: 调用设备的API接口停止容器
		return stopV3Container(deviceIP, containerId, password)
	} else {
		// V0-V2 Docker API: 调用Docker API停止容器
		return stopDockerContainer(deviceIP, containerId, version, password)
	}
}

// startV3Container 启动V3设备的容器
func startV3Container(deviceIP string, containerId string, password string) (bool, string) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/android/start", deviceAddr(deviceIP))

	// 发送HTTP POST请求，包含容器名称
	reqBody := map[string]string{
		"name": containerId,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Sprintf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 添加认证头
	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("读取V3 API响应失败: %v", err)
	}

	log.Printf("V3 Start API响应: %s", string(body))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Sprintf("解析V3 API响应失败: %v", err)
	}

	// 检查响应状态
	if result["code"] == float64(0) {
		return true, "容器启动成功"
	}

	return false, fmt.Sprintf("容器启动失败: %v", result["message"])
}

// stopV3Container 停止V3设备的容器
func stopV3Container(deviceIP string, containerId string, password string) (bool, string) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/android/stop", deviceAddr(deviceIP))

	// 发送HTTP POST请求，包含容器名称
	reqBody := map[string]string{
		"name": containerId,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Sprintf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 添加认证头
	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("读取V3 API响应失败: %v", err)
	}

	log.Printf("V3 Stop API响应: %s", string(body))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Sprintf("解析V3 API响应失败: %v", err)
	}

	// 检查响应状态
	if result["code"] == float64(0) {
		return true, "容器停止成功"
	}

	return false, fmt.Sprintf("容器停止失败: %v", result["message"])
}

// startDockerContainer 启动Docker容器
func startDockerContainer(deviceIP string, containerId string, version string, password string) (bool, string) {
	resp, err := callDockerAPI(deviceIP, version, "/containers/"+containerId+"/start", "POST", nil, password)
	if err != nil {
		return false, fmt.Sprintf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Docker Start API响应状态码: %d", resp.StatusCode)

	// 检查响应状态码
	if resp.StatusCode == http.StatusNoContent {
		return true, "容器启动成功"
	}

	// 读取错误响应
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("容器启动失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// stopDockerContainer 停止Docker容器
func stopDockerContainer(deviceIP string, containerId string, version string, password string) (bool, string) {
	resp, err := callDockerAPI(deviceIP, version, "/containers/"+containerId+"/stop", "POST", nil, password)
	if err != nil {
		return false, fmt.Sprintf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Docker Stop API响应状态码: %d", resp.StatusCode)

	// 检查响应状态码
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotModified {
		return true, "容器停止成功"
	}

	// 读取错误响应
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("容器停止失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// DeleteContainer 删除容器
// Todo: 实现删除容器的完整功能
func (a *App) DeleteContainer(deviceIP string, version string, containerId string) map[string]interface{} {
	log.Printf("[IPC] 收到 DeleteContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, containerId=%s", deviceIP, version, containerId)

	// 功能正在开发中，返回提示信息
	result := map[string]interface{}{
		"success": false,
		"message": "功能正在开发中",
	}
	log.Printf("[IPC] DeleteContainer 返回结果: %+v", result)
	return result
}

// CreateContainer 创建容器
// Todo: 实现创建容器的完整功能
func (a *App) CreateContainer(deviceIP string, version string, params map[string]interface{}) map[string]interface{} {
	log.Printf("[IPC] 收到 CreateContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, params=%+v", deviceIP, version, params)

	// 功能正在开发中，返回提示信息
	result := map[string]interface{}{
		"success": false,
		"message": "功能正在开发中",
	}
	log.Printf("[IPC] CreateContainer 返回结果: %+v", result)
	return result
}

// 应用名称，用于统一管理目录
const appName = "魔云腾"

// storageSavePathKey 自定义保存路径配置文件名
const storageSavePathKey = "storage_path.txt"

// sharedDirPathKey 自定义共享目录（上传文件来源）配置文件名
const sharedDirPathKey = "shared_dir_path.txt"

// getDefaultSharedDirPath 获取默认共享目录路径
func getDefaultSharedDirPath() string {
	userDataDir, err := os.UserCacheDir()
	if err != nil {
		userDataDir = os.TempDir()
	}
	dir := filepath.Join(userDataDir, "edgeclient", "shared")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// getCustomSharedDirPath 获取实际使用的共享目录（优先使用用户自定义路径）
func getCustomSharedDirPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return getDefaultSharedDirPath()
	}
	cfgFile := filepath.Join(configDir, "edgeclient", sharedDirPathKey)
	data, err := os.ReadFile(cfgFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return getDefaultSharedDirPath()
	}
	customPath := strings.TrimSpace(string(data))
	if err := os.MkdirAll(customPath, 0755); err != nil {
		log.Printf("[getCustomSharedDirPath] 自定义路径无效(%s): %v，使用默认路径", customPath, err)
		return getDefaultSharedDirPath()
	}
	return customPath
}

// GetSharedDirPath 获取当前共享目录路径（IPC）
func (a *App) GetSharedDirPath() map[string]interface{} {
	log.Printf("[IPC] 收到 GetSharedDirPath 调用")
	configDir, err := os.UserConfigDir()
	defaultPath := getDefaultSharedDirPath()
	if err != nil {
		return map[string]interface{}{
			"success":     true,
			"path":        defaultPath,
			"isDefault":   true,
			"defaultPath": defaultPath,
		}
	}
	cfgFile := filepath.Join(configDir, "edgeclient", sharedDirPathKey)
	data, err := os.ReadFile(cfgFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return map[string]interface{}{
			"success":     true,
			"path":        defaultPath,
			"isDefault":   true,
			"defaultPath": defaultPath,
		}
	}
	customPath := strings.TrimSpace(string(data))
	if err2 := os.MkdirAll(customPath, 0755); err2 != nil {
		log.Printf("[GetSharedDirPath] 自定义路径无效(%s): %v，自动恢复默认路径", customPath, err2)
		_ = os.Remove(cfgFile)
		return map[string]interface{}{
			"success":     true,
			"path":        defaultPath,
			"isDefault":   true,
			"defaultPath": defaultPath,
		}
	}
	return map[string]interface{}{
		"success":     true,
		"path":        customPath,
		"isDefault":   false,
		"defaultPath": defaultPath,
	}
}

// SetSharedDirPath 设置共享目录路径（IPC）
func (a *App) SetSharedDirPath(path string) map[string]interface{} {
	log.Printf("[IPC] 收到 SetSharedDirPath 调用: path=%s", path)
	path = strings.TrimSpace(path)
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return map[string]interface{}{"success": false, "message": "获取配置目录失败"}
		}
		cfgFile := filepath.Join(configDir, "edgeclient", sharedDirPathKey)
		_ = os.Remove(cfgFile)
		return map[string]interface{}{"success": true, "message": "已恢复默认路径", "path": getDefaultSharedDirPath()}
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("路径无效或无法创建: %v", err)}
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return map[string]interface{}{"success": false, "message": "获取配置目录失败"}
	}
	cfgDir := filepath.Join(configDir, "edgeclient")
	_ = os.MkdirAll(cfgDir, 0777)
	cfgFile := filepath.Join(cfgDir, sharedDirPathKey)
	if err := os.WriteFile(cfgFile, []byte(path), 0666); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("保存路径失败: %v", err)}
	}
	log.Printf("[SetSharedDirPath] 保存成功: %s", path)
	return map[string]interface{}{"success": true, "message": "保存路径成功", "path": path}
}

// getDefaultStorageBaseDir 获取默认存储根目录（os.UserConfigDir/edgeclient）
func getDefaultStorageBaseDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		if home, e := os.UserHomeDir(); e == nil {
			configDir = home
		} else {
			configDir = os.TempDir()
		}
	}
	dir := filepath.Join(configDir, "edgeclient")
	_ = os.MkdirAll(dir, 0777)
	return dir
}

// getStorageBaseDir 获取实际使用的存储根目录（优先使用用户自定义路径）
func getStorageBaseDir() string {
	// 读取配置文件中的自定义路径
	configDir, err := os.UserConfigDir()
	if err != nil {
		return getDefaultStorageBaseDir()
	}
	cfgFile := filepath.Join(configDir, "edgeclient", storageSavePathKey)
	data, err := os.ReadFile(cfgFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return getDefaultStorageBaseDir()
	}
	customPath := strings.TrimSpace(string(data))
	// 验证路径存在或可创建
	if err := os.MkdirAll(customPath, 0777); err != nil {
		log.Printf("[getStorageBaseDir] 自定义路径无效(%s): %v，使用默认路径", customPath, err)
		return getDefaultStorageBaseDir()
	}
	return customPath
}

// GetStoragePath 获取当前保存路径（IPC）
func (a *App) GetStoragePath() map[string]interface{} {
	log.Printf("[IPC] 收到 GetStoragePath 调用")
	configDir, err := os.UserConfigDir()
	if err != nil {
		return map[string]interface{}{
			"success":      true,
			"path":         getDefaultStorageBaseDir(),
			"isDefault":    true,
			"defaultPath":  getDefaultStorageBaseDir(),
		}
	}
	cfgFile := filepath.Join(configDir, "edgeclient", storageSavePathKey)
	data, err := os.ReadFile(cfgFile)
	defaultPath := getDefaultStorageBaseDir()
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return map[string]interface{}{
			"success":     true,
			"path":        defaultPath,
			"isDefault":   true,
			"defaultPath": defaultPath,
		}
	}
	customPath := strings.TrimSpace(string(data))
	// 验证自定义路径是否有效（路径不存在且无法创建时自动恢复默认）
	if err2 := os.MkdirAll(customPath, 0777); err2 != nil {
		log.Printf("[GetStoragePath] 自定义路径无效(%s): %v，自动恢复默认路径", customPath, err2)
		// 清除无效配置
		_ = os.Remove(cfgFile)
		return map[string]interface{}{
			"success":     true,
			"path":        defaultPath,
			"isDefault":   true,
			"defaultPath": defaultPath,
		}
	}
	return map[string]interface{}{
		"success":     true,
		"path":        customPath,
		"isDefault":   false,
		"defaultPath": defaultPath,
	}
}

// SetStoragePath 设置保存路径（IPC）
func (a *App) SetStoragePath(path string) map[string]interface{} {
	log.Printf("[IPC] 收到 SetStoragePath 调用: path=%s", path)
	path = strings.TrimSpace(path)
	if path == "" {
		// 清空配置，恢复默认
		configDir, err := os.UserConfigDir()
		if err != nil {
			return map[string]interface{}{"success": false, "message": "获取配置目录失败"}
		}
		cfgFile := filepath.Join(configDir, "edgeclient", storageSavePathKey)
		_ = os.Remove(cfgFile)
		return map[string]interface{}{"success": true, "message": "已恢复默认路径", "path": getDefaultStorageBaseDir()}
	}
	// 验证并创建目录
	if err := os.MkdirAll(path, 0777); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("路径无效或无法创建: %v", err)}
	}
	// 保存配置
	configDir, err := os.UserConfigDir()
	if err != nil {
		return map[string]interface{}{"success": false, "message": "获取配置目录失败"}
	}
	cfgDir := filepath.Join(configDir, "edgeclient")
	_ = os.MkdirAll(cfgDir, 0777)
	cfgFile := filepath.Join(cfgDir, storageSavePathKey)
	if err := os.WriteFile(cfgFile, []byte(path), 0666); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("保存路径失败: %v", err)}
	}
	log.Printf("[SetStoragePath] 保存成功: %s", path)
	return map[string]interface{}{"success": true, "message": "保存路径成功", "path": path}
}

// SelectDirectory 打开目录选择对话框（IPC）
func (a *App) SelectDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 SelectDirectory 调用")
	if runtime.GOOS != "windows" {
		return map[string]interface{}{"success": false, "message": "仅支持 Windows 平台"}
	}
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = "请选择文件保存路径"
$dialog.ShowNewFolderButton = $true
$result = $dialog.ShowDialog()
if ($result -eq [System.Windows.Forms.DialogResult]::OK) {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($dialog.SelectedPath)
    [Convert]::ToBase64String($bytes)
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[SelectDirectory] 打开目录选择对话框失败: %v", err)
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("打开目录选择对话框失败: %v", err)}
	}
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return map[string]interface{}{"success": false, "message": "用户取消选择"}
	}
	pathBytes, err := ConvertBase64ToBytes(outputStr)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("路径解码失败: %v", err)}
	}
	selectedPath := string(pathBytes)
	log.Printf("[SelectDirectory] 用户选择目录: %s", selectedPath)
	return map[string]interface{}{"success": true, "path": selectedPath}
}

// 下载模板相关功能

// getConfigPath 获取配置文件路径
func getConfigPath(conf string) string {
	// 获取用户的应用支持目录（例如：/Users/username/Library/Application Support/MyApp）
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(configDir, "edgeclient") // 替换为你的应用名

	// 确保目录存在
	if err := os.MkdirAll(appDir, 0777); err != nil {
		panic(err)
	}

	// 返回配置文件的完整路径
	return filepath.Join(appDir, conf)
}

// getConfigPath 获取机型文件路径
func getModelPath(name string) string {
	modelDir := filepath.Join(getStorageBaseDir(), "model")

	// 确保目录存在
	if err := os.MkdirAll(modelDir, 0777); err != nil {
		panic(err)
	}

	// 返回配置文件的完整路径
	return filepath.Join(modelDir, name)
}

// DownloadTemplateResponse 下载URL响应结构体
type DownloadTemplateResponse struct {
	CodeID int    `json:"code_id"`
	Msg    string `json:"msg"`
	Data   struct {
		DownloadURL string `json:"download_url"`
	} `json:"data"`
}

// GetTemplateDownloadURL 获取模板下载URL
func (a *App) GetTemplateDownloadURL(modelID string) (string, error) {
	url := fmt.Sprintf("https://newapi.moyunteng.com/api/v1/template/download-url/%s", modelID)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取下载URL失败: %v", err)
	}
	defer resp.Body.Close()

	var result DownloadTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析下载URL响应失败: %v", err)
	}

	if result.CodeID != 200 {
		return "", fmt.Errorf("获取下载URL失败: %s", result.Msg)
	}

	return result.Data.DownloadURL, nil
}

// DownloadTemplate 下载模板文件
func (a *App) DownloadTemplate(modelID string) (map[string]interface{}, error) {
	// 获取下载URL
	downloadURL, err := a.GetTemplateDownloadURL(modelID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, err
	}

	// 解析文件名
	fileName := filepath.Base(downloadURL)
	// 使用getModelPath函数获取保存路径
	savePath := getModelPath(fileName)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 发起下载请求
	resp, err := client.Get(downloadURL)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载文件失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// 创建文件
	out, err := os.Create(savePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建文件失败: %v", err),
		}, err
	}
	defer out.Close()

	// 复制文件内容
	if _, err = io.Copy(out, resp.Body); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存文件失败: %v", err),
		}, err
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "下载成功",
		"filePath": savePath,
		"fileName": fileName,
	}, nil
}

// getBackupModelPath 获取备份机型文件保存路径
func getBackupModelPath(name string) string {
	baseDir := getStorageBaseDir()
	backupDir := filepath.Join(baseDir, "BackupModel")
	if err := os.MkdirAll(backupDir, 0777); err != nil {
		panic(err)
	}
	return filepath.Join(backupDir, name)
}

// ExportBackupModel 导出备份机型
func (a *App) ExportBackupModel(deviceIP, modelName string) map[string]interface{} {
	log.Printf("[ExportBackupModel] 开始导出备份机型: deviceIP=%s, modelName=%s", deviceIP, modelName)

	// 获取设备认证信息（密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	apiURL := fmt.Sprintf("http://%s/android/backup/modelExport?name=%s", deviceAddr(deviceIP), url.QueryEscape(modelName))
	log.Printf("[ExportBackupModel] API URL: %s", apiURL)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("[ExportBackupModel] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头（如果有密码）
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[ExportBackupModel] 已添加认证头")
	} else {
		log.Printf("[ExportBackupModel] 警告: 未获取到设备密码，可能导致401错误")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ExportBackupModel] 下载失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载失败: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ExportBackupModel] 接口返回错误状态码: %d", resp.StatusCode)
		
		// 如果是401认证失败，提供更明确的错误信息
		if resp.StatusCode == 401 {
			return map[string]interface{}{
				"success": false,
				"message": "认证失败，请检查设备密码是否正确",
			}
		}
		
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("接口返回错误: %d", resp.StatusCode),
		}
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	fileName := modelName + ".zip"
	if contentDisposition != "" {
		if strings.Contains(contentDisposition, "filename=") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				fileName = strings.Trim(parts[1], "\" ")
			}
		}
	}

	savePath := getBackupModelPath(fileName)
	log.Printf("[ExportBackupModel] 保存路径: %s", savePath)

	out, err := os.Create(savePath)
	if err != nil {
		log.Printf("[ExportBackupModel] 创建文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建文件失败: %v", err),
		}
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("[ExportBackupModel] 保存文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存文件失败: %v", err),
		}
	}

	log.Printf("[ExportBackupModel] 导出成功: %s", savePath)
	return map[string]interface{}{
		"success":  true,
		"message":  "导出成功",
		"filePath": savePath,
		"fileName": fileName,
	}
}

// CheckBackupModelExists 检查备份机型文件是否存在
func (a *App) CheckBackupModelExists(modelName string) map[string]interface{} {
	log.Printf("[CheckBackupModelExists] 检查机型是否存在: %s", modelName)

	backupDir := filepath.Join(getStorageBaseDir(), "BackupModel")
	filePath := filepath.Join(backupDir, modelName+".zip")

	exists := fileExists(filePath)
	log.Printf("[CheckBackupModelExists] 文件路径: %s, 存在: %v", filePath, exists)

	return map[string]interface{}{
		"success":  true,
		"exists":   exists,
		"filePath": filePath,
	}
}

// GetAllBackupModels 获取所有本地备份机型列表
func (a *App) GetAllBackupModels() map[string]interface{} {
	log.Printf("[GetAllBackupModels] 获取本地备份机型列表")

	backupDir := filepath.Join(getStorageBaseDir(), "BackupModel")

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		log.Printf("[GetAllBackupModels] 备份目录不存在: %s", backupDir)
		return map[string]interface{}{
			"success": true,
			"list":    []string{},
		}
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("[GetAllBackupModels] 读取目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"list":    []string{},
		}
	}

	var modelList []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
			modelName := strings.TrimSuffix(entry.Name(), ".zip")
			modelList = append(modelList, modelName)
		}
	}

	log.Printf("[GetAllBackupModels] 找到 %d 个本地备份机型", len(modelList))
	return map[string]interface{}{
		"success": true,
		"list":    modelList,
	}
}

// ImportBackupModel 导入备份机型
func (a *App) ImportBackupModel(deviceIP, modelName string) map[string]interface{} {
	log.Printf("[ImportBackupModel] 开始导入备份机型: deviceIP=%s, modelName=%s", deviceIP, modelName)

	// 获取设备认证信息（密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("[ImportBackupModel] 获取配置目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "BackupModel")
	filePath := filepath.Join(backupDir, modelName+".zip")

	log.Printf("[ImportBackupModel] 文件路径: %s", filePath)

	if !fileExists(filePath) {
		log.Printf("[ImportBackupModel] 文件不存在")
		return map[string]interface{}{
			"success": false,
			"message": "本地备份文件不存在: " + filePath,
		}
	}

	log.Printf("[ImportBackupModel] 文件存在，开始打开")
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ImportBackupModel] 打开文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件失败: %v", err),
		}
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Printf("[ImportBackupModel] 获取文件信息失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取文件信息失败: %v", err),
		}
	}

	log.Printf("[ImportBackupModel] 文件大小: %d", fileStat.Size())

	apiURL := fmt.Sprintf("http://%s/android/backup/modelImport", deviceAddr(deviceIP))
	log.Printf("[ImportBackupModel] 请求URL: %s", apiURL)

	client := &http.Client{}

	log.Printf("[ImportBackupModel] 创建表单数据")
	body, contentType := createMultipartFormData(modelName, file, fileStat)
	if body == nil {
		log.Printf("[ImportBackupModel] 创建表单数据失败")
		return map[string]interface{}{
			"success": false,
			"message": "创建表单数据失败",
		}
	}

	log.Printf("[ImportBackupModel] Content-Type: %s", contentType)

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		log.Printf("[ImportBackupModel] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}
	req.Header.Set("Content-Type", contentType)
	
	// 添加认证头（如果有密码）
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[ImportBackupModel] 已添加认证头")
	} else {
		log.Printf("[ImportBackupModel] 警告: 未获取到设备密码，可能导致401错误")
	}

	log.Printf("[ImportBackupModel] 发送请求")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ImportBackupModel] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[ImportBackupModel] 响应状态: %d, 响应体: %s", resp.StatusCode, string(respBody))

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[ImportBackupModel] JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v, 响应体: %s", err, string(respBody)),
		}
	}

	log.Printf("[ImportBackupModel] result: %+v", result)

	codeValue := result["code"]
	code := 0
	switch v := codeValue.(type) {
	case float64:
		code = int(v)
	case int:
		code = v
	case string:
		if v == "0" {
			code = 0
		}
	}

	log.Printf("[ImportBackupModel] code=%d, message=%v", code, result["message"])

	if resp.StatusCode == 200 && code == 0 {
		log.Printf("[ImportBackupModel] 导入成功")
		return map[string]interface{}{
			"success": true,
			"message": "导入成功",
		}
	}

	msg := ""
	if m, ok := result["message"].(string); ok {
		msg = m
	} else if m, ok := result["msg"].(string); ok {
		msg = m
	} else if m, ok := result["message"].(float64); ok {
		msg = fmt.Sprintf("%v", m)
	}
	log.Printf("[ImportBackupModel] 导入失败: %s", msg)
	return map[string]interface{}{
		"success": false,
		"message": msg,
	}
}

// DeleteBackupModel 删除备份机型
func (a *App) DeleteBackupModel(deviceIP, modelName string) map[string]interface{} {
	log.Printf("[DeleteBackupModel] 开始删除备份机型: deviceIP=%s, modelName=%s", deviceIP, modelName)

	// 获取设备认证信息（密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	apiURL := fmt.Sprintf("http://%s/android/backup/model?name=%s", deviceAddr(deviceIP), url.QueryEscape(modelName))
	log.Printf("[DeleteBackupModel] API URL: %s", apiURL)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		log.Printf("[DeleteBackupModel] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头（如果有密码）
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[DeleteBackupModel] 已添加认证头")
	} else {
		log.Printf("[DeleteBackupModel] 警告: 未获取到设备密码，可能导致401错误")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DeleteBackupModel] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[DeleteBackupModel] 响应状态: %d, 响应体: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DeleteBackupModel] 接口返回错误状态码: %d", resp.StatusCode)
		
		// 如果是401认证失败，提供更明确的错误信息
		if resp.StatusCode == 401 {
			return map[string]interface{}{
				"success": false,
				"message": "认证失败，请检查设备密码是否正确",
			}
		}
		
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("接口返回错误: %d, 响应: %s", resp.StatusCode, string(respBody)),
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[DeleteBackupModel] JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v", err),
		}
	}

	codeValue := result["code"]
	code := 0
	switch v := codeValue.(type) {
	case float64:
		code = int(v)
	case int:
		code = v
	case string:
		if v == "0" {
			code = 0
		}
	}

	if code != 0 {
		msg := ""
		if m, ok := result["message"].(string); ok {
			msg = m
		} else if m, ok := result["msg"].(string); ok {
			msg = m
		}
		log.Printf("[DeleteBackupModel] 删除失败: %s", msg)
		return map[string]interface{}{
			"success": false,
			"message": msg,
		}
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("[DeleteBackupModel] 获取配置目录失败: %v", err)
		return map[string]interface{}{
			"success": true,
			"message": "设备上删除成功，但清理本地文件失败: " + err.Error(),
		}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "BackupModel")
	localFilePath := filepath.Join(backupDir, modelName+".zip")

	if fileExists(localFilePath) {
		if err := os.Remove(localFilePath); err != nil {
			log.Printf("[DeleteBackupModel] 删除本地文件失败: %v", err)
			return map[string]interface{}{
				"success": true,
				"message": "设备上删除成功，但清理本地文件失败: " + err.Error(),
			}
		}
		log.Printf("[DeleteBackupModel] 本地文件已删除: %s", localFilePath)
	}

	log.Printf("[DeleteBackupModel] 删除成功")
	return map[string]interface{}{
		"success": true,
		"message": "删除成功",
	}
}

func (a *App) DownloadBackupMachine(deviceIP, machineName string) map[string]interface{} {
	log.Printf("[DownloadBackupMachine] 开始下载备份云机: deviceIP=%s, machineName=%s", deviceIP, machineName)

	// 获取设备认证信息（密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")
	if err := os.MkdirAll(backupDir, 0777); err != nil {
		log.Printf("[DownloadBackupMachine] 创建目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建目录失败: %v", err),
		}
	}

	apiURL := fmt.Sprintf("http://%s/backup/download?name=%s", deviceAddr(deviceIP), url.QueryEscape(machineName))
	log.Printf("[DownloadBackupMachine] API URL: %s", apiURL)

	client := &http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("[DownloadBackupMachine] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头（如果有密码）
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[DownloadBackupMachine] 已添加认证头")
	} else {
		log.Printf("[DownloadBackupMachine] 警告: 未获取到设备密码，可能导致401错误")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DownloadBackupMachine] 下载失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载失败: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DownloadBackupMachine] 接口返回错误状态码: %d", resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)
		
		// 如果是401认证失败，提供更明确的错误信息
		if resp.StatusCode == 401 {
			return map[string]interface{}{
				"success": false,
				"message": "认证失败，请检查设备密码是否正确",
			}
		}
		
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("接口返回错误: %d, 响应: %s", resp.StatusCode, string(respBody)),
		}
	}

	fileName := machineName + ".zip"
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		if strings.Contains(contentDisposition, "filename=") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				fileName = strings.Trim(parts[1], "\" ")
			}
		}
	}

	savePath := filepath.Join(backupDir, fileName)
	log.Printf("[DownloadBackupMachine] 保存路径: %s", savePath)

	out, err := os.Create(savePath)
	if err != nil {
		log.Printf("[DownloadBackupMachine] 创建文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建文件失败: %v", err),
		}
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("[DownloadBackupMachine] 保存文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存文件失败: %v", err),
		}
	}

	log.Printf("[DownloadBackupMachine] 下载成功: %s", savePath)
	return map[string]interface{}{
		"success":  true,
		"message":  "下载成功",
		"filePath": savePath,
		"fileName": fileName,
	}
}

func (a *App) CheckBackupMachineFileExists(machineName string) map[string]interface{} {
	log.Printf("[CheckBackupMachineFileExists] 检查文件是否存在: %s", machineName)

	configDir, err := os.UserConfigDir()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"exists":  false,
		}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")

	zipPath := filepath.Join(backupDir, machineName+".zip")
	tarPath := filepath.Join(backupDir, machineName+".tar")
	noExtPath := filepath.Join(backupDir, machineName)

	zipExists := fileExists(zipPath)
	tarExists := fileExists(tarPath)
	noExtExists := fileExists(noExtPath)

	exists := zipExists || tarExists || noExtExists

	var filePath string
	if zipExists {
		filePath = zipPath
	} else if tarExists {
		filePath = tarPath
	} else if noExtExists {
		filePath = noExtPath
	}

	log.Printf("[CheckBackupMachineFileExists] zip: %s, tar: %s, noExt: %s, 存在: %v", zipPath, tarPath, noExtPath, exists)

	return map[string]interface{}{
		"success":  true,
		"exists":   exists,
		"filePath": filePath,
	}
}

// CheckBackupMachineFilesExistBatch 批量检查备份云机文件是否存在，一次读目录完成所有匹配
func (a *App) CheckBackupMachineFilesExistBatch(machineNames []string) map[string]interface{} {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error(), "result": map[string]bool{}}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")

	// 一次性读取目录所有文件名，构建 Set
	entries, err := os.ReadDir(backupDir)
	existingFiles := make(map[string]bool)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				existingFiles[entry.Name()] = true
			}
		}
	}

	result := make(map[string]bool, len(machineNames))
	for _, name := range machineNames {
		result[name] = existingFiles[name] || existingFiles[name+".zip"] || existingFiles[name+".tar"]
	}

	log.Printf("[CheckBackupMachineFilesExistBatch] 批量检查 %d 个文件，目录共 %d 个文件", len(machineNames), len(existingFiles))

	return map[string]interface{}{"success": true, "result": result}
}

func (a *App) ImportBackupMachine(deviceIP, deviceName, machineName string, slot int) map[string]interface{} {
	log.Printf("[ImportBackupMachine] 开始导入备份云机: deviceIP=%s, deviceName=%s, machineName=%s, slot=%d", deviceIP, deviceName, machineName, slot)

	// 获取设备认证信息（密码）
	a.devicePasswordsMutex.RLock()
	password := a.devicePasswords[deviceIP]
	a.devicePasswordsMutex.RUnlock()

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("[ImportBackupMachine] 获取配置目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")

	var filePath string
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("[ImportBackupMachine] 读取目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取目录失败: %v", err),
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if strings.Contains(fileName, deviceName) {
			filePath = filepath.Join(backupDir, fileName)
			log.Printf("[ImportBackupMachine] 找到匹配文件: %s", filePath)
			break
		}
	}

	if filePath == "" {
		log.Printf("[ImportBackupMachine] 未找到包含设备名称 %s 的备份文件", deviceName)
		return map[string]interface{}{
			"success": false,
			"message": "未找到匹配的备份文件: " + deviceName,
		}
	}

	log.Printf("[ImportBackupMachine] 文件路径: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ImportBackupMachine] 打开文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开文件失败: %v", err),
		}
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Printf("[ImportBackupMachine] 获取文件信息失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取文件信息失败: %v", err),
		}
	}

	log.Printf("[ImportBackupMachine] 文件大小: %d", fileStat.Size())

	apiURL := fmt.Sprintf("http://%s/android/import", deviceAddr(deviceIP))
	log.Printf("[ImportBackupMachine] 请求URL: %s", apiURL)

	client := &http.Client{}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		formFile, err := writer.CreateFormFile("file", fileStat.Name())
		if err != nil {
			log.Printf("[ImportBackupMachine] 创建表单文件失败: %v", err)
			return
		}

		if _, err := io.Copy(formFile, file); err != nil {
			log.Printf("[ImportBackupMachine] 复制文件内容失败: %v", err)
			return
		}

		if err := writer.WriteField("name", machineName); err != nil {
			log.Printf("[ImportBackupMachine] 写入name字段失败: %v", err)
			return
		}

		if err := writer.WriteField("indexNum", strconv.Itoa(slot)); err != nil {
			log.Printf("[ImportBackupMachine] 写入indexNum字段失败: %v", err)
			return
		}
	}()

	req, err := http.NewRequest("POST", apiURL, pr)
	if err != nil {
		log.Printf("[ImportBackupMachine] 创建请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// 添加认证头（如果有密码）
	if password != "" {
		req.SetBasicAuth("admin", password)
		log.Printf("[ImportBackupMachine] 已添加认证头")
	} else {
		log.Printf("[ImportBackupMachine] 警告: 未获取到设备密码，可能导致401错误")
	}

	log.Printf("[ImportBackupMachine] 发送请求")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ImportBackupMachine] 请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[ImportBackupMachine] 响应状态: %d, 响应体: %s", resp.StatusCode, string(respBody))

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[ImportBackupMachine] JSON解析失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析响应失败: %v, 响应体: %s", err, string(respBody)),
		}
	}

	codeValue := result["code"]
	code := 0
	switch v := codeValue.(type) {
	case float64:
		code = int(v)
	case int:
		code = v
	case string:
		if v == "0" {
			code = 0
		}
	}

	log.Printf("[ImportBackupMachine] code=%d, message=%v", code, result["message"])

	if resp.StatusCode == 200 && code == 0 {
		log.Printf("[ImportBackupMachine] 导入成功")
		return map[string]interface{}{
			"success": true,
			"message": "导入成功",
		}
	}

	msg := ""
	if m, ok := result["message"].(string); ok {
		msg = m
	} else if m, ok := result["msg"].(string); ok {
		msg = m
	} else if m, ok := result["message"].(float64); ok {
		msg = fmt.Sprintf("%v", m)
	}
	log.Printf("[ImportBackupMachine] 导入失败: %s", msg)
	return map[string]interface{}{
		"success": false,
		"message": msg,
	}
}

func (a *App) DeleteLocalBackupMachine(machineName string) map[string]interface{} {
	log.Printf("[DeleteLocalBackupMachine] 开始删除本地备份云机: machineName=%s", machineName)

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("[DeleteLocalBackupMachine] 获取配置目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	_ = configDir

	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		log.Printf("[DeleteLocalBackupMachine] 目录不存在: %s", backupDir)
		return map[string]interface{}{
			"success": true,
			"message": "目录不存在，无需删除",
		}
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("[DeleteLocalBackupMachine] 读取目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取目录失败: %v", err),
		}
	}

	deletedFiles := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if strings.Contains(fileName, machineName) {
			filePath := filepath.Join(backupDir, fileName)
			if err := os.Remove(filePath); err != nil {
				log.Printf("[DeleteLocalBackupMachine] 删除文件失败: %s, error: %v", filePath, err)
			} else {
				log.Printf("[DeleteLocalBackupMachine] 已删除文件: %s", filePath)
				deletedFiles = append(deletedFiles, fileName)
			}
		}
	}

	log.Printf("[DeleteLocalBackupMachine] 删除完成，共删除 %d 个文件", len(deletedFiles))
	return map[string]interface{}{
		"success":      true,
		"message":      "删除成功",
		"deletedFiles": deletedFiles,
		"deletedCount": len(deletedFiles),
	}
}

func createMultipartFormData(modelName string, file *os.File, fileStat os.FileInfo) (io.Reader, string) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		formFile, err := writer.CreateFormFile("file", fileStat.Name())
		if err != nil {
			log.Printf("[createMultipartFormData] 创建表单文件失败: %v", err)
			return
		}

		if _, err := io.Copy(formFile, file); err != nil {
			log.Printf("[createMultipartFormData] 复制文件内容失败: %v", err)
			return
		}

		if err := writer.WriteField("name", modelName); err != nil {
			log.Printf("[createMultipartFormData] 写入name字段失败: %v", err)
			return
		}
	}()

	return pr, writer.FormDataContentType()
}

// TemplateInfo 模板信息结构体
type TemplateInfo struct {
	ModelName string
	Version   string
	FilePath  string
}

// ParseTemplateFileName 解析模板文件名
func ParseTemplateFileName(fileName string) TemplateInfo {
	// 移除文件扩展名
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 使用下划线分割
	parts := strings.Split(nameWithoutExt, "_")

	if len(parts) < 2 {
		return TemplateInfo{
			ModelName: nameWithoutExt,
			Version:   "",
			FilePath:  "",
		}
	}

	// 前面的部分是机型名称，最后一个是版本
	modelName := strings.Join(parts[:len(parts)-1], "_")
	version := parts[len(parts)-1]

	return TemplateInfo{
		ModelName: modelName,
		Version:   version,
		FilePath:  "",
	}
}

// GetLocalTemplateInfo 获取本地模板信息
func (a *App) GetLocalTemplateInfo(modelName string) (TemplateInfo, error) {
	appDir := getStorageBaseDir()

	// 定义需要搜索的目录：model目录和custoModel目录
	dirsToSearch := []string{
		filepath.Join(appDir, "model"),
		filepath.Join(appDir, "custoModel"),
	}

	var latestTemplate TemplateInfo
	latestVersion := ""

	// 遍历所有需要搜索的目录
	for _, modelDir := range dirsToSearch {
		// 读取目录中的所有文件
		files, err := os.ReadDir(modelDir)
		if err != nil {
			// 如果目录不存在，跳过继续搜索其他目录
			continue
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
				// 创建模板信息，始终使用解析方法获取正确的机型名称和版本号
				templateInfo := ParseTemplateFileName(file.Name())
				templateInfo.FilePath = filepath.Join(modelDir, file.Name())

				// 检查是否完全匹配机型名称
				if templateInfo.ModelName == modelName {
					// 比较版本号，保留最新版本
					if templateInfo.Version > latestVersion {
						latestVersion = templateInfo.Version
						latestTemplate = templateInfo
					}
				} else if strings.Contains(modelName, "_") {
					// 如果机型名称包含下划线，逐步去掉末尾的 _xxx 后缀进行匹配
					// 例如 V2408A_16_610 依次尝试 V2408A_16、V2408A
					remaining := modelName
					for strings.Contains(remaining, "_") {
						lastIdx := strings.LastIndex(remaining, "_")
						remaining = remaining[:lastIdx]
						if templateInfo.ModelName == remaining {
							if templateInfo.Version > latestVersion {
								latestVersion = templateInfo.Version
								latestTemplate = templateInfo
							}
							break
						}
					}
				}
			}
		}
	}

	// 返回最新版本的模板
	return latestTemplate, nil
}

// NeedShowDownloadButton 检查是否需要显示下载按钮
func (a *App) NeedShowDownloadButton(modelID string, modelName string) (bool, error) {
	// 获取最新的下载URL
	downloadURL, err := a.GetTemplateDownloadURL(modelID)
	if err != nil {
		return false, err
	}

	// 解析远程文件名的版本
	remoteFileName := filepath.Base(downloadURL)
	remoteTemplate := ParseTemplateFileName(remoteFileName)

	// 获取本地模板信息
	localTemplate, err := a.GetLocalTemplateInfo(modelName)
	if err != nil {
		return true, nil // 如果获取本地模板失败，显示下载按钮
	}

	// 如果本地没有模板，显示下载按钮
	if localTemplate.FilePath == "" {
		return true, nil
	}

	// 比较本地版本和远程版本
	return localTemplate.Version != remoteTemplate.Version, nil
}

// HasLocalTemplate 检查本地是否存在指定机型的模板文件，并且本地版本和线上版本一致
func (a *App) HasLocalTemplate(modelID string, modelName string) (bool, error) {
	// 获取本地模板信息
	localTemplate, err := a.GetLocalTemplateInfo(modelName)
	if err != nil {
		return false, err
	}

	// 如果本地没有模板，不显示编辑按钮
	if localTemplate.FilePath == "" {
		return false, nil
	}

	// 获取线上模板下载URL
	downloadURL, err := a.GetTemplateDownloadURL(modelID)
	if err != nil {
		// 如果获取线上版本失败，不显示编辑按钮
		return false, nil
	}

	// 解析远程文件名的版本
	remoteFileName := filepath.Base(downloadURL)
	remoteTemplate := ParseTemplateFileName(remoteFileName)

	// 比较本地版本和远程版本，只有版本一致时才显示编辑按钮
	return localTemplate.Version == remoteTemplate.Version, nil
}

// ModelButtonStatus 单个机型的按钮状态
type ModelButtonStatus struct {
	NeedDownload bool `json:"needDownload"`
	HasLocal     bool `json:"hasLocal"`
}

// CheckModelButtonStatusBatch 批量检查所有机型的按钮状态
// 一次扫描本地目录 + 并发请求所有机型的下载URL，避免 N 次串行 IPC
func (a *App) CheckModelButtonStatusBatch(models []map[string]interface{}) map[string]interface{} {
	if len(models) == 0 {
		return map[string]interface{}{"success": true, "result": map[string]ModelButtonStatus{}}
	}

	// 1. 一次性扫描本地目录，构建 modelName → TemplateInfo Map
	appDir := getStorageBaseDir()
	dirsToSearch := []string{
		filepath.Join(appDir, "model"),
		filepath.Join(appDir, "custoModel"),
	}
	// modelName → 最新版本 TemplateInfo
	localTemplates := make(map[string]TemplateInfo)
	for _, modelDir := range dirsToSearch {
		files, err := os.ReadDir(modelDir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
				info := ParseTemplateFileName(file.Name())
				info.FilePath = filepath.Join(modelDir, file.Name())
				existing, ok := localTemplates[info.ModelName]
				if !ok || info.Version > existing.Version {
					localTemplates[info.ModelName] = info
				}
			}
		}
	}

	// 2. 并发请求所有机型的 downloadURL
	type urlResult struct {
		modelID string
		url     string
		err     error
	}
	urlCh := make(chan urlResult, len(models))
	for _, m := range models {
		modelID := fmt.Sprintf("%v", m["id"])
		go func(id string) {
			url, err := a.GetTemplateDownloadURL(id)
			urlCh <- urlResult{modelID: id, url: url, err: err}
		}(modelID)
	}

	// 收集所有 URL 结果
	urlMap := make(map[string]string, len(models))
	for range models {
		r := <-urlCh
		if r.err == nil {
			urlMap[r.modelID] = r.url
		}
	}

	// 3. 组装每个机型的按钮状态
	result := make(map[string]ModelButtonStatus, len(models))
	for _, m := range models {
		modelID := fmt.Sprintf("%v", m["id"])
		modelName := fmt.Sprintf("%v", m["name"])

		// 查找本地模板：先精确匹配，再尝试 underscore 前缀匹配
		localInfo, hasLocal := localTemplates[modelName]
		if !hasLocal && strings.Contains(modelName, "_") {
			origName := strings.Split(modelName, "_")[0]
			localInfo, hasLocal = localTemplates[origName]
		}

		downloadURL, hasURL := urlMap[modelID]
		var remoteVersion string
		if hasURL {
			remoteVersion = ParseTemplateFileName(filepath.Base(downloadURL)).Version
		}

		needDownload := true
		showEdit := false
		if hasLocal && localInfo.FilePath != "" {
			needDownload = hasURL && localInfo.Version != remoteVersion
			showEdit = hasURL && localInfo.Version == remoteVersion
		}

		result[modelID] = ModelButtonStatus{
			NeedDownload: needDownload,
			HasLocal:     showEdit,
		}
	}

	log.Printf("[CheckModelButtonStatusBatch] 批量检查 %d 个机型完成", len(models))
	return map[string]interface{}{"success": true, "result": result}
}

// GetLocalModels 获取本地存储的机型列表
func (a *App) GetLocalModels() ([]map[string]string, error) {
	appDir := getStorageBaseDir()

	// 定义需要搜索的目录：model目录和custoModel目录
	dirsToSearch := []string{
		// filepath.Join(appDir, "model"),
		filepath.Join(appDir, "custoModel"),
	}

	// 存储唯一的机型名称
	modelNames := make(map[string]bool)
	var models []map[string]string

	// 遍历所有需要搜索的目录
	for _, modelDir := range dirsToSearch {
		// 读取目录中的所有文件
		files, err := os.ReadDir(modelDir)
		if err != nil {
			// 如果目录不存在，跳过继续搜索其他目录
			continue
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
				// 获取文件名（不带后缀）
				fileNameWithoutExt := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

				// 检查是否已经存在
				if !modelNames[fileNameWithoutExt] {
					modelNames[fileNameWithoutExt] = true
					// 添加到机型列表
					models = append(models, map[string]string{
						"name": fileNameWithoutExt,
					})
				}
			}
		}
	}

	return models, nil
}

// OpenLocalModelDirectory 打开本地机型管理目录
func (a *App) OpenLocalModelDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenLocalModelDirectory 调用")

	// 目标目录: custoModel
	modelDir := filepath.Join(getStorageBaseDir(), "custoModel")

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建本地机型目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", modelDir)
	case "darwin":
		cmd = exec.Command("open", modelDir)
	default:
		cmd = exec.Command("xdg-open", modelDir)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("打开本地机型目录成功: %s", modelDir)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    modelDir,
	}
}

// OpenDownloadModelDirectory 打开线上机型下载目录
func (a *App) OpenDownloadModelDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenDownloadModelDirectory 调用")

	// 目标目录: model (与 getModelPath 保持一致)
	modelDir := filepath.Join(getStorageBaseDir(), "model")

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建线上机型下载目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", modelDir)
	case "darwin":
		cmd = exec.Command("open", modelDir)
	default:
		cmd = exec.Command("xdg-open", modelDir)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("打开线上机型下载目录成功: %s", modelDir)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    modelDir,
	}
}

// OpenKnowledgeExportDirectory 打开 edgeclient 数据目录
func (a *App) OpenKnowledgeExportDirectory() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenKnowledgeExportDirectory 调用")

	targetDir := getStorageBaseDir()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", targetDir)
	case "darwin":
		cmd = exec.Command("open", targetDir)
	default:
		cmd = exec.Command("xdg-open", targetDir)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("[KnowledgeBase] 打开数据目录: %s", targetDir)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    targetDir,
	}
}

// OpenBackupModelDir 打开本地备份机型目录
func (a *App) OpenBackupModelDir() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenBackupModelDir 调用")

	// 目标目录: BackupModel
	modelDir := filepath.Join(getStorageBaseDir(), "BackupModel")

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建本地备份机型目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", modelDir)
	case "darwin":
		cmd = exec.Command("open", modelDir)
	default:
		cmd = exec.Command("xdg-open", modelDir)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("打开本地备份机型目录成功: %s", modelDir)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    modelDir,
	}
}

// OpenBackupMachineDir 打开本地备份云机目录
func (a *App) OpenBackupMachineDir() map[string]interface{} {
	log.Printf("[IPC] 收到 OpenBackupMachineDir 调用")

	// 目标目录: cloudMachineBackup
	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建本地备份云机目录失败: %v", err),
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("C:\\Windows\\explorer.exe", backupDir)
	case "darwin":
		cmd = exec.Command("open", backupDir)
	default:
		cmd = exec.Command("xdg-open", backupDir)
	}

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开目录失败: %v", err),
		}
	}

	log.Printf("打开本地备份云机目录成功: %s", backupDir)
	return map[string]interface{}{
		"success": true,
		"message": "目录打开成功",
		"path":    backupDir,
	}
}

// GetModelConfig 获取机型配置信息
func (a *App) GetModelConfig(modelName string) (map[string]interface{}, error) {
	// 获取本地模板信息
	localTemplate, err := a.GetLocalTemplateInfo(modelName)
	if err != nil {
		return nil, err
	}

	if localTemplate.FilePath == "" {
		return nil, fmt.Errorf("no local template found for model: %s", modelName)
	}

	// 打开zip文件
	r, err := zip.OpenReader(localTemplate.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// 查找并读取cfg.json
	var cfgData []byte
	for _, f := range r.File {
		if f.Name == "cfg.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open cfg.json: %v", err)
			}
			cfgData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read cfg.json: %v", err)
			}
			break
		}
	}

	if len(cfgData) == 0 {
		return nil, fmt.Errorf("cfg.json not found in zip file")
	}

	// 解析cfg.json
	var cfgMap map[string]interface{}
	if err := json.Unmarshal(cfgData, &cfgMap); err != nil {
		return nil, fmt.Errorf("failed to parse cfg.json: %v", err)
	}

	return cfgMap, nil
}

// SDKDownloadResponse SDK下载URL响应结构体
type SDKDownloadResponse struct {
	CodeID int    `json:"code_id"`
	Msg    string `json:"msg"`
	Data   struct {
		DownloadURL string `json:"download_url"`
		SDKType     string `json:"sdk_type"`
	} `json:"data"`
}

// getUpdateSDKDir 获取updateSDK目录路径
func getUpdateSDKDir() string {
	updateSDKDir := filepath.Join(getStorageBaseDir(), "updateSDK")
	if err := os.MkdirAll(updateSDKDir, 0777); err != nil {
		panic(err)
	}
	return updateSDKDir
}

// UpgradeDeviceWithNewAPI 升级设备（新API）
func (a *App) UpgradeDeviceWithNewAPI(deviceIP interface{}, latestVersion interface{}, password interface{}) map[string]interface{} {
	log.Printf("[IPC] 收到 UpgradeDeviceWithNewAPI 调用")
	log.Printf("[IPC] 参数: deviceIP=%v, latestVersion=%v, password=%v", deviceIP, latestVersion, password)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("升级设备时发生panic: %v", r)
		}
	}()

	// 转换参数类型
	deviceIPStr, ok := deviceIP.(string)
	if !ok {
		deviceIPStr = fmt.Sprintf("%v", deviceIP)
	}

	latestVersionStr, ok := latestVersion.(string)
	if !ok {
		latestVersionStr = fmt.Sprintf("%v", latestVersion)
	}

	passwordStr, ok := password.(string)
	if !ok {
		passwordStr = fmt.Sprintf("%v", password)
	}

	// 1. 调用API获取下载URL
	log.Printf("调用API获取下载URL，版本: %s", latestVersionStr)
	sdkURL := fmt.Sprintf("https://newapi.moyunteng.com/api/v1/sdk/download-url?version=%s&filename=myt-sdk.zip&sdk_type=box_sdk", latestVersionStr)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(sdkURL)
	if err != nil {
		log.Printf("获取下载URL失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取下载URL失败: %v", err),
		}
	}
	defer resp.Body.Close()

	var sdkResp SDKDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&sdkResp); err != nil {
		log.Printf("解析下载URL响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析下载URL响应失败: %v", err),
		}
	}

	if sdkResp.CodeID != 200 {
		log.Printf("获取下载URL失败: %s", sdkResp.Msg)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取下载URL失败: %s", sdkResp.Msg),
		}
	}

	downloadURL := sdkResp.Data.DownloadURL
	log.Printf("获取下载URL成功: %s", downloadURL)

	// 2. 下载zip包到updateSDK目录
	log.Printf("开始下载SDK包: %s", downloadURL)
	updateSDKDir := getUpdateSDKDir()
	zipFilePath := filepath.Join(updateSDKDir, "myt-sdk.zip")

	// 发起下载请求
	dlResp, err := client.Get(downloadURL)
	if err != nil {
		log.Printf("下载SDK包失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载SDK包失败: %v", err),
		}
	}
	defer dlResp.Body.Close()

	// 创建文件
	out, err := os.Create(zipFilePath)
	if err != nil {
		log.Printf("创建SDK包文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建SDK包文件失败: %v", err),
		}
	}
	defer out.Close()

	// 复制文件内容
	if _, err = io.Copy(out, dlResp.Body); err != nil {
		log.Printf("保存SDK包失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存SDK包失败: %v", err),
		}
	}

	log.Printf("SDK包下载成功，保存路径: %s", zipFilePath)

	// 3. 调用设备的升级接口上传文件
	log.Printf("开始上传SDK包到设备: %s", deviceIPStr)
	uploadURL := fmt.Sprintf("http://%s/server/upgrade/upload", deviceAddr(deviceIPStr))

	// 打开要上传的文件
	file, err := os.Open(zipFilePath)
	if err != nil {
		log.Printf("打开SDK包文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开SDK包文件失败: %v", err),
		}
	}
	defer file.Close()

	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(zipFilePath))
	if err != nil {
		log.Printf("创建multipart请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建multipart请求失败: %v", err),
		}
	}

	// 拷贝文件内容
	if _, err = io.Copy(part, file); err != nil {
		log.Printf("拷贝文件内容失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("拷贝文件内容失败: %v", err),
		}
	}
	writer.Close()

	// 创建请求
	uploadReq, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		log.Printf("创建上传请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建上传请求失败: %v", err),
		}
	}

	// 设置请求头
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 添加认证头
	if passwordStr != "" {
		uploadReq.SetBasicAuth("admin", passwordStr)
	}

	// 发送请求
	uploadClient := &http.Client{Timeout: 120 * time.Second} // 增加超时时间
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		log.Printf("上传SDK包失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("上传SDK包失败: %v", err),
		}
	}
	defer uploadResp.Body.Close()

	// 读取响应
	uploadRespBody, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		log.Printf("读取上传响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取上传响应失败: %v", err),
		}
	}

	log.Printf("上传响应状态码: %d, 响应内容: %s", uploadResp.StatusCode, string(uploadRespBody))

	// 检查响应状态
	if uploadResp.StatusCode != http.StatusOK {
		// 处理设备空间不足的情况
		respStr := string(uploadRespBody)
		if strings.Contains(strings.ToLower(respStr), "no space left on device") {
			return map[string]interface{}{
				"success":   false,
				"message":   fmt.Sprintf("设备 %s 升级失败: 设备存储空间不足，请清理设备存储空间后重试", deviceIPStr),
				"errorType": "storage_full",
			}
		}
		// 处理认证失败的情况
		if uploadResp.StatusCode == http.StatusUnauthorized {
			log.Printf("上传SDK包认证失败，需要认证")
			return map[string]interface{}{
				"success":   false,
				"message":   "认证失败，请输入正确的设备密码",
				"errorType": "auth_required",
			}
		}
		return map[string]interface{}{
			"success":   false,
			"message":   fmt.Sprintf("上传SDK包失败，状态码: %d, 响应: %s", uploadResp.StatusCode, respStr),
			"errorType": "upload_failed",
		}
	}

	// 解析上传响应
	var uploadResult map[string]interface{}
	if err := json.Unmarshal(uploadRespBody, &uploadResult); err != nil {
		// 如果无法解析JSON，直接返回成功，因为有些设备可能返回非JSON响应
		log.Printf("解析上传响应失败，但状态码为200，可能是设备返回非JSON响应: %v", err)
		return map[string]interface{}{
			"success":     true,
			"message":     "SDK包上传成功，设备正在升级中...",
			"rawResponse": string(uploadRespBody),
		}
	}

	// 检查V3 API响应格式
	if code, ok := uploadResult["code"].(float64); ok {
		if code != 0 {
			message := "未知错误"
			if msg, ok := uploadResult["message"].(string); ok {
				message = msg
			}
			log.Printf("设备升级API返回错误: code=%f, message=%s", code, message)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("设备升级失败: %s", message),
				"code":    code,
			}
		}
	}

	log.Printf("设备升级成功")
	return map[string]interface{}{
		"success": true,
		"message": "设备升级成功",
		"data":    uploadResult,
	}
}

// calculateOptimalConcurrency 根据设备数量计算最优并发数
// 策略:
//   - 1-10台: 全并发
//   - 11-50台: CPU核心数 × 4
//   - 51-500台: 200并发
//   - 501-2000台: 500并发
//   - 2001-5000台: 1000并发
//   - >5000台: 1500并发
func calculateOptimalConcurrency(deviceCount int) int {
	cpuCount := runtime.NumCPU()
	
	switch {
	case deviceCount <= 10:
		return deviceCount // 小规模全并发
	case deviceCount <= 50:
		return cpuCount * 4 // 中小规模: CPU核心数×4
	case deviceCount <= 500:
		return 200 // 中等规模
	case deviceCount <= 2000:
		return 500 // 大规模
	case deviceCount <= 5000:
		return 1000 // 超大规模
	default:
		return 1500 // 海量规模
	}
}

// BatchUpgradeRequest 批量升级请求参数
type BatchUpgradeRequest struct {
	DeviceIP      string `json:"deviceIP"`
	LatestVersion string `json:"latestVersion"`
	Password      string `json:"password"`
}

// BatchUpgradeResult 批量升级结果
type BatchUpgradeResult struct {
	DeviceIP  string `json:"deviceIP"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorType string `json:"errorType,omitempty"` // auth_required, storage_full, upload_failed
}

// BatchUpgradeDevices 批量升级设备（并发处理）
func (a *App) BatchUpgradeDevices(devices interface{}) map[string]interface{} {
	log.Printf("[批量升级] 收到 BatchUpgradeDevices 调用")
	
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[批量升级] 发生panic: %v", r)
		}
	}()
	
	// ========== 1. 解析设备列表 ==========
	devicesJSON, err := json.Marshal(devices)
	if err != nil {
		log.Printf("[批量升级] ❌ 序列化设备列表失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析设备列表失败: %v", err),
		}
	}
	
	var upgradeRequests []BatchUpgradeRequest
	if err := json.Unmarshal(devicesJSON, &upgradeRequests); err != nil {
		log.Printf("[批量升级] ❌ 解析设备列表失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("解析设备列表失败: %v", err),
		}
	}
	
	if len(upgradeRequests) == 0 {
		log.Printf("[批量升级] ⚠️ 设备列表为空")
		return map[string]interface{}{
			"success": false,
			"message": "设备列表为空",
		}
	}
	
	log.Printf("[批量升级] 开始批量升级 %d 个设备", len(upgradeRequests))
	
	// ========== 2. 预下载SDK包到本地(只下载一次,所有设备共享) ==========
	// 假设所有设备升级到同一个版本
	latestVersion := upgradeRequests[0].LatestVersion
	zipFilePath, err := a.downloadSDKPackage(latestVersion)
	if err != nil {
		log.Printf("[批量升级] ❌ 预下载SDK包失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("预下载SDK包失败: %v", err),
		}
	}
	log.Printf("[批量升级] ✅ SDK包预下载成功: %s", zipFilePath)
	
	// ========== 3. 并发上传升级包到各个设备 ==========
	// 🚀 根据设备数量动态调整并发数(支持5000台设备)
	maxConcurrency := calculateOptimalConcurrency(len(upgradeRequests))
	
	log.Printf("[批量升级] 使用 %d 个并发处理 %d 个设备", maxConcurrency, len(upgradeRequests))
	
	// 创建结果切片和互斥锁
	results := make([]BatchUpgradeResult, len(upgradeRequests))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)
	
	// 统计变量
	var successCount, failCount int32
	var resultsMutex sync.Mutex
	
	startTime := time.Now()
	
	// 🚀 进度追踪(每完成10%输出一次日志,避免5000台设备时日志爆炸)
	progressInterval := len(upgradeRequests) / 10
	if progressInterval == 0 {
		progressInterval = 1
	}
	var completedCount int32
	
	// 并发处理每个设备
	for i, req := range upgradeRequests {
		wg.Add(1)
		go func(index int, request BatchUpgradeRequest) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 上传SDK包到设备
			result := a.uploadSDKToDevice(request.DeviceIP, request.Password, zipFilePath)
			
			// 保存结果
			resultsMutex.Lock()
			results[index] = result
			if result.Success {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
			resultsMutex.Unlock()
			
			// 🚀 优化日志输出: 只输出失败的设备和进度节点
			completed := atomic.AddInt32(&completedCount, 1)
			
			// 输出失败设备日志
			if !result.Success {
				log.Printf("[批量升级] ❌ [%d/%d] 设备 %s 升级失败: %s (错误类型: %s)", 
					completed, len(upgradeRequests), request.DeviceIP, 
					result.Message, result.ErrorType)
			}
			
			// 每完成10%输出进度日志
			if int(completed) % progressInterval == 0 || int(completed) == len(upgradeRequests) {
				progress := float64(completed) / float64(len(upgradeRequests)) * 100
				log.Printf("[批量升级] 📊 进度: %.1f%% (%d/%d) | 成功: %d | 失败: %d", 
					progress, completed, len(upgradeRequests), 
					atomic.LoadInt32(&successCount), atomic.LoadInt32(&failCount))
			}
			
		}(i, req)
	}
	
	// 等待所有升级完成
	wg.Wait()
	
	totalTime := time.Since(startTime).Seconds()
	log.Printf("[批量升级] ========== 批量升级完成 ==========")
	log.Printf("[批量升级] 总耗时: %.2f秒", totalTime)
	log.Printf("[批量升级] 成功: %d, 失败: %d", successCount, failCount)
	
	// ========== 4. 刷新成功设备的心跳状态(立即触发/info查询) ==========
	go func() {
		time.Sleep(2 * time.Second) // 等待2秒让设备升级完成
		log.Printf("[批量升级] 开始刷新成功设备的心跳状态...")
		
		// 🚀 批量重置心跳状态(避免5000台设备逐个加锁)
		successDevices := make([]string, 0, len(results))
		for _, result := range results {
			if result.Success {
				successDevices = append(successDevices, result.DeviceIP)
			}
		}
		
		// 批量重置LastAPICheckTime
		a.deviceStatusMutex.Lock()
		for _, deviceIP := range successDevices {
			if status := a.deviceStatusMap[deviceIP]; status != nil {
				status.LastAPICheckTime = time.Time{}      // 重置为零值
				status.LastStorageCheckTime = time.Time{} // 重置存储查询时间
			}
		}
		a.deviceStatusMutex.Unlock()
		
		// log.Printf("[批量升级] ✅ 已批量重置 %d 个成功设备的心跳状态", len(successDevices))
		
		// 🔧 立即触发API版本检查(不等待TCP Ping心跳)
		// log.Printf("[批量升级] 开始立即查询成功设备的API版本...")
		for _, deviceIP := range successDevices {
			a.checkDeviceAPIVersion(deviceIP)
			a.checkDeviceStorage(deviceIP)
		}
		// log.Printf("[批量升级] ✅ 已触发 %d 个成功设备的API版本查询", len(successDevices))
	}()
	
	// 返回结果
	return map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("批量升级完成，成功 %d 台，失败 %d 台", successCount, failCount),
		"totalDevices": len(upgradeRequests),
		"successCount": int(successCount),
		"failCount":    int(failCount),
		"totalTime":    fmt.Sprintf("%.2f秒", totalTime),
		"results":      results,
	}
}

// downloadSDKPackage 下载SDK包到本地(辅助函数)
// 优化: 检查本地是否已存在相同版本的SDK包,避免重复下载
func (a *App) downloadSDKPackage(version string) (string, error) {
	updateSDKDir := getUpdateSDKDir()
	zipFilePath := filepath.Join(updateSDKDir, "myt-sdk.zip")
	versionFilePath := filepath.Join(updateSDKDir, "myt-sdk.version")
	
	// 🔧 检查本地是否已存在SDK包
	if _, err := os.Stat(zipFilePath); err == nil {
		// SDK包存在,检查版本是否匹配
		if versionData, err := os.ReadFile(versionFilePath); err == nil {
			cachedVersion := string(versionData)
			if cachedVersion == version {
				log.Printf("[SDK下载] ✅ 使用本地缓存的SDK包(版本: %s)", version)
				return zipFilePath, nil
			}
			log.Printf("[SDK下载] 本地SDK版本(%s)与目标版本(%s)不匹配,需要重新下载", cachedVersion, version)
		}
	}
	
	// 1. 调用API获取下载URL
	log.Printf("[SDK下载] 获取下载URL，版本: %s", version)
	sdkURL := fmt.Sprintf("https://newapi.moyunteng.com/api/v1/sdk/download-url?version=%s&filename=myt-sdk.zip&sdk_type=box_sdk", version)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(sdkURL)
	if err != nil {
		return "", fmt.Errorf("获取下载URL失败: %v", err)
	}
	defer resp.Body.Close()
	
	var sdkResp SDKDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&sdkResp); err != nil {
		return "", fmt.Errorf("解析下载URL响应失败: %v", err)
	}
	
	if sdkResp.CodeID != 200 {
		return "", fmt.Errorf("获取下载URL失败: %s", sdkResp.Msg)
	}
	
	downloadURL := sdkResp.Data.DownloadURL
	log.Printf("[SDK下载] 下载URL: %s", downloadURL)
	
	// 2. 下载zip包到updateSDK目录
	dlResp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("下载SDK包失败: %v", err)
	}
	defer dlResp.Body.Close()
	
	out, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("创建SDK包文件失败: %v", err)
	}
	defer out.Close()
	
	if _, err = io.Copy(out, dlResp.Body); err != nil {
		return "", fmt.Errorf("保存SDK包失败: %v", err)
	}
	
	// 🔧 保存版本信息到文件
	if err := os.WriteFile(versionFilePath, []byte(version), 0644); err != nil {
		log.Printf("[SDK下载] ⚠️ 保存版本信息失败: %v", err)
	}
	
	log.Printf("[SDK下载] ✅ SDK包下载成功: %s (版本: %s)", zipFilePath, version)
	return zipFilePath, nil
}

// uploadSDKToDevice 上传SDK包到设备(辅助函数)
func (a *App) uploadSDKToDevice(deviceIP, password, zipFilePath string) BatchUpgradeResult {
	uploadURL := fmt.Sprintf("http://%s/server/upgrade/upload", deviceAddr(deviceIP))
	
	// 打开要上传的文件
	file, err := os.Open(zipFilePath)
	if err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("打开SDK包文件失败: %v", err),
		}
	}
	defer file.Close()
	
	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(zipFilePath))
	if err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("创建multipart请求失败: %v", err),
		}
	}
	
	if _, err = io.Copy(part, file); err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("拷贝文件内容失败: %v", err),
		}
	}
	writer.Close()
	
	// 创建请求
	uploadReq, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("创建上传请求失败: %v", err),
		}
	}
	
	// 设置请求头
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	
	// 🔧 添加认证头(处理401)
	if password != "" {
		uploadReq.SetBasicAuth("admin", password)
	}
	
	// 发送请求
	uploadClient := &http.Client{Timeout: 120 * time.Second}
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("上传SDK包失败: %v", err),
		}
	}
	defer uploadResp.Body.Close()
	
	// 读取响应
	uploadRespBody, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  false,
			Message:  fmt.Sprintf("读取上传响应失败: %v", err),
		}
	}
	
	// 🔧 处理401认证失败
	if uploadResp.StatusCode == http.StatusUnauthorized {
		log.Printf("[批量升级] 设备 %s 认证失败(401)", deviceIP)
		return BatchUpgradeResult{
			DeviceIP:  deviceIP,
			Success:   false,
			Message:   "认证失败，请输入正确的设备密码",
			ErrorType: "auth_required",
		}
	}
	
	// 检查响应状态
	if uploadResp.StatusCode != http.StatusOK {
		respStr := string(uploadRespBody)
		
		// 处理设备空间不足
		if strings.Contains(strings.ToLower(respStr), "no space left on device") {
			return BatchUpgradeResult{
				DeviceIP:  deviceIP,
				Success:   false,
				Message:   "设备存储空间不足，请清理设备存储空间后重试",
				ErrorType: "storage_full",
			}
		}
		
		return BatchUpgradeResult{
			DeviceIP:  deviceIP,
			Success:   false,
			Message:   fmt.Sprintf("上传失败，状态码: %d, 响应: %s", uploadResp.StatusCode, respStr),
			ErrorType: "upload_failed",
		}
	}
	
	// 解析上传响应
	var uploadResult map[string]interface{}
	if err := json.Unmarshal(uploadRespBody, &uploadResult); err != nil {
		// 如果无法解析JSON，直接返回成功
		return BatchUpgradeResult{
			DeviceIP: deviceIP,
			Success:  true,
			Message:  "SDK包上传成功，设备正在升级中...",
		}
	}
	
	// 检查V3 API响应格式
	if code, ok := uploadResult["code"].(float64); ok {
		if code != 0 {
			message := "未知错误"
			if msg, ok := uploadResult["message"].(string); ok {
				message = msg
			}
			return BatchUpgradeResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("设备升级失败: %s", message),
			}
		}
	}
	
	return BatchUpgradeResult{
		DeviceIP: deviceIP,
		Success:  true,
		Message:  "设备升级成功",
	}
}

// SaveModelConfig 保存机型配置信息并重新生成zip包
func (a *App) SaveModelConfig(modelName string, config map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("[IPC] 收到 SaveModelConfig 调用，modelName: %s", modelName)

	// 获取本地模板信息
	localTemplate, err := a.GetLocalTemplateInfo(modelName)
	if err != nil {
		log.Printf("获取本地模板信息失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取本地模板信息失败: %v", err),
		}, err
	}

	if localTemplate.FilePath == "" {
		log.Printf("未找到本地模板: %s", modelName)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("未找到本地模板: %s", modelName),
		}, fmt.Errorf("no local template found for model: %s", modelName)
	}

	// 创建临时目录用于解压和处理文件
	tempDir, err := os.MkdirTemp("", "model-config-")
	if err != nil {
		log.Printf("创建临时目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建临时目录失败: %v", err),
		}, err
	}
	defer os.RemoveAll(tempDir) // 确保临时目录被清理

	log.Printf("创建临时目录成功: %s", tempDir)

	// 打开原始zip文件
	originalZip, err := zip.OpenReader(localTemplate.FilePath)
	if err != nil {
		log.Printf("打开原始zip文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("打开原始zip文件失败: %v", err),
		}, err
	}
	defer originalZip.Close()

	// 解压所有文件到临时目录
	for _, f := range originalZip.File {
		// 跳过cfg.json，我们会单独处理它
		if f.Name == "cfg.json" {
			continue
		}

		// 确保目录存在
		filePath := filepath.Join(tempDir, f.Name)
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Printf("创建目录失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建目录失败: %v", err),
			}, err
		}

		// 打开文件并写入
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Printf("创建文件失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建文件失败: %v", err),
			}, err
		}

		// 读取原始文件内容
		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			log.Printf("打开源文件失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("打开源文件失败: %v", err),
			}, err
		}

		// 复制文件内容
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			log.Printf("复制文件失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("复制文件失败: %v", err),
			}, err
		}

		srcFile.Close()
		dstFile.Close()
	}
	log.Printf("解压原始文件成功")

	// 保存更新后的cfg.json到临时目录
	cfgPath := filepath.Join(tempDir, "cfg.json")
	cfgData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("序列化cfg.json失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("序列化cfg.json失败: %v", err),
		}, err
	}

	if err := os.WriteFile(cfgPath, cfgData, 0644); err != nil {
		log.Printf("写入cfg.json失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("写入cfg.json失败: %v", err),
		}, err
	}
	log.Printf("写入cfg.json成功: %s", cfgPath)

	// 准备输出目录和文件名
	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("获取用户配置目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取用户配置目录失败: %v", err),
		}, err
	}
	_ = configDir

	// 创建custoModel目录
	custoModelDir := filepath.Join(getStorageBaseDir(), "custoModel")
	if err := os.MkdirAll(custoModelDir, 0755); err != nil {
		log.Printf("创建custoModel目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建custoModel目录失败: %v", err),
		}, err
	}
	log.Printf("创建custoModel目录成功: %s", custoModelDir)

	// 构建新的zip文件名，使用带随机数后缀的新机型名称
	newFileName := fmt.Sprintf("%s.zip", modelName)
	newZipPath := filepath.Join(custoModelDir, newFileName)
	log.Printf("新zip文件路径: %s", newZipPath)

	// 创建新的zip文件
	newZipFile, err := os.Create(newZipPath)
	if err != nil {
		log.Printf("创建新zip文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建新zip文件失败: %v", err),
		}, err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// 将临时目录中的所有文件添加到新zip中
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 计算相对于临时目录的路径
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 在zip文件中创建条目
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// 复制文件内容到zip条目
		if _, err := io.Copy(zipEntry, srcFile); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("创建新zip文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建新zip文件失败: %v", err),
		}, err
	}

	// 确保zipWriter被正确关闭
	if err := zipWriter.Close(); err != nil {
		log.Printf("关闭zipWriter失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("关闭zipWriter失败: %v", err),
		}, err
	}

	log.Printf("创建新zip文件成功: %s", newZipPath)

	return map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("机型配置保存成功，新文件已生成: %s", newFileName),
		"filePath": newZipPath,
		"fileName": newFileName,
	}, nil
}

// PushModelToDevices 推送机型配置到选定的设备
func (a *App) PushModelToDevices(params map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("[IPC] 收到 PushModelToDevices 调用，params: %+v", params)

	// 解析参数
	modelName, ok := params["modelName"].(string)
	if !ok {
		log.Printf("无效的modelName参数: %v", params["modelName"])
		return map[string]interface{}{
			"success": false,
			"message": "无效的modelName参数",
		}, fmt.Errorf("无效的modelName参数")
	}

	devices, ok := params["devices"].([]interface{})
	if !ok {
		log.Printf("无效的devices参数: %v", params["devices"])
		return map[string]interface{}{
			"success": false,
			"message": "无效的devices参数",
		}, fmt.Errorf("无效的devices参数")
	}

	if len(devices) == 0 {
		log.Printf("未选择任何设备")
		return map[string]interface{}{
			"success": false,
			"message": "未选择任何设备",
		}, fmt.Errorf("未选择任何设备")
	}

	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("获取用户配置目录失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取用户配置目录失败: %v", err),
		}, err
	}
	_ = configDir

	// 构建custoModel目录下的zip文件名和路径，直接使用机型名称作为文件名
	custoModelDir := filepath.Join(getStorageBaseDir(), "custoModel")
	zipFileName := modelName + ".zip"
	zipFilePath := filepath.Join(custoModelDir, zipFileName)
	log.Printf("zip文件路径: %s", zipFilePath)

	// 检查zip文件是否存在
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		log.Printf("zip文件不存在: %s", zipFilePath)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("zip文件不存在: %s", zipFileName),
		}, fmt.Errorf("zip文件不存在: %s", zipFilePath)
	}

	// 读取zip文件内容到内存，避免多次打开文件
	zipData, err := os.ReadFile(zipFilePath)
	if err != nil {
		log.Printf("读取zip文件失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取zip文件失败: %v", err),
		}, err
	}
	log.Printf("成功读取zip文件，大小: %d bytes", len(zipData))

	// 定义结果结构
	type PushResult struct {
		DeviceIP string `json:"device_ip"`
		Success  bool   `json:"success"`
		Message  string `json:"message"`
	}

	results := make([]PushResult, 0, len(devices))
	successCount := 0

	// 遍历设备列表，逐个推送整个zip文件
	for _, device := range devices {
		deviceMap, ok := device.(map[string]interface{})
		if !ok {
			log.Printf("无效的设备信息: %v", device)
			results = append(results, PushResult{
				DeviceIP: "",
				Success:  false,
				Message:  "无效的设备信息",
			})
			continue
		}

		// 获取设备IP
		deviceIP, ok := deviceMap["ip"].(string)
		if !ok {
			log.Printf("无效的设备IP: %v", deviceMap["ip"])
			results = append(results, PushResult{
				DeviceIP: "",
				Success:  false,
				Message:  "无效的设备IP",
			})
			continue
		}

		// 获取设备认证信息（密码）
		a.devicePasswordsMutex.RLock()
		password := a.devicePasswords[deviceIP]
		a.devicePasswordsMutex.RUnlock()

		// 构建推送URL
		pushURL := fmt.Sprintf("http://%s/phoneModel/import", deviceAddr(deviceIP))
		log.Printf("推送URL: %s", pushURL)

		// 创建multipart/form-data请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件字段，直接推送整个zip文件
		fileField, err := writer.CreateFormFile("file", zipFileName)
		if err != nil {
			log.Printf("创建表单文件失败: %v", err)
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("创建表单文件失败: %v", err),
			})
			continue
		}

		// 直接写入zip文件内容到表单字段，不打开压缩包
		if _, err := fileField.Write(zipData); err != nil {
			log.Printf("写入zip文件内容失败: %v", err)
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("写入zip文件内容失败: %v", err),
			})
			continue
		}

		// 关闭writer，完成multipart/form-data构建
		writer.Close()

		// 创建HTTP请求
		req, err := http.NewRequest("POST", pushURL, body)
		if err != nil {
			log.Printf("创建HTTP请求失败: %v", err)
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("创建HTTP请求失败: %v", err),
			})
			continue
		}

		// 设置Content-Type头
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// 添加认证头（如果有密码）
		if password != "" {
			req.SetBasicAuth("admin", password)
			log.Printf("设备 %s 已添加认证头", deviceIP)
		} else {
			log.Printf("警告: 设备 %s 未获取到密码，可能导致401错误", deviceIP)
		}

		// 发送HTTP请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("发送HTTP请求失败: %v", err)
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("发送HTTP请求失败: %v", err),
			})
			continue
		}
		defer resp.Body.Close()

		// 读取响应内容
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取响应内容失败: %v", err)
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("读取响应内容失败: %v", err),
			})
			continue
		}

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			log.Printf("推送失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
			results = append(results, PushResult{
				DeviceIP: deviceIP,
				Success:  false,
				Message:  fmt.Sprintf("推送失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody)),
			})
			continue
		}

		// 推送成功
		log.Printf("推送成功，设备IP: %s, 包: %s, 响应: %s", deviceIP, zipFileName, string(respBody))
		results = append(results, PushResult{
			DeviceIP: deviceIP,
			Success:  true,
			Message:  fmt.Sprintf("推送成功，响应: %s", string(respBody)),
		})
		successCount++
	}

	// 返回推送结果
	return map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("推送完成，成功 %d 台设备，失败 %d 台设备", successCount, len(devices)-successCount),
		"totalDevices": len(devices),
		"successCount": successCount,
		"failCount":    len(devices) - successCount,
		"results":      results,
	}, nil
}

// GetAppDirs 获取应用的各种目录
func (a *App) GetAppDirs() map[string]string {
	log.Printf("[IPC] 收到 GetAppDirs 调用")
	configDir := getConfigDir()
	cacheDir := getCacheDir()
	dataDir := getDataDir()
	sharedDir := getSharedDir()

	result := map[string]string{
		"config": configDir,
		"cache":  cacheDir,
		"data":   dataDir,
		"shared": sharedDir,
	}

	log.Printf("[IPC] GetAppDirs 返回: %+v", result)
	return result
}

// GetUserDataDir 获取用户数据目录（兼容旧接口）
func (a *App) GetUserDataDir() string {
	log.Printf("[IPC] 收到 GetUserDataDir 调用")
	userDataDir := getDataDir()
	log.Printf("[IPC] GetUserDataDir 返回: %s", userDataDir)
	return userDataDir
}

// getDataDir 获取应用数据目录
func getDataDir() string {
	var baseDir string
	var err error

	switch runtime.GOOS {
	case "windows":
		// Windows: %LocalAppData%/应用名
		baseDir, err = os.UserCacheDir()
	case "darwin":
		// macOS: ~/Library/Application Support/应用名
		homeDir, err := os.UserHomeDir()
		if err == nil {
			baseDir = filepath.Join(homeDir, "Library", "Application Support")
		}
	default:
		// Linux: ~/.local/share/应用名
		homeDir, err := os.UserHomeDir()
		if err == nil {
			baseDir = filepath.Join(homeDir, ".local", "share")
		}
	}

	if err != nil {
		log.Printf("获取基础目录失败: %v, 使用默认目录", err)
		baseDir = os.TempDir()
	}

	// 创建应用数据目录
	dataDir := filepath.Join(baseDir, appName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("创建应用数据目录失败: %v", err)
		return ""
	}

	return dataDir
}

// getConfigDir 获取应用配置目录
func getConfigDir() string {
	var configDir string
	var err error

	// 尝试使用标准配置目录
	configDir, err = os.UserConfigDir()
	if err != nil {
		log.Printf("获取配置目录失败: %v, 使用数据目录作为备选", err)
		return getDataDir()
	}

	// 创建应用配置目录
	configDir = filepath.Join(configDir, appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("创建应用配置目录失败: %v", err)
		return getDataDir()
	}

	return configDir
}

// getCacheDir 获取应用缓存目录
func getCacheDir() string {
	var cacheDir string
	var err error

	// 尝试使用标准缓存目录
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		log.Printf("获取缓存目录失败: %v, 使用数据目录作为备选", err)
		return getDataDir()
	}

	// 创建应用缓存目录
	cacheDir = filepath.Join(cacheDir, appName)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("创建应用缓存目录失败: %v", err)
		return getDataDir()
	}

	return cacheDir
}

// getSharedDir 获取应用共享目录
func getSharedDir() string {
	// 共享目录存放在数据目录下
	dataDir := getDataDir()
	sharedDir := filepath.Join(dataDir, "shared")

	// 确保共享目录存在
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		log.Printf("创建共享目录失败: %v", err)
		return ""
	}

	return sharedDir
}

// GetImages 获取镜像列表
func (a *App) GetImages(deviceIP string, version string, password string) interface{} {
	log.Printf("[IPC] 收到 GetImages 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s", deviceIP, version)

	// 调用设备API获取镜像列表
	images, err := getImages(deviceIP, version, password)
	if err != nil {
		log.Printf("[IPC] 获取镜像列表失败: %v", err)
		// 失败时返回空结果
		if version == "v3" {
			return map[string]interface{}{"list": []interface{}{}}
		} else {
			return []map[string]interface{}{}
		}
	}

	log.Printf("[IPC] GetImages 返回结果: %+v", images)
	return images
}

// getImages 调用设备API获取镜像列表
func getImages(deviceIP string, version string, password string) (interface{}, error) {
	// 无论什么版本，都使用Docker API获取镜像列表
	// 因为V3 API调用失败，错误信息：解析V3 API响应失败: invalid character 'N' looking for beginning of value
	return getDockerImages(deviceIP, version, password)
}

// getV3Images 获取V3设备的镜像列表
func getV3Images(deviceIP string) (interface{}, error) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/android/images", deviceAddr(deviceIP))

	// 发送HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取V3 API响应失败: %v", err)
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析V3 API响应失败: %v", err)
	}

	log.Printf("V3 API镜像列表响应: %s", string(body))
	return result, nil
}

// getDockerImages 获取Docker镜像列表
func getDockerImages(deviceIP string, version string, password string) (interface{}, error) {
	resp, err := callDockerAPI(deviceIP, version, "/images/json?all=true", "GET", nil, password)
	if err != nil {
		return nil, fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	// 解析JSON响应
	var images []map[string]interface{}
	if err := json.Unmarshal(body, &images); err != nil {
		return nil, fmt.Errorf("解析Docker API响应失败: %v", err)
	}

	// 处理Docker API返回的镜像数据，转换为前端期望的格式
	var result []map[string]interface{}
	for _, img := range images {
		imageInfo := map[string]interface{}{
			"imageUrl":   img["RepoTags"],
			"Image":      img["RepoTags"],
			"size":       img["Size"],
			"createTime": img["Created"],
		}
		result = append(result, imageInfo)
	}
	return result, nil
}

// 镜像信息结构体
type ImageInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type" json:"ttype"`
	Version     string `json:"version"`
	Size        string `json:"size"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Url         string `json:"url"`
}

// 响应结构体
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// DownloadImageToLocal 下载镜像到本地缓存目录
// getProtocol 根据主机地址选择HTTP或HTTPS协议
func getProtocol(tag *ImageTag) string {
	if net.ParseIP(strings.Split(tag.Host, ":")[0]) != nil {
		return "http"
	} else {
		return "https"
	}
}

// fetchHash 获取镜像哈希值
func fetchHash(this *ImageTag) error {
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", getProtocol(this), this.Host, this.Name, this.Version)

	// 创建HTTP请求
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			if name == "Etag" {
				this.Hash = strings.Trim(value, `"`)
				return nil
			}
		}
	}

	return errors.New("not found head")
}

// doAuth 处理Docker Registry认证
func doAuth(this *ImageTag, info string) (*Manifest, error) {
	// 解析Www-Authenticate头信息
	params := strings.Split(strings.TrimPrefix(info, "Bearer "), ",")

	trimParam := func(input, key string) string {
		input = strings.TrimSpace(input)
		if strings.HasPrefix(input, key) {
			value := strings.TrimPrefix(input, key)
			return strings.Trim(value, `"`)
		}
		return ""
	}

	var realm, service, scope string
	for _, param := range params {
		switch {
		case strings.Contains(param, "realm="):
			realm = trimParam(param, "realm=")
		case strings.Contains(param, "service="):
			service = trimParam(param, "service=")
		case strings.Contains(param, "scope="):
			scope = trimParam(param, "scope=")
		}
	}

	if realm == "" || service == "" || scope == "" {
		return nil, errors.New("info parse fail, " + info)
	}

	url := fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &this.AuthInfo)
	if err != nil {
		return nil, err
	}

	return fetchManifest(this)
}

// fetchManifest 下载镜像清单
func fetchManifest(this *ImageTag) (*Manifest, error) {
	hash := this.Hash
	if hash == "" {
		hash = this.Version
	}
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", getProtocol(this), this.Host, this.Name, hash)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if this.AuthInfo.Token != "" {
		req.Header.Add("Authorization", "Bearer "+this.AuthInfo.Token)
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		for name, values := range resp.Header {
			for _, value := range values {
				if name == "Www-Authenticate" && !this.Authed {
					this.Authed = true
					return doAuth(this, value)
				}
			}
		}
		return nil, fmt.Errorf("failed to fetch manifest with status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

// parseImageTag 解析镜像标签
func parseImageTag(tag string) *ImageTag {
	parts := strings.SplitAfterN(tag, "/", 2)
	result := ImageTag{}
	if len(parts) != 2 {
		// 处理格式不正确的镜像标签，尝试作为完整的镜像名称
		result.Host = "registry-1.docker.io"
		if strings.Contains(tag, ":") {
			imageParts := strings.SplitN(tag, ":", 2)
			result.Name = imageParts[0]
			result.Version = imageParts[1]
		} else {
			result.Name = tag
			result.Version = "latest"
		}
		return &result
	}

	result.Host = strings.Trim(parts[0], "/")

	imageParts := strings.SplitN(parts[1], ":", 2)
	if len(imageParts) == 2 {
		result.Name = imageParts[0]
		result.Version = imageParts[1]
	} else {
		result.Name = imageParts[0]
		result.Version = "latest"
	}

	return &result
}

// addTarEntry 向tar文件添加条目
func addTarEntry(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name: name,
		Size: int64(len(data)),
		Mode: 0600,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write(data)
	return err
}

// downloadAndAddToTar 下载文件并添加到tar文件
func downloadAndAddToTar(tw *tar.Writer, imageTag *ImageTag, descriptor Descriptor, tarName string, progressCallback func(float64)) error {
	url := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", getProtocol(imageTag), imageTag.Host, imageTag.Name, descriptor.Digest)

	req, _ := http.NewRequest("GET", url, nil)

	if imageTag.AuthInfo.Token != "" {
		req.Header.Add("Authorization", "Bearer "+imageTag.AuthInfo.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download blob with status code: %d", resp.StatusCode)
	}

	// 添加到tar文件
	header := &tar.Header{
		Name: tarName,
		Size: descriptor.Size,
		Mode: 0600,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// 创建带进度的Reader
	totalBytes := descriptor.Size

	// 创建进度Reader
	progressReader := &ProgressReader{
		Reader:    resp.Body,
		Size:      totalBytes,
		ReadBytes: 0,
		Callback: func(progress float64) {
			if progressCallback != nil {
				progressCallback(progress)
			}
		},
	}

	_, err = io.Copy(tw, progressReader)
	return err
}

// downloadDockerImage 使用纯Go下载Docker镜像
func downloadDockerImage(tag, outputTar string, progressCallback func(float64)) error {
	imageTag := parseImageTag(tag)
	if imageTag == nil {
		return fmt.Errorf("invalid image tag format: %s", tag)
	}

	// 对于阿里云镜像，跳过fetchHash
	if !strings.Contains(imageTag.Host, "aliyuncs.com") {
		if err := fetchHash(imageTag); err != nil {
			log.Println("fetchHash failed:", err)
		}
	}

	// 下载镜像清单
	manifest, err := fetchManifest(imageTag)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// 创建tar文件
	tarFile, err := os.Create(outputTar)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	configFileName := strings.Trim(manifest.Config.Digest, "sha256:") + ".json"

	// 下载并添加配置文件
	if err := downloadAndAddToTar(tw, imageTag, manifest.Config, configFileName, func(progress float64) {
		if progressCallback != nil {
			// 配置文件占10%的进度
			progressCallback(progress * 0.1)
		}
	}); err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}

	layers := make([]string, 0)
	layerCount := len(manifest.Layers)

	// 下载并添加层文件
	for i, layer := range manifest.Layers {
		layerFileName := strings.Trim(layer.Digest, "sha256:") + "/layer.tar"
		layers = append(layers, layerFileName)

		if err := downloadAndAddToTar(tw, imageTag, layer, layerFileName, func(progress float64) {
			if progressCallback != nil {
				// 层文件占90%的进度，平均分配给每个层
				layerProgress := (progress * 0.9) / float64(layerCount)
				// 当前层的起始进度
				startProgress := 0.1 + (float64(i) * (0.9 / float64(layerCount)))
				// 计算总进度
				totalProgress := startProgress + layerProgress
				progressCallback(totalProgress)
			}
		}); err != nil {
			return fmt.Errorf("failed to download layer: %w", err)
		}
	}

	// 创建manifest.json
	manifests := make([]ManifestFile, 1)
	manifests[0].Config = configFileName
	manifests[0].RepoTags = []string{tag}
	manifests[0].Layers = layers

	data, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	// 添加manifest.json到tar文件
	if err := addTarEntry(tw, "manifest.json", data); err != nil {
		return err
	}

	return nil
}

// DownloadImageToLocal 下载镜像到本地
func (a *App) DownloadImageToLocal(image ImageInfo) Response {
	// 获取缓存目录
	cacheDir := getCacheDir()

	// 创建镜像存储目录
	imageDir := filepath.Join(cacheDir, "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建镜像目录失败: %v", err),
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("创建镜像目录失败: %v", err)}
	}

	// 检查镜像名称是否为空
	if image.Name == "" {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": "镜像名称不能为空",
			"path":    "",
		})
		return Response{Success: false, Message: "镜像名称不能为空"}
	}

	// 构建本地文件路径，使用镜像URL中的镜像名称作为文件名
	// 从URL中提取镜像名称，如：registry.cn-guangzhou.aliyuncs.com/mytos/dobox:P14_v3_all_202601091434
	imageName := image.Url
	// 移除协议部分（如http://或https://）
	imageName = strings.TrimPrefix(imageName, "http://")
	imageName = strings.TrimPrefix(imageName, "https://")
	// 替换文件名中不允许的字符为下划线
	imageName = strings.ReplaceAll(imageName, ":", "_")
	imageName = strings.ReplaceAll(imageName, "?", "_")
	imageName = strings.ReplaceAll(imageName, "&", "_")
	imageName = strings.ReplaceAll(imageName, "=", "_")
	imageName = strings.ReplaceAll(imageName, "#", "_")
	localFilePath := filepath.Join(imageDir, fmt.Sprintf("%s.tar.gz", imageName))

	// 检查镜像URL是否为空
	if image.Url == "" {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": "镜像URL不能为空",
			"path":    "",
		})
		return Response{Success: false, Message: "镜像URL不能为空"}
	}

	// 发送初始进度事件（0%）
	a.progressMutex.Lock()
	a.downloadProgress = 0
	progress := a.downloadProgress
	a.progressMutex.Unlock()
	a.emitEvent("download-progress", map[string]interface{}{
		"progress": progress,
	})

	// 判断是否为Docker镜像仓库地址（没有http/https前缀）
	if !strings.HasPrefix(image.Url, "http://") && !strings.HasPrefix(image.Url, "https://") {
		// 使用纯Go语言实现Docker镜像拉取功能
		log.Printf("使用纯Go语言拉取Docker镜像: %s", image.Url)

		// 更新进度为10%（开始拉取镜像）
		a.progressMutex.Lock()
		a.downloadProgress = 10
		progress = a.downloadProgress
		a.progressMutex.Unlock()
		a.emitEvent("download-progress", map[string]interface{}{
			"progress": progress,
		})

		// 下载Docker镜像，并传递进度回调函数
		err := downloadDockerImage(image.Url, localFilePath, func(progress float64) {
			// 计算当前进度百分比，范围10%-95%
			downloadProgress := 10 + (progress * 85) // 10% 到 95%
			if downloadProgress > 95 {
				downloadProgress = 95
			}

			a.progressMutex.Lock()
			a.downloadProgress = downloadProgress
			currentProgress := a.downloadProgress
			a.progressMutex.Unlock()

			// 发送下载进度事件给前端
			a.emitEvent("download-progress", map[string]interface{}{
				"progress": currentProgress,
			})
		})
		if err != nil {
			// 发送下载完成事件（失败）
			a.emitEvent("download-complete", map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("拉取Docker镜像失败: %v", err),
				"path":    "",
			})
			return Response{Success: false, Message: fmt.Sprintf("拉取Docker镜像失败: %v", err)}
		}

		// 更新进度为95%（拉取完成，正在处理文件）
		a.progressMutex.Lock()
		a.downloadProgress = 95
		progress = a.downloadProgress
		a.progressMutex.Unlock()
		a.emitEvent("download-progress", map[string]interface{}{
			"progress": progress,
		})

		// 检查文件是否存在且大小大于0
		if info, err := os.Stat(localFilePath); err == nil && info.Size() > 0 {
			// 更新进度为100%（完成）
			a.progressMutex.Lock()
			a.downloadProgress = 100
			progress = a.downloadProgress
			a.progressMutex.Unlock()
			a.emitEvent("download-progress", map[string]interface{}{
				"progress": progress,
			})

			log.Printf("Docker镜像下载成功: %s", localFilePath)

			// 保存镜像元数据
			metadataPath := filepath.Join(imageDir, strings.TrimSuffix(filepath.Base(localFilePath), ".tar.gz")+".json")
			metadata := ImageMetadata{
				OnlineURL:       image.Url,
				AvailableModels: []string{}, // 可以根据需要添加可用机型
			}

			metadataBytes, err := json.Marshal(metadata)
			if err != nil {
				log.Printf("保存镜像元数据失败: %v", err)
			} else {
				if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
					log.Printf("写入镜像元数据文件失败: %v", err)
				} else {
					log.Printf("镜像元数据保存成功: %s", metadataPath)
				}
			}

			// 发送下载完成事件（成功）
			a.emitEvent("download-complete", map[string]interface{}{
				"success": true,
				"message": "下载镜像成功",
				"path":    localFilePath,
			})

			return Response{Success: true, Message: fmt.Sprintf("镜像下载成功: %s\n\n保存路径: %s", image.Url, localFilePath)}
		}

		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": "Docker镜像下载失败: 生成的文件无效或为空",
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("Docker镜像下载失败: 生成的文件无效或为空")}
	}

	// 普通HTTP/HTTPS URL，直接下载
	log.Printf("使用HTTP GET下载镜像: %s", image.Url)

	// 更新进度为10%（开始下载）
	a.progressMutex.Lock()
	a.downloadProgress = 10
	progress = a.downloadProgress
	a.progressMutex.Unlock()
	a.emitEvent("download-progress", map[string]interface{}{
		"progress": progress,
	})

	// 下载文件
	resp, err := http.Get(image.Url)
	if err != nil {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载镜像失败: %v", err),
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("下载镜像失败: %v", err)}
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("下载镜像失败: HTTP %d, 响应: %s", resp.StatusCode, string(body)),
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("下载镜像失败: HTTP %d, 响应: %s", resp.StatusCode, string(body))}
	}

	// 获取文件大小
	fileSize := resp.ContentLength

	// 创建本地文件
	out, err := os.Create(localFilePath)
	if err != nil {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建本地文件失败: %v", err),
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("创建本地文件失败: %v", err)}
	}
	defer out.Close()

	// 创建带进度的Reader
	progressReader := &ProgressReader{
		Reader:    resp.Body,
		Size:      fileSize,
		ReadBytes: 0,
		Callback: func(progress float64) {
			// 计算进度百分比，范围10%-95%
			downloadProgress := 10 + (progress * 0.85) // 10% 到 95%
			if downloadProgress > 95 {
				downloadProgress = 95
			}

			a.progressMutex.Lock()
			a.downloadProgress = downloadProgress
			currentProgress := a.downloadProgress
			a.progressMutex.Unlock()

			// 发送下载进度事件给前端
			a.emitEvent("download-progress", map[string]interface{}{
				"progress": currentProgress,
			})
		},
	}

	// 写入文件
	_, err = io.Copy(out, progressReader)
	if err != nil {
		// 发送下载完成事件（失败）
		a.emitEvent("download-complete", map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存镜像失败: %v", err),
			"path":    "",
		})
		return Response{Success: false, Message: fmt.Sprintf("保存镜像失败: %v", err)}
	}

	// 更新进度为100%（完成）
	a.progressMutex.Lock()
	a.downloadProgress = 100
	progress = a.downloadProgress
	a.progressMutex.Unlock()
	a.emitEvent("download-progress", map[string]interface{}{
		"progress": progress,
	})

	log.Printf("HTTP镜像下载成功: %s", localFilePath)

	// 发送下载完成事件（成功）
	a.emitEvent("download-complete", map[string]interface{}{
		"success": true,
		"message": "下载镜像成功",
		"path":    localFilePath,
	})

	return Response{Success: true, Message: fmt.Sprintf("镜像下载成功: %s\n\n保存路径: %s", image.Url, localFilePath)}
}

// ImportImageToSystem 导入镜像到系统
func (a *App) ImportImageToSystem(image ImageInfo) Response {
	// 构建请求URL
	url := "http://10.10.11.3:8000/android/image/import"

	// 打开本地镜像文件
	file, err := os.Open(image.Path)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("打开镜像文件失败: %v", err)}
	}
	defer file.Close()

	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	fileField, err := writer.CreateFormFile("file", filepath.Base(image.Path))
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("创建表单文件失败: %v", err)}
	}

	// 复制文件内容到表单
	_, err = io.Copy(fileField, file)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("复制文件内容失败: %v", err)}
	}

	// 添加其他字段
	_ = writer.WriteField("file", "1")

	// 关闭writer
	writer.Close()

	// 创建请求
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("创建请求失败: %v", err)}
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("发送请求失败: %v", err)}
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("读取响应失败: %v", err)}
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return Response{Success: false, Message: fmt.Sprintf("导入失败: %s, %s", resp.Status, string(body))}
	}

	return Response{Success: true, Message: "镜像导入成功"}
}

// PushImageToDevice 推送镜像到设备
func (a *App) PushImageToDevice(data map[string]interface{}) Response {
	log.Printf("收到推送镜像到设备请求: %+v", data)

	// 获取镜像信息
	imageInfo, ok := data["image"].(map[string]interface{})
	if !ok {
		return Response{Success: false, Message: "无效的镜像信息"}
	}

	// 获取设备信息
	deviceInfo, ok := data["device"].(map[string]interface{})
	if !ok {
		return Response{Success: false, Message: "无效的设备信息"}
	}

	// 获取镜像路径
	imagePath, ok := imageInfo["path"].(string)
	if !ok || imagePath == "" {
		return Response{Success: false, Message: "无效的镜像路径"}
	}

	// 获取设备IP
	deviceIP, ok := deviceInfo["ip"].(string)
	if !ok || deviceIP == "" {
		return Response{Success: false, Message: "无效的设备IP"}
	}

	// 构建请求URL
	url := fmt.Sprintf("http://%s/android/image/import", deviceAddr(deviceIP))
	log.Printf("推送镜像到设备: %s, URL: %s", deviceIP, url)

	// 打开本地镜像文件
	file, err := os.Open(imagePath)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("打开镜像文件失败: %v", err)}
	}
	defer file.Close()

	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	fileField, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("创建表单文件失败: %v", err)}
	}

	// 复制文件内容到表单
	_, err = io.Copy(fileField, file)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("复制文件内容失败: %v", err)}
	}

	// 添加其他字段
	_ = writer.WriteField("file", "1")

	// 关闭writer
	writer.Close()

	// 创建请求
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("创建请求失败: %v", err)}
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("发送请求失败: %v", err)}
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{Success: false, Message: fmt.Sprintf("读取响应失败: %v", err)}
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return Response{Success: false, Message: fmt.Sprintf("推送失败: %s, %s", resp.Status, string(body))}
	}

	return Response{Success: true, Message: fmt.Sprintf("镜像成功推送到设备 %s", deviceIP)}
}

// DeleteLocalImageFromManagement 删除本地镜像
func (a *App) DeleteLocalImageFromManagement(image ImageInfo) Response {
	// 删除本地文件
	if err := os.Remove(image.Path); err != nil {
		return Response{Success: false, Message: fmt.Sprintf("删除镜像文件失败: %v", err)}
	}

	return Response{Success: true, Message: "镜像删除成功"}
}

// GetLocalImagesForManagement 获取本地镜像列表
func (a *App) GetLocalImagesForManagement() Response {
	// 获取缓存目录
	cacheDir := getCacheDir()
	imageDir := filepath.Join(cacheDir, "images")

	// 读取目录中的镜像文件
	files, err := os.ReadDir(imageDir)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在，返回空列表
			return Response{Success: true, Data: []ImageInfo{}}
		}
		return Response{Success: false, Message: fmt.Sprintf("读取镜像目录失败: %v", err)}
	}

	// 构建镜像列表
	var images []ImageInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
			// 获取文件信息
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			// 解析文件名获取镜像信息
			fileName := strings.TrimSuffix(file.Name(), ".tar.gz")

			// 构建镜像信息
			image := ImageInfo{
				Name:    fileName,
				Type:    "android", // 默认类型
				Version: "1.0",     // 默认版本
				Size:    fmt.Sprintf("%.2f MB", float64(fileInfo.Size())/(1024*1024)),
				Path:    filepath.Join(imageDir, file.Name()),
			}

			images = append(images, image)
		}
	}

	return Response{Success: true, Data: images}
}

// GetDockerNetworks 获取Docker网络列表
func (a *App) GetDockerNetworks(deviceIP string, version string, password string) interface{} {
	log.Printf("[IPC] 收到 GetDockerNetworks 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s", deviceIP, version)

	// 调用真实的Docker网络获取逻辑
	networks, err := getDockerNetworks(deviceIP, version, password)
	if err != nil {
		log.Printf("[IPC] 获取Docker网络列表失败: %v", err)
		// 失败时返回空结果
		return []interface{}{}
	}

	// log.Printf("[IPC] GetDockerNetworks 返回结果: %+v", networks)
	return networks
}

// CheckMytSdkContainer 检查myt_sdk容器是否存在
func (a *App) CheckMytSdkContainer(deviceIP string, version string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 CheckMytSdkContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s", deviceIP, version)

	// 调用Docker API检查myt_sdk容器是否存在
	containerExists, _, err := checkMytSdkContainerExists(deviceIP, version, password)
	if err != nil {
		log.Printf("[IPC] 检查myt_sdk容器失败: %v", err)
		return map[string]interface{}{
			"exists":  false,
			"success": false,
			"message": err.Error(),
		}
	}

	result := map[string]interface{}{
		"exists":  containerExists,
		"success": true,
		"message": "检查完成",
	}
	log.Printf("[IPC] CheckMytSdkContainer 返回结果: %+v", result)
	return result
}

// checkMytSdkContainerExists 检查myt_sdk容器是否存在以及运行状态
func checkMytSdkContainerExists(deviceIP string, version string, password string) (bool, bool, error) {
	// 参数验证
	if deviceIP == "" {
		return false, false, fmt.Errorf("设备IP不能为空")
	}

	// 调用Docker API获取容器列表
	containers, err := getDockerContainers(deviceIP, version, password)
	if err != nil {
		return false, false, err
	}

	// 遍历容器列表，检查是否存在确切名称为/myt_sdk的容器
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/myt_sdk" {
				// 检查容器状态是否为运行中（状态字符串包含"Up"表示运行中）
				isRunning := strings.Contains(container.Status, "Up")
				return true, isRunning, nil
			}
		}
	}

	return false, false, nil
}

// SyncAuthorization 同步授权
// 当点击同步授权时调用，发送token到本地UDP服务
func (a *App) SyncAuthorization(token string, deviceID string) map[string]interface{} {
	log.Printf("[IPC] 收到 SyncAuthorization 调用")
	log.Printf("[IPC] 参数: token=%s, deviceID=%s", token, deviceID)

	// 创建UDP连接
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		log.Printf("[IPC] 创建UDP连接失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": "创建UDP连接失败: " + err.Error(),
		}
	}
	defer conn.Close()

	// 构建发送参数
	param := "lgtoken:" + token
	log.Printf("[IPC] 发送UDP消息: %s", param)

	// 发送消息到本地UDP服务
	_, err = conn.WriteToUDP([]byte(param), &net.UDPAddr{
		IP:   net.ParseIP(deviceID),
		Port: 7678,
	})
	if err != nil {
		log.Printf("[IPC] 发送UDP消息失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": "发送UDP消息失败: " + err.Error(),
		}
	}

	log.Printf("[IPC] 同步授权成功")
	return map[string]interface{}{
		"success": true,
		"message": "同步授权成功",
	}
}

// GetPhoneVCode 获取手机验证码
func (a *App) GetPhoneVCode(phone string, token string) map[string]interface{} {
	log.Printf("[IPC] 收到 GetPhoneVCode 调用")
	log.Printf("[IPC] 参数: phone=%s", phone)

	// 构造请求数据
	data := map[string]string{
		"act":   "com",
		"phone": phone,
		"plat":  "",
		"token": token,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{
			"code":    -1,
			"message": "JSON编码失败: " + err.Error(),
		}
	}

	// 构造POST表单数据
	formData := url.Values{}
	formData.Set("data", string(jsonData))
	formData.Set("type", "get_phone_vcode")

	// 发送POST请求
	req, err := http.NewRequest("POST", "https://www.moyunteng.com/api/api.php", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[IPC] 创建请求失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "创建请求失败: " + err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[IPC] 请求验证码接口失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "请求验证码接口失败: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IPC] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "读取响应失败: " + err.Error(),
		}
	}

	log.Printf("[IPC] 验证码接口响应: %s", string(body))

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[IPC] 解析响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "解析响应失败: " + err.Error(),
		}
	}

	return result
}

// Register 用户注册
func (a *App) Register(phone string, password string, vcode string, vkey string) map[string]interface{} {
	log.Printf("[IPC] 收到 Register 调用")
	log.Printf("[IPC] 参数: phone=%s, vkey=%s", phone, vkey)

	// MD5加密密码
	hash := md5.New()
	hash.Write([]byte(password))
	encryptedPassword := hex.EncodeToString(hash.Sum(nil))

	// 构造请求数据
	data := map[string]string{
		"uname": phone,
		"pwd":   encryptedPassword,
		"vcode": vcode,
		"vkey":  vkey,
		"token": "",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{
			"code":    -1,
			"message": "JSON编码失败: " + err.Error(),
		}
	}

	// 构造POST表单数据
	formData := url.Values{}
	formData.Set("data", string(jsonData))
	formData.Set("type", "register")

	// 发送POST请求
	req, err := http.NewRequest("POST", "https://www.moyunteng.com/api/api.php", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[IPC] 创建请求失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "创建请求失败: " + err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[IPC] 请求注册接口失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "请求注册接口失败: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IPC] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "读取响应失败: " + err.Error(),
		}
	}

	log.Printf("[IPC] 注册接口响应: %s", string(body))

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[IPC] 解析响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "解析响应失败: " + err.Error(),
		}
	}

	return result
}

// UnbindHost 解绑主机
func (a *App) UnbindHost(token string, deviceIds []string, vcode string, vkey string) map[string]interface{} {
	log.Printf("[IPC] 收到 UnbindHost 调用")
	log.Printf("[IPC] 参数: token=%s, deviceIds=%v, vcode=%s, vkey=%s", token, deviceIds, vcode, vkey)

	// 1. 构造内部 data JSON
	innerData := map[string]interface{}{
		"hlist": deviceIds,
		"vcode": vcode,
		"vkey":  vkey,
	}
	innerDataBytes, err := json.Marshal(innerData)
	if err != nil {
		return map[string]interface{}{
			"code":    -1,
			"message": "内部数据JSON编码失败: " + err.Error(),
		}
	}

	// 2. 构造外部 data JSON
	outerData := map[string]string{
		"act":   "batchUnbind",
		"data":  string(innerDataBytes),
		"token": token,
	}
	outerDataBytes, err := json.Marshal(outerData)
	if err != nil {
		return map[string]interface{}{
			"code":    -1,
			"message": "外部数据JSON编码失败: " + err.Error(),
		}
	}

	// 3. 构造表单数据
	formData := url.Values{}
	formData.Set("data", string(outerDataBytes))
	formData.Set("type", "user_host_oper")

	// 4. 发送POST请求
	req, err := http.NewRequest("POST", "https://www.moyunteng.com/api/api.php", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("[IPC] 创建请求失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "创建请求失败: " + err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[IPC] 请求解绑接口失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "请求解绑接口失败: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IPC] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "读取响应失败: " + err.Error(),
		}
	}

	log.Printf("[IPC] 解绑接口响应: %s", string(body))

	// 5. 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[IPC] 解析响应失败: %v", err)
		return map[string]interface{}{
			"code":    -1,
			"message": "解析响应失败: " + err.Error(),
		}
	}

	return result
}

// CreateMacvlanNetwork 创建macvlan网络
func (a *App) CreateMacvlanNetwork(deviceIP string, version string, networkConfig map[string]interface{}, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 CreateMacvlanNetwork 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, networkConfig=%+v", deviceIP, version, networkConfig)

	// 调用真实的Docker网络创建逻辑
	result, err := createMacvlanNetwork(deviceIP, version, networkConfig, password)
	if err != nil {
		log.Printf("[IPC] 创建macvlan网络失败: %v", err)
		// 失败时返回错误结果
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	log.Printf("[IPC] CreateMacvlanNetwork 返回结果: %+v", result)
	return result
}

// DeleteDockerNetwork 删除Docker网络
func (a *App) DeleteDockerNetwork(deviceIP string, version string, networkID string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 DeleteDockerNetwork 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, networkID=%s", deviceIP, version, networkID)

	// 验证networkID参数
	if networkID == "" {
		log.Printf("[IPC] 删除Docker网络失败: 无效的网络ID")
		return map[string]interface{}{
			"success": false,
			"message": "无效的网络ID",
		}
	}

	// 调用真实的Docker网络删除逻辑
	result, err := deleteDockerNetwork(deviceIP, version, networkID, password)
	if err != nil {
		log.Printf("[IPC] 删除Docker网络失败: %v", err)
		// 失败时返回错误结果
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	log.Printf("[IPC] DeleteDockerNetwork 返回结果: %+v", result)
	return result
}

// UpdateDockerNetwork 更新Docker网络
func (a *App) UpdateDockerNetwork(deviceIP string, version string, networkID string, networkConfig map[string]interface{}, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 UpdateDockerNetwork 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s, networkID=%s, networkConfig=%+v", deviceIP, version, networkID, networkConfig)

	// 验证networkID参数
	if networkID == "" {
		log.Printf("[IPC] 更新Docker网络失败: 无效的网络ID")
		return map[string]interface{}{
			"success": false,
			"message": "无效的网络ID",
		}
	}

	// 调用真实的Docker网络更新逻辑
	result, err := updateDockerNetwork(deviceIP, version, networkID, networkConfig, password)
	if err != nil {
		log.Printf("[IPC] 更新Docker网络失败: %v", err)
		// 失败时返回错误结果
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// 确保返回结果的所有字段都是有效的
	if result["network"] == nil {
		result["network"] = map[string]interface{}{}
	}

	log.Printf("[IPC] UpdateDockerNetwork 返回结果: %+v", result)
	return result
}

// getDockerNetworks 获取Docker网络列表
func getDockerNetworks(deviceIP string, version string, password string) (interface{}, error) {
	resp, err := callDockerAPI(deviceIP, version, "/networks", "GET", nil, password)
	if err != nil {
		return nil, fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	// 解析JSON响应
	var networks []interface{}
	if err := json.Unmarshal(body, &networks); err != nil {
		return nil, fmt.Errorf("解析Docker API响应失败: %v", err)
	}

	log.Printf("Docker API网络响应: %s", string(body))
	return networks, nil
}

// createMacvlanNetwork 创建macvlan网络（使用通用createNetwork函数）
func createMacvlanNetwork(deviceIP string, version string, networkConfig map[string]interface{}, password string) (map[string]interface{}, error) {
	// 设置网络驱动为macvlan
	networkConfig["driver"] = "macvlan"
	// 调用通用的createNetwork函数
	return createNetwork(deviceIP, version, networkConfig, password)
}

// deleteDockerNetwork 删除Docker网络
func deleteDockerNetwork(deviceIP string, version string, networkID string, password string) (map[string]interface{}, error) {
	// 参数验证
	if deviceIP == "" || networkID == "" {
		return nil, fmt.Errorf("设备IP和网络ID不能为空")
	}

	resp, err := callDockerAPI(deviceIP, version, "/networks/"+networkID, "DELETE", nil, password)
	if err != nil {
		return nil, fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	log.Printf("Docker API删除网络响应: %s", string(body))

	// 检查响应状态码
	if resp.StatusCode == http.StatusNoContent {
		return map[string]interface{}{
			"success": true,
			"message": "网络删除成功",
		}, nil
	}

	return nil, fmt.Errorf("删除网络失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// createNetwork 创建网络的通用函数
func createNetwork(deviceIP string, version string, networkConfig map[string]interface{}, password string) (map[string]interface{}, error) {
	// 参数验证
	networkName, ok := networkConfig["networkName"].(string)
	if !ok || networkName == "" {
		return nil, fmt.Errorf("网络名称不能为空")
	}

	driver, ok := networkConfig["driver"].(string)
	if !ok || driver == "" {
		return nil, fmt.Errorf("网络驱动不能为空")
	}

	subnet, ok := networkConfig["subnet"].(string)
	if !ok || subnet == "" {
		return nil, fmt.Errorf("网段不能为空")
	}

	gateway, ok := networkConfig["gateway"].(string)
	if !ok || gateway == "" {
		return nil, fmt.Errorf("网关不能为空")
	}

	// 构造请求体
	reqBody := map[string]interface{}{
		"Name":   networkName,
		"Driver": driver,
		"IPAM": map[string]interface{}{
			"Driver": "default",
			"Config": []map[string]interface{}{
				{
					"Subnet":  subnet,
					"Gateway": gateway,
				},
			},
		},
		"Internal": networkConfig["isPrivate"],
	}

	// 添加IP范围（如果有）
	if ipRange, ok := networkConfig["ipRange"].(string); ok && ipRange != "" {
		ipamConfig := reqBody["IPAM"].(map[string]interface{})["Config"].([]map[string]interface{})[0]
		ipamConfig["IPRange"] = ipRange
	}

	// 仅对macvlan网络添加parent选项
	if driver == "macvlan" {
		if parentInterface, ok := networkConfig["parentInterface"].(string); ok && parentInterface != "" {
			reqBody["Options"] = map[string]interface{}{
				"parent": parentInterface,
			}
		}
	}

	// 序列化请求体
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 调用Docker API创建网络
	resp, err := callDockerAPI(deviceIP, version, "/networks/create", "POST", data, password)
	if err != nil {
		return nil, fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	log.Printf("Docker API创建网络响应: %s", string(body))

	// 检查响应状态码
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("创建网络失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析Docker API响应失败: %v", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "网络创建成功",
		"network": result,
	}, nil
}

// updateDockerNetwork 更新Docker网络
func updateDockerNetwork(deviceIP string, version string, networkID string, networkConfig map[string]interface{}, password string) (map[string]interface{}, error) {
	// 参数验证
	if networkID == "" {
		return nil, fmt.Errorf("网络ID不能为空")
	}

	// 1. 获取当前网络信息，包括连接的容器
	networks, err := getDockerNetworks(deviceIP, version, password)
	if err != nil {
		return nil, fmt.Errorf("获取网络信息失败: %v", err)
	}

	var networkInfo map[string]interface{}
	var connectedContainers []string

	// 查找目标网络
	for _, net := range networks.([]interface{}) {
		netMap := net.(map[string]interface{})
		netID, ok := netMap["ID"].(string)
		if !ok {
			// 尝试获取小写的Id字段
			netID, ok = netMap["Id"].(string)
			if !ok {
				continue
			}
		}

		if netID == networkID {
			networkInfo = netMap
			// 记录连接的容器
			if containers, ok := netMap["Containers"].(map[string]interface{}); ok {
				for containerID := range containers {
					connectedContainers = append(connectedContainers, containerID)
				}
			}
			break
		}
	}

	if networkInfo == nil {
		return nil, fmt.Errorf("未找到网络: %s", networkID)
	}

	// 2. 断开所有连接的容器
	disconnectedContainers := []string{}
	for _, containerID := range connectedContainers {
		if err := disconnectContainerFromNetwork(deviceIP, version, networkID, containerID, password); err != nil {
			log.Printf("断开容器 %s 失败: %v", containerID, err)
			// 继续尝试断开其他容器，不中断整个过程
			continue
		}
		disconnectedContainers = append(disconnectedContainers, containerID)
	}

	// 3. 更新网络配置（Docker API不支持直接更新网络配置，需要先删除再创建）
	// 保存原有网络的基本信息
	originalName := ""
	if name, ok := networkInfo["Name"].(string); ok {
		originalName = name
	} else {
		return nil, fmt.Errorf("网络名称获取失败")
	}

	originalDriver := ""
	if driver, ok := networkInfo["Driver"].(string); ok {
		originalDriver = driver
	} else {
		return nil, fmt.Errorf("网络驱动获取失败")
	}

	// 安全获取Options，处理nil情况
	originalOptions := make(map[string]interface{})
	if opts, ok := networkInfo["Options"].(map[string]interface{}); ok {
		originalOptions = opts
	}

	originalInternal := false
	if internal, ok := networkInfo["Internal"].(bool); ok {
		originalInternal = internal
	}

	// 构造新的网络配置
	// 提取并验证subnet
	subnet, ok := networkConfig["subnet"].(string)
	if !ok || subnet == "" {
		return nil, fmt.Errorf("网段不能为空")
	}

	// 提取并验证gateway
	gateway, ok := networkConfig["gateway"].(string)
	if !ok || gateway == "" {
		return nil, fmt.Errorf("网关不能为空")
	}

	newConfig := map[string]interface{}{
		"networkName": originalName,
		"driver":      originalDriver,
		"subnet":      subnet,
		"gateway":     gateway,
		"isPrivate":   originalInternal,
	}

	// 添加IP范围（如果有）
	if ipRange, ok := networkConfig["ipRange"].(string); ok && ipRange != "" {
		newConfig["ipRange"] = ipRange
	}

	// 仅对macvlan网络添加parent选项
	if originalDriver == "macvlan" {
		if parent, ok := originalOptions["parent"].(string); ok && parent != "" {
			newConfig["parentInterface"] = parent
		}
	}

	// 4. 删除原有网络
	if _, err := deleteDockerNetwork(deviceIP, version, networkID, password); err != nil {
		// 删除失败，尝试恢复容器连接
		for _, containerID := range disconnectedContainers {
			_ = connectContainerToNetwork(deviceIP, version, networkID, containerID, password)
		}
		return nil, fmt.Errorf("删除原有网络失败: %v", err)
	}

	// 5. 创建新网络（重试5次）
	var newNetworkResult map[string]interface{}
	var newNetworkErr error
	for i := 0; i < 5; i++ {
		newNetworkResult, newNetworkErr = createNetwork(deviceIP, version, newConfig, password)
		if newNetworkErr == nil {
			break
		}
		log.Printf("创建新网络失败，第 %d 次重试: %v", i+1, newNetworkErr)
		time.Sleep(1 * time.Second)
	}

	if newNetworkErr != nil {
		// 创建新网络失败，尝试恢复原有网络
		// 这里需要重新创建原有网络配置，暂时简化处理
		return nil, fmt.Errorf("创建新网络失败: %v", newNetworkErr)
	}

	// 6. 重新连接所有容器（重试5次）
	newNetworkID := ""
	if network, ok := newNetworkResult["network"].(map[string]interface{}); ok {
		if id, ok := network["ID"].(string); ok {
			newNetworkID = id
		} else if id, ok := network["Id"].(string); ok {
			newNetworkID = id
		}
	}

	if newNetworkID == "" {
		return nil, fmt.Errorf("获取新网络ID失败")
	}

	for _, containerID := range disconnectedContainers {
		var connectErr error
		for i := 0; i < 5; i++ {
			connectErr = connectContainerToNetwork(deviceIP, version, newNetworkID, containerID, password)
			if connectErr == nil {
				break
			}
			log.Printf("重新连接容器 %s 失败，第 %d 次重试: %v", containerID, i+1, connectErr)
			time.Sleep(1 * time.Second)
		}
		if connectErr != nil {
			log.Printf("最终未能重新连接容器 %s: %v", containerID, connectErr)
			// 继续尝试连接其他容器，不中断整个过程
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "网络更新成功",
		"network": newNetworkResult["network"],
	}, nil
}

// disconnectContainerFromNetwork 断开容器与网络的连接
func disconnectContainerFromNetwork(deviceIP string, version string, networkID string, containerID string, password string) error {
	// 参数验证
	if deviceIP == "" || networkID == "" || containerID == "" {
		return fmt.Errorf("参数不能为空: deviceIP=%s, networkID=%s, containerID=%s", deviceIP, networkID, containerID)
	}

	reqBody := map[string]interface{}{
		"Container": containerID,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := callDockerAPI(deviceIP, version, "/networks/"+networkID+"/disconnect", "POST", data, password)
	if err != nil {
		return fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("断开容器失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// connectContainerToNetwork 连接容器到网络
func connectContainerToNetwork(deviceIP string, version string, networkID string, containerID string, password string) error {
	// 参数验证
	if deviceIP == "" || networkID == "" || containerID == "" {
		return fmt.Errorf("参数不能为空: deviceIP=%s, networkID=%s, containerID=%s", deviceIP, networkID, containerID)
	}

	reqBody := map[string]interface{}{
		"Container": containerID,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := callDockerAPI(deviceIP, version, "/networks/"+networkID+"/connect", "POST", data, password)
	if err != nil {
		return fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("连接容器失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetVPCProxies 获取VPC代理列表
// Todo: 实现获取VPC代理列表的完整功能
func (a *App) GetVPCProxies() []VPCProxy {
	log.Printf("[IPC] 收到 GetVPCProxies 调用")

	// 返回模拟VPC代理列表，添加功能开发中的提示
	result := mockVPCProxies
	log.Printf("[IPC] GetVPCProxies 返回代理数量: %d", len(result))
	log.Printf("[IPC] GetVPCProxies 返回结果: %+v", result)
	return result
}

// GetVPCHosts 获取VPC主机列表
// Todo: 实现获取VPC主机列表的完整功能
func (a *App) GetVPCHosts() []VPCHost {
	log.Printf("[IPC] 收到 GetVPCHosts 调用")

	// 返回模拟VPC主机列表，添加功能开发中的提示
	result := mockVPCHosts
	log.Printf("[IPC] GetVPCHosts 返回主机数量: %d", len(result))
	log.Printf("[IPC] GetVPCHosts 返回结果: %+v", result)
	return result
}

// GetPhoneTemplates 获取手机模板列表
// 代理调用外部API: https://newapi.moyunteng.com/api/v1/template/list
func (a *App) GetPhoneTemplates(page int, androidVersion string) map[string]interface{} {
	log.Printf("[IPC] 收到 GetPhoneTemplates 调用, page: %d, androidVersion: %s", page, androidVersion)

	// 调用外部API
	versionParam := "all"
	if androidVersion != "" {
		versionParam = androidVersion
	}
	url := fmt.Sprintf("https://newapi.moyunteng.com/api/v1/template/list?page=%d&android_version=%s", page, versionParam)
	log.Printf("[IPC] 调用外部API: %s", url)

	// 发送GET请求
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[IPC] 调用API失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"message": "调用外部API失败",
			"data":    nil,
			"error":   err.Error(),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IPC] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code_id": 500,
			"message": "读取响应失败",
			"data":    nil,
			"error":   err.Error(),
		}
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[IPC] 解析JSON失败: %v, 响应内容: %s", err, string(body))
		// 返回原始响应，让前端处理
		return map[string]interface{}{
			"code_id": 200,
			"message": "OK",
			"data": map[string]interface{}{
				"raw": string(body),
			},
		}
	}

	log.Printf("[IPC] GetPhoneTemplates 返回结果: %+v", result)
	return result
}

// GetMirrorList 获取镜像列表
// 代理调用外部API: http://api.moyunteng.com/api.php?type=get_mirror_list2
func (a *App) GetMirrorList() map[string]interface{} {
	log.Printf("[IPC] 收到 GetMirrorList 调用")

	// 调用外部API
	url := "http://api.moyunteng.com/api.php?type=get_mirror_list2"
	log.Printf("[IPC] 调用外部API: %s", url)

	// 发送POST请求
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Printf("[IPC] 调用API失败: %v", err)
		return map[string]interface{}{
			"code":  "500",
			"msg":   "调用外部API失败",
			"data":  nil,
			"error": err.Error(),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[IPC] 读取响应失败: %v", err)
		return map[string]interface{}{
			"code":  "500",
			"msg":   "读取响应失败",
			"data":  nil,
			"error": err.Error(),
		}
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[IPC] 解析JSON失败: %v, 响应内容: %s", err, string(body))
		// 返回原始响应，让前端处理
		return map[string]interface{}{
			"code": "200",
			"msg":  "OK",
			"data": map[string]interface{}{
				"raw": string(body),
			},
		}
	}

	log.Printf("[IPC] GetMirrorList 返回结果: %+v", result)
	return result
}

// LoadImageToDevice 从本地tar.gz文件导入镜像到设备
func (a *App) LoadImageToDevice(deviceIP string, imagePath string, version string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 LoadImageToDevice 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, imagePath=%s, version=%s", deviceIP, imagePath, version)

	// 重置进度
	a.progressMutex.Lock()
	a.imagePullProgress = 0
	a.progressMutex.Unlock()

	// 创建上传上下文
	a.uploadMutex.Lock()
	uploadCtx, uploadCancel := context.WithCancel(context.Background())
	a.uploadCtx = uploadCtx
	a.uploadCancel = uploadCancel
	a.uploadMutex.Unlock()

	// 调用镜像加载逻辑
	imageName, err := loadImageToDevice(a, deviceIP, version, imagePath, password)
	
	// 清理上传上下文
	a.uploadMutex.Lock()
	a.uploadCancel = nil
	a.uploadCtx = nil
	a.uploadMutex.Unlock()
	
	if err != nil {
		log.Printf("[IPC] 加载镜像失败: %v", err)
		// 检查是否是因为取消导致的失败
		if uploadCtx.Err() == context.Canceled {
			log.Printf("[IPC] 上传已被取消")
			return map[string]interface{}{
				"success": false,
				"message": "上传已取消",
				"canceled": true,
			}
		}
		// 发送失败事件
		a.emitEvent("upload-complete", map[string]interface{}{
			"success":  false,
			"message":  err.Error(),
			"deviceIP": deviceIP,
		})
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// 更新进度为100%
	a.progressMutex.Lock()
	a.imagePullProgress = 100
	a.progressMutex.Unlock()

	// 发送完成事件（包含deviceIP用于前端按设备显示进度）
	a.emitEvent("upload-complete", map[string]interface{}{
		"success":   true,
		"message":   "镜像加载成功",
		"deviceIP":  deviceIP,
		"imageName": imageName,
	})

	result := map[string]interface{}{
		"success":   true,
		"message":   "镜像加载成功",
		"imageName": imageName,
	}
	log.Printf("[IPC] LoadImageToDevice 返回结果: %+v", result)
	return result
}

// loadImageToDevice 从本地tar.gz文件导入镜像到设备
func loadImageToDevice(a *App, deviceIP string, version string, imagePath string, password string) (string, error) {
	// 检查是否已取消
	a.uploadMutex.Lock()
	uploadCtx := a.uploadCtx
	a.uploadMutex.Unlock()
	
	if uploadCtx != nil && uploadCtx.Err() == context.Canceled {
		return "", fmt.Errorf("上传已取消")
	}
	
	// 参数验证
	if deviceIP == "" || imagePath == "" {
		return "", fmt.Errorf("设备IP和镜像路径不能为空")
	}

	// 打开本地tar.gz文件
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("打开镜像文件失败: %v", err)
	}
	defer file.Close()

	// 再次检查是否已取消
	if uploadCtx != nil && uploadCtx.Err() == context.Canceled {
		return "", fmt.Errorf("上传已取消")
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := fileInfo.Size()

	// 创建带进度和取消检查的读取器
	progressReader := &ProgressReader{
		Reader: file,
		Size:   fileSize,
		Callback: func(progress float64) {
			// 检查是否已取消
			a.uploadMutex.Lock()
			ctx := a.uploadCtx
			a.uploadMutex.Unlock()
			
			if ctx != nil && ctx.Err() == context.Canceled {
				log.Printf("[上传镜像] 进度回调检测到取消信号")
				return
			}
			
			// 更新全局进度
			a.progressMutex.Lock()
			a.imagePullProgress = progress
			a.progressMutex.Unlock()

			// 发送进度事件（包含deviceIP用于前端按设备显示进度）
			a.emitEvent("upload-progress", map[string]interface{}{
				"progress": progress,
				"deviceIP": deviceIP,
			})
		},
	}

	// 上传前最后检查
	if uploadCtx != nil && uploadCtx.Err() == context.Canceled {
		return "", fmt.Errorf("上传已取消")
	}

	// 调用Docker API上传镜像(传入context以支持取消)
	resp, err := callDockerAPIWithReader(uploadCtx, deviceIP, version, "/images/load", "POST", progressReader, "application/x-tar", password)
	if err != nil {
		// 检查是否是context取消导致的错误
		if uploadCtx != nil && uploadCtx.Err() == context.Canceled {
			return "", fmt.Errorf("上传已取消")
		}
		return "", fmt.Errorf("调用Docker API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("加载镜像失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析Docker API响应，提取加载的镜像名称
	var response struct {
		Stream string `json:"stream"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		// 解析失败，返回空字符串
		log.Printf("解析Docker API响应失败: %v", err)
		return "", nil
	}

	// 从响应流中提取镜像名称
	// 响应格式类似: {"stream":"Loaded image: registry.cn-guangzhou.aliyuncs.com/mytos/dobox:Q12_base_202601051937\n"}
	stream := response.Stream
	imageName := ""
	if strings.Contains(stream, "Loaded image: ") {
		parts := strings.Split(stream, "Loaded image: ")
		if len(parts) > 1 {
			imageName = strings.TrimSpace(parts[1])
			// 移除可能的换行符
			imageName = strings.TrimSuffix(imageName, "\n")
		}
	}

	log.Printf("Docker API加载镜像响应: %s", string(body))
	log.Printf("提取的镜像名称: %s", imageName)
	return imageName, nil
}

// CreateMytSdkContainer 创建并启动myt_sdk容器
func (a *App) CreateMytSdkContainer(deviceIP string, version string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 CreateMytSdkContainer 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, version=%s", deviceIP, version)

	// 调用容器创建逻辑
	if err := createMytSdkContainer(deviceIP, version, password); err != nil {
		log.Printf("[IPC] 创建myt_sdk容器失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	result := map[string]interface{}{
		"success": true,
		"message": "myt_sdk容器创建成功",
	}
	log.Printf("[IPC] CreateMytSdkContainer 返回结果: %+v", result)
	return result
}

// createMytSdkContainer 创建并启动myt_sdk容器
func createMytSdkContainer(deviceIP string, version string, password string) error {
	// 构造请求体
	reqBody := map[string]interface{}{
		"Image": "registry.magicloud.tech/magicloud/myt_sdk:latest",
		"Cmd": []string{
			"bash", "-c", "while true; do sleep 3600; done",
		},
		"HostConfig": map[string]interface{}{
			"Privileged":  true,
			"NetworkMode": "host",
		},
		"Name": "myt_sdk",
	}

	// 序列化请求体
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 调用Docker API创建容器
	resp, err := callDockerAPI(deviceIP, version, "/containers/create", "POST", data, password)
	if err != nil {
		return fmt.Errorf("调用Docker API创建容器失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Docker API响应失败: %v", err)
	}

	log.Printf("Docker API创建myt_sdk容器响应: %s", string(body))

	// 检查响应状态码
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("创建容器失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应，获取容器ID
	var containerInfo map[string]interface{}
	if err := json.Unmarshal(body, &containerInfo); err != nil {
		return fmt.Errorf("解析Docker API响应失败: %v", err)
	}

	containerID, ok := containerInfo["Id"].(string)
	if !ok {
		return fmt.Errorf("获取容器ID失败")
	}

	// 启动容器
	startResp, err := callDockerAPI(deviceIP, version, "/containers/"+containerID+"/start", "POST", nil, password)
	if err != nil {
		return fmt.Errorf("调用Docker API启动容器失败: %v", err)
	}
	defer startResp.Body.Close()

	log.Printf("Docker API启动myt_sdk容器响应状态码: %d", startResp.StatusCode)

	// 检查响应状态码
	if startResp.StatusCode != http.StatusNoContent {
		startBody, _ := io.ReadAll(startResp.Body)
		return fmt.Errorf("启动容器失败，状态码: %d, 响应: %s", startResp.StatusCode, string(startBody))
	}

	return nil
}

// PullDockerImage 拉取Docker镜像
func (a *App) PullDockerImage(deviceIP string, imageUrl string, version string, password string) map[string]interface{} {
	log.Printf("[IPC] 收到 PullDockerImage 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, imageUrl=%s, version=%s", deviceIP, imageUrl, version)

	// 重置进度
	a.progressMutex.Lock()
	a.imagePullProgress = 0
	a.progressMutex.Unlock()

	// 调用pullDockerImage方法拉取镜像
	success, message := a.pullDockerImage(deviceIP, imageUrl, version, password)

	result := map[string]interface{}{
		"success": success,
		"message": message,
	}
	log.Printf("[IPC] PullDockerImage 返回结果: %+v", result)
	return result
}

// pullDockerImage 拉取Docker镜像，内部方法（失败时最多重试2次）
func (a *App) pullDockerImage(deviceIP string, imageUrl string, version string, password string) (bool, string) {
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ok, msg := a.pullDockerImageOnce(deviceIP, imageUrl, version, password)
		if ok {
			return true, msg
		}
		// DNS 解析失败 / 连接失败时自动重试
		if attempt < maxRetries && isDNSOrConnError(msg) {
			log.Printf("[镜像拉取] 第 %d 次失败（DNS/连接问题），2秒后重试: %s", attempt, msg)
			time.Sleep(2 * time.Second)
			continue
		}
		return false, msg
	}
	return false, "镜像拉取失败"
}

// isDNSOrConnError 判断是否为 DNS 解析或网络连接类错误
func isDNSOrConnError(msg string) bool {
	keywords := []string{
		"no such host",
		"dial tcp",
		"connection refused",
		"i/o timeout",
		"EOF",
		"network is unreachable",
	}
	msgLower := strings.ToLower(msg)
	for _, kw := range keywords {
		if strings.Contains(msgLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// pullDockerImageOnce 拉取Docker镜像，单次执行
func (a *App) pullDockerImageOnce(deviceIP string, imageUrl string, version string, password string) (bool, string) {
	// 参数验证
	if deviceIP == "" || imageUrl == "" {
		return false, "设备IP和镜像地址不能为空"
	}

	// 构造Docker API URL
	endpoint := fmt.Sprintf("/images/create?fromImage=%s", imageUrl)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 设置请求头，接受流式响应
	req.Header.Set("Accept", "application/json")

	// 调用Docker API
	// 注意：由于需要处理流式响应，这里不能直接使用callDockerAPI，需要单独实现
	// 对于v3设备，优先使用8000端口的docker端点
	var url string
	var resp *http.Response

	if version == "v3" {
		url = fmt.Sprintf("http://%s/docker%s", deviceAddr(deviceIP), endpoint)
		req, err = http.NewRequest("POST", url, nil)
		if err != nil {
			return false, fmt.Sprintf("创建请求失败: %v", err)
		}

		// 添加认证头
		if password != "" {
			req.SetBasicAuth("admin", password)
		}

		// 设置请求头，接受流式响应
		req.Header.Set("Accept", "application/json")

		// 发送请求
		client := &http.Client{
			Timeout: 0, // 不设置超时，因为拉取镜像可能需要很长时间
		}
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			// 成功，使用这个响应
		} else {
			// 失败，回退到2375端口
			log.Printf("v3 docker端点调用失败，回退到2375端口: %v", err)
			url = fmt.Sprintf("http://%s:2375%s", deviceIP, endpoint)
			req, err = http.NewRequest("POST", url, nil)
			if err != nil {
				return false, fmt.Sprintf("创建请求失败: %v", err)
			}

			// 设置请求头，接受流式响应
			req.Header.Set("Accept", "application/json")

			// 发送请求
			client := &http.Client{
				Timeout: 0, // 不设置超时，因为拉取镜像可能需要很长时间
			}
			resp, err = client.Do(req)
			if err != nil {
				return false, fmt.Sprintf("调用Docker API失败: %v", err)
			}
		}
	} else {
		// 非v3设备，直接使用2375端口
		url = fmt.Sprintf("http://%s:2375%s", deviceIP, endpoint)
		req, err = http.NewRequest("POST", url, nil)
		if err != nil {
			return false, fmt.Sprintf("创建请求失败: %v", err)
		}

		// 设置请求头，接受流式响应
		req.Header.Set("Accept", "application/json")

		// 发送请求
		client := &http.Client{
			Timeout: 0, // 不设置超时，因为拉取镜像可能需要很长时间
		}
		resp, err = client.Do(req)
		if err != nil {
			return false, fmt.Sprintf("调用Docker API失败: %v", err)
		}
	}
	defer resp.Body.Close()

	log.Printf("拉取Docker镜像URL: %s", url)

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 读取错误响应
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Sprintf("镜像拉取失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	decoder := json.NewDecoder(resp.Body)
	var lastProgress float64 = 0
	var streamError string

	for {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				// 流结束
				break
			}
			log.Printf("解析流式响应失败: %v", err)
			continue
		}

		// 检查 Docker 流式错误（如 DNS 失败、认证失败等）
		if errMsg, ok := event["error"].(string); ok && errMsg != "" {
			log.Printf("镜像拉取流式错误: %s", errMsg)
			streamError = errMsg
			continue
		}

		// 解析进度信息
		if status, ok := event["status"].(string); ok {
			log.Printf("镜像拉取状态: %s", status)

			// 检查是否有进度信息
			if progress, ok := event["progressDetail"].(map[string]interface{}); ok {
				// 提取已完成和总大小
				if current, ok := progress["current"].(float64); ok {
					if total, ok := progress["total"].(float64); ok && total > 0 {
						// 计算进度百分比
						currentProgress := (current / total) * 100

						// 每隔1%或超过1秒才记录日志，避免日志刷屏
						if currentProgress-lastProgress >= 1 {
							log.Printf("镜像拉取进度: %.1f%%", currentProgress)
							lastProgress = currentProgress

							// 更新全局进度变量
							a.progressMutex.Lock()
							a.imagePullProgress = currentProgress
							a.progressMutex.Unlock()
						}
					}
				}
			}
		}
	}

	// 流式响应中有错误，返回失败（由外层重试逻辑决定是否重试）
	if streamError != "" {
		return false, streamError
	}

	// 镜像拉取成功，更新进度为100%
	a.progressMutex.Lock()
	a.imagePullProgress = 100
	a.progressMutex.Unlock()

	log.Printf("镜像拉取成功，总进度: 100%%")
	return true, "镜像拉取成功"
}

// GetImagePullProgress 获取当前镜像拉取进度
func (a *App) GetImagePullProgress() map[string]interface{} {
	a.progressMutex.Lock()
	// 同时检查downloadProgress和imagePullProgress，返回较大的值
	progress := a.imagePullProgress
	if a.downloadProgress > progress {
		progress = a.downloadProgress
	}
	a.progressMutex.Unlock()

	return map[string]interface{}{
		"progress": progress,
	}
}

// SetDevicePassword 设置设备认证密码
func (a *App) SetDevicePassword(deviceIP string, password string, currentPassword string) map[string]interface{} {
	log.Printf("[IPC] 收到 SetDevicePassword 调用")
	log.Printf("[IPC] 参数: deviceIP=%s", deviceIP)

	// 调用设置密码逻辑
	success, message := setDevicePassword(deviceIP, password, currentPassword)

	result := map[string]interface{}{
		"success": success,
		"message": message,
	}
	log.Printf("[IPC] SetDevicePassword 返回结果: %+v", result)
	return result
}

// CloseDevicePassword 关闭设备认证密码
func (a *App) CloseDevicePassword(deviceIP string, currentPassword string) map[string]interface{} {
	log.Printf("[IPC] 收到 CloseDevicePassword 调用")
	log.Printf("[IPC] 参数: deviceIP=%s", deviceIP)

	// 调用关闭密码逻辑
	success, message := closeDevicePassword(deviceIP, currentPassword)

	result := map[string]interface{}{
		"success": success,
		"message": message,
	}
	log.Printf("[IPC] CloseDevicePassword 返回结果: %+v", result)
	return result
}

// setDevicePassword 设置设备认证密码
func setDevicePassword(deviceIP string, password string, currentPassword string) (bool, string) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/auth/password", deviceAddr(deviceIP))

	// 发送HTTP POST请求，包含密码
	reqBody := map[string]string{
		"newPassword":     password,
		"confirmPassword": password,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Sprintf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 添加认证头
	if currentPassword != "" {
		req.SetBasicAuth("admin", currentPassword)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("读取V3 API响应失败: %v", err)
	}

	log.Printf("V3 Set Password API响应: %s", string(body))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Sprintf("解析V3 API响应失败: %v", err)
	}

	// 检查响应状态
	if result["code"] == float64(0) {
		return true, "密码设置成功"
	}

	return false, fmt.Sprintf("密码设置失败: %v", result["message"])
}

// closeDevicePassword 关闭设备认证密码
func closeDevicePassword(deviceIP string, currentPassword string) (bool, string) {
	// 构造V3 API URL
	url := fmt.Sprintf("http://%s/auth/close", deviceAddr(deviceIP))

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 添加认证头
	if currentPassword != "" {
		req.SetBasicAuth("admin", currentPassword)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("调用V3 API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("读取V3 API响应失败: %v", err)
	}

	log.Printf("V3 Close Password API响应: %s", string(body))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Sprintf("解析V3 API响应失败: %v", err)
	}

	// 检查响应状态
	if result["code"] == float64(0) {
		return true, "密码关闭成功"
	}

	return false, fmt.Sprintf("密码关闭失败: %v", result["message"])
}

// ImageMetadata 镜像元数据结构体
type ImageMetadata struct {
	OnlineURL       string   `json:"online_url"`
	ImageName       string   `json:"image_name"`
	LocalPath       string   `json:"local_path"`
	Size            string   `json:"size"`
	CreateTime      int64    `json:"create_time"`
	AvailableModels []string `json:"available_models"`

	// 在线镜像完整元数据
	ID        string   `json:"id"`
	Spid      string   `json:"spid"`
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Url       string   `json:"url"`
	Sort      string   `json:"sort"`
	State     string   `json:"state"`
	Ttype     string   `json:"ttype"`
	OsVer     string   `json:"os_ver"`
	Udesc     string   `json:"udesc"`
	SysVer    string   `json:"sys_ver"`
	Ttype2    []string `json:"ttype2"`
	SysVerDes string   `json:"sys_ver_des"`
}

// getStringFromMap 从map中获取字符串值，如果不存在或类型不正确则返回空字符串
func getStringFromMap(data interface{}, key string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}

// getStringSliceFromMap 从map中获取字符串切片值，如果不存在或类型不正确则返回空切片
func getStringSliceFromMap(data interface{}, key string) []string {
	if m, ok := data.(map[string]interface{}); ok {
		if val, ok := m[key]; ok {
			if slice, ok := val.([]interface{}); ok {
				result := make([]string, 0, len(slice))
				for _, item := range slice {
					if str, ok := item.(string); ok {
						result = append(result, str)
					}
				}
				return result
			}
		}
	}
	return []string{}
}

// GetLocalImages 获取本地镜像列表
func (a *App) GetLocalImages() []map[string]interface{} {
	log.Printf("[IPC] 收到 GetLocalImages 调用")

	// 获取存储目录（支持用户自定义）
	edgeclientDir := getStorageBaseDir()

	// 确保目录存在
	if _, err := os.Stat(edgeclientDir); os.IsNotExist(err) {
		log.Printf("[IPC] 存储目录不存在，创建目录: %s", edgeclientDir)
		if err := os.MkdirAll(edgeclientDir, 0777); err != nil {
			log.Printf("[IPC] 创建存储目录失败: %v", err)
			return []map[string]interface{}{}
		}
		return []map[string]interface{}{}
	}

	// 读取目录下的所有.tar.gz文件
	files, err := os.ReadDir(edgeclientDir)
	if err != nil {
		log.Printf("[IPC] 读取edgeclient目录失败: %v", err)
		return []map[string]interface{}{}
	}

	// 过滤出.tar.gz文件
	imageFiles := []os.DirEntry{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
			imageFiles = append(imageFiles, file)
		}
	}

	// 构造返回结果
	imageList := []map[string]interface{}{}
	for _, file := range imageFiles {
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("[IPC] 获取文件信息失败: %v", err)
			continue
		}

		// 检查是否有对应的元数据文件
		metadataPath := filepath.Join(edgeclientDir, strings.TrimSuffix(file.Name(), ".tar.gz")+".json")
		metadata := map[string]interface{}{}

		if _, err := os.Stat(metadataPath); err == nil {
			// 读取元数据文件
			metadataBytes, err := os.ReadFile(metadataPath)
			if err == nil {
				var meta ImageMetadata
				if err := json.Unmarshal(metadataBytes, &meta); err == nil {
					metadata["online_url"] = meta.OnlineURL
					metadata["available_models"] = meta.AvailableModels
					// 添加镜像名称字段
					metadata["image_name"] = meta.Name
				}
			}
		}

		imageItem := map[string]interface{}{
			"name":       file.Name(),
			"path":       filepath.Join(edgeclientDir, file.Name()),
			"size":       formatFileSize(fileInfo.Size()),
			"createTime": fileInfo.ModTime().UnixMilli(),
		}

		// 合并元数据
		for k, v := range metadata {
			imageItem[k] = v
		}

		imageList = append(imageList, imageItem)
	}

	log.Printf("[IPC] GetLocalImages 返回结果: %+v", imageList)
	return imageList
}

// DeleteLocalImage 删除本地缓存镜像（包括zip包和json文件）
func (a *App) DeleteLocalImage(imagePath string) map[string]interface{} {
	log.Printf("[IPC] 收到 DeleteLocalImage 调用，删除镜像: '%s'", imagePath)

	// 检查imagePath是否为空
	if imagePath == "" {
		log.Printf("[IPC] DeleteLocalImage 调用失败: imagePath为空")
		return map[string]interface{}{"code": 1, "message": "删除镜像失败: 无效的文件路径"}
	}

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); err != nil {
		log.Printf("[IPC] DeleteLocalImage 调用失败: 文件不存在: %s, 错误: %v", imagePath, err)
		return map[string]interface{}{"code": 1, "message": fmt.Sprintf("删除镜像失败: 文件不存在: %v", err)}
	}

	// 获取文件目录和文件名
	imageDir := filepath.Dir(imagePath)
	imageName := filepath.Base(imagePath)
	log.Printf("[IPC] DeleteLocalImage 处理: imageDir=%s, imageName=%s", imageDir, imageName)

	// 构建对应的json文件名
	jsonPath := filepath.Join(imageDir, strings.TrimSuffix(imageName, ".tar.gz")+".json")
	log.Printf("[IPC] DeleteLocalImage 处理: 对应的json文件路径=%s", jsonPath)

	// 删除tar.gz文件
	if err := os.Remove(imagePath); err != nil {
		log.Printf("[IPC] 删除镜像文件失败: %v", err)
		return map[string]interface{}{"code": 1, "message": fmt.Sprintf("删除镜像文件失败: %v", err)}
	}

	// 删除json文件（如果存在）
	if _, err := os.Stat(jsonPath); err == nil {
		if err := os.Remove(jsonPath); err != nil {
			log.Printf("[IPC] 删除镜像元数据文件失败: %v", err)
			return map[string]interface{}{"code": 1, "message": fmt.Sprintf("删除镜像元数据文件失败: %v", err)}
		}
		log.Printf("[IPC] 删除镜像元数据文件成功: %s", jsonPath)
	} else {
		log.Printf("[IPC] 镜像元数据文件不存在，跳过删除: %s", jsonPath)
	}

	// 删除对应名称的文件夹（如果存在）
	folderName := strings.TrimSuffix(imageName, ".tar.gz")
	folderPath := filepath.Join(imageDir, folderName)
	if _, err := os.Stat(folderPath); err == nil {
		if err := os.RemoveAll(folderPath); err != nil {
			log.Printf("[IPC] 删除镜像文件夹失败: %v", err)
		} else {
			log.Printf("[IPC] 删除镜像文件夹成功: %s", folderPath)
		}
	} else {
		log.Printf("[IPC] 镜像文件夹不存在，跳过删除: %s", folderPath)
	}

	log.Printf("[IPC] 删除本地镜像成功: %s", imagePath)
	return map[string]interface{}{"code": 0, "message": "OK"}
}

// DownloadImage 下载镜像到本地
func (a *App) DownloadImage(metadata interface{}) map[string]interface{} {
	log.Printf("[IPC] 收到 DownloadImage 调用")

	// 取消之前的下载任务
	a.downloadMutex.Lock()
	if a.downloadCancel != nil {
		log.Printf("[IPC] 取消之前的下载任务")
		a.downloadCancel()
		// 等待一段时间确保旧任务完全停止
		time.Sleep(500 * time.Millisecond)
	}
	
	// 创建新的下载上下文
	ctx, cancel := context.WithCancel(context.Background())
	a.downloadCtx = ctx
	a.downloadCancel = cancel
	a.downloadMutex.Unlock()

	var imageUrl string

	// 处理不同类型的参数
	switch meta := metadata.(type) {
	case map[string]interface{}:
		// 完整的元数据对象
		if url, ok := meta["url"].(string); ok {
			imageUrl = url
		} else {
			log.Printf("[IPC] 无效的参数: 缺少url字段")
			return map[string]interface{}{"success": false, "message": "无效的参数: 缺少url字段"}
		}
	case string:
		// 兼容旧的调用方式，直接传入imageUrl字符串
		imageUrl = meta
	default:
		log.Printf("[IPC] 无效的参数类型: %T", metadata)
		return map[string]interface{}{"success": false, "message": "无效的参数类型"}
	}

	log.Printf("[IPC] 参数: imageUrl=%s", imageUrl)
	log.Printf("[IPC] 镜像元数据: %+v", metadata)

	// 获取存储目录（支持用户自定义）
	edgeclientDir := getStorageBaseDir()
	if err := os.MkdirAll(edgeclientDir, 0777); err != nil {
		log.Printf("[IPC] 创建存储目录失败: %v", err)
		return map[string]interface{}{"success": false, "message": "创建存储目录失败: " + err.Error()}
	}

	// 解析镜像URL，获取镜像名称和标签
	// 镜像URL格式：registry.magicloud.tech/magicloud/dobox-android13:Q1
	parts := strings.Split(imageUrl, "/")
	imageName := parts[len(parts)-1]

	// 替换文件名中的冒号为下划线，因为Windows文件名不能包含冒号
	imageName = strings.ReplaceAll(imageName, ":", "_")

	// 构建目标文件路径
	targetPath := filepath.Join(edgeclientDir, imageName+(".tar.gz"))

	// 重置下载进度
	a.progressMutex.Lock()
	a.downloadProgress = 0
	a.progressMutex.Unlock()

	// 创建下载客户端
	dgetClient := dget.Client{}
	dgetClient.SetClient(http.DefaultClient)

	// 创建取消标志位,确保取消日志只输出一次
	var cancelLogged int32 = 0
	
	// 设置进度回调函数，将真实下载进度通过Wails事件发送给前端
	dgetClient.SetProgressCallback(func(progress float64) {
		// 检查是否已取消
		select {
		case <-ctx.Done():
			// 使用原子操作确保只输出一次取消日志
			if atomic.CompareAndSwapInt32(&cancelLogged, 0, 1) {
				log.Printf("[IPC] 下载已取消，停止发送进度")
			}
			return
		default:
		}
		
		// 更新全局进度，限制在0-95%之间，95-100%留作后续处理
		a.progressMutex.Lock()
		if progress < 95 {
			a.downloadProgress = progress
		} else {
			a.downloadProgress = 95
		}
		currentProgress := a.downloadProgress
		a.progressMutex.Unlock()

		// 发送下载进度事件给前端
		a.emitEvent("download-progress", map[string]interface{}{
			"progress": currentProgress,
		})
	})

	// 解析镜像URL，提取registry、package和tag
	var registry, pkg, tag string
	if strings.Contains(imageUrl, "/") {
		registryParts := strings.Split(imageUrl, "/")
		registry = registryParts[0]
		pkgParts := strings.Split(imageUrl[len(registry)+1:], ":")
		pkg = pkgParts[0]
		if len(pkgParts) > 1 {
			tag = pkgParts[1]
		} else {
			tag = "latest"
		}
	} else {
		registry = "registry-1.docker.io"
		pkgParts := strings.Split(imageUrl, ":")
		pkg = "library/" + pkgParts[0]
		if len(pkgParts) > 1 {
			tag = pkgParts[1]
		} else {
			tag = "latest"
		}
	}

	log.Printf("[IPC] 解析镜像URL: registry=%s, pkg=%s, tag=%s", registry, pkg, tag)

	// 构建目标目录路径（不含.tar.gz后缀）
	targetDir := strings.TrimSuffix(targetPath, ".tar.gz")

	log.Printf("[IPC] 准备下载镜像，targetPath: %s", targetPath)
	log.Printf("[IPC] 准备下载镜像，targetDir: %s", targetDir)

	// 清理旧的下载目录和临时文件，避免文件被占用
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		log.Printf("[IPC] 检测到旧的下载目录，尝试清理: %s", targetDir)
		// 尝试删除目录中的所有.part文件
		filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".part") {
				log.Printf("[IPC] 删除旧的临时文件: %s", path)
				os.Remove(path)
			}
			return nil
		})
		// 尝试删除整个目录
		if err := os.RemoveAll(targetDir); err != nil {
			log.Printf("[IPC] 删除旧下载目录失败: %v，继续下载", err)
		} else {
			log.Printf("[IPC] 成功删除旧下载目录")
		}
		// 等待文件系统释放
		time.Sleep(200 * time.Millisecond)
	}

	// 确保targetDir目录存在
	if err := os.MkdirAll(targetDir, 0777); err != nil {
		log.Printf("[IPC] 创建targetDir目录失败: %v", err)
		return map[string]interface{}{"success": false, "message": "创建targetDir目录失败: " + err.Error()}
	}
	log.Printf("[IPC] targetDir目录已创建: %s", targetDir)

	// 调用dget库下载镜像，使用目标目录作为工作目录，真实进度会通过回调函数返回
	log.Printf("[IPC] 开始调用InstallWithTargetDir，参数: registry=%s, pkg=%s, tag=%s, arch=linux/amd64, targetDir=%s", registry, pkg, tag, targetDir)
	
	// 在goroutine中执行下载，以便可以检测取消信号
	downloadErr := make(chan error, 1)
	go func() {
		downloadErr <- dgetClient.InstallWithTargetDir(3, registry, pkg, tag, "linux/amd64", false, false, "", "", targetDir)
	}()
	
	// 等待下载完成或取消
	select {
	case <-ctx.Done():
		log.Printf("[IPC] 下载被取消")
		// 清理下载目录
		os.RemoveAll(targetDir)
		return map[string]interface{}{"success": false, "message": "下载已取消"}
	case downloadResult := <-downloadErr:
		err := downloadResult
		log.Printf("[IPC] InstallWithTargetDir执行完成，err: %v", err)

		// 再次检查targetDir目录是否存在
		if _, statErr := os.Stat(targetDir); os.IsNotExist(statErr) {
			log.Printf("[IPC] InstallWithTargetDir执行后，targetDir目录不存在: %s", targetDir)
		} else {
			// 列出targetDir目录下的文件
			files, readErr := os.ReadDir(targetDir)
			if readErr != nil {
				log.Printf("[IPC] 读取targetDir目录失败: %v", readErr)
			} else {
				log.Printf("[IPC] InstallWithTargetDir执行后，targetDir目录下的文件:")
				for _, file := range files {
					log.Printf("[IPC] - %s (目录: %t)", file.Name(), file.IsDir())
				}
			}
		}

		// 检查是否在当前目录下生成了tar.gz文件
		currentDir, _ := os.Getwd()
		currentDirFiles, _ := os.ReadDir(currentDir)
		log.Printf("[IPC] 当前目录下的tar.gz文件:")
		for _, file := range currentDirFiles {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
				log.Printf("[IPC] - %s", file.Name())
			}
		}

		if err != nil {
			log.Printf("[IPC] 下载镜像失败: %v", err)
			// 发送下载完成事件（失败）
			a.emitEvent("download-complete", map[string]interface{}{
				"success": false,
				"message": "下载镜像失败: " + err.Error(),
				"path":    "",
			})
			return map[string]interface{}{"success": false, "message": "下载镜像失败: " + err.Error()}
		}
	}

	// 更新进度为95%，表示下载完成，正在处理文件
	a.progressMutex.Lock()
	a.downloadProgress = 95
	progress := a.downloadProgress
	a.progressMutex.Unlock()

	// 发送下载进度事件给前端
	a.emitEvent("download-progress", map[string]interface{}{
		"progress": progress,
	})

	// 检查tar.gz文件是否存在
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		log.Printf("[IPC] 镜像文件不存在: %s", targetPath)
		// 列出targetDir目录下的文件，查看下载结果
		files, err := os.ReadDir(targetDir)
		if err != nil {
			log.Printf("[IPC] 读取targetDir目录失败: %v", err)
		} else {
			log.Printf("[IPC] targetDir目录下的文件:")
			for _, file := range files {
				log.Printf("[IPC] - %s (目录: %t)", file.Name(), file.IsDir())
			}
		}
		// 尝试查找targetDir目录下的tar.gz文件
		gzipFiles, _ := filepath.Glob(filepath.Join(targetDir, "*.tar.gz"))
		if len(gzipFiles) > 0 {
			log.Printf("[IPC] targetDir目录下找到tar.gz文件: %v", gzipFiles)
			// 如果找到tar.gz文件，尝试重命名到目标路径
			if err := os.Rename(gzipFiles[0], targetPath); err == nil {
				log.Printf("[IPC] 成功将文件重命名为: %s", targetPath)
			} else {
				log.Printf("[IPC] 重命名文件失败: %v", err)
				// 发送下载完成事件（失败）
				a.emitEvent("download-complete", map[string]interface{}{
					"success": false,
					"message": "重命名镜像文件失败: " + err.Error(),
					"path":    "",
				})
				return map[string]interface{}{"success": false, "message": "重命名镜像文件失败: " + err.Error()}
			}
		} else {
			// 发送下载完成事件（失败）
			a.emitEvent("download-complete", map[string]interface{}{
				"success": false,
				"message": "镜像文件不存在",
				"path":    "",
			})
			return map[string]interface{}{"success": false, "message": "镜像文件不存在"}
		}
	}

	// 可选：删除临时工作目录（如果需要）
	if err := os.RemoveAll(targetDir); err != nil {
		log.Printf("[IPC] 删除临时工作目录失败: %v", err)
		// 不影响主流程，继续执行
	}

	// 更新进度为100%
	a.progressMutex.Lock()
	a.downloadProgress = 100
	progress = a.downloadProgress
	a.progressMutex.Unlock()

	// 发送下载进度事件给前端
	a.emitEvent("download-progress", map[string]interface{}{
		"progress": progress,
	})

	// 创建镜像元数据
	fileInfo, err := os.Stat(targetPath)
	if err == nil {
		// 提取可用型号信息（优先从元数据中获取，其次从镜像名称中提取）
		availableModels := []string{}

		// 从元数据中获取ttype和ttype2作为可用型号
		if ttype := getStringFromMap(metadata, "ttype"); ttype != "" {
			availableModels = append(availableModels, ttype)
		}
		if ttype2 := getStringSliceFromMap(metadata, "ttype2"); len(ttype2) > 0 {
			availableModels = append(availableModels, ttype2...)
		}

		// 如果从元数据中没有获取到型号，从镜像名称中提取
		imageNameWithoutExt := strings.TrimSuffix(filepath.Base(targetPath), ".tar.gz")
		if len(availableModels) == 0 {
			// 简单的型号提取逻辑
			if strings.Contains(imageNameWithoutExt, "P14") {
				availableModels = append(availableModels, "P14")
			}
			if strings.Contains(imageNameWithoutExt, "v3") {
				availableModels = append(availableModels, "v3")
			}
			if strings.Contains(imageNameWithoutExt, "android") {
				availableModels = append(availableModels, "android")
			}
		}

		// 构造元数据
		imageMetadata := ImageMetadata{
			OnlineURL:       imageUrl,
			ImageName:       imageNameWithoutExt,
			LocalPath:       targetPath,
			Size:            formatFileSize(fileInfo.Size()),
			CreateTime:      fileInfo.ModTime().UnixMilli(),
			AvailableModels: availableModels,

			// 从参数中提取在线镜像完整元数据
			ID:        getStringFromMap(metadata, "id"),
			Spid:      getStringFromMap(metadata, "spid"),
			Type:      getStringFromMap(metadata, "type"),
			Name:      getStringFromMap(metadata, "name"),
			Url:       getStringFromMap(metadata, "url"),
			Sort:      getStringFromMap(metadata, "sort"),
			State:     getStringFromMap(metadata, "state"),
			Ttype:     getStringFromMap(metadata, "ttype"),
			OsVer:     getStringFromMap(metadata, "os_ver"),
			Udesc:     getStringFromMap(metadata, "udesc"),
			SysVer:    getStringFromMap(metadata, "sys_ver"),
			Ttype2:    getStringSliceFromMap(metadata, "ttype2"),
			SysVerDes: getStringFromMap(metadata, "sys_ver_des"),
		}

		// 保存元数据到JSON文件
		metadataPath := strings.TrimSuffix(targetPath, ".tar.gz") + ".json"
		metadataBytes, err := json.MarshalIndent(imageMetadata, "", "  ")
		if err == nil {
			if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
				log.Printf("[IPC] 保存元数据文件失败: %v", err)
			} else {
				log.Printf("[IPC] 保存元数据文件成功: %s", metadataPath)
			}
		} else {
			log.Printf("[IPC] 序列化元数据失败: %v", err)
		}
	}

	// 发送下载完成事件（成功）
	a.emitEvent("download-complete", map[string]interface{}{
		"success": true,
		"message": "下载镜像成功",
		"path":    targetPath,
	})

	log.Printf("[IPC] 下载镜像成功，保存到: %s", targetPath)
	return map[string]interface{}{"success": true, "message": "下载镜像成功", "path": targetPath}
}

// GetDownloadProgress 获取当前镜像下载进度
func (a *App) GetDownloadProgress() map[string]interface{} {
	a.progressMutex.Lock()
	progress := a.downloadProgress
	a.progressMutex.Unlock()

	return map[string]interface{}{
		"progress": progress,
	}
}

// CancelImageDownload 取消当前的镜像下载任务
func (a *App) CancelImageDownload() map[string]interface{} {
	log.Printf("[IPC] 收到取消下载请求")
	
	a.downloadMutex.Lock()
	defer a.downloadMutex.Unlock()
	
	if a.downloadCancel != nil {
		log.Printf("[IPC] 执行取消下载")
		a.downloadCancel()
		a.downloadCancel = nil
		a.downloadCtx = nil
		
		// 重置进度
		a.progressMutex.Lock()
		a.downloadProgress = 0
		a.progressMutex.Unlock()
		
		return map[string]interface{}{"success": true, "message": "下载已取消"}
	}
	
	log.Printf("[IPC] 没有正在进行的下载任务")
	return map[string]interface{}{"success": false, "message": "没有正在进行的下载任务"}
}

// CancelImageUpload 取消当前的镜像上传任务
func (a *App) CancelImageUpload() map[string]interface{} {
	log.Printf("[IPC] 收到取消上传请求")
	
	a.uploadMutex.Lock()
	defer a.uploadMutex.Unlock()
	
	if a.uploadCancel != nil {
		log.Printf("[IPC] 执行取消上传")
		a.uploadCancel()
		a.uploadCancel = nil
		a.uploadCtx = nil
		
		// 重置进度
		a.progressMutex.Lock()
		a.imagePullProgress = 0
		a.progressMutex.Unlock()
		
		return map[string]interface{}{"success": true, "message": "上传已取消"}
	}
	
	log.Printf("[IPC] 没有正在进行的上传任务")
	return map[string]interface{}{"success": false, "message": "没有正在进行的上传任务"}
}

// IsImageDownloaded 检查在线镜像是否已下载
func (a *App) IsImageDownloaded(onlineURL string) map[string]interface{} {
	log.Printf("[IPC] 收到 IsImageDownloaded 调用")
	log.Printf("[IPC] 参数: onlineURL=%s", onlineURL)

	// 方法1: 检查images目录下的文件（GetLocalImagesForManagement使用的目录）
	cacheDir := getCacheDir()
	imageDir := filepath.Join(cacheDir, "images")

	// 读取images目录下的所有.tar.gz文件
	files, err := os.ReadDir(imageDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
				// 提取镜像名称
				imageName := strings.TrimSuffix(file.Name(), ".tar.gz")
				tarGzPath := filepath.Join(imageDir, file.Name())

				// 方式1: 直接比较URL和文件名
				if strings.Contains(onlineURL, imageName) || strings.Contains(imageName, filepath.Base(onlineURL)) {
					log.Printf("[IPC] 镜像已下载到images目录: %s", onlineURL)
					return map[string]interface{}{
						"downloaded": true,
						"local_path": tarGzPath,
					}
				}

				// 方式2: 对于Docker镜像，比较tag部分
				if strings.Contains(onlineURL, ":") {
					// 提取Docker镜像的tag部分
					onlineTag := strings.Split(onlineURL, ":")[1]
					if strings.Contains(imageName, onlineTag) || strings.Contains(onlineTag, imageName) {
						log.Printf("[IPC] 镜像已下载到images目录（通过tag匹配）: %s", onlineURL)
						return map[string]interface{}{
							"downloaded": true,
							"local_path": tarGzPath,
						}
					}
				}

				// 方式3: 比较文件名和URL的最后一部分
				onlineBase := filepath.Base(onlineURL)
				onlineBaseNoExt := strings.TrimSuffix(onlineBase, filepath.Ext(onlineBase))
				if strings.Contains(imageName, onlineBaseNoExt) || strings.Contains(onlineBaseNoExt, imageName) {
					log.Printf("[IPC] 镜像已下载到images目录（通过basename匹配）: %s", onlineURL)
					return map[string]interface{}{
						"downloaded": true,
						"local_path": tarGzPath,
					}
				}

				// 方式4: 对于包含特定关键词的镜像名称，使用更宽松的匹配
				// 例如，"特别版-CQR14-ALL-v1.0.0" 和 "Q14_v3_all_202601091619"
				// 提取在线URL中的关键部分
				onlineParts := strings.Split(onlineURL, "/")
				if len(onlineParts) > 1 {
					// 获取URL的最后一部分，通常是镜像名称:tag
					lastPart := onlineParts[len(onlineParts)-1]
					// 进一步拆分得到tag
					tagParts := strings.Split(lastPart, ":")
					if len(tagParts) > 1 {
						tag := tagParts[1]
						// 检查tag中的关键部分是否在文件名中
						// 例如，"Q14_v3_all_202601091619" 中的 "Q14" 和 "all" 在 "特别版-CQR14-ALL-v1.0.0" 中
						tagLower := strings.ToLower(tag)
						imageNameLower := strings.ToLower(imageName)
						// 检查多个关键词匹配
						keywordMatches := 0
						keywords := []string{"q14", "cqr14", "all", "v3", "v1.0.0"}
						for _, keyword := range keywords {
							if strings.Contains(tagLower, keyword) && strings.Contains(imageNameLower, keyword) {
								keywordMatches++
							}
						}
						// 如果匹配到多个关键词，认为是同一个镜像
						if keywordMatches >= 2 {
							log.Printf("[IPC] 镜像已下载到images目录（通过关键词匹配）: %s", onlineURL)
							return map[string]interface{}{
								"downloaded": true,
								"local_path": tarGzPath,
							}
						}
					}
				}
			}
		}
	}

	// 方法2: 检查存储目录（支持自定义路径）
	edgeclientDir := getStorageBaseDir()

	// 检查目录是否存在
	if _, err := os.Stat(edgeclientDir); os.IsNotExist(err) {
		log.Printf("[IPC] 存储目录不存在")
		return map[string]interface{}{"downloaded": false}
	}

	// 读取目录下的所有.json文件（元数据文件）
	files, err = os.ReadDir(edgeclientDir)
	if err != nil {
		log.Printf("[IPC] 读取存储目录失败: %v", err)
		return map[string]interface{}{"downloaded": false}
	}

	// 检查每个元数据文件
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			metadataPath := filepath.Join(edgeclientDir, file.Name())
			metadataBytes, err := os.ReadFile(metadataPath)
			if err != nil {
				continue
			}

			var metadata ImageMetadata
			if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
				continue
			}

			if metadata.OnlineURL == onlineURL {
				// 检查对应的tar.gz文件是否存在
				tarGzPath := strings.TrimSuffix(metadataPath, ".json") + ".tar.gz"
				if _, err := os.Stat(tarGzPath); err == nil {
					log.Printf("[IPC] 镜像已下载到edgeclient目录: %s", onlineURL)
					return map[string]interface{}{
						"downloaded": true,
						"local_path": tarGzPath,
						"metadata":   metadata,
					}
				}
			}
		}
	}

	log.Printf("[IPC] 镜像未下载: %s", onlineURL)
	return map[string]interface{}{"downloaded": false}
}

// simulateDownloadProgress 模拟下载进度更新
func (a *App) simulateDownloadProgress() {
	// 模拟进度更新，实际应该从dget库获取进度
	// 这里我们使用模拟进度，每500毫秒更新一次，直到进度达到100%
	for {
		a.progressMutex.Lock()
		currentProgress := a.downloadProgress
		a.progressMutex.Unlock()

		if currentProgress >= 100 {
			break
		}

		// 更新进度，每次增加1%
		a.progressMutex.Lock()
		a.downloadProgress += 1
		currentProgress = a.downloadProgress
		a.progressMutex.Unlock()

		// 发送下载进度事件给前端
		a.emitEvent("download-progress", map[string]interface{}{
			"progress": currentProgress,
		})

		// 等待500毫秒
		time.Sleep(500 * time.Millisecond)
	}
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CreateV0V2Device 实现v0-v2设备的创建流程
func (a *App) CreateV0V2Device(deviceIP string, createParams map[string]interface{}) map[string]interface{} {
	log.Printf("[IPC] 收到 CreateV0V2Device 调用")
	log.Printf("[IPC] 参数: deviceIP=%s, createParams=%+v", deviceIP, createParams)

	// 对于v0-v2设备，版本默认为v0
	version := "v0"
	password := "" // 暂时使用空密码，后续可以从参数中获取

	// 步骤1: 检查myt_sdk容器是否存在以及运行状态
	containerExists, isRunning, err := checkMytSdkContainerExists(deviceIP, version, password)
	if err != nil {
		log.Printf("[IPC] 检查myt_sdk容器失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("检查myt_sdk容器失败: %v", err),
		}
	}

	if !containerExists {
		log.Printf("[IPC] myt_sdk容器不存在，开始创建流程")

		// 步骤2: 加载镜像到设备
		imagePath := "oldversion_api_refrence/myt_supser_sdk_v1.0.14.30.42.tar.gz"
		// 检查镜像文件是否存在
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			log.Printf("[IPC] 镜像文件不存在: %s", imagePath)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("镜像文件不存在: %s", imagePath),
			}
		}

		// 加载镜像
		_, err := loadImageToDevice(a, deviceIP, version, imagePath, password)
		if err != nil {
			log.Printf("[IPC] 加载镜像失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("加载镜像失败: %v", err),
			}
		}

		// 步骤3: 创建并启动myt_sdk容器
		if err := createMytSdkContainer(deviceIP, version, password); err != nil {
			log.Printf("[IPC] 创建myt_sdk容器失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建myt_sdk容器失败: %v", err),
			}
		}
	} else if !isRunning {
		log.Printf("[IPC] myt_sdk容器存在但未运行，开始启动流程")
		// 步骤3: 启动myt_sdk容器
		if success, message := startDockerContainer(deviceIP, "myt_sdk", version, password); !success {
			log.Printf("[IPC] 启动myt_sdk容器失败: %s", message)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("启动myt_sdk容器失败: %s", message),
			}
		}
	} else {
		log.Printf("[IPC] myt_sdk容器已存在且运行正常，跳过创建流程")
	}

	// 步骤4: 如果传入了创建参数，调用batch_create API
	if createParams != nil && len(createParams) > 0 {
		log.Printf("[IPC] 开始调用batch_create API创建云机")
		result, err := batchCreateV0V2Device(deviceIP, createParams)
		if err != nil {
			log.Printf("[IPC] 调用batch_create API失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建云机失败: %v", err),
			}
		}
		return result
	}

	result := map[string]interface{}{
		"success": true,
		"message": "v0-v2设备创建流程完成",
	}
	log.Printf("[IPC] CreateV0V2Device 返回结果: %+v", result)
	return result
}

// batchCreateV0V2Device 调用v0-v2设备的batch_create API创建云机
func batchCreateV0V2Device(deviceIP string, params map[string]interface{}) (map[string]interface{}, error) {
	// 提取参数
	num := "1"
	if count, ok := params["count"].(float64); ok {
		num = fmt.Sprintf("%d", int(count))
	} else if count, ok := params["count"].(string); ok {
		num = count
	}

	preName := "T000"
	if name, ok := params["name"].(string); ok && name != "" {
		preName = name
	}

	force := "1"
	if forceVal, ok := params["force"].(bool); ok && forceVal {
		force = "1"
	}

	// 构造查询参数
	queryParams := make(map[string]string)

	// 沙盒模式处理
	// sandboxSize > 0 时启用沙盒模式，否则不启用
	sandbox := "0"
	if sandboxSize, ok := params["sandboxSize"].(float64); ok && sandboxSize > 0 {
		sandbox = "1"
		queryParams["sandbox"] = sandbox
		queryParams["sandbox_size"] = fmt.Sprintf("%d", int(sandboxSize))
	} else {
		// 沙盒大小为0或未设置时，不传递sandbox_size参数
		queryParams["sandbox"] = sandbox
	}

	// 镜像地址
	imageAddr := ""
	if imageSelect, ok := params["imageSelect"].(string); ok && imageSelect != "" {
		if imageSelect == "custom" {
			if customUrl, ok := params["customImageUrl"].(string); ok && customUrl != "" {
				imageAddr = customUrl
			}
		} else {
			imageAddr = imageSelect
		}
	}
	if imageAddr != "" {
		queryParams["image_addr"] = imageAddr
	}

	// 分辨率参数 (0=720P, 1=1080P, 2=自定义)
	resolution := "0"
	width := "720"
	height := "1280"
	if res, ok := params["resolution"].(string); ok && res != "" {
		parts := strings.Split(res, "x")
		if len(parts) == 2 {
			width = parts[0]
			height = parts[1]
			// 根据分辨率判断参数值
			if width == "720" && height == "1280" {
				resolution = "0"
			} else if width == "1080" && height == "1920" {
				resolution = "1"
			} else {
				resolution = "2"
			}
		}
	}
	queryParams["resolution"] = resolution
	queryParams["width"] = width
	queryParams["height"] = height
	queryParams["dpi"] = "320"
	queryParams["fps"] = "24"

	// DNS
	if dns, ok := params["dns"].(string); ok && dns != "" {
		queryParams["dns"] = dns
	} else {
		queryParams["dns"] = "223.5.5.5"
	}

	// 随机设备信息
	queryParams["random_dev"] = "1"

	// 强制模式
	queryParams["enforce"] = "1"

	// 构造URL
	url := fmt.Sprintf("http://%s:81/dc_api/v1/batch_create/%s/%s/%s/%s", deviceIP, deviceIP, num, preName, force)

	// 添加查询参数
	if len(queryParams) > 0 {
		url += "?"
		first := true
		for key, value := range queryParams {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	log.Printf("batch_create API URL: %s", url)

	// 发送HTTP POST请求
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("调用batch_create API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取batch_create API响应失败: %v", err)
	}

	log.Printf("batch_create API响应: %s", string(body))

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("batch_create API失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果解析失败，返回原始响应
		return map[string]interface{}{
			"success":  true,
			"message":  "创建请求已发送",
			"response": string(body),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"message": "云机创建成功",
		"result":  result,
	}, nil
}

// GetContainerPort 获取容器端口映射
// 返回值：publicPort - 映射后的公共端口，如果没有找到映射则返回privatePort
//
//	found - 是否找到端口映射
func getContainerPort(info interface{}, privatePort int) (publicPort int, found bool) {
	// 默认返回原始端口和未找到映射标记
	publicPort = privatePort
	found = false

	switch v := info.(type) {
	case []DockerContainer:
		// Docker容器列表，假设只有一个容器
		if len(v) > 0 {
			container := v[0]
			// 遍历Docker端口列表，查找匹配的privatePort
			for _, port := range container.Ports {
				if port.PrivatePort == privatePort && port.PublicPort > 0 {
					publicPort = port.PublicPort
					found = true
					return
				}
			}

			// 尝试从NetworkSettings中查找端口映射
			if container.NetworkSettings != nil && container.NetworkSettings.Ports != nil {
				portKey := fmt.Sprintf("%d/tcp", privatePort)
				if bindings, ok := container.NetworkSettings.Ports[portKey]; ok && len(bindings) > 0 {
					if hostPort, err := strconv.Atoi(bindings[0].HostPort); err == nil {
						publicPort = hostPort
						found = true
						return
					}
				}
			}
		}

	case map[string]interface{}:
		// V3 API响应，包含list字段
		if list, ok := v["list"].([]interface{}); ok && len(list) > 0 {
			item := list[0]
			if containerInfo, ok := item.(map[string]interface{}); ok {
				// 首先尝试从PortBindings中查找实际的端口映射
				if portBindings, ok := containerInfo["portBindings"].(map[string]interface{}); ok {
					portKey := fmt.Sprintf("%d/tcp", privatePort)
					if bindings, ok := portBindings[portKey].([]interface{}); ok && len(bindings) > 0 {
						if binding, ok := bindings[0].(map[string]interface{}); ok {
							if hostPort, ok := binding["HostPort"].(string); ok {
								if parsedPort, err := strconv.Atoi(hostPort); err == nil {
									publicPort = parsedPort
									found = true
									return
								}
							}
						}
					}
				}

				// 如果没有找到实际的端口映射，再使用硬编码的规则作为后备
				if indexNum, ok := containerInfo["indexNum"].(float64); ok {
					if privatePort == 9083 {
						publicPort = 11000 + int(indexNum)*10
						found = true
						return
					} else if privatePort == 9082 {
						publicPort = 11000 + int(indexNum)*10 - 1
						found = true
						return
					}
				}
			}
		}
	}

	return
}

// StartProjectionWindow 使用 Wails V3 多窗口创建投屏窗口
func (a *App) StartProjectionWindow(config ProjectionConfig) map[string]interface{} {
	log.Printf("[IPC] 收到 StartProjectionWindow 调用")
	log.Printf("[IPC] 参数: %+v", config)

	extractContainerName := func(fullName string) string {
		if fullName == "" {
			return fullName
		}
		if idx := strings.LastIndex(fullName, "_"); idx != -1 && idx < len(fullName)-1 {
			return fullName[idx+1:]
		}
		return fullName
	}

	if a.wailsApp == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Wails应用未初始化",
		}
	}

	if config.DeviceIP == "" && config.List == "" {
		return map[string]interface{}{
			"success": false,
			"message": "设备IP不能为空",
		}
	}

	if config.Width == 0 {
		config.Width = 360
	}
	if config.Height == 0 {
		config.Height = 640
	}

	// 使用 ip_容器名 作为唯一标识（而不是容器ID）
	containerName := config.ContainerName
	if containerName == "" {
		if config.List != "" {
			containerName = "batch_control"
		} else if config.ContainerID != "" {
			containerName = config.ContainerID
		} else {
			containerName = "unknown"
		}
	}
	
	// 生成唯一的窗口标识：ip_容器名
	windowID := config.DeviceIP + "_" + containerName

	windowTitle := "投屏"
	if config.ContainerName != "" {
		windowTitle = fmt.Sprintf("投屏 - %s", extractContainerName(config.ContainerName))
	} else if config.ContainerID != "" {
		windowTitle = fmt.Sprintf("投屏 - %s", config.ContainerID)
	}

	if runtime.GOOS == "windows" {
		return a.startWindowsProjectionProcess(config, windowID, windowTitle)
	}

	if config.TCPPort == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "TCP端口不能为空",
		}
	}

	existingWindow, exists := a.projectionWindows[windowID]
	if exists && existingWindow != nil {
		// 检查窗口是否仍然有效（可见）
		if existingWindow.IsVisible() {
			log.Printf("[IPC] 投屏窗口已存在，聚焦窗口: %s", windowID)
			existingWindow.Focus()
			return map[string]interface{}{
				"success":  true,
				"message":  "投屏窗口已聚焦",
				"windowID": windowID,
				"focused":  true,
				"deviceIP": config.DeviceIP,
				"tcpPort":  config.TCPPort,
				"udpPort":  config.UDPPort,
			}
		}
		// 窗口已关闭但引用仍在map中，清理后创建新窗口
		log.Printf("[IPC] 投屏窗口已关闭但引用存在，清理并创建新窗口: %s", windowID)
		delete(a.projectionWindows, windowID)
	}

	width := config.Width
	height := config.Height

	playerURL := fmt.Sprintf(
		"/webplayer/play.html?shost=%s&sport=%d&q=1&v=h264&rtc_i=%s&rtc_p=%d&container_name=%s",
		config.DeviceIP,
		config.TCPPort,
		config.DeviceIP,
		config.UDPPort,
		url.QueryEscape(config.ContainerName),
	)

	log.Printf("[IPC] 创建投屏窗口，URL: %s", playerURL)

	// 检查是否启用调试模式
	enableDebug := os.Getenv("APP_DEBUG")
	isDebugMode := strings.ToLower(enableDebug) == "true" || enableDebug == "1"

	window := a.wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       windowID,
		Title:                      windowTitle,
		Width:                      width,
		Height:                     height,
		BackgroundColour:           application.NewRGB(0, 0, 0),
		URL:                        playerURL,
		DisableResize:              false,
		Frameless:                  false,
		StartState:                 application.WindowStateNormal,
		DevToolsEnabled:            isDebugMode,
		DefaultContextMenuDisabled: !isDebugMode,
	})

	window.Show()

	// 注册窗口关闭事件监听，自动从 map 中清理
	window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		log.Printf("[IPC] 投屏窗口 %s 正在关闭", windowID)
		delete(a.projectionWindows, windowID)
		delete(a.windowAlwaysOnTop, windowID)
	})

	a.projectionWindows[windowID] = window
	a.windowAlwaysOnTop[windowID] = false

	log.Printf("[IPC] 投屏窗口创建成功: %s", windowID)

	return map[string]interface{}{
		"success":     true,
		"message":     "投屏窗口已创建",
		"windowID":    windowID,
		"focused":     false,
		"windowTitle": windowTitle,
		"deviceIP":    config.DeviceIP,
		"tcpPort":     config.TCPPort,
		"udpPort":     config.UDPPort,
	}
}

func (a *App) startWindowsProjectionProcess(config ProjectionConfig, windowID, windowTitle string) map[string]interface{} {
	log.Printf("[投屏] ========== 开始处理投屏请求 ==========")
	log.Printf("[投屏] windowID=%s, IP=%s, 容器=%s", windowID, config.DeviceIP, config.ContainerName)
	
	projectionLock.Lock()
	log.Printf("[投屏] 当前runningProjections中有 %d 个进程", len(runningProjections))
	for key, proc := range runningProjections {
		log.Printf("[投屏] - %s: PID=%d", key, proc.Pid)
	}
	
	if proc, exists := runningProjections[windowID]; exists {
		log.Printf("[投屏] 找到已存在的进程记录，PID=%d", proc.Pid)
		if isProcessRunning(proc) {
			projectionLock.Unlock()
			log.Printf("[投屏] ✓ 投屏窗口已存在且运行中(PID=%d)，尝试聚焦", proc.Pid)
			// 通过PID精确查找并置顶窗口
			if err := bringProcessToFront(proc.Pid); err != nil {
				log.Printf("[投屏] ✗ 将投屏进程置前失败: %v", err)
			} else {
				log.Printf("[投屏] ✓ 成功通过PID %d 将投屏窗口置前", proc.Pid)
			}
			return map[string]interface{}{
				"success":  true,
				"message":  "投屏窗口已聚焦",
				"windowID": windowID,
				"focused":  true,
				"deviceIP": config.DeviceIP,
				"tcpPort":  config.TCPPort,
				"udpPort":  config.UDPPort,
			}
		}
		// 进程已退出，从map中清理
		log.Printf("[投屏] ✗ 检测到投屏进程已退出，清理记录: %s", windowID)
		delete(runningProjections, windowID)
	} else {
		log.Printf("[投屏] 未找到已存在的进程记录，准备启动新进程")
	}

	// 检查是否正在启动中（解压/查找player阶段）
	if pendingProjections[windowID] {
		projectionLock.Unlock()
		log.Printf("[投屏] 投屏正在启动中，忽略重复请求: %s", windowID)
		return map[string]interface{}{
			"success":  true,
			"message":  "投屏正在启动中，请稍候...",
			"windowID": windowID,
			"focused":  false,
			"deviceIP": config.DeviceIP,
			"tcpPort":  config.TCPPort,
			"udpPort":  config.UDPPort,
		}
	}

	// 标记为启动中，防止重复启动
	pendingProjections[windowID] = true
	projectionLock.Unlock()

	videoPort := config.TCPPort
	controlPort := config.ControlPort
	if controlPort == 0 {
		controlPort = config.UDPPort
	}

	if videoPort == 0 || controlPort == 0 {
		projectionLock.Lock()
		delete(pendingProjections, windowID)
		projectionLock.Unlock()
		return map[string]interface{}{
			"success": false,
			"message": "端口映射信息不完整，无法启动投屏",
		}
	}

	orient := config.Orient
	if orient != 0 && orient != 1 {
		if config.Width >= config.Height {
			orient = 1
		} else {
			orient = 0
		}
	}

	// 横屏模式下，如果宽<高则交换宽高，确保传给player的尺寸与方向一致
	if orient == 1 && config.Width < config.Height {
		config.Width, config.Height = config.Height, config.Width
	}

	term := config.Term
	if term == "" {
		term = windowTitle
	}

	playerExe, err := findPlayerExecutable()
	if err != nil {
		log.Printf("[投屏] 未找到投屏程序: %v", err)
		projectionLock.Lock()
		delete(pendingProjections, windowID)
		projectionLock.Unlock()
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	args := []string{
		"-ip", config.DeviceIP,
		"-vport", strconv.Itoa(videoPort),
		"-cport", strconv.Itoa(controlPort),
	}
	if orient == 1 {
		args = append(args, "-landscape")
	}
	if config.List != "" {
		args = append(args, "-list", config.List)
	}
	if term != "" {
		args = append(args, "-title", term)
	}

	log.Printf("[投屏] 启动新投屏进程: %s %v", playerExe, args)

	// Log extracted directory contents for remote diagnostics
	playerDir := filepath.Dir(playerExe)
	var fileList []string
	if entries, err := os.ReadDir(playerDir); err == nil {
		for _, e := range entries {
			if info, err := e.Info(); err == nil {
				fileList = append(fileList, fmt.Sprintf("%s(%d)", e.Name(), info.Size()))
			}
		}
		log.Printf("[Player] 投屏目录文件: %s", strings.Join(fileList, ", "))
	}

	cmd := exec.Command(playerExe, args...)
	cmd.Dir = playerDir
	// CREATE_NO_WINDOW (0x08000000) prevents console allocation for non-GUI builds,
	// without hiding the actual player GUI window (unlike HideWindow which hides everything).
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

	// Capture player stderr for remote diagnostics
	stderrPipe, pipeErr := cmd.StderrPipe()
	if pipeErr != nil {
		log.Printf("[Player] 无法捕获stderr: %v", pipeErr)
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[投屏] 启动投屏失败: %v", err)
		projectionLock.Lock()
		delete(pendingProjections, windowID)
		projectionLock.Unlock()
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("启动投屏失败: %v", err),
		}
	}

	log.Printf("[投屏] 投屏进程启动成功，PID=%d, windowID=%s", cmd.Process.Pid, windowID)

	// 注册进程到map并清除pending标记
	projectionLock.Lock()
	runningProjections[windowID] = cmd.Process
	delete(pendingProjections, windowID)
	projectionLock.Unlock()

	// Log player stderr output for remote diagnostics
	if stderrPipe != nil {
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				log.Printf("[Player-%s] %s", windowID, scanner.Text())
			}
		}()
	}

	// Wait briefly to detect immediate crash (e.g., missing DLL)
	// 仅用 500ms 窗口，足以抓住真正的启动失败（缺少 DLL 约 50-100ms 即崩溃），
	// 但不会误捕用户手动关闭窗口（窗口出现后用户操作通常 > 500ms）
	crashCh := make(chan error, 1)
	go func() { crashCh <- cmd.Wait() }()
	select {
	case err := <-crashCh:
		// Player exited within 500ms
		projectionLock.Lock()
		delete(runningProjections, windowID)
		projectionLock.Unlock()

		if err != nil {
			log.Printf("[Player] 投屏进程快速退出 (exit: %v)。目录: %s, 文件: %s",
				err, playerDir, strings.Join(fileList, ", "))
		} else {
			log.Printf("[Player] 投屏进程已正常退出 (exit: 0), windowID=%s", windowID)
		}
		// 无论退出码如何，都不向用户弹错误——可能是用户快速关闭、重复实例等
		return map[string]interface{}{
			"success":  true,
			"message":  "投屏已启动",
			"windowID": windowID,
			"focused":  false,
			"deviceIP": config.DeviceIP,
			"tcpPort":  config.TCPPort,
			"udpPort":  config.UDPPort,
		}
	case <-time.After(500 * time.Millisecond):
		// Player survived 500ms - it's running OK
		log.Printf("[Player] 投屏进程启动正常 (500ms检查通过)")
	}

	// Monitor for later exit (reuse the goroutine that already called cmd.Wait)
	go func() {
		err := <-crashCh // wait for the already-running cmd.Wait goroutine
		projectionLock.Lock()
		defer projectionLock.Unlock()
		if _, exists := runningProjections[windowID]; exists {
			delete(runningProjections, windowID)
			log.Printf("[投屏] 投屏进程退出，已清理记录: %s (PID=%d)", windowID, cmd.Process.Pid)
		}
		if err != nil {
			log.Printf("[投屏] 投屏进程异常退出: windowID=%s, error=%v", windowID, err)
		}
	}()

	return map[string]interface{}{
		"success":  true,
		"message":  "投屏已启动",
		"windowID": windowID,
		"focused":  false,
		"deviceIP": config.DeviceIP,
		"tcpPort":  config.TCPPort,
		"udpPort":  config.UDPPort,
	}
}

func getPlayerExtractRoot() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		if cacheDir, err := os.UserCacheDir(); err == nil && cacheDir != "" {
			base = cacheDir
		}
	}
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "edgeclient", "player")
}

func getPlayerExtractVersionDir(hash string) string {
	return filepath.Join(getPlayerExtractRoot(), hash)
}

// 缓存嵌入 player 资源哈希，避免每次都遍历并读取 ~67MB 嵌入文件计算哈希
var (
	cachedPlayerHash     string
	cachedPlayerHashErr  error
	cachedPlayerHashOnce sync.Once
)

func embeddedPlayerHash() (string, error) {
	cachedPlayerHashOnce.Do(func() {
		start := time.Now()
		log.Printf("[Player] 开始计算嵌入投屏资源哈希...")
		h := sha256.New()
		cachedPlayerHashErr = fs.WalkDir(playerAssets, "player_dist", func(embedPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if _, err := h.Write([]byte(embedPath)); err != nil {
				return err
			}
			if _, err := h.Write([]byte{0}); err != nil {
				return err
			}
			file, err := playerAssets.Open(embedPath)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(h, file); err != nil {
				return err
			}
			if _, err := h.Write([]byte{0}); err != nil {
				return err
			}
			return nil
		})
		if cachedPlayerHashErr == nil {
			cachedPlayerHash = fmt.Sprintf("%x", h.Sum(nil))
		}
		log.Printf("[Player] 嵌入投屏资源哈希计算完成，耗时 %v", time.Since(start))
	})
	return cachedPlayerHash, cachedPlayerHashErr
}

func acquireExtractLock(lockPath string, timeout time.Duration) (*os.File, error) {
	start := time.Now()
	for {
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
		if err == nil {
			return lockFile, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		if info, statErr := os.Stat(lockPath); statErr == nil {
			if time.Since(info.ModTime()) > 2*time.Minute {
				_ = os.Remove(lockPath)
			}
		}
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("获取投屏解压锁超时")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func releaseExtractLock(lockFile *os.File, lockPath string) {
	if lockFile != nil {
		_ = lockFile.Close()
	}
	_ = os.Remove(lockPath)
}

func isPlayerExtractReady(dir string) bool {
	exePath := filepath.Join(dir, "player.exe")
	readyPath := filepath.Join(dir, ".ready")
	if st, err := os.Stat(exePath); err == nil && st.Size() > 0 {
		if _, err := os.Stat(readyPath); err == nil {
			return true
		}
	}
	return false
}

func ensureEmbeddedPlayerExtracted() (string, error) {
	extractStart := time.Now()
	embeddedExe := "player_dist/player.exe"
	if _, err := fs.Stat(playerAssets, embeddedExe); err != nil {
		return "", fmt.Errorf("嵌入的 player.exe 不存在，请将 player.exe 放入 player_dist/ 目录并重新构建")
	}

	hash, err := embeddedPlayerHash()
	if err != nil {
		return "", fmt.Errorf("计算投屏资源版本失败: %v", err)
	}
	log.Printf("[Player] 哈希计算阶段耗时 %v", time.Since(extractStart))

	baseDir := getPlayerExtractRoot()
	versionDir := getPlayerExtractVersionDir(hash)
	exePath := filepath.Join(versionDir, "player.exe")
	readyPath := filepath.Join(versionDir, ".ready")

	if isPlayerExtractReady(versionDir) {
		log.Printf("[Player] 投屏资源已就绪(跳过解压)，总耗时 %v", time.Since(extractStart))
		return exePath, nil
	}

	log.Printf("[Player] 投屏资源未就绪，开始解压...")

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("创建投屏解压目录失败: %v", err)
	}

	lockPath := filepath.Join(baseDir, ".extract.lock")
	lockFile, err := acquireExtractLock(lockPath, 30*time.Second)
	if err != nil {
		if isPlayerExtractReady(versionDir) {
			return exePath, nil
		}
		return "", err
	}
	defer releaseExtractLock(lockFile, lockPath)

	if isPlayerExtractReady(versionDir) {
		return exePath, nil
	}

	tmpDir := filepath.Join(baseDir, fmt.Sprintf(".tmp-%s-%d", hash, os.Getpid()))
	_ = os.RemoveAll(tmpDir)
	copyStart := time.Now()
	if err := extractEmbeddedDir("player_dist", tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("解压投屏资源失败: %v", err)
	}
	log.Printf("[Player] 文件解压阶段耗时 %v", time.Since(copyStart))

	if err := os.WriteFile(filepath.Join(tmpDir, ".ready"), []byte(hash+"\n"+time.Now().Format(time.RFC3339)), 0644); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("写入投屏资源标记失败: %v", err)
	}

	if isPlayerExtractReady(versionDir) {
		_ = os.RemoveAll(tmpDir)
		return exePath, nil
	}

	if err := os.RemoveAll(versionDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("清理旧投屏目录失败: %v", err)
	}

	if err := os.Rename(tmpDir, versionDir); err != nil {
		if isPlayerExtractReady(versionDir) {
			_ = os.RemoveAll(tmpDir)
			return exePath, nil
		}
		return "", fmt.Errorf("投屏资源目录替换失败: %v", err)
	}

	if _, err := os.Stat(exePath); err != nil {
		return "", fmt.Errorf("投屏程序不存在: %v", err)
	}
	if _, err := os.Stat(readyPath); err != nil {
		return "", fmt.Errorf("投屏资源标记缺失: %v", err)
	}

	log.Printf("[Player] 投屏资源解压完成，总耗时 %v", time.Since(extractStart))
	return exePath, nil
}

func extractEmbeddedDir(embedRoot, targetDir string) error {
	return fs.WalkDir(playerAssets, embedRoot, func(embedPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(embedPath, embedRoot)
		if rel == "" {
			return nil
		}
		if strings.HasPrefix(rel, "/") {
			rel = rel[1:]
		}
		dstPath := filepath.Join(targetDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		src, err := playerAssets.Open(embedPath)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			return err
		}
		if err := dst.Close(); err != nil {
			return err
		}
		_ = os.Chmod(dstPath, 0755)
		return nil
	})
}

func findPlayerExecutable() (string, error) {
	// Check local paths FIRST (allows dev override without rebuilding main app)
	// Search for webview_demo.exe (new) and player.exe (legacy) in local paths
	candidates := []string{}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "player", "bin", "webview_demo.exe"),
			filepath.Join(cwd, "player", "webview_demo.exe"),
			filepath.Join(cwd, "player", "bin", "player.exe"),
			filepath.Join(cwd, "player", "player.exe"),
			filepath.Join(cwd, "player.exe"),
		)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "player", "bin", "webview_demo.exe"),
			filepath.Join(exeDir, "player", "webview_demo.exe"),
			filepath.Join(exeDir, "player", "bin", "player.exe"),
			filepath.Join(exeDir, "player", "player.exe"),
			filepath.Join(exeDir, "player.exe"),
			filepath.Join(exeDir, "..", "player", "bin", "webview_demo.exe"),
			filepath.Join(exeDir, "..", "player", "player.exe"),
			filepath.Join(exeDir, "..", "..", "player", "bin", "webview_demo.exe"),
			filepath.Join(exeDir, "..", "..", "player", "player.exe"),
		)
	}

	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		absPath, err := filepath.Abs(cleaned)
		if err == nil {
			cleaned = absPath
		}
		if _, err := os.Stat(cleaned); err == nil {
			log.Printf("[Player] 使用本地投屏程序: %s", cleaned)
			return cleaned, nil
		}
	}

	// Fall back to embedded player（失败后自动重试一次）
	var embedErr error
	for attempt := 0; attempt < 2; attempt++ {
		if exePath, err := ensureEmbeddedPlayerExtracted(); err == nil && exePath != "" {
			if _, statErr := os.Stat(exePath); statErr == nil {
				return exePath, nil
			}
		} else if err != nil {
			log.Printf("[Player] 内嵌投屏资源提取失败(第%d次): %v", attempt+1, err)
			embedErr = err
			if attempt == 0 {
				// 第一次失败，清理可能残留的锁文件和临时目录后重试
				extractRoot := getPlayerExtractRoot()
				lockPath := filepath.Join(extractRoot, ".extract.lock")
				_ = os.Remove(lockPath)
				// 清理可能残留的 .tmp- 临时目录
				if entries, readErr := os.ReadDir(extractRoot); readErr == nil {
					for _, entry := range entries {
						if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
							_ = os.RemoveAll(filepath.Join(extractRoot, entry.Name()))
						}
					}
				}
				log.Printf("[Player] 已清理残留文件，准备重试解压...")
				continue
			}
		}
		break
	}

	// 最后兜底：扫描 edgeclient/player/ 已有的解压目录
	extractRoot := getPlayerExtractRoot()
	if entries, err := os.ReadDir(extractRoot); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dir := filepath.Join(extractRoot, entry.Name())
			// 优先找 webview_demo.exe（新版播放器）
			wvPath := filepath.Join(dir, "webview_demo.exe")
			if st, err := os.Stat(wvPath); err == nil && st.Size() > 0 {
				log.Printf("[Player] 使用已解压的投屏程序(兜底): %s", wvPath)
				return wvPath, nil
			}
			// 其次找 player.exe
			pPath := filepath.Join(dir, "player.exe")
			if st, err := os.Stat(pPath); err == nil && st.Size() > 0 {
				log.Printf("[Player] 使用已解压的投屏程序(兜底): %s", pPath)
				return pPath, nil
			}
		}
	}

	// 将解压失败的具体原因包含在错误信息中，方便用户反馈
	if embedErr != nil {
		return "", fmt.Errorf("未找到投屏程序 player.exe (解压失败: %v)", embedErr)
	}
	return "", fmt.Errorf("未找到投屏程序 player.exe")
}

// CloseProjectionWindow 关闭指定投屏窗口
func (a *App) CloseProjectionWindow(windowIDOrContainerID string) map[string]interface{} {
	log.Printf("[投屏] 收到关闭投屏请求, windowIDOrContainerID=%s", windowIDOrContainerID)

	// 首先尝试精确匹配 Webview 窗口
	window, exists := a.projectionWindows[windowIDOrContainerID]
	if exists {
		window.Close()
		delete(a.projectionWindows, windowIDOrContainerID)
		delete(a.windowAlwaysOnTop, windowIDOrContainerID)
		log.Printf("[投屏] Webview窗口已关闭: %s", windowIDOrContainerID)
		return map[string]interface{}{
			"success":  true,
			"message":  "投屏窗口已关闭",
			"windowID": windowIDOrContainerID,
		}
	}

	// 尝试精确匹配 Windows 进程
	projectionLock.Lock()
	proc, procExists := runningProjections[windowIDOrContainerID]
	if procExists && proc != nil {
		log.Printf("[投屏] 找到精确匹配的进程，PID=%d", proc.Pid)
		proc.Kill()
		delete(runningProjections, windowIDOrContainerID)
		projectionLock.Unlock()
		log.Printf("[投屏] 投屏进程已关闭: %s", windowIDOrContainerID)
		return map[string]interface{}{
			"success":  true,
			"message":  "投屏窗口已关闭",
			"windowID": windowIDOrContainerID,
		}
	}
	projectionLock.Unlock()

	// 如果精确匹配失败，尝试查找包含 containerID 的 Webview 窗口
	closedCount := 0
	for id, win := range a.projectionWindows {
		if strings.HasSuffix(id, "_"+windowIDOrContainerID) {
			win.Close()
			delete(a.projectionWindows, id)
			delete(a.windowAlwaysOnTop, id)
			closedCount++
			log.Printf("[投屏] Webview窗口已关闭(模糊匹配): %s", id)
		}
	}
	if closedCount > 0 {
		return map[string]interface{}{
			"success":  true,
			"message":  fmt.Sprintf("已关闭 %d 个投屏窗口", closedCount),
			"windowID": windowIDOrContainerID,
			"count":    closedCount,
		}
	}

	// 尝试查找包含 containerID 的 Windows 进程
	projectionLock.Lock()
	log.Printf("[投屏] 当前运行中的投屏进程:")
	for id, p := range runningProjections {
		log.Printf("[投屏] - %s: PID=%d", id, p.Pid)
	}
	
	for id, p := range runningProjections {
		// 支持后缀匹配：ip_batch_control 匹配 batch_control
		// 或者包含匹配（用于批量投屏控制）
		if strings.HasSuffix(id, "_"+windowIDOrContainerID) || 
		   (windowIDOrContainerID == "batch_control" && strings.Contains(id, "批量")) {
			if p != nil {
				log.Printf("[投屏] 找到匹配的进程(模糊匹配)，ID=%s, PID=%d", id, p.Pid)
				p.Kill()
			}
			delete(runningProjections, id)
			closedCount++
			log.Printf("[投屏] 投屏进程已关闭(模糊匹配): %s", id)
		}
	}
	projectionLock.Unlock()

	if closedCount > 0 {
		return map[string]interface{}{
			"success":  true,
			"message":  fmt.Sprintf("已关闭 %d 个投屏窗口", closedCount),
			"windowID": windowIDOrContainerID,
			"count":    closedCount,
		}
	}

	log.Printf("[投屏] 未找到匹配的投屏窗口: %s", windowIDOrContainerID)
	return map[string]interface{}{
		"success": false,
		"message": "投屏窗口不存在",
	}
}

// CloseAllProjectionWindows 关闭所有投屏窗口
func (a *App) CloseAllProjectionWindows() map[string]interface{} {
	log.Printf("[IPC] 收到 CloseAllProjectionWindows 调用")

	count := 0
	for windowID, window := range a.projectionWindows {
		window.Close()
		delete(a.projectionWindows, windowID)
		delete(a.windowAlwaysOnTop, windowID)
		count++
	}

	projectionLock.Lock()
	for windowID, proc := range runningProjections {
		if proc != nil {
			proc.Kill()
		}
		delete(runningProjections, windowID)
		count++
	}
	projectionLock.Unlock()

	log.Printf("[IPC] 已关闭 %d 个投屏窗口", count)

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("已关闭 %d 个投屏窗口", count),
		"count":   count,
	}
}

// ProjectionWindowInfo 投屏窗口信息
type ProjectionWindowInfo struct {
	WindowID    string
	WindowTitle string
	DeviceIP    string
	TCPPort     int
	UDPPort     int
	ContainerID string
	Width       int
	Height      int
}

// GetProjectionWindows 获取所有投屏窗口信息
func (a *App) GetProjectionWindows() []map[string]interface{} {
	log.Printf("[IPC] 收到 GetProjectionWindows 调用")

	windows := make([]map[string]interface{}, 0, len(a.projectionWindows))
	for windowID, window := range a.projectionWindows {
		windows = append(windows, map[string]interface{}{
			"windowID": windowID,
			"name":     window.Name(),
			"width":    window.Width(),
			"height":   window.Height(),
			"focused":  window.IsFocused(),
		})
	}

	return windows
}

// CleanupProjectionWindows 清理所有投屏窗口 (应用退出时调用)
func (a *App) CleanupProjectionWindows() {
	log.Printf("[IPC] 收到 CleanupProjectionWindows 调用，清理所有投屏窗口")

	// 先隐藏所有窗口，防止用户在清理过程中操作
	for _, window := range a.projectionWindows {
		if window != nil && window.IsVisible() {
			window.Hide()
		}
	}

	// 关闭所有窗口
	for windowID, window := range a.projectionWindows {
		if window != nil {
			log.Printf("[IPC] 关闭投屏窗口: %s", windowID)
			window.Close()
		}
	}

	a.projectionWindows = make(map[string]*application.WebviewWindow)
	a.windowAlwaysOnTop = make(map[string]bool)
	log.Printf("[IPC] 所有投屏窗口已清理完毕")
}

// ToggleProjectionWindowTop 切换投屏窗口的置顶状态
func (a *App) ToggleProjectionWindowTop(windowID string) map[string]interface{} {
	log.Printf("[IPC] 收到 ToggleProjectionWindowTop 调用, windowID=%s", windowID)

	window, exists := a.projectionWindows[windowID]
	if !exists || window == nil {
		return map[string]interface{}{
			"success":  false,
			"message":  "投屏窗口不存在",
			"windowID": windowID,
		}
	}

	// 获取当前置顶状态（默认为false）
	currentTop := a.windowAlwaysOnTop[windowID]
	// 切换状态
	newTop := !currentTop

	// 设置新的置顶状态
	window.SetAlwaysOnTop(newTop)
	a.windowAlwaysOnTop[windowID] = newTop

	if newTop {
		log.Printf("[IPC] 投屏窗口已置顶: %s", windowID)
	} else {
		log.Printf("[IPC] 投屏窗口已取消置顶: %s", windowID)
	}

	return map[string]interface{}{
		"success":     true,
		"message":     "投屏窗口置顶状态已切换",
		"windowID":    windowID,
		"alwaysOnTop": newTop,
	}
}

// ArrangeProjectionWindows 智能排列所有投屏窗口
func (a *App) ArrangeProjectionWindows(params map[string]interface{}) map[string]interface{} {
	log.Printf("[IPC] 收到 ArrangeProjectionWindows 调用")

	windows := make([]string, 0, len(a.projectionWindows))
	for windowID := range a.projectionWindows {
		if a.projectionWindows[windowID] != nil && a.projectionWindows[windowID].IsVisible() {
			windows = append(windows, windowID)
		}
	}

	if len(windows) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "没有投屏窗口需要排列",
		}
	}

	log.Printf("[IPC] 排列 %d 个投屏窗口", len(windows))

	monitorWidth := 1920
	monitorHeight := 1080
	if width, ok := params["screenWidth"].(float64); ok {
		monitorWidth = int(width)
	}
	if height, ok := params["screenHeight"].(float64); ok {
		monitorHeight = int(height)
	}
	log.Printf("[IPC] 显示器尺寸: %dx%d", monitorWidth, monitorHeight)

	margin := 2
	aspectRatio := 9.0 / 16.0

	cols := int(math.Ceil(math.Sqrt(float64(len(windows)))))
	rows := (len(windows) + cols - 1) / cols

	rowHeight := monitorHeight/rows - 2*margin
	windowHeight := rowHeight
	windowWidth := int(float64(windowHeight) * aspectRatio)

	maxWidth := monitorWidth/cols - 2*margin
	if windowWidth > maxWidth {
		windowWidth = maxWidth
		windowHeight = int(float64(windowWidth) / aspectRatio)
	}

	log.Printf("[IPC] 排列布局: %d 列 x %d 行, 窗口尺寸: %dx%d (9:16)", cols, rows, windowWidth, windowHeight)

	for i, windowID := range windows {
		window := a.projectionWindows[windowID]
		if window == nil || !window.IsVisible() {
			continue
		}

		col := i % cols
		row := i / cols

		x := margin + col*windowWidth
		y := margin + row*windowHeight

		window.SetSize(windowWidth, windowHeight)
		window.SetPosition(x, y)
		window.Focus()

		log.Printf("[IPC] 窗口 %s -> 位置(%d,%d) 尺寸(%dx%d)", windowID, x, y, windowWidth, windowHeight)
	}

	return map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("已排列 %d 个投屏窗口", len(windows)),
		"windowCount": len(windows),
		"layout":      fmt.Sprintf("%dx%d", cols, rows),
		"windowSize":  fmt.Sprintf("%dx%d", windowWidth, windowHeight),
	}
}

// StartProjection 启动投屏 (旧版，使用外部进程)
func (a *App) StartProjection(deviceIP string, containerInfo map[string]interface{}) map[string]interface{} {
	log.Printf("[投屏] 收到 StartProjection 调用(旧版)")
	log.Printf("[投屏] 参数: deviceIP=%s, containerInfo=%+v", deviceIP, containerInfo)

	// 检查deviceIP是否为空
	if deviceIP == "" {
		return map[string]interface{}{
			"success": false,
			"message": "设备IP不能为空",
		}
	}

	// 获取容器ID
	containerID, ok := containerInfo["ID"].(string)
	if !ok {
		// V3设备使用name作为容器ID
		containerID, _ = containerInfo["name"].(string)
	}

	if containerID == "" {
		// 尝试使用容器名称作为ID
		if names, ok := containerInfo["Names"].([]interface{}); ok && len(names) > 0 {
			if nameStr, ok := names[0].(string); ok {
				containerID = nameStr
			}
		}
	}

	if containerID == "" {
		return map[string]interface{}{
			"success": false,
			"message": "容器ID不能为空",
		}
	}

	// 获取容器名称
	containerName := containerID
	if names, ok := containerInfo["Names"].([]interface{}); ok && len(names) > 0 {
		containerName = names[0].(string)
	} else if name, ok := containerInfo["name"].(string); ok {
		containerName = name
	}

	// 使用 ip_容器名 作为唯一标识
	windowID := deviceIP + "_" + containerName
	log.Printf("[投屏] windowID=%s", windowID)

	// 检查是否已有投屏进程
	projectionLock.Lock()
	if proc, exists := runningProjections[windowID]; exists {
		// 检查进程是否真的在运行
		if isProcessRunning(proc) {
			projectionLock.Unlock()
			log.Printf("[投屏] 投屏已在运行(PID=%d)，尝试置顶", proc.Pid)
			// 尝试将已运行的进程窗口置于前台
			if err := bringProcessToFront(proc.Pid); err != nil {
				log.Printf("[投屏] 将投屏进程置顶失败: %v", err)
			}
			return map[string]interface{}{
				"success": true,
				"message": "投屏已在运行，已将窗口置于前台",
				"pid":     proc.Pid,
			}
		}
		// 进程已结束则清理记录
		log.Printf("[投屏] 检测到进程已退出，清理记录: %s", windowID)
		delete(runningProjections, windowID)
	}
	projectionLock.Unlock()

	// 获取端口信息
	port, _ := getContainerPort(map[string]interface{}{"list": []interface{}{containerInfo}}, 9083)

	// 获取备用端口
	backport := 7100
	if indexNum, ok := containerInfo["indexNum"].(float64); ok {
		backport = 7100 + int(indexNum)
	}

	// 获取容器宽度和高度
	width := 720
	height := 1280
	if w, ok := containerInfo["width"].(float64); ok {
		width = int(w)
	}
	if h, ok := containerInfo["height"].(float64); ok {
		height = int(h)
	}

	// 获取uploadPort
	uploadPort, found := getContainerPort(map[string]interface{}{"list": []interface{}{containerInfo}}, 9082)
	uploadPortStr := strconv.Itoa(uploadPort)

	// 如果没有找到映射，使用云机容器IP:9082
	if !found {
		// 从containerInfo中获取容器IP
		containerIP := ""
		if ip, ok := containerInfo["ip"].(string); ok && ip != "" {
			containerIP = ip
		} else if ip, ok := containerInfo["IP"].(string); ok && ip != "" {
			containerIP = ip
		} else if ip, ok := containerInfo["Ip"].(string); ok && ip != "" {
			containerIP = ip
		}

		if containerIP != "" {
			uploadPortStr = fmt.Sprintf("%s:%d", containerIP, 9082)
		}
	}

	// 准备启动参数
	args := []string{
		"-ip", deviceIP,
		"-port", strconv.Itoa(port),
		"-name", containerName,
		"-uploadPort", uploadPortStr,
		"-containerID", containerID,
		"-width", strconv.Itoa(width),
		"-height", strconv.Itoa(height),
		"-bakport", strconv.Itoa(backport),
	}

	screenexe := "./screen"
	switch runtime.GOOS {
	case "windows":
		screenexe = "./screen.exe"
	case "darwin":
		// 对于MacOS，需要考虑不同架构和应用程序包结构
		exePath, err := os.Executable()
		if err == nil {
			// 获取应用程序目录路径
			macosDir := filepath.Dir(exePath)
			// 构建screen的绝对路径
			screenexe = filepath.Join(macosDir, "screen")
		}
	default:
		screenexe = "./screen"
	}

	// 检查screen可执行文件是否存在
	if _, err := os.Stat(screenexe); os.IsNotExist(err) {
		// 尝试从screen目录查找
		fallbackPath := "./screen/screen"
		if runtime.GOOS == "windows" {
			fallbackPath = "./screen/screen.exe"
		}
		if _, err := os.Stat(fallbackPath); err == nil {
			screenexe = fallbackPath
		}
	}

	log.Printf("[投屏] 启动投屏程序: %s %v", screenexe, args)

	// 启动进程
	cmd := exec.Command(screenexe, args...)
	if err := cmd.Start(); err != nil {
		log.Printf("[投屏] 启动投屏失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("启动投屏失败: %v", err),
		}
	}

	log.Printf("[投屏] 投屏进程启动成功，PID=%d, windowID=%s", cmd.Process.Pid, windowID)
	
	// 记录进程并启动监控
	projectionLock.Lock()
	runningProjections[windowID] = cmd.Process
	projectionLock.Unlock()

	// 启动监控协程
	go monitorProjectionProcess(windowID, cmd)

	return map[string]interface{}{
		"success": true,
		"message": "投屏已启动",
		"pid":     cmd.Process.Pid,
	}
}

// 进程监控协程
func monitorProjectionProcess(windowID string, cmd *exec.Cmd) {
	err := cmd.Wait()
	projectionLock.Lock()
	defer projectionLock.Unlock()

	// 清理进程记录
	if _, exists := runningProjections[windowID]; exists {
		delete(runningProjections, windowID)
		log.Printf("[投屏] 投屏进程退出，已清理记录: %s (PID=%d)", windowID, cmd.Process.Pid)
	}

	// 处理异常退出
	if err != nil {
		log.Printf("[投屏] 投屏进程异常退出: windowID=%s, error=%v", windowID, err)
	} else {
		log.Printf("[投屏] 投屏进程正常退出: windowID=%s", windowID)
	}
}

// 检查进程是否运行
func isProcessRunning(proc *os.Process) bool {
	if proc == nil {
		return false
	}

	// Windows和Unix使用不同的检测方式
	if runtime.GOOS == "windows" {
		return isWindowsProcessRunning(proc.Pid)
	}

	// Unix/Linux：发送信号0检查进程是否存在
	err := proc.Signal(syscall.Signal(0))
	return err == nil
}

// GetScreenshotProxy 代理获取截图并转换为base64
func (a *App) GetScreenshotProxy(url string) map[string]interface{} {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("HTTP错误: %d", resp.StatusCode),
		}
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取数据失败: %v", err),
		}
	}

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 尝试从数据判断类型
		if len(imageData) > 3 {
			if imageData[0] == 0xFF && imageData[1] == 0xD8 {
				contentType = "image/jpeg"
			} else if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47 {
				contentType = "image/png"
			} else {
				contentType = "image/jpeg" // 默认JPEG
			}
		} else {
			contentType = "image/jpeg"
		}
	}

	// 转换为base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)

	return map[string]interface{}{
		"success": true,
		"data":    dataURL,
	}
}

// ---------------- 流媒体相关方法 ----------------

// GetLocalIp 获取本机IP
func (a *App) GetLocalIp() string {
	return GetLocalIP()
}

// StartRtmpServer 启动流媒体服务
func (a *App) StartRtmpServer() {
	if a.RtmpService != nil {
		a.RtmpService.StartRtmpServer()
	}
}

// GetActiveStreams 获取活跃流 (包含保存的流和当前活跃的流)
func (a *App) GetActiveStreams() []StreamInfo {
	if a.RtmpService != nil {
		// 这里改为调用 GetSavedStreams，它已经包含了合并逻辑
		return a.RtmpService.GetSavedStreams()
	}
	return []StreamInfo{}
}

// AddStreamName 添加房间号
func (a *App) AddStreamName(name string) {
	if a.RtmpService != nil {
		a.RtmpService.AddStreamName(name)
	}
}

// DeleteStreamName 删除房间号
func (a *App) DeleteStreamName(name string) {
	if a.RtmpService != nil {
		a.RtmpService.DeleteStreamName(name)
	}
}

// PushStreamToDevices 推送流到设备
func (a *App) PushStreamToDevices(streamName string, deviceIPs []string) map[string]interface{} {
	if a.RtmpService != nil {
		return a.RtmpService.PushStreamToDevices(streamName, deviceIPs)
	}
	return map[string]interface{}{"success": false, "message": "服务未启动"}
}

// StopDevicePlay 停止设备拉流
func (a *App) StopDevicePlay(deviceIPs []string) map[string]interface{} {
	if a.RtmpService != nil {
		return a.RtmpService.StopDevicePlay(deviceIPs)
	}
	return map[string]interface{}{"success": false, "message": "服务未启动"}
}

// StopPushSession 停止推流
func (a *App) StopPushSession(streamName string) map[string]interface{} {
	if a.RtmpService != nil {
		return a.RtmpService.StopPushSession(streamName)
	}
	return map[string]interface{}{"success": false, "message": "服务未启动"}
}

// GetStreamPullers 获取流的拉流者
func (a *App) GetStreamPullers(streamName string) []PullerInfo {
	if a.RtmpService != nil {
		return a.RtmpService.GetStreamPullers(streamName)
	}
	return []PullerInfo{}
}

// ProxyHttpGet 代理 HTTP GET 请求，解决前端跨域问题
func (a *App) ProxyHttpGet(requestUrl string) (string, error) {
	resp, err := http.Get(requestUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// DeviceShake 摇一摇
func (a *App) DeviceShake(deviceIP string, port int, password string) error {
	log.Println("DeviceShake", deviceIP, port)
	url := fmt.Sprintf("http://%s:%d/modifydev?cmd=17&shake=true", deviceIP, port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if password != "" {
		req.SetBasicAuth("admin", password)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	return nil
}

// SetDeviceGPS 设置设备定位
func (a *App) SetDeviceGPS(host string, port int, deviceIP string, language string) error {
	log.Println("SetDeviceGPS", host, port, deviceIP, language)
	url := fmt.Sprintf("http://%s:%d/modifydev?cmd=11&ip=%s&launage=%s", host, port, deviceIP, language)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	return nil
}

// UploadLLMModel 上传LLM模型到设备
func (a *App) UploadLLMModel(deviceIP string, filePath string, token string) map[string]interface{} {
	log.Printf("[UploadLLMModel] 开始上传模型: deviceIP=%s, filePath=%s", deviceIP, filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[UploadLLMModel] 文件不存在: %s", filePath)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件不存在: %s", filePath),
		}
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[UploadLLMModel] 获取文件信息失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("获取文件信息失败: %v", err),
		}
	}
	fileSize := fileInfo.Size()
	log.Printf("[UploadLLMModel] 文件大小: %d bytes (%.2f MB)", fileSize, float64(fileSize)/1024/1024)

	// 使用 io.Pipe 流式上传，避免将整个文件读入内存
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// 在后台 goroutine 中流式写入文件内容，写完后关闭 pw
	go func() {
		var gErr error
		defer func() {
			writer.Close()       // 写完 multipart 结尾边界
			pw.CloseWithError(gErr) // 通知 HTTP 发送端：写入结束（或出错）
		}()

		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			gErr = fmt.Errorf("创建multipart字段失败: %v", err)
			return
		}
		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			gErr = fmt.Errorf("打开文件失败: %v", err)
			return
		}
		defer file.Close()
		// 流式拷贝，内存占用仅为 io.Copy 的默认 buffer（32KB）
		if _, err = io.Copy(part, file); err != nil {
			gErr = fmt.Errorf("拷贝文件内容失败: %v", err)
		}
	}()

	// 构建URL
	uploadURL := fmt.Sprintf("http://%s/lm/import", deviceAddr(deviceIP))
	log.Printf("[UploadLLMModel] 上传URL: %s", uploadURL)

	// 创建请求（body 为流式 pipe，无需等待文件全部读取）
	req, err := http.NewRequest("POST", uploadURL, pr)
	if err != nil {
		pr.CloseWithError(err)
		log.Printf("[UploadLLMModel] 创建上传请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建上传请求失败: %v", err),
		}
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", token)
		log.Printf("[UploadLLMModel] 已添加认证头")
	}

	// 发送请求（不设置超时，适合大文件）
	client := &http.Client{
		Timeout: 0, // 不设置超时
	}
	
	log.Printf("[UploadLLMModel] 开始发送请求...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[UploadLLMModel] 发送请求失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("发送请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[UploadLLMModel] 读取响应失败: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	log.Printf("[UploadLLMModel] 响应状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 如果无法解析为JSON，根据状态码判断成功与否
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return map[string]interface{}{
				"success": true,
				"message": "模型上传成功",
				"code":    0,
			}
		}
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("服务器返回错误 (状态码: %d): %s", resp.StatusCode, string(respBody)),
			"code":    resp.StatusCode,
		}
	}

	// 检查返回的code字段
	if code, ok := result["code"]; ok {
		if codeInt, ok := code.(float64); ok && codeInt == 0 {
			result["success"] = true
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result["success"] = true
	}

	return result
}


// Logger 简单的日志记录器
type Logger struct {
	prefix string
}

func (l *Logger) Info(format string, args ...interface{}) {
	log.Printf("[%s] "+format, append([]interface{}{l.prefix}, args...)...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	log.Printf("[%s] WARN: "+format, append([]interface{}{l.prefix}, args...)...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	log.Printf("[%s] ERROR: "+format, append([]interface{}{l.prefix}, args...)...)
}

// ======================== 批量任务服务绑定方法 ========================

// ExecuteBatchCommand 立即批量执行命令（前端绑定）
func (a *App) ExecuteBatchCommand(targets []Target, command string, taskName string) map[string]interface{} {
	log.Printf("[IPC] 收到 ExecuteBatchCommand 调用: taskName=%s, targets数量=%d", taskName, len(targets))
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	history, err := a.BatchTaskService.ExecuteBatchCommand(targets, command, taskName)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"history": history,
	}
}

// CreateScheduledTask 创建定时任务（前端绑定）
func (a *App) CreateScheduledTask(task BatchTask) map[string]interface{} {
	log.Printf("[IPC] 收到 CreateScheduledTask 调用: taskName=%s", task.Name)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	err := a.BatchTaskService.CreateScheduledTask(&task)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"task":    task,
	}
}

// UpdateScheduledTask 更新定时任务（前端绑定）
func (a *App) UpdateScheduledTask(task BatchTask) map[string]interface{} {
	log.Printf("[IPC] 收到 UpdateScheduledTask 调用: taskID=%s", task.ID)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	err := a.BatchTaskService.UpdateScheduledTask(&task)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"task":    task,
	}
}

// DeleteScheduledTask 删除定时任务（前端绑定）
func (a *App) DeleteScheduledTask(taskID string) map[string]interface{} {
	log.Printf("[IPC] 收到 DeleteScheduledTask 调用: taskID=%s", taskID)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	err := a.BatchTaskService.DeleteScheduledTask(taskID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
	}
}

// GetScheduledTasks 获取所有定时任务（前端绑定）
func (a *App) GetScheduledTasks() map[string]interface{} {
	log.Printf("[IPC] 收到 GetScheduledTasks 调用")
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
			"tasks":   []interface{}{},
		}
	}
	
	tasks := a.BatchTaskService.GetScheduledTasks()
	
	return map[string]interface{}{
		"success": true,
		"tasks":   tasks,
	}
}

// GetScheduledTask 获取单个定时任务（前端绑定）
func (a *App) GetScheduledTask(taskID string) map[string]interface{} {
	log.Printf("[IPC] 收到 GetScheduledTask 调用: taskID=%s", taskID)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	task, err := a.BatchTaskService.GetScheduledTask(taskID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"task":    task,
	}
}

// SaveCommandTemplate 保存命令模板（前端绑定）
func (a *App) SaveCommandTemplate(template CommandTemplate) map[string]interface{} {
	log.Printf("[IPC] 收到 SaveCommandTemplate 调用: templateName=%s", template.Name)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	err := a.BatchTaskService.SaveCommandTemplate(&template)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success":  true,
		"template": template,
	}
}

// GetCommandTemplates 获取所有命令模板（前端绑定）
func (a *App) GetCommandTemplates() map[string]interface{} {
	log.Printf("[IPC] 收到 GetCommandTemplates 调用")
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success":   false,
			"message":   "批量任务服务未初始化",
			"templates": []interface{}{},
		}
	}
	
	templates := a.BatchTaskService.GetCommandTemplates()
	
	return map[string]interface{}{
		"success":   true,
		"templates": templates,
	}
}

// DeleteCommandTemplate 删除命令模板（前端绑定）
func (a *App) DeleteCommandTemplate(templateID string) map[string]interface{} {
	log.Printf("[IPC] 收到 DeleteCommandTemplate 调用: templateID=%s", templateID)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	err := a.BatchTaskService.DeleteCommandTemplate(templateID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success": true,
	}
}

// GetExecutionHistory 获取执行历史（前端绑定）
func (a *App) GetExecutionHistory(limit int, offset int) map[string]interface{} {
	log.Printf("[IPC] 收到 GetExecutionHistory 调用: limit=%d, offset=%d", limit, offset)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
			"history": []interface{}{},
			"total":   0,
		}
	}
	
	history, total, err := a.BatchTaskService.GetExecutionHistory(limit, offset)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"history": []interface{}{},
			"total":   0,
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"history": history,
		"total":   total,
	}
}

// ExportExecutionHistory 导出执行历史（前端绑定）
func (a *App) ExportExecutionHistory(format string) map[string]interface{} {
	log.Printf("[IPC] 收到 ExportExecutionHistory 调用: format=%s", format)
	
	if a.BatchTaskService == nil {
		return map[string]interface{}{
			"success": false,
			"message": "批量任务服务未初始化",
		}
	}
	
	filePath, err := a.BatchTaskService.ExportExecutionHistory(format)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"success":  true,
		"filePath": filePath,
	}
}

// ==================== 批量导入云机备份功能 ====================

// ListBackupFiles 列出所有备份文件
func (a *App) ListBackupFiles() map[string]interface{} {
	files, err := a.BatchTaskService.ListBackupFiles()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"files":   files,
	}
}

// DeleteBackupFile 删除备份文件
func (a *App) DeleteBackupFile(fileName string) map[string]interface{} {
	err := a.BatchTaskService.DeleteBackupFile(fileName)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "删除成功",
	}
}

// StartBatchImport 开始批量导入
func (a *App) StartBatchImport(backupFileName string, devicesConfig []DeviceSlotConfig) map[string]interface{} {
	task, err := a.BatchTaskService.StartBatchImport(backupFileName, devicesConfig)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	
	// 保存任务引用以便前端轮询
	a.currentBatchImportTask = task

	return map[string]interface{}{
		"success": true,
		"task":    task,
		"task_id": task.ID,
	}
}

// GetBatchImportProgress 获取批量导入进度（供前端轮询）
func (a *App) GetBatchImportProgress() map[string]interface{} {
	if a.currentBatchImportTask == nil {
		return map[string]interface{}{
			"success": false,
			"message": "没有正在进行的任务",
		}
	}
	
	task := a.currentBatchImportTask
	
	// 序列化 details 数组
	detailsArray := make([]map[string]interface{}, len(task.Progress.Details))
	for i, d := range task.Progress.Details {
		detailsArray[i] = map[string]interface{}{
			"device_ip":    d.DeviceIP,
			"slot_number":  d.SlotNumber,
			"machine_name": d.MachineName,
			"success":      d.Success,
			"message":      d.Message,
			"duration":     d.Duration,
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"task_id": task.ID,
		"status":  task.Status,
		"progress": map[string]interface{}{
			"total_tasks":     task.Progress.TotalTasks,
			"completed_tasks": task.Progress.CompletedTasks,
			"failed_tasks":    task.Progress.FailedTasks,
			"current_device":  task.Progress.CurrentDevice,
			"current_slot":    task.Progress.CurrentSlot,
			"details":         detailsArray,
		},
	}
}

// UploadGoogleCert 上传 Google 证书（keybox）到云机
// fileData: 文件内容的 base64 编码字符串
// fileName: 原始文件名（含扩展名 .pem / .xml）
// host: 目标设备 IP
// port: 目标端口（通常为 9082 或映射端口）
func (a *App) UploadGoogleCert(host string, port int, fileName string, fileData string) map[string]interface{} {
	log.Printf("[IPC] UploadGoogleCert host=%s port=%d file=%s", host, port, fileName)

	// base64 解码文件内容
	decoded, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("文件解码失败: %v", err),
		}
	}

	// 构建 multipart/form-data 请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建表单失败: %v", err),
		}
	}
	if _, err = io.Copy(part, bytes.NewReader(decoded)); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("写入文件内容失败: %v", err),
		}
	}
	writer.Close()

	// 发送请求
	targetURL := fmt.Sprintf("http://%s:%d/uploadkeybox", host, port)
	req, err := http.NewRequest("POST", targetURL, body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("创建请求失败: %v", err),
		}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[UploadGoogleCert] status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("服务器返回 %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	// 尝试解析 JSON 响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"success": true,
			"body":    string(respBody),
		}
	}

	// 判断业务层是否成功（code==0 视为成功）
	if code, ok := result["code"]; ok {
		codeVal := fmt.Sprintf("%v", code)
		if codeVal != "0" {
			msg, _ := result["message"].(string)
			if msg == "" {
				msg, _ = result["msg"].(string)
			}
			return map[string]interface{}{
				"success": false,
				"message": msg,
				"body":    result,
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"body":    result,
	}
}

// min 辅助函数（Go 1.21 之前没有内置 min）
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
