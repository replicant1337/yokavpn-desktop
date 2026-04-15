package main

import (
	"embed"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"YokaVPN/internal/utils"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var icon []byte

func main() {
	// Hide console immediately on Windows
	if runtime.GOOS == "windows" {
		utils.HideConsoleWindow()
	}

	app := NewApp()

	// Robust cleanup
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		app.Cleanup()
		os.Exit(0)
	}()

	// System Tray
	go func() {
		systray.Run(func() {
			systray.SetIcon(icon)
			systray.SetTitle("YokaVPN")
			systray.SetTooltip("YokaVPN")

			mShow := systray.AddMenuItem("Show YokaVPN", "Show the main window")
			mHide := systray.AddMenuItem("Hide to Tray", "Hide the main window")
			systray.AddSeparator()
			mQuit := systray.AddMenuItem("Quit", "Quit the application")

			for {
				select {
				case <-mShow.ClickedCh:
					if app.ctx != nil {
						wailsruntime.WindowShow(app.ctx)
						wailsruntime.WindowUnminimise(app.ctx)
					}
				case <-mHide.ClickedCh:
					if app.ctx != nil {
						wailsruntime.WindowHide(app.ctx)
					}
				case <-mQuit.ClickedCh:
					if app.ctx != nil {
						wailsruntime.Quit(app.ctx)
					}
					systray.Quit()
					return
				}
			}
		}, func() {
			app.Cleanup()
		})
	}()

	err := wails.Run(&options.App{
		Title:  "YokaVPN",
		Width:  380,
		Height: 640,
		MinWidth: 320,
		MinHeight: 500,
		MaxWidth: 450,
		MaxHeight: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		// Single Instance Lock
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "yokavpn-unique-id-2026",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				if app.ctx != nil {
					wailsruntime.WindowShow(app.ctx)
					wailsruntime.WindowUnminimise(app.ctx)
					wailsruntime.WindowCenter(app.ctx)
				}
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
