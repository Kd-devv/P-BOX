package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// ShadowsocksConfig Shadowsocks配置结构
// 注意：字段名需要与 Mihomo 配置格式一致
type ShadowsocksConfig struct {
	Cipher     string `json:"cipher"` // Mihomo 使用 cipher 而不是 method
	Password   string `json:"password"`
	Plugin     string `json:"plugin,omitempty"`
	PluginOpts string `json:"plugin-opts,omitempty"` // Mihomo 使用连字符
	UDP        bool   `json:"udp,omitempty"`
}

// ParseShadowsocksURL 解析Shadowsocks链接
// 格式: ss://base64(method:password)@server:port?plugin=xxx#name
// 或: ss://base64(method:password@server:port)#name
func ParseShadowsocksURL(ssURL string) (*ProxyNode, error) {
	if !strings.HasPrefix(ssURL, "ss://") {
		return nil, errors.New("无效的Shadowsocks URL")
	}

	// 移除协议前缀
	urlStr := strings.TrimPrefix(ssURL, "ss://")

	// 提取节点名称
	name := "Shadowsocks"
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		name, _ = url.QueryUnescape(urlStr[idx+1:])
		urlStr = urlStr[:idx]
	}

	// 分离主要部分和查询参数
	mainPart := urlStr
	queryString := ""
	if idx := strings.Index(urlStr, "?"); idx != -1 {
		mainPart = urlStr[:idx]
		queryString = urlStr[idx+1:]
	}

	var method, password, server string
	var port int

	// 尝试新格式: method:password@server:port (Base64编码)
	if strings.Contains(mainPart, "@") {
		parts := strings.SplitN(mainPart, "@", 2)

		// 解码第一部分
		decoded, err := DecodeBase64(parts[0])
		if err != nil {
			return nil, errors.New("shadowsocks Base64解码失败")
		}

		// 解析 method:password
		methodPassword := strings.SplitN(decoded, ":", 2)
		if len(methodPassword) != 2 {
			return nil, errors.New("shadowsocks格式错误")
		}
		method = methodPassword[0]
		password = methodPassword[1]

		// 解析服务器和端口（移除可能的路径分隔符）
		serverPort := strings.TrimSuffix(parts[1], "/")
		serverParts := strings.Split(serverPort, ":")
		if len(serverParts) != 2 {
			return nil, errors.New("shadowsocks服务器地址格式错误")
		}
		server = serverParts[0]
		port = ParseInt(serverParts[1], 8388)

	} else {
		// 尝试旧格式: 整个URL都是Base64编码
		decoded, err := DecodeBase64(mainPart)
		if err != nil {
			return nil, errors.New("shadowsocks Base64解码失败")
		}

		// 解析 method:password@server:port
		parts := strings.SplitN(decoded, "@", 2)
		if len(parts) != 2 {
			return nil, errors.New("shadowsocks格式错误")
		}

		// 解析 method:password
		methodPassword := strings.SplitN(parts[0], ":", 2)
		if len(methodPassword) != 2 {
			return nil, errors.New("shadowsocks格式错误")
		}
		method = methodPassword[0]
		password = methodPassword[1]

		// 解析服务器和端口
		serverParts := strings.Split(parts[1], ":")
		if len(serverParts) != 2 {
			return nil, errors.New("shadowsocks服务器地址格式错误")
		}
		server = serverParts[0]
		port = ParseInt(serverParts[1], 8388)
	}

	// 解析查询参数（插件）
	params := ParseQueryParams(queryString)

	// 构建配置 - 使用 Mihomo 要求的字段名
	config := ShadowsocksConfig{
		Cipher:     method, // SS URL 中叫 method，Mihomo 配置中叫 cipher
		Password:   password,
		Plugin:     params["plugin"],
		PluginOpts: params["plugin-opts"],
		UDP:        true, // 默认启用 UDP
	}

	// 转换为JSON字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("shadowsocks配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "ss", // Mihomo 使用 "ss" 作为类型标识
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *ShadowsocksConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
