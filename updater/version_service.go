package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

type VersionService struct {
	config     *UpdateConfig
	httpClient *http.Client
}

func NewVersionService(config *UpdateConfig) *VersionService {
	return &VersionService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *VersionService) CheckForUpdate(currentVersion, channel string) (*ReleaseAsset, error) {
	log.Printf("[VersionService] 开始检查更新, 当前版本: %s, 渠道: %s", currentVersion, channel)

	s.config.Channel = channel

	checkURL, err := s.buildCheckURL(currentVersion, channel)
	if err != nil {
		return nil, fmt.Errorf("构建检查URL失败: %w", err)
	}

	log.Printf("[VersionService] 发送版本检查请求: %s", checkURL)

	resp, err := s.httpClient.Get(checkURL)
	if err != nil {
		return nil, fmt.Errorf("请求版本检查接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("版本检查接口返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var checkResp ClientUpdateResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return nil, fmt.Errorf("解析响应JSON失败: %w", err)
	}

	if checkResp.CodeID != 200 {
		return nil, fmt.Errorf("版本检查接口返回错误: %s", checkResp.Msg)
	}

	if checkResp.Data == nil {
		log.Printf("[VersionService] 当前已是最新版本，无可用更新")
		return nil, nil
	}

	releaseAsset := ReleaseAsset{
		Version:      checkResp.Data.Version,
		DownloadURL:  checkResp.Data.DownloadURL,
		Checksum:     checkResp.Data.Checksum,
		FileSize:     checkResp.Data.FileSize,
		ReleaseNotes: checkResp.Data.ReleaseNotes,
		Mandatory:    checkResp.Data.Mandatory,
		Platform:     checkResp.Data.Platform,
		Architecture: checkResp.Data.Arch,
	}

	log.Printf("[VersionService] 获取到更新信息: 版本=%s, 下载URL=%s",
		releaseAsset.Version, releaseAsset.DownloadURL)

	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		log.Printf("[VersionService] 无法解析当前版本号: %v", err)
		return &releaseAsset, nil
	}

	latestVer, err := semver.NewVersion(releaseAsset.Version)
	if err != nil {
		log.Printf("[VersionService] 无法解析远程版本号: %v", err)
		return &releaseAsset, nil
	}

	if latestVer.GreaterThan(currentVer) {
		log.Printf("[VersionService] 发现新版本: %s > %s", latestVer.String(), currentVer.String())
		return &releaseAsset, nil
	}

	log.Printf("[VersionService] 当前已是最新版本: %s", currentVersion)
	return nil, nil
}

func (s *VersionService) buildCheckURL(currentVersion, channel string) (string, error) {
	baseURL := s.config.CheckURL

	platform := GetPlatform()
	arch := GetArchitecture()
	platformArch := platform + "-" + arch

	params := url.Values{}
	params.Add("version", currentVersion)
	params.Add("platform", platformArch)
	params.Add("channel", channel)

	if strings.Contains(baseURL, "?") {
		return baseURL + "&" + params.Encode(), nil
	}
	return baseURL + "?" + params.Encode(), nil
}

func (s *VersionService) GetVersionInfo() (*VersionInfo, error) {
	return &VersionInfo{
		Version:   AppVersion,
		BuildTime: BuildTime,
	}, nil
}

func (s *VersionService) CompareVersions(v1, v2 string) (int, error) {
	semver1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("无法解析版本号 %s: %w", v1, err)
	}

	semver2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("无法解析版本号 %s: %w", v2, err)
	}

	if semver1.Equal(semver2) {
		return 0, nil
	} else if semver1.GreaterThan(semver2) {
		return 1, nil
	} else {
		return -1, nil
	}
}

func (s *VersionService) ParseReleaseNotes(notes string) []string {
	if notes == "" {
		return []string{}
	}

	lines := strings.Split(notes, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (s *VersionService) FetchWithProxy(proxyURL string) (*HTTPResponse, error) {
	var client *http.Client

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("解析代理URL失败: %w", err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		}
	} else {
		client = s.httpClient
	}

	_ = client

	return nil, fmt.Errorf("此方法已废弃，请使用 CheckForUpdate")
}

var (
	AppVersion string
	BuildTime  string
)

func init() {
	if AppVersion == "" {
		AppVersion = "1.0.0"
	}
	if BuildTime == "" {
		BuildTime = time.Now().Format(time.RFC3339)
	}
}
