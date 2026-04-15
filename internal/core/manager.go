package core

import (
	"archive/zip"
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type ProgressFunc func(current, total int64, message string)

type Manager struct {
	BaseDir      string
	isInstalling bool
	mu           sync.Mutex
	onProgress   ProgressFunc
}

func NewManager(baseDir string) *Manager {
	os.MkdirAll(filepath.Join(baseDir, "core"), 0755)
	os.MkdirAll(filepath.Join(baseDir, "tun2"), 0755)
	return &Manager{BaseDir: baseDir}
}

func (m *Manager) SetProgressCallback(fn ProgressFunc) {
	m.onProgress = fn
}

func (m *Manager) IsInstalling() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isInstalling
}

func (m *Manager) GetCoreDir() string { return filepath.Join(m.BaseDir, "core") }
func (m *Manager) GetTunDir() string  { return filepath.Join(m.BaseDir, "tun2") }

func (m *Manager) GetCorePath() string {
	name := "xray"
	if runtime.GOOS == "windows" { name += ".exe" }
	return filepath.Join(m.GetCoreDir(), name)
}

func (m *Manager) GetTunPath() string {
	name := "tun2proxy-bin"
	if runtime.GOOS == "windows" { name += ".exe" }
	return filepath.Join(m.GetTunDir(), name)
}

func (m *Manager) GetAssetDir() string { return m.GetCoreDir() }

func (m *Manager) IsInstalled() bool {
	info, err := os.Stat(m.GetCorePath())
	return err == nil && info.Size() > 1000000 // Basic validation: must exist and be > 1MB
}

func (m *Manager) IsTunInstalled() bool {
	info, err := os.Stat(m.GetTunPath())
	return err == nil && info.Size() > 500000 // Basic validation
}

func (m *Manager) InstallXray(version string) error {
	var osName, arch string
	switch runtime.GOOS {
	case "windows": osName = "windows"
	case "darwin":  osName = "macos"
	case "linux":   osName = "linux"
	}
	switch runtime.GOARCH {
	case "amd64": arch = "64"
	case "386":   arch = "32"
	case "arm64": arch = "arm64-v8a"
	}
	
	url := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip", version, osName, arch)
	err := m.installFromArchive(url, m.GetCoreDir(), "Xray Core")
	if err == nil {
		os.Chmod(m.GetCorePath(), 0755)
	}
	return err
}

func (m *Manager) InstallTun2Proxy(version string) error {
	var target, ext string
	ext = "zip"
	switch runtime.GOOS {
	case "windows":
		target = "x86_64-pc-windows-msvc"
	case "darwin":
		target = "x86_64-apple-darwin"
		if runtime.GOARCH == "arm64" { target = "aarch64-apple-darwin" }
		ext = "tar.gz"
	case "linux":
		target = "x86_64-unknown-linux-gnu"
		ext = "tar.gz"
	}
	
	url := fmt.Sprintf("https://github.com/tun2proxy/tun2proxy/releases/download/%s/tun2proxy-%s.%s", version, target, ext)
	err := m.installFromArchive(url, m.GetTunDir(), "Tun2Proxy")
	if err != nil { return err }

	entries, _ := os.ReadDir(m.GetTunDir())
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "tun2proxy") && !strings.HasSuffix(e.Name(), ".zip") && !strings.HasSuffix(e.Name(), ".gz") {
			os.Rename(filepath.Join(m.GetTunDir(), e.Name()), m.GetTunPath())
			os.Chmod(m.GetTunPath(), 0755)
			break
		}
	}
	return nil
}

func (m *Manager) installFromArchive(url string, destDir string, label string) error {
	m.mu.Lock()
	if m.isInstalling { m.mu.Unlock(); return fmt.Errorf("busy") }
	m.isInstalling = true
	m.mu.Unlock()
	defer func() { m.mu.Lock(); m.isInstalling = false; m.mu.Unlock() }()

	resp, err := http.Get(url)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	// Wrapper to track progress
	reader := &progressReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
		Label:  label,
		OnProgress: m.onProgress,
	}

	if strings.HasSuffix(url, ".zip") {
		return m.unzipStream(reader, destDir)
	}
	return m.untargzStream(reader, destDir)
}

func (m *Manager) unzipStream(body io.Reader, dest string) error {
	tmp, _ := os.CreateTemp("", "xray-*.zip")
	defer os.Remove(tmp.Name())
	io.Copy(tmp, body)
	tmp.Close()

	r, err := zip.OpenReader(tmp.Name())
	if err != nil { return err }
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) { continue }
		if f.FileInfo().IsDir() { os.MkdirAll(fpath, 0755); continue }
		os.MkdirAll(filepath.Dir(fpath), 0755)
		out, _ := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		rc, _ := f.Open()
		io.Copy(out, rc)
		out.Close(); rc.Close()
	}
	return nil
}

func (m *Manager) untargzStream(body io.Reader, dest string) error {
	gzr, err := gzip.NewReader(body)
	if err != nil { return err }
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF { break }
		fpath := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) { continue }
		switch header.Typeflag {
		case tar.TypeDir: os.MkdirAll(fpath, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(fpath), 0755)
			out, _ := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			io.Copy(out, tr)
			out.Close()
		}
	}
	return nil
}

type progressReader struct {
	io.Reader
	Current    int64
	Total      int64
	Label      string
	OnProgress ProgressFunc
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.OnProgress != nil {
		pr.OnProgress(pr.Current, pr.Total, pr.Label)
	}
	return
}
