package tun

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"YokaVPN/internal/log"
)

type TunManager struct {
	mu         sync.Mutex
	deviceName string
	serverIP   string
	tunnelIP   string
	gateway    string
	started    bool
	logger     *log.Logger
}

func NewTunManager() *TunManager {
	logger, _ := log.NewLogger("tunnel", log.GetLogDir())
	logger.SetConsole(false)

	return &TunManager{
		logger: logger,
	}
}

func (t *TunManager) Setup(serverIP string, tunnelIP string, gateway string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		return fmt.Errorf("tunnel already active")
	}

	t.serverIP = serverIP
	t.tunnelIP = tunnelIP
	t.gateway = gateway

	if t.logger != nil {
		t.logger.Info("Setting up tunnel - IP: %s, Gateway: %s, Server: %s", tunnelIP, gateway, serverIP)
	}

	switch runtime.GOOS {
	case "windows":
		return t.setupWindows()
	case "linux":
		return t.setupLinux()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (t *TunManager) setupWindows() error {
	t.deviceName = "wintun"

	if t.logger != nil {
		t.logger.Info("Windows: Setting up WINTUN interface")
	}

	exec.Command("netsh", "interface", "ipv4", "set", "address",
		t.deviceName, "static", t.tunnelIP, "255.255.255.0").Run()

	exec.Command("netsh", "interface", "ipv4", "set", "dns",
		t.deviceName, "static", "1.1.1.1", "index=1").Run()
	exec.Command("netsh", "interface", "ipv4", "add", "dns",
		t.deviceName, "8.8.8.8", "index=2").Run()

	if t.serverIP != "" {
		exec.Command("netsh", "interface", "ipv4", "add", "route",
			fmt.Sprintf("%s/32", t.serverIP), t.deviceName).Run()
		if t.logger != nil {
			t.logger.Info("Excluded server IP from tunnel: %s", t.serverIP)
		}
	}

	exec.Command("netsh", "interface", "ipv4", "set", "route",
		"0.0.0.0/0", t.gateway, t.deviceName, "store=persistent").Run()

	t.started = true

	if t.logger != nil {
		t.logger.Info("Tunnel setup complete - Device: %s", t.deviceName)
	}

	return nil
}

func (t *TunManager) setupLinux() error {
	t.deviceName = "tun0"

	if t.logger != nil {
		t.logger.Info("Linux: Setting up TUN interface")
	}

	exec.Command("ip", "tuntap", "add", "dev", t.deviceName, "mode", "tun").Run()
	exec.Command("ip", "addr", "add", t.tunnelIP+"/24", "dev", t.deviceName).Run()
	exec.Command("ip", "link", "set", "dev", t.deviceName, "up").Run()

	exec.Command("ip", "route", "add", "default", "via", t.gateway, "dev", t.deviceName).Run()

	if t.serverIP != "" {
		exec.Command("ip", "route", "add", t.serverIP, "via", t.gateway).Run()
		if t.logger != nil {
			t.logger.Info("Excluded server IP from tunnel: %s", t.serverIP)
		}
	}

	t.started = true

	if t.logger != nil {
		t.logger.Info("Tunnel setup complete - Device: %s", t.deviceName)
	}

	return nil
}

func (t *TunManager) Cleanup() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.started {
		return nil
	}

	if t.logger != nil {
		t.logger.Info("Cleaning up tunnel...")
	}

	switch runtime.GOOS {
	case "windows":
		exec.Command("netsh", "interface", "ipv4", "set", "route",
			"0.0.0.0/0", "store=persistent").Run()
		exec.Command("netsh", "interface", "set", "interface",
			t.deviceName, "admin=disabled").Run()
	case "linux":
		exec.Command("ip", "link", "del", t.deviceName).Run()
	}

	t.started = false

	if t.logger != nil {
		t.logger.Info("Tunnel cleanup complete")
	}

	return nil
}

func (t *TunManager) IsActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.started
}

func (t *TunManager) GetDeviceName() string {
	return t.deviceName
}

func (t *TunManager) CheckPrivileges() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		check := exec.Command("net", "session")
		if err := check.Run(); err != nil {
			if t.logger != nil {
				t.logger.Warn("Admin privileges check failed")
			}
			return false, fmt.Errorf("admin privileges required")
		}
		return true, nil
	case "linux":
		out, err := exec.Command("id", "-u").Output()
		if err != nil || string(out) != "0\n" {
			if t.logger != nil {
				t.logger.Warn("Root privileges check failed")
			}
			return false, fmt.Errorf("root required")
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported platform")
	}
}

func ParseServerIP(config string) string {
	if config == "" {
		return ""
	}

	parts := strings.Split(config, "\"address\"")
	if len(parts) < 2 {
		return ""
	}

	segments := strings.Split(parts[1], "\"")
	if len(segments) < 2 {
		return ""
	}

	addr := segments[1]
	if ip := net.ParseIP(addr); ip != nil {
		return ip.String()
	}

	host := strings.Split(addr, ":")[0]
	ips, _ := net.LookupIP(host)
	if len(ips) > 0 {
		return ips[0].String()
	}

	return addr
}
