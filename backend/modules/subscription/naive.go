package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// NaiveConfig Naive 配置结构
type NaiveConfig struct {
	Username string     `json:"username"`
	Password string     `json:"password"`
	Protocol string     `json:"protocol"` // https, quic
	TLS      *TLSConfig `json:"tls,omitempty"`
}

// ParseNaiveURL 解析 Naive 链接
// 格式: naive+https://username:password@server:port?params#name
// 或: naive+quic://username:password@server:port?params#name
func ParseNaiveURL(naiveURL string) (*ProxyNode, error) {
	// 检测协议类型
	protocol := "https"
	urlStr := ""

	if strings.HasPrefix(naiveURL, "naive+https://") {
		protocol = "https"
		urlStr = strings.TrimPrefix(naiveURL, "naive+https://")
	} else if strings.HasPrefix(naiveURL, "naive+quic://") {
		protocol = "quic"
		urlStr = strings.TrimPrefix(naiveURL, "naive+quic://")
	} else if strings.HasPrefix(naiveURL, "naive://") {
		protocol = "https"
		urlStr = strings.TrimPrefix(naiveURL, "naive://")
	} else {
		return nil, errors.New("不支持的 Naive 协议格式")
	}

	// 解析 URL
	u, err := url.Parse("https://" + urlStr)
	if err != nil {
		return nil, errors.New("naive URL 格式错误")
	}

	// 提取服务器和端口
	server := u.Hostname()
	port := ParseInt(u.Port(), 443)

	// 提取节点名称
	name := u.Fragment
	if name == "" {
		name = "NaiveProxy"
	}
	name, _ = url.QueryUnescape(name)

	// 提取用户名和密码
	username := u.User.Username()
	password, _ := u.User.Password()

	// 解析查询参数
	query := u.Query()

	// 构建配置
	config := NaiveConfig{
		Username: username,
		Password: password,
		Protocol: protocol,
	}

	// 解析 TLS 配置
	tlsConfig := &TLSConfig{
		Enabled:    true,
		ServerName: query.Get("sni"),
	}

	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = server
	}

	if insecure := query.Get("insecure"); insecure == "1" || insecure == "true" {
		tlsConfig.Insecure = true
	}

	config.TLS = tlsConfig

	// 转换为 JSON 字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("naive 配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "naive",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现 NodeConfig 接口
func (c *NaiveConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
