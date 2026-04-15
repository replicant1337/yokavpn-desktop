package update

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type UpdateInfo struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Update  bool   `json:"update"`
	URL     string `json:"url"`
	Body    string `json:"body"`
}

type Checker struct {
	version string
	apiURL  string
}

func NewChecker(version, apiURL string) *Checker {
	return &Checker{
		version: version,
		apiURL:  apiURL,
	}
}

func (u *Checker) Check() (*UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var release map[string]interface{}
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	tagName, _ := release["tag_name"].(string)
	latest := strings.TrimPrefix(tagName, "v")

	return &UpdateInfo{
		Current: u.version,
		Latest:  latest,
		Update:  latest != u.version,
		URL:     release["html_url"].(string),
		Body:    release["body"].(string),
	}, nil
}

type UpdateChecker struct {
	version string
	apiURL  string
}

func NewUpdateChecker(version, apiURL string) *UpdateChecker {
	return &UpdateChecker{
		version: version,
		apiURL:  apiURL,
	}
}

func (u *UpdateChecker) Check() (*UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var release map[string]interface{}
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	tagName, _ := release["tag_name"].(string)
	latest := strings.TrimPrefix(tagName, "v")

	return &UpdateInfo{
		Current: u.version,
		Latest:  latest,
		Update:  latest != u.version,
		URL:     release["html_url"].(string),
		Body:    release["body"].(string),
	}, nil
}
