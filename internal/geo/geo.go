package geo

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func GetLogsPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "YokoVPN", "logs")
}

func OpenLogsFolder() error {
	path := GetLogsPath()
	os.MkdirAll(path, 0755)
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	}
	return fmt.Errorf("unsupported platform")
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	local := conn.LocalAddr().(*net.UDPAddr)
	return local.IP.String()
}

func UpdateGeoAssets(targetDir string) error {
	assets := map[string]string{
		"geoip.dat":   "https://github.com/runetfreedom/russia-v2ray-rules-dat/releases/latest/download/geoip.dat",
		"geosite.dat": "https://github.com/runetfreedom/russia-v2ray-rules-dat/releases/latest/download/geosite.dat",
	}

	os.MkdirAll(targetDir, 0755)

	for name, url := range assets {
		if err := downloadFile(filepath.Join(targetDir, name), url); err != nil {
			return fmt.Errorf("failed to download %s: %v", name, err)
		}
	}
	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(filepath + ".tmp")
		return err
	}

	return os.Rename(filepath+".tmp", filepath)
}

type GeoIPInfo struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	ASN         string `json:"asn"`
	ISP         string `json:"isp"`
	IP          string `json:"ip"`
}

func Lookup(ip string) (*GeoIPInfo, error) {
	info := &GeoIPInfo{IP: ip}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	if isPrivate(ipAddr) {
		info.CountryCode = "PRIVATE"
		info.CountryName = "Private Network"
		return info, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,isp,org,as", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geoData struct {
		Status      string `json:"status"`
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
		ISP         string `json:"isp"`
		Org         string `json:"org"`
		AS          string `json:"as"`
	}

	if json.NewDecoder(resp.Body).Decode(&geoData) != nil {
		return nil, fmt.Errorf("decode error")
	}

	if geoData.Status != "success" {
		return nil, fmt.Errorf("lookup failed")
	}

	info.CountryCode = geoData.CountryCode
	info.CountryName = geoData.Country
	info.ISP = geoData.ISP
	info.ASN = geoData.AS

	return info, nil
}

func isPrivate(ip net.IP) bool {
	privateBlocks := []*net.IPNet{
		parseCIDR("10.0.0.0/8"),
		parseCIDR("172.16.0.0/12"),
		parseCIDR("192.168.0.0/16"),
		parseCIDR("127.0.0.0/8"),
		parseCIDR("169.254.0.0/16"),
	}
	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDR(cidr string) *net.IPNet {
	_, ipNet, _ := net.ParseCIDR(cidr)
	return ipNet
}
