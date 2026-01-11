package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// AnyTLSConfig AnyTLS 配置结构
type AnyTLSConfig struct {
	Password string         `json:"password"`
	TLS      *TLSConfig     `json:"tls,omitempty"`
	Reality  *RealityConfig `json:"reality,omitempty"`
}

// ParseAnyTLSURL 解析 AnyTLS 链接
// 格式: anytls://password@server:port?params#name
func ParseAnyTLSURL(anytlsURL string) (*ProxyNode, error) {
	// 移除协议前缀
	urlStr := strings.TrimPrefix(anytlsURL, "anytls://")

	// 分离密码和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("anytls URL 格式错误")
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
	name := "AnyTLS"
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
		return nil, errors.New("anytls 服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)

	// 构建配置
	config := AnyTLSConfig{
		Password: password,
	}

	// 解析 TLS 配置
	tlsConfig := &TLSConfig{
		Enabled:    true,
		ServerName: params["sni"],
	}

	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = params["peer"]
	}

	if alpn := params["alpn"]; alpn != "" {
		tlsConfig.ALPN = strings.Split(alpn, ",")
	}

	if insecure := params["insecure"]; insecure == "1" || insecure == "true" {
		tlsConfig.Insecure = true
	}

	// 解析 fingerprint (uTLS)
	if fp := params["fp"]; fp != "" {
		tlsConfig.UTLS = &UTLSConfig{
			Enabled:     true,
			Fingerprint: fp,
		}
	}

	config.TLS = tlsConfig

	// 解析 Reality 配置
	security := params["security"]
	if security == "reality" || params["pbk"] != "" {
		shortID := params["sid"]
		if shortID == "None" || shortID == "null" {
			shortID = ""
		}

		config.Reality = &RealityConfig{
			Enabled:   true,
			PublicKey: params["pbk"],
			ShortID:   shortID,
		}
	}

	// 转换为 JSON 字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("anytls 配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "anytls",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现 NodeConfig 接口
func (c *AnyTLSConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
