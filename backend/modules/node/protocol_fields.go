package node

// FieldDefinition 字段定义
type FieldDefinition struct {
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Type         string      `json:"type"` // text, number, password, select, boolean, textarea
	Required     bool        `json:"required"`
	Default      interface{} `json:"default,omitempty"`
	Placeholder  string      `json:"placeholder,omitempty"`
	Description  string      `json:"description,omitempty"`
	Options      []Option    `json:"options,omitempty"`
	Min          int         `json:"min,omitempty"`
	Max          int         `json:"max,omitempty"`
	DependsOn    string      `json:"depends_on,omitempty"`
	DependsValue interface{} `json:"depends_value,omitempty"`
}

// Option 选项
type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// GetProtocolFieldDefinitions 获取协议字段定义
func GetProtocolFieldDefinitions(protocol string) []FieldDefinition {
	switch protocol {
	case "vmess":
		return getVMessFields()
	case "vless":
		return getVLESSFields()
	case "trojan":
		return getTrojanFields()
	case "shadowsocks", "ss":
		return getShadowsocksFields()
	case "socks", "socks5":
		return getSOCKS5Fields()
	case "hysteria":
		return getHysteriaFields()
	case "hysteria2", "hy2":
		return getHysteria2Fields()
	case "tuic":
		return getTUICFields()
	case "wireguard", "wg":
		return getWireGuardFields()
	case "ssh":
		return getSSHFields()
	default:
		return nil
	}
}

// getVMessFields VMess 字段定义
func getVMessFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "uuid", Label: "UUID", Type: "text", Required: true, Placeholder: "00000000-0000-0000-0000-000000000000"},
		{Name: "alter_id", Label: "Alter ID", Type: "number", Required: false, Default: 0, Min: 0, Max: 255},
		{Name: "security", Label: "加密方式", Type: "select", Required: true, Default: "auto", Options: []Option{
			{Label: "auto", Value: "auto"},
			{Label: "aes-128-gcm", Value: "aes-128-gcm"},
			{Label: "chacha20-poly1305", Value: "chacha20-poly1305"},
			{Label: "none", Value: "none"},
		}},
		{Name: "network", Label: "传输协议", Type: "select", Required: false, Default: "tcp", Options: []Option{
			{Label: "TCP", Value: "tcp"},
			{Label: "WebSocket", Value: "ws"},
			{Label: "HTTP/2", Value: "http"},
			{Label: "gRPC", Value: "grpc"},
		}},
		{Name: "tls_enabled", Label: "启用 TLS", Type: "boolean", Required: false, Default: false},
		{Name: "tls_server_name", Label: "TLS Server Name (SNI)", Type: "text", Required: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "ws_path", Label: "WebSocket 路径", Type: "text", Required: false, Placeholder: "/", DependsOn: "network", DependsValue: "ws"},
		{Name: "ws_host", Label: "WebSocket Host", Type: "text", Required: false, DependsOn: "network", DependsValue: "ws"},
		{Name: "grpc_service_name", Label: "gRPC Service Name", Type: "text", Required: false, DependsOn: "network", DependsValue: "grpc"},
	}
}

// getVLESSFields VLESS 字段定义
func getVLESSFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "uuid", Label: "UUID", Type: "text", Required: true, Placeholder: "00000000-0000-0000-0000-000000000000"},
		{Name: "flow", Label: "Flow", Type: "select", Required: false, Options: []Option{
			{Label: "无", Value: ""},
			{Label: "xtls-rprx-vision", Value: "xtls-rprx-vision"},
			{Label: "xtls-rprx-vision-udp443", Value: "xtls-rprx-vision-udp443"},
		}},
		{Name: "network", Label: "传输协议", Type: "select", Required: false, Default: "tcp", Options: []Option{
			{Label: "TCP", Value: "tcp"},
			{Label: "WebSocket", Value: "ws"},
			{Label: "HTTP/2", Value: "http"},
			{Label: "gRPC", Value: "grpc"},
		}},
		{Name: "ws_path", Label: "WebSocket 路径", Type: "text", Required: false, Placeholder: "/", DependsOn: "network", DependsValue: "ws"},
		{Name: "ws_host", Label: "WebSocket Host", Type: "text", Required: false, DependsOn: "network", DependsValue: "ws"},
		{Name: "grpc_service_name", Label: "gRPC Service Name", Type: "text", Required: false, DependsOn: "network", DependsValue: "grpc"},
		{Name: "tls_enabled", Label: "启用 TLS", Type: "boolean", Required: false, Default: true},
		{Name: "tls_server_name", Label: "TLS Server Name (SNI)", Type: "text", Required: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "alpn", Label: "ALPN", Type: "text", Required: false, Placeholder: "h2,http/1.1", DependsOn: "tls_enabled", DependsValue: true},
		{Name: "fingerprint", Label: "uTLS 指纹", Type: "select", Required: false, DependsOn: "tls_enabled", DependsValue: true, Options: []Option{
			{Label: "无", Value: ""},
			{Label: "Chrome", Value: "chrome"},
			{Label: "Firefox", Value: "firefox"},
			{Label: "Safari", Value: "safari"},
			{Label: "iOS", Value: "ios"},
			{Label: "Edge", Value: "edge"},
			{Label: "Random", Value: "random"},
		}},
		{Name: "reality_enabled", Label: "启用 REALITY", Type: "boolean", Required: false, Default: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "reality_public_key", Label: "REALITY Public Key", Type: "text", Required: false, DependsOn: "reality_enabled", DependsValue: true},
		{Name: "reality_short_id", Label: "REALITY Short ID", Type: "text", Required: false, DependsOn: "reality_enabled", DependsValue: true},
	}
}

// getTrojanFields Trojan 字段定义
func getTrojanFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "password", Label: "密码", Type: "password", Required: true},
		{Name: "network", Label: "传输协议", Type: "select", Required: false, Default: "tcp", Options: []Option{
			{Label: "TCP", Value: "tcp"},
			{Label: "WebSocket", Value: "ws"},
			{Label: "gRPC", Value: "grpc"},
		}},
		{Name: "tls_enabled", Label: "启用 TLS", Type: "boolean", Required: false, Default: true},
		{Name: "tls_server_name", Label: "TLS Server Name", Type: "text", Required: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "fingerprint", Label: "uTLS 指纹", Type: "select", Required: false, DependsOn: "tls_enabled", DependsValue: true, Options: []Option{
			{Label: "无", Value: ""},
			{Label: "Chrome", Value: "chrome"},
			{Label: "Firefox", Value: "firefox"},
			{Label: "Safari", Value: "safari"},
		}},
		{Name: "ws_path", Label: "WebSocket 路径", Type: "text", Required: false, Placeholder: "/", DependsOn: "network", DependsValue: "ws"},
		{Name: "grpc_service_name", Label: "gRPC Service Name", Type: "text", Required: false, DependsOn: "network", DependsValue: "grpc"},
	}
}

// getShadowsocksFields Shadowsocks 字段定义
func getShadowsocksFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "method", Label: "加密方式", Type: "select", Required: true, Default: "aes-256-gcm", Options: []Option{
			{Label: "aes-128-gcm", Value: "aes-128-gcm"},
			{Label: "aes-192-gcm", Value: "aes-192-gcm"},
			{Label: "aes-256-gcm", Value: "aes-256-gcm"},
			{Label: "chacha20-ietf-poly1305", Value: "chacha20-ietf-poly1305"},
			{Label: "xchacha20-ietf-poly1305", Value: "xchacha20-ietf-poly1305"},
			{Label: "2022-blake3-aes-128-gcm", Value: "2022-blake3-aes-128-gcm"},
			{Label: "2022-blake3-aes-256-gcm", Value: "2022-blake3-aes-256-gcm"},
			{Label: "2022-blake3-chacha20-poly1305", Value: "2022-blake3-chacha20-poly1305"},
		}},
		{Name: "password", Label: "密码", Type: "password", Required: true},
		{Name: "plugin", Label: "插件", Type: "select", Required: false, Options: []Option{
			{Label: "无", Value: ""},
			{Label: "obfs-local", Value: "obfs-local"},
			{Label: "v2ray-plugin", Value: "v2ray-plugin"},
		}},
		{Name: "plugin_opts", Label: "插件选项", Type: "text", Required: false, Placeholder: "obfs=http;obfs-host=www.bing.com"},
	}
}

// getSOCKS5Fields SOCKS5 字段定义
func getSOCKS5Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "version", Label: "SOCKS 版本", Type: "select", Required: false, Default: "5", Options: []Option{
			{Label: "SOCKS4", Value: "4"},
			{Label: "SOCKS4A", Value: "4a"},
			{Label: "SOCKS5", Value: "5"},
		}, Description: "SOCKS 协议版本，默认为 SOCKS5"},
		{Name: "username", Label: "用户名", Type: "text", Required: false, Placeholder: "留空表示无需认证", Description: "认证用户名"},
		{Name: "password", Label: "密码", Type: "password", Required: false, Placeholder: "留空表示无需认证", Description: "认证密码"},
		{Name: "udp_over_tcp", Label: "UDP over TCP", Type: "boolean", Required: false, Default: false, Description: "启用 UDP over TCP 传输"},
		{Name: "tls_enabled", Label: "启用 TLS", Type: "boolean", Required: false, Default: false, Description: "为 SOCKS 连接启用 TLS 加密"},
		{Name: "tls_server_name", Label: "TLS Server Name", Type: "text", Required: false, DependsOn: "tls_enabled", DependsValue: true},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false, DependsOn: "tls_enabled", DependsValue: true},
	}
}

// getHysteriaFields Hysteria v1 字段定义
func getHysteriaFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "auth_str", Label: "认证字符串", Type: "password", Required: false},
		{Name: "obfs", Label: "混淆密码", Type: "password", Required: false},
		{Name: "up_mbps", Label: "上传速度 (Mbps)", Type: "number", Required: false, Default: 10, Min: 1},
		{Name: "down_mbps", Label: "下载速度 (Mbps)", Type: "number", Required: false, Default: 50, Min: 1},
		{Name: "tls_server_name", Label: "TLS Server Name", Type: "text", Required: false},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false},
		{Name: "alpn", Label: "ALPN", Type: "text", Required: false, Placeholder: "h3"},
	}
}

// getHysteria2Fields Hysteria2 字段定义
func getHysteria2Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "password", Label: "密码", Type: "password", Required: true},
		{Name: "obfs_type", Label: "混淆类型", Type: "select", Required: false, Options: []Option{
			{Label: "无", Value: ""},
			{Label: "salamander", Value: "salamander"},
		}},
		{Name: "obfs_password", Label: "混淆密码", Type: "password", Required: false, DependsOn: "obfs_type", DependsValue: "salamander"},
		{Name: "up_mbps", Label: "上传速度 (Mbps)", Type: "number", Required: false, Min: 0, Description: "0 表示不限速"},
		{Name: "down_mbps", Label: "下载速度 (Mbps)", Type: "number", Required: false, Min: 0, Description: "0 表示不限速"},
		{Name: "tls_server_name", Label: "TLS Server Name", Type: "text", Required: false},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false},
		{Name: "alpn", Label: "ALPN", Type: "text", Required: false, Placeholder: "h3"},
	}
}

// getTUICFields TUIC 字段定义
func getTUICFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "uuid", Label: "UUID", Type: "text", Required: true, Placeholder: "00000000-0000-0000-0000-000000000000"},
		{Name: "password", Label: "密码", Type: "password", Required: true},
		{Name: "congestion_control", Label: "拥塞控制", Type: "select", Required: false, Default: "cubic", Options: []Option{
			{Label: "BBR", Value: "bbr"},
			{Label: "CUBIC", Value: "cubic"},
			{Label: "New Reno", Value: "new_reno"},
		}},
		{Name: "udp_relay_mode", Label: "UDP 中继模式", Type: "select", Required: false, Default: "native", Options: []Option{
			{Label: "Native", Value: "native"},
			{Label: "QUIC", Value: "quic"},
		}},
		{Name: "zero_rtt_handshake", Label: "启用 0-RTT", Type: "boolean", Required: false, Default: false},
		{Name: "tls_server_name", Label: "TLS Server Name", Type: "text", Required: false},
		{Name: "tls_insecure", Label: "跳过证书验证", Type: "boolean", Required: false, Default: false},
		{Name: "alpn", Label: "ALPN", Type: "text", Required: false, Placeholder: "h3"},
	}
}

// getWireGuardFields WireGuard 字段定义
func getWireGuardFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "private_key", Label: "Private Key", Type: "password", Required: true, Description: "Base64 编码的私钥"},
		{Name: "peer_public_key", Label: "Peer Public Key", Type: "text", Required: true, Description: "Base64 编码的对端公钥"},
		{Name: "pre_shared_key", Label: "Pre-shared Key", Type: "password", Required: false, Description: "Base64 编码的预共享密钥"},
		{Name: "local_address", Label: "本地地址", Type: "text", Required: true, Placeholder: "10.0.0.2/32", Description: "IPv4/IPv6 地址，用逗号分隔"},
		{Name: "mtu", Label: "MTU", Type: "number", Required: false, Default: 1420, Min: 1280, Max: 1500},
		{Name: "reserved", Label: "Reserved", Type: "text", Required: false, Placeholder: "0,0,0", Description: "三个十进制数字"},
	}
}

// getSSHFields SSH 字段定义
func getSSHFields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "user", Label: "用户名", Type: "text", Required: true, Default: "root"},
		{Name: "password", Label: "密码", Type: "password", Required: false, Description: "密码或私钥至少填一个"},
		{Name: "private_key", Label: "私钥", Type: "textarea", Required: false, Placeholder: "-----BEGIN OPENSSH PRIVATE KEY-----\n..."},
		{Name: "private_key_passphrase", Label: "私钥密码", Type: "password", Required: false, DependsOn: "private_key"},
	}
}

// GetSupportedProtocols 获取支持的协议列表
func GetSupportedProtocols() []Option {
	return []Option{
		{Label: "VMess", Value: "vmess"},
		{Label: "VLESS", Value: "vless"},
		{Label: "Trojan", Value: "trojan"},
		{Label: "Shadowsocks", Value: "shadowsocks"},
		{Label: "SOCKS5", Value: "socks"},
		{Label: "Hysteria", Value: "hysteria"},
		{Label: "Hysteria2", Value: "hysteria2"},
		{Label: "TUIC", Value: "tuic"},
		{Label: "WireGuard", Value: "wireguard"},
		{Label: "SSH", Value: "ssh"},
	}
}
