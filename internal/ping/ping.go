package ping

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"runtime"

	"golang.org/x/net/proxy"
	"yokovpn/internal/domain"
	"yokovpn/internal/log"
)

var pingLogger *log.Logger

func init() {
	pingLogger, _ = log.NewLogger("ping", log.GetLogDir())
	pingLogger.SetConsole(false)
}

func findXrayPath() string {
	exeDir := filepath.Dir(os.Args[0])
	paths := []string{
		filepath.Join(exeDir, "core", "xray.exe"),
		filepath.Join(exeDir, "..", "core", "xray.exe"),
		filepath.Join(os.Getenv("APPDATA"), "YokoVPN", "core", "xray.exe"),
		"build/bin/core/xray.exe",
		"build/bin/xray.exe",
		"xray.exe",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return "xray"
}

func getAssetDir() string {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	localCore := filepath.Join(exeDir, "core")
	if _, err := os.Stat(filepath.Join(localCore, "geosite.dat")); err == nil {
		return localCore
	}

	buildPath := "build/bin/core"
	if _, err := os.Stat(filepath.Join(buildPath, "geosite.dat")); err == nil {
		abs, _ := filepath.Abs(buildPath)
		return abs
	}

	appDataAsset := filepath.Join(os.Getenv("APPDATA"), "YokoVPN", "core")
	return appDataAsset
}

var testTargets = []string{
	"https://www.cloudflare.com/cdn-cgi/trace",
	"https://1.1.1.1/cdn-cgi/trace",
	"https://cp.cloudflare.com/generate_204",
	"https://www.google.com/generate_204",
}

func TestProxy(link string, socks5Addr string, timeout time.Duration) *domain.PingResult {
	result := &domain.PingResult{
		TargetURL: testTargets[0],
	}

	proxyURL, err := url.Parse(socks5Addr)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("proxy URL error: %v", err)
		return result
	}

	var auth *proxy.Auth
	if proxyURL.User != nil {
		auth = &proxy.Auth{
			User: proxyURL.User.Username(),
		}
		auth.Password, _ = proxyURL.User.Password()
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("dialer error: %v", err)
		return result
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	start := time.Now()

	for _, target := range testTargets {
		resp, err := httpClient.Get(target)
		if err != nil {
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 400 {
			result.Success = true
			result.Latency = time.Since(start).Milliseconds()
			result.TargetURL = target
			return result
		}
	}

	result.Success = false
	result.Error = "connection failed"
	result.Latency = -1
	return result
}

func getFreePort() int {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestViaXray(link string, index int, name string) *domain.PingResult {
	result := &domain.PingResult{
		Index: index,
		Name:  name,
	}

	socksPort := getFreePort()
	if socksPort == 0 {
		socksPort = 20000 + index
	}

	cfg, err := BuildPingConfig(link, socksPort)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Latency = -1
		return result
	}

	cfgJSON, _ := json.Marshal(cfg)
	tmpDir := os.TempDir()
	cfgPath := filepath.Join(tmpDir, fmt.Sprintf("xray_test_%d.json", time.Now().UnixNano()))
	os.WriteFile(cfgPath, cfgJSON, 0644)
	defer os.Remove(cfgPath)

	xrayPath := findXrayPath()
	assetDir := getAssetDir()

	xrayProcess := exec.Command(xrayPath, "run", "-c", cfgPath)
	xrayProcess.Dir = assetDir
	xrayProcess.Env = append(os.Environ(), 
		"XRAY_LOCATION_ASSET="+assetDir,
		"V2RAY_LOCATION_ASSET="+assetDir,
	)
	if runtime.GOOS == "windows" {
		xrayProcess.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	var stderr strings.Builder
	xrayProcess.Stderr = &stderr

	if err := xrayProcess.Start(); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Latency = -1
		return result
	}

	// Defer forceful cleanup for safety
	defer func() {
		if xrayProcess.Process != nil {
			if runtime.GOOS == "windows" {
				exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", xrayProcess.Process.Pid)).Run()
			} else {
				xrayProcess.Process.Kill()
			}
			xrayProcess.Wait()
		}
	}()

	ready := false
	for i := 0; i < 60; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !ready {
		result.Success = false
		result.Error = "xray start timeout"
		if pingLogger != nil {
			pingLogger.Debug("Xray test timeout. Stderr: %s", stderr.String())
		}
		result.Latency = -1
		return result
	}

	socks5Addr := fmt.Sprintf("socks5://127.0.0.1:%d", socksPort)
	pingResult := TestProxy(link, socks5Addr, 10*time.Second)

	if pingResult.Success {
		result.Success = true
		result.Latency = pingResult.Latency
		result.TargetURL = pingResult.TargetURL
	} else {
		result.Success = false
		result.Error = pingResult.Error
		result.Latency = -1
	}

	return result
}

type XrayConfig struct {
	Log       interface{}    `json:"log"`
	DNS       *DNSConfig     `json:"dns,omitempty"`
	Inbounds  []Inbound      `json:"inbounds"`
	Outbounds []Outbound     `json:"outbounds"`
	Routing   *RoutingConfig `json:"routing,omitempty"`
	API       *APIConfig     `json:"api,omitempty"`
	Stats     interface{}    `json:"stats,omitempty"`
	Policy    *PolicyConfig  `json:"policy,omitempty"`
}

type DNSConfig struct {
	Servers []interface{} `json:"servers"`
}

type RoutingConfig struct {
	DomainStrategy string        `json:"domainStrategy"`
	Rules          []RoutingRule `json:"rules"`
}

type RoutingRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag"`
	IP          []string `json:"ip,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	Port        string   `json:"port,omitempty"`
}

type Inbound struct {
	Tag      string          `json:"tag"`
	Protocol string          `json:"protocol"`
	Port     string          `json:"port"`
	Listen   string          `json:"listen"`
	Settings json.RawMessage `json:"settings,omitempty"`
	Sniffing *SniffingConfig `json:"sniffing,omitempty"`
}

type SniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

type PolicyConfig struct {
	Levels map[string]PolicyLevel `json:"levels"`
	System *SystemPolicy          `json:"system"`
}

type PolicyLevel struct {
	StatsUplink   bool `json:"statsUserUplink"`
	StatsDownlink bool `json:"statsUserDownlink"`
}

type SystemPolicy struct {
	StatsIncoming bool `json:"statsInboundUplink"`
	StatsInboundDownlink bool `json:"statsInboundDownlink"`
}

type Outbound struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
}

func BuildPingConfig(link string, socksPort int) (*XrayConfig, error) {
	link = strings.TrimSpace(link)
	sniffing := &SniffingConfig{Enabled: true, DestOverride: []string{"http", "tls"}}
	cfg := &XrayConfig{
		Log: map[string]interface{}{"access": "", "error": "", "loglevel": "debug"},
		Inbounds: []Inbound{{
			Tag: "socks-in", Protocol: "socks", Port: strconv.Itoa(socksPort), Listen: "127.0.0.1",
			Settings: json.RawMessage(`{"auth": "noauth", "udp": true}`), Sniffing: sniffing,
		}},
		Outbounds: []Outbound{},
	}
	var outbound Outbound
	var err error
	switch {
	case strings.HasPrefix(link, "vless://"): outbound, err = parseVLESS(link)
	case strings.HasPrefix(link, "vmess://"): outbound, err = parseVMess(link)
	case strings.HasPrefix(link, "trojan://"): outbound, err = parseTrojan(link)
	default: return nil, fmt.Errorf("unsupported protocol")
	}
	if err != nil { return nil, err }
	outbound.Tag = "proxy"
	cfg.Outbounds = append(cfg.Outbounds, outbound)
	return cfg, nil
}

func BuildXrayConfig(link string, socksPort, httpPort, apiPort int, user, pass string) (*XrayConfig, error) {
	link = strings.TrimSpace(link)
	sniffing := &SniffingConfig{Enabled: true, DestOverride: []string{"http", "tls"}}
	socksIn := Inbound{Tag: "socks-in", Protocol: "socks", Port: strconv.Itoa(socksPort), Listen: "127.0.0.1", Sniffing: sniffing}
	httpIn := Inbound{Tag: "http-in", Protocol: "http", Port: strconv.Itoa(httpPort), Listen: "127.0.0.1", Sniffing: sniffing}
	if user != "" && pass != "" {
		socksIn.Settings = json.RawMessage(fmt.Sprintf(`{"auth": "password", "accounts": [{"user": "%s", "pass": "%s"}], "udp": true, "ip": "127.0.0.1"}`, user, pass))
		httpIn.Settings = json.RawMessage(fmt.Sprintf(`{"accounts": [{"user": "%s", "pass": "%s"}]}`, user, pass))
	} else {
		socksIn.Settings = json.RawMessage(`{"auth": "noauth", "udp": true}`)
		httpIn.Settings = json.RawMessage(`{}`)
	}
	cfg := &XrayConfig{
		Log: map[string]interface{}{"access": "", "error": "", "loglevel": "debug"},
		DNS: &DNSConfig{Servers: []interface{}{
			map[string]interface{}{"address": "1.1.1.1", "domains": []string{"geosite:geolocation-!cn", "geosite:google", "geosite:youtube"}},
			map[string]interface{}{"address": "localhost", "domains": []string{"geosite:category-ru", "geosite:private", "domain:ru", "domain:su", "domain:xn--p1ai"}},
			"8.8.8.8",
		}},
		Inbounds: []Inbound{socksIn, httpIn, {Tag: "api", Protocol: "dokodemo-door", Port: strconv.Itoa(apiPort), Listen: "127.0.0.1", Settings: json.RawMessage(`{"address": "127.0.0.1"}`)}},
		Outbounds: []Outbound{},
	}
	var outbound Outbound
	var err error
	switch {
	case strings.HasPrefix(link, "vless://"): outbound, err = parseVLESS(link)
	case strings.HasPrefix(link, "vmess://"): outbound, err = parseVMess(link)
	case strings.HasPrefix(link, "trojan://"): outbound, err = parseTrojan(link)
	default: return nil, fmt.Errorf("unsupported protocol")
	}
	if err != nil { return nil, err }
	outbound.Tag = "proxy"
	cfg.Outbounds = append(cfg.Outbounds, outbound, Outbound{Tag: "direct", Protocol: "freedom"}, Outbound{Tag: "block", Protocol: "blackhole"})
	cfg.Routing = &RoutingConfig{DomainStrategy: "IPIfNonMatch", Rules: []RoutingRule{
		{Type: "field", InboundTag: []string{"api"}, OutboundTag: "api"},
		{Type: "field", IP: []string{"geoip:private", "geoip:ru"}, OutboundTag: "direct"},
		{Type: "field", Domain: []string{"geosite:private", "geosite:category-ru", "domain:ru", "domain:su", "domain:xn--p1ai"}, OutboundTag: "direct"},
		{Type: "field", Port: "53", OutboundTag: "proxy"},
	}}
	cfg.API = &APIConfig{Tag: "api", Services: []string{"HandlerService", "StatsService", "LoggerService"}}
	cfg.Stats = map[string]interface{}{}
	cfg.Policy = &PolicyConfig{Levels: map[string]PolicyLevel{"0": {StatsUplink: true, StatsDownlink: true}}, System: &SystemPolicy{StatsIncoming: true, StatsInboundDownlink: true}}
	return cfg, nil
}

func parseVLESS(link string) (Outbound, error) {
	u, err := url.Parse(link)
	if err != nil { return Outbound{}, err }
	port, _ := strconv.Atoi(u.Port())
	params := u.Query()
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{{
			"address": u.Hostname(), "port": port,
			"users": []map[string]interface{}{{"id": u.User.Username(), "encryption": "none", "flow": params.Get("flow")}},
		}},
	}
	settingsJSON, _ := json.Marshal(settings)
	outbound := Outbound{Protocol: "vless", Settings: settingsJSON}
	stream := buildStreamSettings(params.Get("type"), params, u.Hostname())
	streamJSON, _ := json.Marshal(stream)
	outbound.StreamSettings = streamJSON
	return outbound, nil
}

func parseVMess(link string) (Outbound, error) {
	s := strings.TrimPrefix(link, "vmess://")
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil { return Outbound{}, err }
	var v map[string]interface{}
	json.Unmarshal(data, &v)
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{{
			"address": v["add"], "port": v["port"],
			"users": []map[string]interface{}{{"id": v["id"], "alterId": v["aid"]}},
		}},
	}
	settingsJSON, _ := json.Marshal(settings)
	return Outbound{Protocol: "vmess", Settings: settingsJSON}, nil
}

func parseTrojan(link string) (Outbound, error) {
	u, err := url.Parse(link)
	if err != nil { return Outbound{}, err }
	port, _ := strconv.Atoi(u.Port())
	settings := map[string]interface{}{
		"servers": []map[string]interface{}{{"address": u.Hostname(), "port": port, "password": u.User.Username()}},
	}
	settingsJSON, _ := json.Marshal(settings)
	return Outbound{Protocol: "trojan", Settings: settingsJSON}, nil
}

func buildStreamSettings(transport string, params url.Values, addr string) map[string]interface{} {
	stream := map[string]interface{}{"network": transport}
	if transport == "" { transport = "tcp"; stream["network"] = "tcp" }
	sni := params.Get("sni"); if sni == "" { sni = addr }
	if transport == "xhttp" {
		stream["xhttpSettings"] = map[string]interface{}{"path": params.Get("path"), "mode": params.Get("mode")}
	} else if transport == "ws" {
		stream["wsSettings"] = map[string]interface{}{"path": params.Get("path"), "headers": map[string]string{"Host": params.Get("host")}}
	}
	if params.Get("security") == "reality" {
		stream["security"] = "reality"
		stream["realitySettings"] = map[string]interface{}{"enabled": true, "serverName": sni, "publicKey": params.Get("pbk"), "fingerprint": params.Get("fp"), "shortId": params.Get("sid")}
	} else if params.Get("security") == "tls" {
		stream["security"] = "tls"
		stream["tlsSettings"] = map[string]interface{}{"enabled": true, "serverName": sni}
	}
	return stream
}
