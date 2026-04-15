package utils

import (
	"strconv"
	"strings"
)

var countryFlags = map[string]string{
	"US": "🇺🇸", "GB": "🇬🇧", "DE": "🇩🇪", "FR": "🇫🇷", "NL": "🇳🇱",
	"RU": "🇷🇺", "UA": "🇺🇦", "KZ": "🇰🇿", "JP": "🇯🇵", "KR": "🇰🇷",
	"SG": "🇸🇬", "HK": "🇭🇰", "TW": "🇹🇼", "AU": "🇦🇺", "CA": "🇨🇦",
	"BR": "🇧🇷", "IN": "🇮🇳", "TR": "🇹🇷", "AE": "🇦🇪",
}

var countryKeywords = []struct {
	keywords []string
	flag     string
}{
	{[]string{"RUSSIA", "MOSCOW", "MOW", "SPB", "RF", "РФ"}, "🇷🇺"},
	{[]string{"GERMANY", "DEUTSCHLAND", "DE."}, "🇩🇪"},
	{[]string{"FRANCE", "FR."}, "🇫🇷"},
	{[]string{"NETHERLANDS", "HOLLAND", "NL."}, "🇳🇱"},
	{[]string{"UNITED STATES", "USA", "AMERICA", "US."}, "🇺🇸"},
	{[]string{"UKRAINE", "UA."}, "🇺🇦"},
	{[]string{"JAPAN", "TOKYO", "JP."}, "🇯🇵"},
	{[]string{"KOREA", "SEOUL", "KR."}, "🇰🇷"},
	{[]string{"SINGAPORE", "SG."}, "🇸🇬"},
	{[]string{"HONG KONG", "HK."}, "🇭🇰"},
	{[]string{"TAIWAN", "TW."}, "🇹🇼"},
	{[]string{"AUSTRALIA", "SYDNEY", "AU."}, "🇦🇺"},
	{[]string{"CANADA", "TORONTO", "CA."}, "🇨🇦"},
	{[]string{"BRAZIL", "BR."}, "🇧🇷"},
	{[]string{"INDIA", "IN."}, "🇮🇳"},
	{[]string{"TURKEY", "TÜRKİYE", "TR."}, "🇹🇷"},
	{[]string{"UAE", "DUBAI", "AE."}, "🇦🇪"},
	{[]string{"UK", "BRITAIN", "ENGLAND", "LONDON", "GB."}, "🇬🇧"},
}

func GetFlag(name string) string {
	nameUpper := strings.ToUpper(name)

	for _, c := range countryKeywords {
		for _, kw := range c.keywords {
			if strings.Contains(nameUpper, kw) {
				return c.flag
			}
		}
	}

	for code, flag := range countryFlags {
		if strings.Contains(nameUpper, code) {
			return flag
		}
	}

	return "🌐"
}

func GetFlagText(name string) string {
	nameUpper := strings.ToUpper(name)

	if strings.Contains(nameUpper, "RUSSIA") || strings.Contains(nameUpper, "MOSCOW") || strings.Contains(nameUpper, "MOW") {
		return "[RU]"
	}
	if strings.Contains(nameUpper, "GERMANY") || strings.Contains(nameUpper, "DE.") {
		return "[DE]"
	}
	if strings.Contains(nameUpper, "FRANCE") || strings.Contains(nameUpper, "FR.") {
		return "[FR]"
	}
	if strings.Contains(nameUpper, "NETHERLANDS") || strings.Contains(nameUpper, "HOLLAND") {
		return "[NL]"
	}
	if strings.Contains(nameUpper, "USA") || strings.Contains(nameUpper, "AMERICA") || strings.Contains(nameUpper, "US.") {
		return "[US]"
	}
	if strings.Contains(nameUpper, "UKRAINE") {
		return "[UA]"
	}
	if strings.Contains(nameUpper, "JAPAN") || strings.Contains(nameUpper, "TOKYO") {
		return "[JP]"
	}
	if strings.Contains(nameUpper, "KOREA") || strings.Contains(nameUpper, "SEOUL") {
		return "[KR]"
	}
	if strings.Contains(nameUpper, "SINGAPORE") {
		return "[SG]"
	}
	if strings.Contains(nameUpper, "HONG KONG") || strings.Contains(nameUpper, "HK.") {
		return "[HK]"
	}
	if strings.Contains(nameUpper, "TAIWAN") {
		return "[TW]"
	}
	if strings.Contains(nameUpper, "AUSTRALIA") || strings.Contains(nameUpper, "SYDNEY") {
		return "[AU]"
	}
	if strings.Contains(nameUpper, "CANADA") || strings.Contains(nameUpper, "TORONTO") {
		return "[CA]"
	}
	if strings.Contains(nameUpper, "BRAZIL") {
		return "[BR]"
	}
	if strings.Contains(nameUpper, "INDIA") {
		return "[IN]"
	}
	if strings.Contains(nameUpper, "TURKEY") {
		return "[TR]"
	}
	if strings.Contains(nameUpper, "UAE") || strings.Contains(nameUpper, "DUBAI") {
		return "[AE]"
	}
	if strings.Contains(nameUpper, "UK") || strings.Contains(nameUpper, "BRITAIN") || strings.Contains(nameUpper, "ENGLAND") {
		return "[UK]"
	}

	return "[??]"
}

func GetProtocolFlag(cfg string) string {
	if strings.Contains(cfg, "vmess://") {
		return "[V]"
	}
	if strings.Contains(cfg, "vless://") {
		return "[L]"
	}
	if strings.Contains(cfg, "trojan://") {
		return "[T]"
	}
	if strings.Contains(cfg, "ss://") {
		return "[S]"
	}
	return "[?]"
}

func ExtractName(cfg string, index int) string {
	nameParts := strings.Split(cfg, "name:")
	if len(nameParts) > 1 {
		nameEnd := strings.Split(nameParts[1], "\"")
		if len(nameEnd) > 1 {
			return strings.TrimSpace(nameEnd[1])
		}
	}
	return "Server " + strconv.Itoa(index+1)
}

func FormatBytes(bytes uint64) string {
	if bytes < 1024 {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	if bytes < 1024*1024 {
		return strconv.FormatFloat(float64(bytes)/1024, 'f', 1, 64) + " KB"
	}
	if bytes < 1024*1024*1024 {
		return strconv.FormatFloat(float64(bytes)/1024/1024, 'f', 1, 64) + " MB"
	}
	return strconv.FormatFloat(float64(bytes)/1024/1024/1024, 'f', 2, 64) + " GB"
}
