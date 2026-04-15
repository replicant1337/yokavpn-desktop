package domain

import "time"

// VPNState represents the current state of the VPN connection
type VPNState string

const (
	StateIdle         VPNState = "idle"
	StateStarting     VPNState = "starting"
	StateConnected    VPNState = "connected"
	StateDisconnecting VPNState = "disconnecting"
	StateError        VPNState = "error"
)

// Server represents a VPN server configuration
type Server struct {
	Name      string `json:"name"`
	Flag      string `json:"flag"`
	Index     int    `json:"index"`
	Type      string `json:"type"`
	Transport string `json:"transport"`
	Link      string `json:"link"`
	Latency   int64  `json:"latency_ms"`
}

// Stats represents connection statistics
type Stats struct {
	UploadBytes   uint64 `json:"upload_bytes"`
	DownloadBytes uint64 `json:"download_bytes"`
	UploadRate    uint64 `json:"upload_rate"`
	DownloadRate  uint64 `json:"download_rate"`
	Connections   uint32 `json:"connections"`
}

// PingResult represents the result of a server ping test
type PingResult struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Latency   int64  `json:"latency_ms"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	TargetURL string `json:"target"`
}

// SubscriptionInfo contains metadata about a subscription
type SubscriptionInfo struct {
	URL       string    `json:"url"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	Total     int64     `json:"total"`
	Expire    int64     `json:"expire"`
	ExpiresAt string    `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
