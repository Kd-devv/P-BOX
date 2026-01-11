package subscription

import (
	"encoding/json"
	"errors"
	"strings"
)

// VMessConfig VMess配置结构
type VMessConfig struct {
	UUID                string                 `json:"uuid"`
	AlterID             int                    `json:"alter_id"`
	Security            string                 `json:"security"`
	GlobalPadding       bool                   `json:"global_padding"`
	AuthenticatedLength bool                   `json:"authenticated_length"`
	Network             string                 `json:"network,omitempty"`
	PacketEncoding      string                 `json:"packet_encoding,omitempty"`
	TLS                 *TLSConfig             `json:"tls,omitempty"`
	Transport           map[string]interface{} `json:"transport,omitempty"`
	ServerName          string                 `json:"server_name,omitempty"`
}

// TLSConfig 和 UTLSConfig 已在 base.go 中定义

// VMessJSON VMess分享链接的JSON结构
type VMessJSON struct {
	V    string      `json:"v"`
	PS   string      `json:"ps"`
	Add  string      `json:"add"`
	Port interface{} `json:"port"`
	ID   string      `json:"id"`
	Aid  interface{} `json:"aid"`
	Net  string      `json:"net"`
	Type string      `json:"type"`
	Host string      `json:"host"`
	Path string      `json:"path"`
	TLS  string      `json:"tls"`
	SNI  string      `json:"sni"`
	ALPN string      `json:"alpn"`
	FP   string      `json:"fp"`
}

// ParseVMessURL 解析VMess链接
func ParseVMessURL(vmessURL string) (*ProxyNode, error) {
	if !strings.HasPrefix(vmessURL, "vmess://") {
		return nil, errors.New("无效的VMess URL")
	}

	// 移除前缀并解码Base64
	base64Str := strings.TrimPrefix(vmessURL, "vmess://")
	jsonStr, err := DecodeBase64(base64Str)
	if err != nil {
		return nil, errors.New("vmess Base64解码失败")
	}

	// 解析JSON
	var vmessJSON VMessJSON
	if err := json.Unmarshal([]byte(jsonStr), &vmessJSON); err != nil {
		return nil, errors.New("vmess JSON解析失败")
	}

	// 提取端口
	port := 443
	switch v := vmessJSON.Port.(type) {
	case string:
		port = ParseInt(v, 443)
	case float64:
		port = int(v)
	case int:
		port = v
	}

	// 提取AlterID
	alterID := 0
	switch v := vmessJSON.Aid.(type) {
	case string:
		alterID = ParseInt(v, 0)
	case float64:
		alterID = int(v)
	case int:
		alterID = v
	}

	// 🆕 将 "raw" 视为 "tcp"（无传输层）
	// 某些客户端使用 net=raw 表示纯 TCP 连接
	networkType := vmessJSON.Net
	if networkType == "raw" || networkType == "none" {
		networkType = "tcp"
	}

	// 构建配置
	config := VMessConfig{
		UUID:                vmessJSON.ID,
		AlterID:             alterID,
		Security:            "auto",
		GlobalPadding:       false,
		AuthenticatedLength: true,
		Network:             networkType,
		PacketEncoding:      "",
	}

	// 解析传输层配置
	if networkType != "" && networkType != "tcp" {
		transport := make(map[string]interface{})
		transport["type"] = networkType

		switch networkType {
		case "ws":
			// WebSocket 配置直接放在 transport 层级
			if vmessJSON.Path != "" {
				transport["path"] = vmessJSON.Path
			}
			if vmessJSON.Host != "" {
				headers := map[string]interface{}{"Host": vmessJSON.Host}
				transport["headers"] = headers
			}

		case "grpc":
			grpcOpts := make(map[string]interface{})
			if vmessJSON.Path != "" {
				grpcOpts["service_name"] = vmessJSON.Path
			}
			transport["grpc_options"] = grpcOpts

		case "http", "h2":
			httpOpts := make(map[string]interface{})
			if vmessJSON.Host != "" {
				httpOpts["host"] = []string{vmessJSON.Host}
			}
			if vmessJSON.Path != "" {
				httpOpts["path"] = vmessJSON.Path
			}
			transport["http_options"] = httpOpts

		case "quic":
			quicOpts := make(map[string]interface{})
			transport["quic_options"] = quicOpts
		}

		config.Transport = transport
	}

	// 解析TLS配置
	if vmessJSON.TLS == "tls" {
		tlsConfig := &TLSConfig{
			Enabled:    true,
			ServerName: vmessJSON.SNI,
		}

		if vmessJSON.SNI == "" && vmessJSON.Host != "" {
			tlsConfig.ServerName = vmessJSON.Host
		}

		if vmessJSON.ALPN != "" {
			tlsConfig.ALPN = strings.Split(vmessJSON.ALPN, ",")
		}

		if vmessJSON.FP != "" {
			tlsConfig.Fingerprint = vmessJSON.FP
		}

		config.TLS = tlsConfig
		config.ServerName = tlsConfig.ServerName
	}

	// 转换为JSON字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("vmess配置序列化失败")
	}

	node := &ProxyNode{
		Name:       vmessJSON.PS,
		Type:       "vmess",
		Server:     vmessJSON.Add,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *VMessConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
