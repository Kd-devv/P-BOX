package subscription

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// TrojanConfig Trojan配置结构
type TrojanConfig struct {
	Password   string                 `json:"password"`
	Network    string                 `json:"network,omitempty"`
	TLS        *TLSConfig             `json:"tls,omitempty"`
	Transport  map[string]interface{} `json:"transport,omitempty"`
	ServerName string                 `json:"server_name,omitempty"`
}

// ParseTrojanURL 解析Trojan链接
// 格式: trojan://password@server:port?params#name
func ParseTrojanURL(trojanURL string) (*ProxyNode, error) {
	if !strings.HasPrefix(trojanURL, "trojan://") {
		return nil, errors.New("无效的Trojan URL")
	}

	// 移除协议前缀
	urlStr := strings.TrimPrefix(trojanURL, "trojan://")

	// 分离密码和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("trojan URL格式错误")
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
	name := "Trojan"
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
		return nil, errors.New("trojan服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)
	
	// 调试日志：打印所有解析的参数
	fmt.Printf("🔍 [Trojan解析] URL: %s\n", trojanURL)
	fmt.Printf("🔍 [Trojan解析] 密码: %s\n", password)
	fmt.Printf("🔍 [Trojan解析] 服务器: %s:%d\n", server, port)
	fmt.Printf("🔍 [Trojan解析] 查询参数:\n")
	for k, v := range params {
		fmt.Printf("    %s = %s\n", k, v)
	}

	// 构建配置
	// 🆕 将 "raw" 视为 "tcp"（无传输层）
	// 某些客户端使用 type=raw 表示纯 TCP 连接
	networkType := params["type"]
	if networkType == "raw" || networkType == "none" {
		networkType = "tcp"
	}

	config := TrojanConfig{
		Password: password,
		Network:  networkType,
	}

	// 解析传输层配置
	if network := networkType; network != "" && network != "tcp" {
		transport := make(map[string]interface{})
		transport["type"] = network

		switch network {
		case "ws", "websocket":
			// WebSocket 配置直接放在 transport 层级，不需要 ws_options 嵌套
			if path := params["path"]; path != "" {
				// URL解码路径 (例如: %2F%3Fed%3D2048 -> /?ed=2048)
				decodedPath, err := url.QueryUnescape(path)
				if err == nil {
					transport["path"] = decodedPath
				} else {
					transport["path"] = path
				}
			}
			if host := params["host"]; host != "" {
				headers := map[string]interface{}{"Host": host}
				transport["headers"] = headers
			}
			// 添加 early_data_header_name 支持
			if ed := params["ed"]; ed != "" {
				transport["max_early_data"] = ParseInt(ed, 0)
				transport["early_data_header_name"] = "Sec-WebSocket-Protocol"
			}

		case "grpc":
			grpcOpts := make(map[string]interface{})
			if serviceName := params["serviceName"]; serviceName != "" {
				grpcOpts["service_name"] = serviceName
			}
			transport["grpc_options"] = grpcOpts

		case "http", "h2":
			httpOpts := make(map[string]interface{})
			if host := params["host"]; host != "" {
				httpOpts["host"] = []string{host}
			}
			if path := params["path"]; path != "" {
				decodedPath, err := url.QueryUnescape(path)
				if err == nil {
					httpOpts["path"] = decodedPath
				} else {
					httpOpts["path"] = path
				}
			}
			transport["http_options"] = httpOpts
		}

		config.Transport = transport
	}

	// Trojan默认启用TLS
	tlsConfig := &TLSConfig{
		Enabled:    true,
		ServerName: params["sni"],
	}

	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = params["peer"]
	}
	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = params["host"]
	}

	if alpn := params["alpn"]; alpn != "" {
		tlsConfig.ALPN = strings.Split(alpn, ",")
	}

	// 处理浏览器指纹伪装 (fp=chrome)
	if fp := params["fp"]; fp != "" {
		// sing-box 使用 utls 进行浏览器指纹伪装
		tlsConfig.UTLS = &UTLSConfig{
			Enabled:     true,
			Fingerprint: fp,
		}
		// 同时保留 fingerprint 字段以兼容
		tlsConfig.Fingerprint = fp
	}

	if allowInsecure := params["allowInsecure"]; allowInsecure == "1" || allowInsecure == "true" {
		tlsConfig.Insecure = true
	}
	
	// 如果没有设置 allowInsecure，默认跳过证书验证（兼容性）
	if params["allowInsecure"] == "" {
		tlsConfig.Insecure = false // 默认验证证书
	}

	config.TLS = tlsConfig
	config.ServerName = tlsConfig.ServerName

	// 转换为JSON字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("trojan配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "trojan",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	// 调试日志
	fmt.Printf("🔍 [Trojan解析] 节点名称: %s\n", node.Name)
	fmt.Printf("🔍 [Trojan解析] 服务器: %s:%d\n", node.Server, node.ServerPort)
	fmt.Printf("🔍 [Trojan解析] 网络类型: %s\n", config.Network)
	fmt.Printf("🔍 [Trojan解析] TLS SNI: %s\n", config.ServerName)
	if config.TLS != nil {
		fmt.Printf("🔍 [Trojan解析] TLS Fingerprint: %s\n", config.TLS.Fingerprint)
	}
	if config.Transport != nil {
		fmt.Printf("🔍 [Trojan解析] Transport: %+v\n", config.Transport)
	}
	fmt.Printf("🔍 [Trojan解析] 完整配置: %s\n", configJSON)

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *TrojanConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
