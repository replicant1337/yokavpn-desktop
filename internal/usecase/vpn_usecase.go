package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"YokaVPN/internal/core"
	"YokaVPN/internal/domain"
	"YokaVPN/internal/ping"
	"YokaVPN/internal/utils"
	"YokaVPN/internal/xray"
)

type VPNOrchestrator struct {
	ctx      context.Context
	mu       sync.Mutex
	state    domain.VPNState
	
	client   *xray.Client
	coreMgr  *core.Manager
	alerts   *AlertManager
	
	tunProcess *exec.Cmd
	
	useTun        bool
	systemProxy   bool
	socksPort     int
	httpPort      int
	apiPort       int

	lastAppDataDir string
	reconnectCount int
}

func NewVPNOrchestrator(client *xray.Client, coreMgr *core.Manager, alerts *AlertManager) *VPNOrchestrator {
	return &VPNOrchestrator{
		client:  client,
		coreMgr: coreMgr,
		alerts:  alerts,
		state:   domain.StateIdle,
	}
}

func (o *VPNOrchestrator) SetContext(ctx context.Context) {
	o.ctx = ctx
}

func (o *VPNOrchestrator) setState(s domain.VPNState) {
	o.mu.Lock()
	o.state = s
	o.mu.Unlock()
	if o.ctx != nil {
		wailsruntime.EventsEmit(o.ctx, "vpn-state-changed", string(s))
	}
}

// loadClientConfig reads the client_config.json file from the application data directory.
func (o *VPNOrchestrator) loadClientConfig() (*domain.ClientConfig, error) {
	configPath := filepath.Join(o.lastAppDataDir, "client_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If the file doesn't exist, it's not an error, just means default settings will be used.
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read client config: %w", err)
	}

	var config domain.ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse client config: %w", err)
	}
	return &config, nil
}

func (o *VPNOrchestrator) Connect(useTun, systemProxy bool, appDataDir string) (string, error) {
	o.mu.Lock()
	if o.state != domain.StateIdle && o.state != domain.StateError {
		o.mu.Unlock()
		return "", fmt.Errorf("app is busy")
	}
	o.lastAppDataDir = appDataDir
	o.mu.Unlock()

	o.ForceCleanup()
	o.setState(domain.StateStarting)

	o.useTun = useTun
	o.systemProxy = systemProxy
	o.socksPort, o.httpPort, o.apiPort = 10808, 10809, 62789

	// Load client configuration
	clientConfig, err := o.loadClientConfig()
	if err != nil {
		o.rollback(fmt.Sprintf("Configuration error: %v", err))
		return "", err
	}

	// Start Xray
	link, err := o.client.GetActiveConfig()
	if err != nil { o.rollback(err.Error()); return "", err }

	cfg, err := ping.BuildXrayConfig(link, o.socksPort, o.httpPort, o.apiPort, "", "")
	if err != nil { o.rollback(fmt.Sprintf("Failed to build Xray config: %v", err)); return "", err}
	cfgJSON, _ := json.Marshal(cfg)
	configPath := filepath.Join(appDataDir, "xray_config.json")
	os.WriteFile(configPath, cfgJSON, 0644)

	o.client.SetPaths(o.coreMgr.GetCorePath(), o.coreMgr.GetAssetDir())
	o.client.SetConfig(string(cfgJSON), configPath)

	if err := o.client.Start(useTun); err != nil { o.rollback(err.Error()); return "", err }

	// Start TUN or Proxy
	if useTun {
		tun2Path := o.coreMgr.GetTunPath()
		tunArgs := []string{"--proxy", fmt.Sprintf("socks5://127.0.0.1:%d", o.socksPort), "--setup", "--dns", "virtual"}

		if runtime.GOOS == "windows" {
			defaultTunDevice := "gatewaytun" // Default to Wintun name
			if clientConfig != nil && clientConfig.AdapterType == "tap" {
				if clientConfig.AdapterName != "" {
					tunArgs = append(tunArgs, "--tun", clientConfig.AdapterName)
				} else {
					// Use a default TAP adapter name if none is provided
					tunArgs = append(tunArgs, "--tun", "tap0") 
				}
			} else {
				// Use default Wintun name if not TAP or config is missing/invalid
				tunArgs = append(tunArgs, "--tun", defaultTunDevice)
			}
		} else {
			// For non-Windows, use the default tun device name
			tunArgs = append(tunArgs, "--tun", "tun0")
		}

		o.tunProcess = exec.Command(tun2Path, tunArgs...)
		if runtime.GOOS == "windows" { o.tunProcess.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		o.tunProcess.Dir = o.coreMgr.GetTunDir()
		if err := o.tunProcess.Start(); err != nil {
			o.rollback(fmt.Sprintf("Failed to start tun2proxy: %v", err))
			return "", err
		}
	} else if systemProxy {
		utils.EnableSystemProxy(fmt.Sprintf("127.0.0.1:%d", o.httpPort))
	}

	o.setState(domain.StateConnected)
	if o.alerts != nil {
		o.alerts.Notify("YokaVPN", "Connected")
	}
	go o.monitor()
	return "Connected", nil
}

func (o *VPNOrchestrator) Disconnect() {
	o.setState(domain.StateDisconnecting)
	go func() {
		o.ForceCleanup()
		o.setState(domain.StateIdle)
		wailsruntime.EventsEmit(o.ctx, "connect-status", "app.disconnect")
	}()
}

func (o *VPNOrchestrator) ForceCleanup() {
	utils.DisableSystemProxy()
	o.client.Stop()
	
	if o.tunProcess != nil && o.tunProcess.Process != nil {
		if runtime.GOOS == "windows" {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", o.tunProcess.Process.Pid))
			killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			killCmd.Run()
		} else {
			o.tunProcess.Process.Kill()
		}
		o.tunProcess = nil
	}

	// Hard kill by name
	if runtime.GOOS == "windows" {
		killXray := exec.Command("taskkill", "/F", "/IM", "xray.exe", "/T")
		killXray.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		killXray.Run()
		
		killTun := exec.Command("taskkill", "/F", "/IM", "tun2proxy-bin.exe", "/T")
		killTun.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		killTun.Run()
	} else {
		exec.Command("killall", "-9", "xray").Run()
		exec.Command("killall", "-9", "tun2proxy-bin").Run()
	}
}

func (o *VPNOrchestrator) rollback(reason string) {
	o.ForceCleanup()
	o.setState(domain.StateError)
	if o.alerts != nil {
		o.alerts.Alert("YokaVPN", "Error: "+reason)
	}
}

func (o *VPNOrchestrator) monitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-o.ctx.Done(): return
		case <-ticker.C:
			o.mu.Lock()
			if o.state != domain.StateConnected { o.mu.Unlock(); return }
			isRunning := o.client.IsRunning()
			o.mu.Unlock()

			if !isRunning {
				o.rollback("Connection lost")
				return
			}
		}
	}
}

func (o *VPNOrchestrator) SwitchServer(newIndex int, useTun, systemProxy bool, appDataDir string) (string, error) {
	o.Disconnect()
	time.Sleep(1 * time.Second)
	o.client.SetActiveServer(newIndex)
	return o.Connect(useTun, systemProxy, appDataDir)
}

func (o *VPNOrchestrator) GetStatus() map[string]interface{} {
	o.mu.Lock()
	defer o.mu.Unlock()
	return map[string]interface{}{
		"state": string(o.state),
	}
}
