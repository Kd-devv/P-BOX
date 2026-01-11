package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// ShadowTLSConfig ShadowTLS 配置结构
type ShadowTLSConfig struct {
	Version    int        `json:"version"`
	Password   string     `json:"password,omitempty"`
	ServerName string     `json:"server_name"`
	TLS        *TLSConfig `json:"tls,omitempty"`
}

// ParseShadowTLSURL 解析 ShadowTLS 链接
// 格式: shadowtls://password@server:port?params#name
func ParseShadowTLSURL(shadowtlsURL string) (*ProxyNode, error) {
	// 移除协议前缀
	urlStr := strings.TrimPrefix(shadowtlsURL, "shadowtls://")

	// 分离密码和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("shadowtls URL 格式错误")
	}

	password, _ := url.QueryUnescape(parts[0])
	remaining := parts[1]

	// 分离服务器和参数
	serverAndParams := strings.SplitN(remaining, "?", 2)
	serverPart := serverAndParams[0]
	queryString := ""
	if len(serverAndParams) == 2 {
		queryString = serverAndParams[1]
	}

	// 提取节点名称
	name := "ShadowTLS"
	if idx := strings.Index(queryString, "#"); idx != -1 {
		name, _ = url.QueryUnescape(queryString[idx+1:])
		queryString = queryString[:idx]
	} else if idx := strings.Index(serverPart, "#"); idx != -1 {
		name, _ = url.QueryUnescape(serverPart[idx+1:])
		serverPart = serverPart[:idx]
	}

	// 解析服务器和端口（移除可能的路径分隔符）
	serverPart = strings.TrimSuffix(serverPart, "/")
	serverParts := strings.Split(serverPart, ":")
	if len(serverParts) != 2 {
		return nil, errors.New("shadowtls 服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)

	// 构建配置
	config := ShadowTLSConfig{
		Version:    ParseInt(params["version"], 3),
		Password:   password,
		ServerName: params["sni"],
	}

	if config.ServerName == "" {
		config.ServerName = params["peer"]
	}

	// 解析 TLS 配置
	tlsConfig := &TLSConfig{
		Enabled:    true,
		ServerName: config.ServerName,
	}

	if fp := params["fingerprint"]; fp != "" {
		tlsConfig.Fingerprint = fp
	}

	if insecure := params["insecure"]; insecure == "1" || insecure == "true" {
		tlsConfig.Insecure = true
	}

	config.TLS = tlsConfig

	// 转换为 JSON 字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("shadowtls 配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "shadowtls",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现 NodeConfig 接口
func (c *ShadowTLSConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
