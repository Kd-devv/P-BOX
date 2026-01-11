package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// Hysteria2Config Hysteria2配置结构
type Hysteria2Config struct {
	Password   string         `json:"password"`
	Obfs       *Hysteria2Obfs `json:"obfs,omitempty"`
	TLS        *TLSConfig     `json:"tls,omitempty"`
	ServerName string         `json:"server_name,omitempty"`
	UpMbps     int            `json:"up_mbps,omitempty"`
	DownMbps   int            `json:"down_mbps,omitempty"`
}

// Hysteria2Obfs Hysteria2混淆配置
type Hysteria2Obfs struct {
	Type     string `json:"type"`
	Password string `json:"password,omitempty"`
}

// ParseHysteria2URL 解析Hysteria2链接
// 格式: hysteria2://password@server:port?params#name
// 或: hy2://password@server:port?params#name
func ParseHysteria2URL(hy2URL string) (*ProxyNode, error) {
	// 移除协议前缀
	urlStr := strings.TrimPrefix(hy2URL, "hysteria2://")
	urlStr = strings.TrimPrefix(urlStr, "hy2://")

	// 分离密码和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("hysteria2 URL格式错误")
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
	name := "Hysteria2"
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
		return nil, errors.New("hysteria2服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)

	// 构建配置
	config := Hysteria2Config{
		Password: password,
		UpMbps:   ParseInt(params["up"], 0),
		DownMbps: ParseInt(params["down"], 0),
	}

	// 解析混淆
	if obfsType := params["obfs"]; obfsType != "" {
		config.Obfs = &Hysteria2Obfs{
			Type:     obfsType,
			Password: params["obfs-password"],
		}
	}

	// 解析TLS配置
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

	config.TLS = tlsConfig
	config.ServerName = tlsConfig.ServerName

	// 转换为JSON字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("hysteria2配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "hysteria2",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *Hysteria2Config) ToJSON() (string, error) {
	return ToJSONString(c)
}
