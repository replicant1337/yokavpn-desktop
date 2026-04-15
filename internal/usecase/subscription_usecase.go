package usecase

import (
	"fmt"
	"strings"

	"yokovpn/internal/domain"
	"yokovpn/internal/subscription"
)

type SubscriptionManager struct {
	subscriptions map[string]*subscription.RemnawaveClient
	active        string
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]*subscription.RemnawaveClient),
	}
}

func (sm *SubscriptionManager) AddSubscription(name string, url string) {
	sm.subscriptions[name] = subscription.NewRemnawaveClient(url)
}

func (sm *SubscriptionManager) RemoveSubscription(name string) {
	delete(sm.subscriptions, name)
	if sm.active == name {
		sm.active = ""
	}
}

func (sm *SubscriptionManager) SetActive(name string) {
	sm.active = name
}

func (sm *SubscriptionManager) GetActive() string {
	return sm.active
}

func (sm *SubscriptionManager) GetConfigs(name string) ([]string, *domain.SubscriptionInfo, error) {
	if client, ok := sm.subscriptions[name]; ok {
		return client.FetchConfigs()
	}
	return nil, nil, fmt.Errorf("subscription not found: %s", name)
}

func (sm *SubscriptionManager) GetActiveConfigs() ([]string, *domain.SubscriptionInfo, error) {
	if sm.active == "" {
		return nil, nil, fmt.Errorf("no active subscription")
	}
	return sm.GetConfigs(sm.active)
}

func (sm *SubscriptionManager) ListSubscriptions() []string {
	var names []string
	for name := range sm.subscriptions {
		names = append(names, name)
	}
	return names
}

func ParseServers(configs []string) []domain.Server {
	var servers []domain.Server
	for _, cfg := range configs {
		cfg = strings.TrimSpace(cfg)
		if cfg == "" {
			continue
		}

		// Strictly check for supported protocols
		isVless := strings.HasPrefix(cfg, "vless://")
		isVmess := strings.HasPrefix(cfg, "vmess://")
		isTrojan := strings.HasPrefix(cfg, "trojan://")

		if !isVless && !isVmess && !isTrojan {
			continue
		}

		name, flag, transport := subscription.ParseServerInfo(cfg)
		srvType := "vless"
		if isVmess {
			srvType = "vmess"
		} else if isTrojan {
			srvType = "trojan"
		}

		servers = append(servers, domain.Server{
			Name:      name,
			Flag:      flag,
			Index:     len(servers),
			Type:      srvType,
			Transport: transport,
			Link:      cfg,
			Latency:   -1,
		})
	}
	return servers
}
