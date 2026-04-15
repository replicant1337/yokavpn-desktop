package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"YokaVPN/internal/core"
	"YokaVPN/internal/domain"
	"YokaVPN/internal/geo"
	"YokaVPN/internal/log"
	"YokaVPN/internal/ping"
	"YokaVPN/internal/subscription"
	"YokaVPN/internal/usecase"
	xray "YokaVPN/internal/xray"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	client  *xray.Client
	subMgr  *usecase.SubscriptionManager
	coreMgr *core.Manager
	vpnSvc  *usecase.VPNOrchestrator
	alerts  *usecase.AlertManager
	logger  *log.Logger
}

func NewApp() *App {
	configDir, _ := os.UserConfigDir()
	appDataDir := filepath.Join(configDir, "YokaVPN")
	logger, _ := log.NewLogger("app", filepath.Join(appDataDir, "logs"))
	client := xray.NewClient()
	coreMgr := core.NewManager(appDataDir)
	alerts := usecase.NewAlertManager(appDataDir)
	vpnSvc := usecase.NewVPNOrchestrator(client, coreMgr, alerts)

	return &App{
		client:  client,
		subMgr:  usecase.NewSubscriptionManager(),
		coreMgr: coreMgr,
		vpnSvc:  vpnSvc,
		alerts:  alerts,
		logger:  logger,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.vpnSvc.SetContext(ctx)
	os.MkdirAll(a.AppDataDir(), 0755)

	// Aggressive startup cleanup
	a.vpnSvc.ForceCleanup()

	a.coreMgr.SetProgressCallback(func(current, total int64, label string) {
		percentage := 0
		if total > 0 {
			percentage = int((float64(current) / float64(total)) * 100)
		}
		wailsruntime.EventsEmit(a.ctx, "core-install-progress", map[string]interface{}{
			"percentage": percentage, "label": label,
		})
	})

	go a.EnsureCore()

	// Load and auto-select best server
	go func() {
		time.Sleep(1 * time.Second)
		configPath := filepath.Join(a.AppDataDir(), "config.json")
		data, err := os.ReadFile(configPath)
		if err == nil {
			var links []string
			if err := json.Unmarshal(data, &links); err == nil && len(links) > 0 {
				a.client.SetLinks(links)
				a.AutoSelectBestServer(links)
			}
		}
	}()
}

func (a *App) AutoSelectBestServer(links []string) {
	if len(links) == 0 {
		return
	}

	wailsruntime.EventsEmit(a.ctx, "connect-status", "app.status_optimizing")

	type result struct {
		index   int
		latency int64
	}
	results := make(chan result, len(links))

	// Fast concurrent ping
	for i, link := range links {
		go func(idx int, l string) {
			res := ping.TestViaXray(l, idx, "ping")
			if res.Success {
				results <- result{idx, res.Latency}
			} else {
				results <- result{idx, 9999}
			}
		}(i, link)
	}

	all := make([]result, 0, len(links))
	for i := 0; i < len(links); i++ {
		all = append(all, <-results)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].latency < all[j].latency })

	bestIdx := all[0].index
	a.client.SetActiveServer(bestIdx)
	wailsruntime.EventsEmit(a.ctx, "best-server-selected", bestIdx)
	wailsruntime.EventsEmit(a.ctx, "connect-status", "")
}

func (a *App) EnsureCore() {
	const xrayVersion = "v24.12.31"
	const tunVersion = "v0.7.20"

	// Force show install view if anything is missing
	if !a.coreMgr.IsInstalled() || !a.coreMgr.IsTunInstalled() {
		wailsruntime.EventsEmit(a.ctx, "core-install-state", true)
		if !a.coreMgr.IsInstalled() {
			a.coreMgr.InstallXray(xrayVersion)
		}
		if !a.coreMgr.IsTunInstalled() {
			a.coreMgr.InstallTun2Proxy(tunVersion)
		}
	}

	wailsruntime.EventsEmit(a.ctx, "connect-status", "app.status_checking_assets")
	geo.UpdateGeoAssets(a.coreMgr.GetAssetDir())

	wailsruntime.EventsEmit(a.ctx, "core-install-state", false)
	wailsruntime.EventsEmit(a.ctx, "connect-status", "")
}

func (a *App) Connect(useTun bool, systemProxy bool) string {
	if a.coreMgr.IsInstalling() {
		return "Error: Initializing components..."
	}
	msg, err := a.vpnSvc.Connect(useTun, systemProxy, a.AppDataDir())
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return msg
}

func (a *App) SwitchServer(index int, useTun, systemProxy bool) string {
	msg, err := a.vpnSvc.SwitchServer(index, useTun, systemProxy, a.AppDataDir())
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return msg
}

func (a *App) Disconnect() string {
	a.vpnSvc.Disconnect()
	return "Disconnected"
}

func (a *App) GetStatus() map[string]interface{} {
	status := a.vpnSvc.GetStatus()
	status["appDataDir"] = a.AppDataDir()
	status["isInstalling"] = a.coreMgr.IsInstalling()
	return status
}

func (a *App) AppDataDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "YokaVPN")
}

func (a *App) GetStats() *domain.Stats         { return a.client.GetStats() }
func (a *App) GetOutboundIP() string           { return geo.GetOutboundIP() }
func (a *App) SetActiveServer(index int) error { return a.client.SetActiveServer(index) }

func (a *App) FetchSubscription(name string, url string) (map[string]interface{}, error) {
	client := subscription.NewRemnawaveClient(url)
	configs, info, err := client.FetchConfigs()
	if err != nil {
		return nil, err
	}

	a.subMgr.AddSubscription(name, url)
	a.subMgr.SetActive(name)

	linksJSON, _ := json.Marshal(configs)
	os.WriteFile(filepath.Join(a.AppDataDir(), "config.json"), linksJSON, 0644)
	a.client.SetLinks(configs)

	// After fetch, auto select best
	go a.AutoSelectBestServer(configs)

	return map[string]interface{}{
		"success": true,
		"servers": usecase.ParseServers(configs),
		"upload":  info.Upload, "download": info.Download, "total": info.Total,
	}, nil
}

func (a *App) TestServer(link string, index int, name string) *domain.PingResult {
	return ping.TestViaXray(link, index, name)
}

func (a *App) Cleanup()                     { a.vpnSvc.ForceCleanup() }
func (a *App) shutdown(ctx context.Context) { a.Cleanup() }
