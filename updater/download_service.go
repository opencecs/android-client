package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadProgress struct {
	DownloadedBytes int64
	TotalBytes      int64
	Progress        float64
}

type DownloadService struct {
	config     *UpdateConfig
	httpClient *http.Client
	progressCh chan<- DownloadProgress
}

func NewDownloadService(config *UpdateConfig) *DownloadService {
	return &DownloadService{
		config: config,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (s *DownloadService) SetProgressChannel(ch chan<- DownloadProgress) {
	s.progressCh = ch
}

func (s *DownloadService) DownloadFile(url, destination string) error {
	log.Printf("[DownloadService] 开始下载文件: %s -> %s", url, destination)

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("EdgeClient/%s (%s; %s)",
		AppVersion, GetPlatform(), GetArchitecture()))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败, 状态码: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	log.Printf("[DownloadService] 文件大小: %d bytes", contentLength)

	writer := &ProgressWriter{
		Writer:     file,
		Total:      contentLength,
		Progress:   s.progressCh,
		Downloaded: 0,
	}

	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	log.Printf("[DownloadService] 下载完成, 共写入: %d bytes", written)

	if s.progressCh != nil {
		s.progressCh <- DownloadProgress{
			DownloadedBytes: written,
			TotalBytes:      contentLength,
			Progress:        100.0,
		}
	}

	return nil
}

func (s *DownloadService) DownloadWithResume(url, destination string) error {
	tempPath := destination + ".tmp"

	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(tempPath)
			if err != nil {
				return fmt.Errorf("创建临时文件失败: %w", err)
			}
		} else {
			return fmt.Errorf("打开临时文件失败: %w", err)
		}
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	offset := fileInfo.Size()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		log.Printf("[DownloadService] 断点续传, 从 %d bytes 开始", offset)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		file.Truncate(0)
		file.Seek(0, 0)
		log.Printf("[DownloadService] 服务器不支持断点续传, 重新下载")
	case http.StatusPartialContent:
		log.Printf("[DownloadService] 支持断点续传, 从 %d bytes 继续", offset)
	default:
		return fmt.Errorf("下载失败, 状态码: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength + offset
	log.Printf("[DownloadService] 剩余下载大小: %d bytes, 总大小: %d bytes",
		resp.ContentLength, contentLength)

	writer := &ProgressWriter{
		Writer:     file,
		Total:      contentLength,
		Progress:   s.progressCh,
		Downloaded: offset,
	}

	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	log.Printf("[DownloadService] 断点续传完成, 本次写入: %d bytes", written)

	os.Rename(tempPath, destination)

	return nil
}

func (s *DownloadService) GetTempDir() string {
	if s.config.TemporaryPath != "" {
		os.MkdirAll(s.config.TemporaryPath, 0755)
		return s.config.TemporaryPath
	}
	tmpDir, _ := os.MkdirTemp("", "edgeclient-update")
	return tmpDir
}

func (s *DownloadService) CleanupTempDir() error {
	if s.config.TemporaryPath != "" {
		return os.RemoveAll(s.config.TemporaryPath)
	}
	return nil
}

type ProgressWriter struct {
	Writer     io.Writer
	Total      int64
	Progress   chan<- DownloadProgress
	Downloaded int64
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if n > 0 {
		pw.Downloaded += int64(n)
		if pw.Progress != nil && pw.Total > 0 {
			progress := float64(pw.Downloaded) / float64(pw.Total) * 100
			pw.Progress <- DownloadProgress{
				DownloadedBytes: pw.Downloaded,
				TotalBytes:      pw.Total,
				Progress:        progress,
			}
		}
	}
	return n, err
}

func (s *DownloadService) ValidateURL(url string) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("无效的URL格式: %s", url)
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("EdgeClient/%s", AppVersion))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("URL验证失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("URL不可访问, 状态码: %d", resp.StatusCode)
	}

	return nil
}

func (s *DownloadService) GetFileSize(url string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("获取文件大小失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("无法获取文件信息, 状态码: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	return contentLength, nil
}
