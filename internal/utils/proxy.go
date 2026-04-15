package utils

import (
	"fmt"
	"runtime"
)

func EnableSystemProxy(proxyServer string) error {
	switch runtime.GOOS {
	case "windows":
		return enableWindowsProxy(proxyServer)
	case "darwin":
		return enableMacOSProxy(proxyServer)
	case "linux":
		return enableLinuxProxy(proxyServer)
	}
	return fmt.Errorf("unsupported platform")
}

func DisableSystemProxy() error {
	switch runtime.GOOS {
	case "windows":
		return disableWindowsProxy()
	case "darwin":
		return disableMacOSProxy()
	case "linux":
		return disableLinuxProxy()
	}
	return nil
}
