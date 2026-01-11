package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// TUICConfig TUIC配置结构
type TUICConfig struct {
	UUID              string     `json:"uuid"`
	Password          string     `json:"password"`
	CongestionControl string     `json:"congestion_control,omitempty"`
	UDPRelayMode      string     `json:"udp_relay_mode,omitempty"`
	ZeroRTTHandshake  bool       `json:"zero_rtt_handshake,omitempty"`
	TLS               *TLSConfig `json:"tls,omitempty"`
	ServerName        string     `json:"server_name,omitempty"`
}

// ParseTUICURL 解析TUIC链接
// 格式: tuic://uuid:password@server:port?params#name
func ParseTUICURL(tuicURL string) (*ProxyNode, error) {
	if !strings.HasPrefix(tuicURL, "tuic://") {
		return nil, errors.New("无效的TUIC URL")
	}

	// 移除协议前缀
	urlStr := strings.TrimPrefix(tuicURL, "tuic://")

	// 分离UUID:密码和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("tuic URL格式错误")
	}

	// 解析UUID和密码
	uuidPassword := strings.SplitN(parts[0], ":", 2)
	if len(uuidPassword) != 2 {
		return nil, errors.New("tuic UUID或密码格式错误")
	}
	uuid, _ := url.QueryUnescape(uuidPassword[0])
	password, _ := url.QueryUnescape(uuidPassword[1])

	remaining := parts[1]

	// 分离服务器和参数
	serverAndParams := strings.SplitN(remaining, "?", 2)
	serverPart := serverAndParams[0]
	queryString := ""
	if len(serverAndParams) == 2 {
		queryString = serverAndParams[1]
	}

	// 提取节点名称
	name := "TUIC"
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
		return nil, errors.New("tuic服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)

	// 构建配置
	config := TUICConfig{
		UUID:              uuid,
		Password:          password,
		CongestionControl: params["congestion_control"],
		UDPRelayMode:      params["udp_relay_mode"],
		ZeroRTTHandshake:  ParseBool(params["zero_rtt_handshake"], false),
	}

	// 默认值
	if config.CongestionControl == "" {
		config.CongestionControl = "cubic"
	}
	if config.UDPRelayMode == "" {
		config.UDPRelayMode = "native"
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
		return nil, errors.New("tuic配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "tuic",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *TUICConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
