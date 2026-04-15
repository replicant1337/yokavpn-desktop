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

	"github.com/gen2brain/beeep"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"yokovpn/internal/core"
	"yokovpn/internal/domain"
	"yokovpn/internal/ping"
	"yokovpn/internal/utils"
	"yokovpn/internal/xray"
)

type VPNOrchestrator struct {
	ctx      context.Context
	mu       sync.Mutex
	state    domain.VPNState
	
	client   *xray.Client
	coreMgr  *core.Manager
	
	tunProcess *exec.Cmd
	
	useTun        bool
	systemProxy   bool
	socksPort     int
	httpPort      int
	apiPort       int

	lastAppDataDir string
	reconnectCount int
}

func NewVPNOrchestrator(client *xray.Client, coreMgr *core.Manager) *VPNOrchestrator {
	return &VPNOrchestrator{
		client:  client,
		coreMgr: coreMgr,
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

	// Start Xray
	link, err := o.client.GetActiveConfig()
	if err != nil { o.rollback(err.Error()); return "", err }

	cfg, _ := ping.BuildXrayConfig(link, o.socksPort, o.httpPort, o.apiPort, "", "")
	cfgJSON, _ := json.Marshal(cfg)
	configPath := filepath.Join(appDataDir, "xray_config.json")
	os.WriteFile(configPath, cfgJSON, 0644)

	o.client.SetPaths(o.coreMgr.GetCorePath(), o.coreMgr.GetAssetDir())
	o.client.SetConfig(string(cfgJSON), configPath)

	if err := o.client.Start(useTun); err != nil { o.rollback(err.Error()); return "", err }

	// Start TUN or Proxy
	if useTun {
		tun2Path := o.coreMgr.GetTunPath()
		o.tunProcess = exec.Command(tun2Path, "--tun", "gatewaytun", "--proxy", fmt.Sprintf("socks5://127.0.0.1:%d", o.socksPort), "--setup", "--dns", "virtual")
		if runtime.GOOS == "windows" { o.tunProcess.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		o.tunProcess.Dir = o.coreMgr.GetTunDir()
		o.tunProcess.Start()
	} else if systemProxy {
		utils.EnableSystemProxy(fmt.Sprintf("127.0.0.1:%d", o.httpPort))
	}

	o.setState(domain.StateConnected)
	beeep.Notify("YokoVPN", "Connected", "")
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
	beeep.Alert("YokoVPN", "Error: "+reason, "")
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
