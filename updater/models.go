package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	UpdateStateIdle        = "idle"
	UpdateStateChecking    = "checking"
	UpdateStateDownloading = "downloading"
	UpdateStateVerifying   = "verifying"
	UpdateStateInstalling  = "installing"
	UpdateStateComplete    = "complete"
	UpdateStateFailed      = "failed"
)

type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
}

type ReleaseAsset struct {
	Version      string `json:"version"`
	DownloadURL  string `json:"downloadUrl"`
	Checksum     string `json:"checksum"`
	FileSize     int64  `json:"fileSize"`
	ReleaseNotes string `json:"releaseNotes"`
	Mandatory    bool   `json:"mandatory"`
	Channel      string `json:"channel"`
	PublishedAt  string `json:"publishedAt"`
	Platform     string `json:"platform"`
	Architecture string `json:"arch"`
}

type UpdateConfig struct {
	CheckURL           string    `json:"checkUrl"`
	Channel            string    `json:"channel"`
	AutoCheck          bool      `json:"autoCheck"`
	AutoUpdate         bool      `json:"autoUpdate"`
	CheckInterval      int       `json:"checkInterval"`
	LastCheckTime      time.Time `json:"lastCheckTime"`
	LastCheckedVersion string    `json:"lastCheckedVersion"`
	ProxyURL           string    `json:"proxyUrl"`
	TemporaryPath      string    `json:"-"`
	UpdateLogPath      string    `json:"-"`
	PublicKeyPath      string    `json:"-"`
}

type UpdateState struct {
	State            string  `json:"state"`
	CurrentVersion   string  `json:"currentVersion"`
	LatestVersion    string  `json:"latestVersion"`
	DownloadProgress float64 `json:"downloadProgress"`
	ErrorMessage     string  `json:"errorMessage"`
	UpdateLog        string  `json:"updateLog"`
}

type UpdateCheckResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    ReleaseAsset `json:"data"`
}

type ClientUpdateResponse struct {
	CodeID int               `json:"code_id"`
	Msg    string            `json:"msg"`
	Data   *ClientUpdateData `json:"data"`
}

type ClientUpdateData struct {
	Version      string `json:"version"`
	DownloadURL  string `json:"downloadUrl"`
	Checksum     string `json:"checksum"`
	FileSize     int64  `json:"fileSize"`
	ReleaseNotes string `json:"releaseNotes"`
	Mandatory    bool   `json:"mandatory"`
	Channel      string `json:"channel"`
	PublishedAt  string `json:"publishedAt"`
	Platform     string `json:"platform"`
	Arch         string `json:"arch"`
}

func DefaultUpdateConfig() *UpdateConfig {
	userConfigDir, _ := os.UserConfigDir()
	temporaryPath := filepath.Join(os.TempDir(), "edgeclient-update")
	updateLogPath := filepath.Join(userConfigDir, "EdgeClient", "update.log")

	return &UpdateConfig{
		CheckURL:           "https://newapi.moyunteng.com/api/v1/client/update",
		Channel:            "published",
		AutoCheck:          true,
		AutoUpdate:         false,
		CheckInterval:      3600,
		LastCheckTime:      time.Time{},
		LastCheckedVersion: "",
		ProxyURL:           "",
		TemporaryPath:      temporaryPath,
		UpdateLogPath:      updateLogPath,
		PublicKeyPath:      "",
	}
}

func (c *UpdateConfig) Save() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(userConfigDir, "EdgeClient")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "update_config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func LoadUpdateConfig() (*UpdateConfig, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultUpdateConfig(), err
	}

	configPath := filepath.Join(userConfigDir, "EdgeClient", "update_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultUpdateConfig()
			cfg.Save()
			return cfg, nil
		}
		return nil, err
	}

	var config UpdateConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.TemporaryPath == "" {
		config.TemporaryPath = filepath.Join(os.TempDir(), "edgeclient-update")
	}
	if config.UpdateLogPath == "" {
		config.UpdateLogPath = filepath.Join(userConfigDir, "EdgeClient", "update.log")
	}

	return &config, nil
}

func GetPlatform() string {
	return runtime.GOOS
}

func GetArchitecture() string {
	return runtime.GOARCH
}

func (r *ReleaseAsset) IsPlatformMatch() bool {
	platform := GetPlatform()
	arch := GetArchitecture()

	switch platform {
	case "windows":
		return (r.Platform == "windows" || r.Platform == "win") &&
			(r.Architecture == "amd64" || r.Architecture == "x64" || r.Architecture == arch)
	case "darwin":
		return r.Platform == "darwin" && (r.Architecture == "amd64" || r.Architecture == "x64" || r.Architecture == arch || r.Architecture == "arm64")
	case "linux":
		return r.Platform == "linux" &&
			(r.Architecture == "amd64" || r.Architecture == "x64" || r.Architecture == arch ||
				r.Architecture == "arm64" || r.Architecture == "arm")
	default:
		return false
	}
}
