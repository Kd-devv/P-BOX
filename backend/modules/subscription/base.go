package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// NodeConfig 节点配置接口
type NodeConfig interface {
	ToJSON() (string, error)
}

// ProxyNode 代理节点结构
type ProxyNode struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort int    `json:"serverPort"`
	Config     string `json:"config,omitempty"`   // JSON格式的协议配置
	ShareURL   string `json:"shareUrl,omitempty"` // 原始分享链接
}

// ParseURL 解析单个代理URL
func ParseURL(proxyURL string) (*ProxyNode, error) {
	if strings.HasPrefix(proxyURL, "vmess://") {
		return ParseVMessURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "vless://") {
		return ParseVLESSURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "trojan://") {
		return ParseTrojanURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "ss://") {
		return ParseShadowsocksURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "hysteria2://") || strings.HasPrefix(proxyURL, "hy2://") {
		return ParseHysteria2URL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "hysteria://") {
		return ParseHysteriaURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "tuic://") {
		return ParseTUICURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "shadowtls://") {
		return ParseShadowTLSURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "wireguard://") {
		return nil, errors.New("wireguard 服务端不支持链接导入，请使用入口管理功能")
	} else if strings.HasPrefix(proxyURL, "ssh://") {
		return ParseSSHURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "naive") {
		return ParseNaiveURL(proxyURL)
	} else if strings.HasPrefix(proxyURL, "anytls://") {
		return ParseAnyTLSURL(proxyURL)
	}

	return nil, errors.New("不支持的协议格式")
}

// ParseQueryParams 解析URL查询参数
func ParseQueryParams(query string) map[string]string {
	params := make(map[string]string)
	if query == "" {
		return params
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key, _ := url.QueryUnescape(parts[0])
			value, _ := url.QueryUnescape(parts[1])
			params[key] = value
		}
	}
	return params
}

// DecodeBase64 解码Base64字符串
func DecodeBase64(encoded string) (string, error) {
	// 尝试标准Base64
	if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
		return string(decoded), nil
	}

	// 尝试URL安全的Base64
	if decoded, err := base64.URLEncoding.DecodeString(encoded); err == nil {
		return string(decoded), nil
	}

	// 尝试不带填充的Base64
	if decoded, err := base64.RawStdEncoding.DecodeString(encoded); err == nil {
		return string(decoded), nil
	}

	if decoded, err := base64.RawURLEncoding.DecodeString(encoded); err == nil {
		return string(decoded), nil
	}

	return "", errors.New("base64解码失败")
}

// ToJSONString 将结构体转换为JSON字符串
func ToJSONString(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseInt 安全解析整数
func ParseInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

// ParseBool 安全解析布尔值
func ParseBool(s string, defaultValue bool) bool {
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}
	return defaultValue
}

// TLSConfig TLS配置结构 (共享)
type TLSConfig struct {
	Enabled     bool        `json:"enabled"`
	ServerName  string      `json:"server_name,omitempty"`
	Insecure    bool        `json:"insecure,omitempty"`
	ALPN        []string    `json:"alpn,omitempty"`
	Fingerprint string      `json:"fingerprint,omitempty"`
	UTLS        *UTLSConfig `json:"utls,omitempty"`
}

// UTLSConfig uTLS配置结构 (共享)
type UTLSConfig struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint,omitempty"`
}
