package subscription

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"YokaVPN/internal/domain"
	"YokaVPN/internal/log"
	"YokaVPN/internal/utils"
)

type RemnawaveClient struct {
	baseURL string
	client  *http.Client
	logger  *log.Logger
}

type APIResponse struct {
	Links       []string         `json:"links"`
	SSConfLinks map[string][]any `json:"ssConfLinks"`
	User        UserInfo         `json:"user"`
}

type UserInfo struct {
	Username     string `json:"username"`
	DaysLeft     int    `json:"daysLeft"`
	TrafficUsed  string `json:"trafficUsed"`
	TrafficLimit string `json:"trafficLimit"`
	ExpiresAt    string `json:"expiresAt"`
	IsActive     bool   `json:"isActive"`
}

func NewRemnawaveClient(baseURL string) *RemnawaveClient {
	logger, _ := log.NewLogger("subscription", log.GetLogDir())
	logger.SetConsole(false)

	return &RemnawaveClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
		logger:  logger,
	}
}

func (r *RemnawaveClient) FetchConfigs() ([]string, *domain.SubscriptionInfo, error) {
	info := &domain.SubscriptionInfo{URL: r.baseURL}

	if r.logger != nil {
		r.logger.Info("Fetching subscription from: %s", r.baseURL)
	}

	req, err := http.NewRequest("GET", r.baseURL, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", "v2rayN/6.31")
	req.Header.Set("Subscription-User-Agent", "v2rayN/6.31")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if userinfo := resp.Header.Get("subscription-userinfo"); userinfo != "" {
		parseUserInfo(info, userinfo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	raw := string(body)
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b {
		body, err = decompressGzip(body)
		if err == nil {
			raw = string(body)
		}
	}

	raw = strings.TrimSpace(raw)

	var apiResp APIResponse
	if err := json.Unmarshal([]byte(raw), &apiResp); err == nil && len(apiResp.Links) > 0 {
		parseUserInfoFromUser(info, apiResp.User)
		return apiResp.Links, info, nil
	}

	var linkArray []string
	if err := json.Unmarshal([]byte(raw), &linkArray); err == nil && len(linkArray) > 0 {
		return linkArray, info, nil
	}

	decoded, err := decodeBase64Safe(raw)
	if err == nil {
		raw = decoded
	}

	configs := parseLinks(raw)
	return configs, info, nil
}

func decodeBase64Safe(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s)%4 != 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(b), nil
	}
	b, err = base64.URLEncoding.DecodeString(s)
	if err == nil {
		return string(b), nil
	}
	return "", err
}

func parseUserInfoFromUser(info *domain.SubscriptionInfo, user UserInfo) {
	info.Upload, _ = parseTraffic(user.TrafficUsed)
	info.Total, _ = parseTraffic(user.TrafficLimit)
	if user.ExpiresAt != "" {
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", user.ExpiresAt); err == nil {
			info.ExpiresAt = t.Format("2006-01-02 15:04:05")
			info.Expire = t.Unix()
		}
	}
}

func parseTraffic(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	var mult int64 = 1
	if strings.Contains(s, "G") { mult = 1024 * 1024 * 1024 } else if strings.Contains(s, "M") { mult = 1024 * 1024 } else if strings.Contains(s, "K") { mult = 1024 }
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return int64(val * float64(mult)), nil
}

func parseUserInfo(info *domain.SubscriptionInfo, userinfo string) {
	parts := strings.Split(userinfo, ";")
	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), "=")
		if len(kv) != 2 { continue }
		var val int64
		fmt.Sscanf(kv[1], "%d", &val)
		switch kv[0] {
		case "upload": info.Upload = val
		case "download": info.Download = val
		case "total": info.Total = val
		case "expire":
			info.Expire = val
			if val > 0 { info.ExpiresAt = time.Unix(val, 0).Format("2006-01-02 15:04:05") }
		}
	}
}

func parseLinks(content string) []string {
	var configs []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" { configs = append(configs, line) }
	}
	return configs
}

func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil { return nil, err }
	defer reader.Close()
	return io.ReadAll(reader)
}

func ParseServerInfo(cfg string) (name string, flag string, transport string) {
	cfg = strings.TrimSpace(cfg)
	flag = utils.GetProtocolFlag(cfg)
	transport = detectTransport(cfg)
	parts := strings.Split(cfg, "#")
	if len(parts) > 1 {
		name = parts[len(parts)-1]
		unescapedName, _ := url.QueryUnescape(name)
		runes := []rune(unescapedName)
		if len(runes) >= 2 && runes[0] >= 0x1F1E6 && runes[0] <= 0x1F1FF && runes[1] >= 0x1F1E6 && runes[1] <= 0x1F1FF {
			flag = string(runes[0:2])
			name = strings.TrimSpace(string(runes[2:]))
		} else {
			flag = utils.GetFlag(unescapedName)
			name = unescapedName
		}
	}
	if name == "" { name = "Server" }
	return
}

func detectTransport(cfg string) string {
	lower := strings.ToLower(cfg)
	if strings.Contains(lower, "type=xhttp") { return "xhttp" }
	if strings.Contains(lower, "type=grpc") || strings.Contains(lower, "serviceName=") { return "grpc" }
	if strings.Contains(lower, "type=ws") || strings.Contains(lower, "path=") { return "ws" }
	if strings.Contains(lower, "flow=xtls") { return "xtls" }
	if strings.Contains(lower, "security=reality") { return "reality" }
	return "tcp"
}
