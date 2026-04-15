package utils

import (
	"testing"
)

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Russia Moscow", "🇷🇺"},
		{"Germany Berlin", "🇩🇪"},
		{"USA New York", "🇺🇸"},
		{"France Paris", "🇫🇷"},
		{"Japan Tokyo", "🇯🇵"},
		{"Singapore", "🇸🇬"},
		{"Hong Kong", "🇭🇰"},
		{"Unknown Place", "🌐"},
	}

	for _, tt := range tests {
		got := GetFlag(tt.name)
		if got != tt.want {
			t.Errorf("GetFlag(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGetProtocolFlag(t *testing.T) {
	tests := []struct {
		cfg  string
		want string
	}{
		{"vmess://config", "⚡"},
		{"vless://uuid@host:443", "🔮"},
		{"trojan://uuid@host:443", "🐴"},
		{"ss://config", "🔒"},
		{"unknown://config", "🌐"},
	}

	for _, tt := range tests {
		got := GetProtocolFlag(tt.cfg)
		if got != tt.want {
			t.Errorf("GetProtocolFlag(%q) = %q, want %q", tt.cfg, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}
