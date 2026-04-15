//go:build windows
// +build windows

package utils

import (
	"strings"
	"syscall"
	"golang.org/x/sys/windows/registry"
)

var (
	modwininet            = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption = modwininet.NewProc("InternetSetOptionW")
)

const (
	INTERNET_OPTION_SETTINGS_CHANGED = 39
	INTERNET_OPTION_REFRESH          = 37
)

func enableMacOSProxy(s string) error { return nil }
func disableMacOSProxy() error        { return nil }
func enableLinuxProxy(s string) error  { return nil }
func disableLinuxProxy() error         { return nil }

func enableWindowsProxy(proxyServer string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	cleanProxy := strings.TrimSpace(proxyServer)
	cleanProxy = strings.TrimPrefix(cleanProxy, "http://")
	cleanProxy = strings.TrimPrefix(cleanProxy, "https://")
	cleanProxy = strings.TrimPrefix(cleanProxy, "socks=")

	k.SetDWordValue("ProxyEnable", 1)
	k.SetStringValue("ProxyServer", cleanProxy)
	k.SetStringValue("ProxyOverride", "localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*;<local>")
	
	k.SetStringValue("AutoConfigURL", "")
	k.SetDWordValue("ProxyHttp1.1", 1)
	
	refreshProxy()
	return nil
}

func disableWindowsProxy() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	k.SetDWordValue("ProxyEnable", 0)
	k.SetStringValue("ProxyServer", "")
	k.SetStringValue("ProxyOverride", "")
	k.SetStringValue("AutoConfigURL", "")
	
	refreshProxy()
	refreshProxy()
	return nil
}

func refreshProxy() {
	procInternetSetOption.Call(0, INTERNET_OPTION_SETTINGS_CHANGED, 0, 0)
	procInternetSetOption.Call(0, INTERNET_OPTION_REFRESH, 0, 0)
}
