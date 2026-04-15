package main

// App config
var (
	APP_NAME    = "YokaVPN"
	APP_VERSION = "1.0.3"
)

// Download config
const (
	DOWNLOAD_API_URL = "https://api.github.com/repos/replicant1337/yokavpn-desktop/releases/latest"
	DOWNLOAD_REPO    = "replicant1337/yokavpn-desktop"
)

// Core config
const (
	DEFAULT_SUBSCRIPTION = ""

	XRAY_VERSION = "v24.12.31"
	TUN_VERSION  = "v0.7.20"
)

// Platform-specific downloads
var CoreDownloads = map[string]map[string]string{
	"windows": {
		"xray": "https://github.com/XTLS/Xray-core/releases/download/v24.12.31/Xray-windows-64.zip",
		"tun":  "https://github.com/ICKXING/Tun2Proxy/releases/download/v0.7.20/windows-amd64.zip",
	},
	"linux": {
		"xray": "https://github.com/XTLS/Xray-core/releases/download/v24.12.31/Xray-linux-64.zip",
		"tun":  "https://github.com/ICKXING/Tun2Proxy/releases/download/v0.7.20/linux-amd64.zip",
	},
	"darwin": {
		"xray": "https://github.com/XTLS/Xray-core/releases/download/v24.12.31/Xray-darwin-arm64-v.zip",
		"tun":  "https://github.com/ICKXING/Tun2Proxy/releases/download/v0.7.20/darwin-amd64.zip",
	},
}

// Platform release assets
var ReleaseAssets = map[string]map[string]string{
	"windows": {"exe": "YokaVPN.exe", "zip": ""},
	"linux":   {"exe": "YokaVPN", "zip": "YokaVPN.tar.gz"},
	"darwin":  {"exe": "YokaVPN", "zip": "YokaVPN.dmg"},
}

// Get core download URL for current platform
func GetCoreURL(platform, coreType string) string {
	if urls, ok := CoreDownloads[platform]; ok {
		return urls[coreType]
	}
	return ""
}
