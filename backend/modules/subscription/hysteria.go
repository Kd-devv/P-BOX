package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// HysteriaConfig Hysteria v1 配置结构
type HysteriaConfig struct {
	Password       string     `json:"password,omitempty"`
	AuthStr        string     `json:"auth_str,omitempty"`
	Obfs           string     `json:"obfs,omitempty"`
	TLS            *TLSConfig `json:"tls,omitempty"`
	ServerName     string     `json:"server_name,omitempty"`
	UpMbps         int        `json:"up_mbps,omitempty"`
	DownMbps       int        `json:"down_mbps,omitempty"`
	HopInterval    string     `json:"hop_interval,omitempty"`
	Ports          string     `json:"ports,omitempty"`
	RecvWindowConn int        `json:"recv_window_conn,omitempty"`
	RecvWindow     int        `json:"recv_window,omitempty"`
	DisableMTU     bool       `json:"disable_mtu_discovery,omitempty"`
}

// ParseHysteriaURL 解析 Hysteria v1 链接
// 格式: hysteria://server:port?params#name
func ParseHysteriaURL(hysteriaURL string) (*ProxyNode, error) {
	// 移除协议前缀
	urlStr := strings.TrimPrefix(hysteriaURL, "hysteria://")

	// 解析 URL
	u, err := url.Parse("hysteria://" + urlStr)
	if err != nil {
		return nil, errors.New("hysteria URL 格式错误")
	}

	// 提取服务器和端口
	server := u.Hostname()
	port := ParseInt(u.Port(), 443)

	// 提取节点名称
	name := u.Fragment
	if name == "" {
		name = "Hysteria"
	}
	name, _ = url.QueryUnescape(name)

	// 解析查询参数
	query := u.Query()

	// 构建配置
	config := HysteriaConfig{
		AuthStr:        query.Get("auth"),
		Password:       query.Get("password"),
		Obfs:           query.Get("obfs"),
		UpMbps:         ParseInt(query.Get("upmbps"), 10),
		DownMbps:       ParseInt(query.Get("downmbps"), 50),
		HopInterval:    query.Get("hop_interval"),
		Ports:          query.Get("ports"),
		RecvWindowConn: ParseInt(query.Get("recv_window_conn"), 0),
		RecvWindow:     ParseInt(query.Get("recv_window"), 0),
		DisableMTU:     ParseBool(query.Get("disable_mtu_discovery"), false),
	}

	// 解析 TLS 配置
	tlsConfig := &TLSConfig{
		Enabled:    true,
		ServerName: query.Get("peer"),
	}

	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = query.Get("sni")
	}

	if alpn := query.Get("alpn"); alpn != "" {
		tlsConfig.ALPN = strings.Split(alpn, ",")
	}

	if insecure := query.Get("insecure"); insecure == "1" || insecure == "true" {
		tlsConfig.Insecure = true
	}

	config.TLS = tlsConfig
	config.ServerName = tlsConfig.ServerName

	// 转换为 JSON 字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("hysteria 配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "hysteria",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现 NodeConfig 接口
func (c *HysteriaConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
