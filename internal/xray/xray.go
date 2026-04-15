package xray

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"yokovpn/internal/domain"
	"yokovpn/internal/log"
)

type Client struct {
	mu         sync.Mutex
	process    *exec.Cmd
	configPath string
	xrayPath   string
	assetDir   string
	running    bool
	useTun     bool
	logChan    chan string
	logFile    *os.File
	serverIP   string
	links      []string
	activeIdx  int
	logger     *log.Logger
	apiPort    int
}

func NewClient() *Client {
	logger, _ := log.NewLogger("xray", log.GetLogDir())
	if logger != nil { logger.SetConsole(false) }
	return &Client{
		logChan: make(chan string, 100),
		logger:  logger,
	}
}

func (c *Client) SetPaths(xrayPath string, assetDir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.xrayPath = xrayPath
	c.assetDir = assetDir
}

func (c *Client) SetLinks(links []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.links = links
}

func (c *Client) SetActiveServer(index int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.links) { return fmt.Errorf("invalid index") }
	c.activeIdx = index
	return nil
}

func (c *Client) GetActiveConfig() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.activeIdx < 0 || c.activeIdx >= len(c.links) { return "", fmt.Errorf("no server") }
	return c.links[c.activeIdx], nil
}

func (c *Client) SetConfig(configJSON string, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configPath = path
	c.serverIP = extractServerIP(configJSON)

	// Extract API port
	var cfg struct {
		Inbounds []struct {
			Tag  string `json:"tag"`
			Port string `json:"port"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err == nil {
		for _, inb := range cfg.Inbounds {
			if inb.Tag == "api" {
				fmt.Sscanf(inb.Port, "%d", &c.apiPort)
				break
			}
		}
	}
	if c.apiPort == 0 { c.apiPort = 62789 }

	return nil
}

func (c *Client) Start(useTun bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running { return nil }
	if c.xrayPath == "" || c.configPath == "" { return fmt.Errorf("not configured") }

	c.process = exec.Command(c.xrayPath, "run", "-c", c.configPath)
	c.process.Dir = c.assetDir
	c.process.Env = append(os.Environ(), "XRAY_LOCATION_ASSET="+c.assetDir)
	if runtime.GOOS == "windows" {
		c.process.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	
	stdout, _ := c.process.StdoutPipe()
	stderr, _ := c.process.StderrPipe()

	if err := c.process.Start(); err != nil { return err }

	c.running = true
	go func() {
		c.process.Wait()
		c.mu.Lock()
		c.running = false
		c.mu.Unlock()
	}()

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			if c.logger != nil { c.logger.Debug("[Xray] %s", scanner.Text()) }
		}
	}()

	return nil
}

func (c *Client) Stop() error {
	c.mu.Lock()
	proc := c.process
	c.mu.Unlock()

	if proc != nil && proc.Process != nil {
		if runtime.GOOS == "windows" {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", proc.Process.Pid))
			killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			killCmd.Run()
		} else {
			proc.Process.Kill()
		}
	}

	c.mu.Lock()
	c.running = false
	c.process = nil
	c.mu.Unlock()
	return nil
}

func (c *Client) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

func (c *Client) GetServerIP() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serverIP
}

func (c *Client) GetStats() *domain.Stats {
	stats := &domain.Stats{}
	c.mu.Lock()
	isRunning := c.running
	xrayPath := c.xrayPath
	apiPort := c.apiPort
	assetDir := c.assetDir
	c.mu.Unlock()

	if !isRunning || xrayPath == "" { return stats }

	cmd := exec.Command(xrayPath, "api", "statsquery", "--server", fmt.Sprintf("127.0.0.1:%d", apiPort), "-pattern", "")
	cmd.Dir = assetDir
	cmd.Env = append(os.Environ(), "XRAY_LOCATION_ASSET="+assetDir)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	
	output, err := cmd.Output()
	if err != nil { return stats }

	var result struct {
		Stat []struct {
			Name  string `json:"name"`
			Value int64  `json:"value"`
		} `json:"stat"`
	}

	if err := json.Unmarshal(output, &result); err == nil {
		for _, s := range result.Stat {
			if strings.Contains(s.Name, "uplink") {
				stats.UploadBytes += uint64(s.Value)
			} else if strings.Contains(s.Name, "downlink") {
				stats.DownloadBytes += uint64(s.Value)
			}
		}
	}

	return stats
}

func extractServerIP(configJSON string) string {
	var cfg struct {
		Outbounds []struct {
			Settings struct {
				Vnext []struct { Address string `json:"address"` } `json:"vnext"`
				Servers []struct { Address string `json:"address"` } `json:"servers"`
			} `json:"settings"`
		} `json:"outbounds"`
	}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err == nil {
		for _, o := range cfg.Outbounds {
			if len(o.Settings.Vnext) > 0 { return o.Settings.Vnext[0].Address }
			if len(o.Settings.Servers) > 0 { return o.Settings.Servers[0].Address }
		}
	}
	return ""
}
