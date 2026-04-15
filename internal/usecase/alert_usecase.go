package usecase

import (
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
)

type AlertManager struct {
	iconPath string
}

func NewAlertManager(appDataDir string) *AlertManager {
	var iconPath string

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, "build", "appicon.png"),
		filepath.Join(exeDir, "build", "windows", "icon.ico"),
		filepath.Join(appDataDir, "appicon.png"),
	}

	for _, p := range candidates {
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				iconPath = abs
				break
			}
		}
	}

	return &AlertManager{
		iconPath: iconPath,
	}
}

func (m *AlertManager) Notify(title, message string) {
	// Always ensure YokaVPN is the title to avoid DefaultAppName
	displayTitle := "YokaVPN"
	if title != "" && title != "YokaVPN" {
		displayTitle = "YokaVPN - " + title
	}

	// Use beeep with explicit icon path
	_ = beeep.Notify(displayTitle, message, m.iconPath)
}

func (m *AlertManager) Alert(title, message string) {
	displayTitle := "YokaVPN"
	if title != "" && title != "YokaVPN" {
		displayTitle = "YokaVPN - " + title
	}
	_ = beeep.Alert(displayTitle, message, m.iconPath)
}
