package utils

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
)

func GetFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil { return 10808 }
	l, err := net.ListenTCP("tcp", addr)
	if err != nil { return 10808 }
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func IsPortFree(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil { return false }
	l.Close()
	return true
}

func KillProcessesByName(names []string) {
	for _, name := range names {
		if runtime.GOOS == "windows" {
			exec.Command("taskkill", "/F", "/IM", name, "/T").Run()
		} else {
			exec.Command("killall", "-9", name).Run()
		}
	}
}
