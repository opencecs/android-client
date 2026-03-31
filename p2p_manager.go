package main

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed camdemo/windows_pusher/Release_native/*
var windowsPusherAssets embed.FS

const windowsPusherAssetDir = "camdemo/windows_pusher/Release_native"

type P2PManager struct {
	mu       sync.Mutex
	sessions map[string]*P2PSession
	exePath  string
	exeOnce  sync.Once
	exeErr   error
}

type P2PSession struct {
	DeviceIP   string
	Port       int
	StreamName string
	StreamURL  string
	ListenPort int
	Cmd        *exec.Cmd
	Output     *strings.Builder
	Running    bool
	StartedAt  time.Time
	StoppedAt  time.Time
	LastError  string
}

func NewP2PManager() *P2PManager {
	return &P2PManager{
		sessions: make(map[string]*P2PSession),
	}
}

func (m *P2PManager) Start(deviceIP string, port int, streamName string, streamURL string, listenPort int) map[string]interface{} {
	if deviceIP == "" || port <= 0 || streamName == "" || streamURL == "" || listenPort <= 0 {
		return map[string]interface{}{"success": false, "message": "invalid parameters"}
	}

	exePath, err := m.ensureExe()
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	key := m.key(deviceIP, port, streamName)
	m.mu.Lock()
	if existing, ok := m.sessions[key]; ok && existing.Running {
		m.mu.Unlock()
		return map[string]interface{}{
			"success": true,
			"running": true,
			"message": "already running",
		}
	}
	m.mu.Unlock()

	cmd := exec.Command(exePath, deviceIP, strconv.Itoa(port), strconv.Itoa(listenPort))
	configureHiddenProcess(cmd)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	session := &P2PSession{
		DeviceIP:   deviceIP,
		Port:       port,
		StreamName: streamName,
		StreamURL:  streamURL,
		ListenPort: listenPort,
		Cmd:        cmd,
		Output:     &output,
		Running:    true,
		StartedAt:  time.Now(),
	}

	m.mu.Lock()
	m.sessions[key] = session
	m.mu.Unlock()

	go m.waitForExit(key, cmd)

	return map[string]interface{}{
		"success": true,
		"running": true,
		"message": "started",
	}
}

func (m *P2PManager) Stop(deviceIP string, port int, streamName string) map[string]interface{} {
	key := m.key(deviceIP, port, streamName)
	m.mu.Lock()
	session, ok := m.sessions[key]
	m.mu.Unlock()
	if !ok || session == nil {
		return map[string]interface{}{"success": false, "message": "session not found"}
	}
	if session.Cmd != nil && session.Cmd.Process != nil {
		_ = session.Cmd.Process.Kill()
	}

	m.mu.Lock()
	session.Running = false
	session.StoppedAt = time.Now()
	m.mu.Unlock()

	return map[string]interface{}{"success": true, "message": "stopped"}
}

func (m *P2PManager) Status(deviceIP string, port int, streamName string) map[string]interface{} {
	key := m.key(deviceIP, port, streamName)
	m.mu.Lock()
	session, ok := m.sessions[key]
	m.mu.Unlock()
	if !ok || session == nil {
		return map[string]interface{}{"running": false, "message": "not running"}
	}
	if session.Running {
		return map[string]interface{}{"running": true, "message": "running"}
	}
	msg := session.LastError
	if msg == "" {
		msg = "stopped"
	}
	return map[string]interface{}{"running": false, "message": msg}
}

func (m *P2PManager) StopAll() {
	m.mu.Lock()
	keys := make([]string, 0, len(m.sessions))
	for k := range m.sessions {
		keys = append(keys, k)
	}
	m.mu.Unlock()

	for _, key := range keys {
		parts := splitKey(key)
		if len(parts) != 3 {
			continue
		}
		port, _ := strconv.Atoi(parts[1])
		_ = m.Stop(parts[0], port, parts[2])
	}
}

func (m *P2PManager) waitForExit(key string, cmd *exec.Cmd) {
	err := cmd.Wait()
	m.mu.Lock()
	if session, ok := m.sessions[key]; ok && session != nil {
		session.Running = false
		session.StoppedAt = time.Now()
		if err != nil {
			session.LastError = err.Error()
		}
		if session.Output != nil && session.LastError == "" {
			out := strings.TrimSpace(session.Output.String())
			if out != "" {
				session.LastError = out
			}
		}
	}
	m.mu.Unlock()
}

func (m *P2PManager) ensureExe() (string, error) {
	m.exeOnce.Do(func() {
		if runtime.GOOS != "windows" {
			m.exeErr = fmt.Errorf("p2p WindowsPusher is only available on Windows")
			return
		}
		dir := filepath.Join(os.TempDir(), "edgeclient-p2p")
		_ = os.MkdirAll(dir, 0755)
		if err := extractWindowsPusherAssets(dir); err != nil {
			m.exeErr = err
			return
		}
		m.exePath = filepath.Join(dir, "WindowsPusher.exe")
	})
	return m.exePath, m.exeErr
}

func (m *P2PManager) key(deviceIP string, port int, streamName string) string {
	return fmt.Sprintf("%s:%d:%s", deviceIP, port, streamName)
}

func splitKey(key string) []string {
	return strings.SplitN(key, ":", 3)
}

// extractWindowsPusherAssets extracts all files from windowsPusherAssets
// into destDir, skipping files whose sha256 already matches.
func extractWindowsPusherAssets(destDir string) error {
	return fs.WalkDir(windowsPusherAssets, windowsPusherAssetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(windowsPusherAssetDir, filepath.FromSlash(path))
		if rel == ".gitkeep" {
			return nil
		}
		data, err := windowsPusherAssets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embed %s: %w", path, err)
		}
		if len(data) == 0 {
			return fmt.Errorf("embedded file %s is empty", path)
		}
		want := sha256.Sum256(data)
		dest := filepath.Join(destDir, rel)
		if existing, err := os.ReadFile(dest); err == nil {
			if sha256.Sum256(existing) == want {
				return nil
			}
		}
		return os.WriteFile(dest, data, 0755)
	})
}
