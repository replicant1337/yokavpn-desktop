//go:build !windows
// +build !windows

package utils

import (
	"os/exec"
	"strings"
)

func enableWindowsProxy(s string) error { return nil }
func disableWindowsProxy() error        { return nil }

func enableMacOSProxy(proxyServer string) error {
	parts := strings.Split(proxyServer, ":")
	host := parts[0]
	port := parts[1]

	out, _ := exec.Command("sh", "-c", "networksetup -listallnetworkservices | grep -v '*' | head -n 1").Output()
	iface := strings.TrimSpace(string(out))
	if iface == "" { iface = "Wi-Fi" }

	exec.Command("networksetup", "-setwebproxy", iface, host, port).Run()
	exec.Command("networksetup", "-setsecurewebproxy", iface, host, port).Run()
	return nil
}

func disableMacOSProxy() error {
	out, _ := exec.Command("sh", "-c", "networksetup -listallnetworkservices | grep -v '*' | head -n 1").Output()
	iface := strings.TrimSpace(string(out))
	if iface == "" { iface = "Wi-Fi" }

	exec.Command("networksetup", "-setwebproxystate", iface, "off").Run()
	exec.Command("networksetup", "-setsecurewebproxystate", iface, "off").Run()
	return nil
}

func enableLinuxProxy(proxyServer string) error {
	parts := strings.Split(proxyServer, ":")
	host := parts[0]
	port := parts[1]

	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", host).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "port", port).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "host", host).Run()
	exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "port", port).Run()
	return nil
}

func disableLinuxProxy() error {
	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	return nil
}
