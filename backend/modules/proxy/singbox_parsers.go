package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================================
// 节点解析器 - 将 ProxyNode 转换为 Sing-Box Outbound
// ============================================================================

// ParseNodeToSingBox 将节点转换为 sing-box outbound
func ParseNodeToSingBox(node ProxyNode) (*SBOutbound, error) {
	// 优先使用完整的 Config JSON 解析
	if node.Config != "" {
		return parseFromConfigJSON(node)
	}

	// 没有 Config JSON，使用基础字段构建
	return parseFromBasicFields(node)
}

// parseFromBasicFields 从基础字段构建 outbound
func parseFromBasicFields(node ProxyNode) (*SBOutbound, error) {
	nodeType := strings.ToLower(node.Type)

	outbound := &SBOutbound{
		Tag:        node.Name,
		Type:       nodeType,
		Server:     node.Server,
		ServerPort: node.GetPort(),
	}

	// 根据类型设置默认值
	switch nodeType {
	case "vmess":
		outbound.Security = "auto"
		outbound.PacketEncoding = "xudp"
	case "vless":
		outbound.PacketEncoding = "xudp"
	case "ss", "shadowsocks":
		outbound.Type = "shadowsocks"
	case "trojan":
		// Trojan 默认需要 TLS
		outbound.TLS = &SBTLS{
			Enabled:  true,
			Insecure: true,
		}
	case "hysteria2", "hy2":
		outbound.Type = "hysteria2"
	case "socks", "socks5":
		outbound.Type = "socks"
	case "http":
		outbound.Type = "http"
	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}

	return outbound, nil
}

// parseFromConfigJSON 从 Config JSON 解析
func parseFromConfigJSON(node ProxyNode) (*SBOutbound, error) {
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(node.Config), &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	nodeType := strings.ToLower(node.Type)

	outbound := &SBOutbound{
		Tag:  node.Name,
		Type: nodeType,
	}

	// 提取通用字段 - 优先使用 node 的基础字段，因为 Config 可能不包含 server/port
	outbound.Server = node.Server
	outbound.ServerPort = node.GetPort()

	// 如果 Config JSON 中有 server/port，则使用 Config 中的值
	if server, ok := rawConfig["server"].(string); ok && server != "" {
		outbound.Server = server
	}
	if port, ok := rawConfig["port"].(float64); ok && port > 0 {
		outbound.ServerPort = int(port)
	}

	// 根据类型提取特定字段
	switch nodeType {
	case "vmess":
		parseVMessConfig(rawConfig, outbound)
	case "vless":
		parseVLESSConfig(rawConfig, outbound)
	case "ss", "shadowsocks":
		outbound.Type = "shadowsocks"
		parseShadowsocksConfig(rawConfig, outbound)
	case "trojan":
		parseTrojanConfig(rawConfig, outbound)
	case "hysteria2", "hy2":
		outbound.Type = "hysteria2"
		parseHysteria2Config(rawConfig, outbound)
	case "tuic":
		parseTUICConfig(rawConfig, outbound)
	case "anytls":
		parseAnyTLSConfig(rawConfig, outbound)
	}

	return outbound, nil
}

// ============================================================================
// VMess 解析
// ============================================================================

func parseVMessConfig(config map[string]interface{}, out *SBOutbound) {
	// UUID
	if uuid, ok := config["uuid"].(string); ok {
		out.UUID = uuid
	}
	// Security
	if security, ok := config["cipher"].(string); ok {
		out.Security = security
	} else {
		out.Security = "auto"
	}
	// Alter ID
	if alterId, ok := config["alterId"].(float64); ok {
		out.AlterId = int(alterId)
	}
	// Packet Encoding
	out.PacketEncoding = "xudp"

	// TLS
	if tls, ok := config["tls"].(bool); ok && tls {
		out.TLS = &SBTLS{
			Enabled:  true,
			Insecure: true,
		}
		if sni, ok := config["servername"].(string); ok {
			out.TLS.ServerName = sni
		}
		if skip, ok := config["skip-cert-verify"].(bool); ok {
			out.TLS.Insecure = skip
		}
		// UTLS fingerprint
		if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
			out.TLS.UTLS = &SBUTLS{
				Enabled:     true,
				Fingerprint: fp,
			}
		}
	}

	// Transport
	parseTransportConfig(config, out)
}

// ============================================================================
// VLESS 解析
// ============================================================================

func parseVLESSConfig(config map[string]interface{}, out *SBOutbound) {
	// UUID
	if uuid, ok := config["uuid"].(string); ok {
		out.UUID = uuid
	}
	// Flow
	if flow, ok := config["flow"].(string); ok {
		out.Flow = flow
	}
	// Packet Encoding
	out.PacketEncoding = "xudp"

	// TLS
	if tls, ok := config["tls"].(bool); ok && tls {
		out.TLS = &SBTLS{
			Enabled:  true,
			Insecure: true,
		}
		if sni, ok := config["servername"].(string); ok {
			out.TLS.ServerName = sni
		}
		if skip, ok := config["skip-cert-verify"].(bool); ok {
			out.TLS.Insecure = skip
		}
		// UTLS
		if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
			out.TLS.UTLS = &SBUTLS{
				Enabled:     true,
				Fingerprint: fp,
			}
		}
		// Reality
		if realityOpts, ok := config["reality-opts"].(map[string]interface{}); ok {
			out.TLS.Reality = &SBReality{
				Enabled: true,
			}
			if pk, ok := realityOpts["public-key"].(string); ok {
				out.TLS.Reality.PublicKey = pk
			}
			if sid, ok := realityOpts["short-id"].(string); ok {
				out.TLS.Reality.ShortID = sid
			}
		}
	}

	// Transport
	parseTransportConfig(config, out)
}

// ============================================================================
// Shadowsocks 解析
// ============================================================================

func parseShadowsocksConfig(config map[string]interface{}, out *SBOutbound) {
	// Method
	if method, ok := config["cipher"].(string); ok {
		out.Method = method
	}
	// Password
	if password, ok := config["password"].(string); ok {
		out.Password = password
	}
	// Plugin
	if plugin, ok := config["plugin"].(string); ok {
		out.Plugin = plugin
	}
	if pluginOpts, ok := config["plugin-opts"].(map[string]interface{}); ok {
		opts := []string{}
		for k, v := range pluginOpts {
			opts = append(opts, fmt.Sprintf("%s=%v", k, v))
		}
		out.PluginOpts = strings.Join(opts, ";")
	}
}

// ============================================================================
// Trojan 解析
// ============================================================================

func parseTrojanConfig(config map[string]interface{}, out *SBOutbound) {
	// Password
	if password, ok := config["password"].(string); ok {
		out.Password = password
	}

	// TLS (Trojan 默认开启 TLS)
	out.TLS = &SBTLS{
		Enabled:  true,
		Insecure: true,
	}
	if sni, ok := config["sni"].(string); ok {
		out.TLS.ServerName = sni
	}
	if skip, ok := config["skip-cert-verify"].(bool); ok {
		out.TLS.Insecure = skip
	}
	if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
		out.TLS.UTLS = &SBUTLS{
			Enabled:     true,
			Fingerprint: fp,
		}
	}

	// Transport
	parseTransportConfig(config, out)
}

// ============================================================================
// Hysteria2 解析
// ============================================================================

func parseHysteria2Config(config map[string]interface{}, out *SBOutbound) {
	// Password
	if password, ok := config["password"].(string); ok {
		out.Password = password
	}
	// Up/Down Mbps
	if up, ok := config["up"].(string); ok {
		if val, err := strconv.Atoi(strings.TrimSuffix(up, " Mbps")); err == nil {
			out.UpMbps = val
		}
	}
	if down, ok := config["down"].(string); ok {
		if val, err := strconv.Atoi(strings.TrimSuffix(down, " Mbps")); err == nil {
			out.DownMbps = val
		}
	}

	// Obfs
	if obfs, ok := config["obfs"].(string); ok && obfs != "" {
		out.Obfs = &SBObfs{
			Type: obfs,
		}
		if obfsPwd, ok := config["obfs-password"].(string); ok {
			out.Obfs.Password = obfsPwd
		}
	}

	// TLS
	out.TLS = &SBTLS{
		Enabled:  true,
		Insecure: true,
	}
	if sni, ok := config["sni"].(string); ok {
		out.TLS.ServerName = sni
	}
	if skip, ok := config["skip-cert-verify"].(bool); ok {
		out.TLS.Insecure = skip
	}
}

// ============================================================================
// TUIC 解析
// ============================================================================

func parseTUICConfig(config map[string]interface{}, out *SBOutbound) {
	// UUID
	if uuid, ok := config["uuid"].(string); ok {
		out.UUID = uuid
	}
	// Password
	if password, ok := config["password"].(string); ok {
		out.Password = password
	}
	// Congestion Control
	if cc, ok := config["congestion-controller"].(string); ok {
		out.CongestionControl = cc
	}
	// UDP Relay Mode
	if udpMode, ok := config["udp-relay-mode"].(string); ok {
		out.UDPRelayMode = udpMode
	}

	// TLS
	out.TLS = &SBTLS{
		Enabled:  true,
		Insecure: true,
	}
	if sni, ok := config["sni"].(string); ok {
		out.TLS.ServerName = sni
	}
	if skip, ok := config["skip-cert-verify"].(bool); ok {
		out.TLS.Insecure = skip
	}
	if alpn, ok := config["alpn"].([]interface{}); ok {
		for _, a := range alpn {
			if s, ok := a.(string); ok {
				out.TLS.ALPN = append(out.TLS.ALPN, s)
			}
		}
	}
}

// ============================================================================
// AnyTLS 解析
// ============================================================================

func parseAnyTLSConfig(config map[string]interface{}, out *SBOutbound) {
	out.Type = "anytls"

	// Password
	if password, ok := config["password"].(string); ok {
		out.Password = password
	}

	// Session 配置
	if interval, ok := config["idle-session-check-interval"].(float64); ok {
		out.IdleSessionCheckInterval = fmt.Sprintf("%ds", int(interval))
	}
	if timeout, ok := config["idle-session-timeout"].(float64); ok {
		out.IdleSessionTimeout = fmt.Sprintf("%ds", int(timeout))
	}
	if minSession, ok := config["min-idle-session"].(float64); ok {
		out.MinIdleSession = int(minSession)
	}

	// TLS 配置
	out.TLS = &SBTLS{
		Enabled:  true,
		Insecure: true,
	}
	if sni, ok := config["sni"].(string); ok {
		out.TLS.ServerName = sni
	} else if sni, ok := config["servername"].(string); ok {
		out.TLS.ServerName = sni
	}
	if skip, ok := config["skip-cert-verify"].(bool); ok {
		out.TLS.Insecure = skip
	}
	if alpn, ok := config["alpn"].([]interface{}); ok {
		for _, a := range alpn {
			if s, ok := a.(string); ok {
				out.TLS.ALPN = append(out.TLS.ALPN, s)
			}
		}
	}
	// UTLS fingerprint
	if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
		out.TLS.UTLS = &SBUTLS{
			Enabled:     true,
			Fingerprint: fp,
		}
	}
	// Reality 配置
	if realityOpts, ok := config["reality-opts"].(map[string]interface{}); ok {
		out.TLS.Reality = &SBReality{
			Enabled: true,
		}
		if pk, ok := realityOpts["public-key"].(string); ok {
			out.TLS.Reality.PublicKey = pk
		}
		if sid, ok := realityOpts["short-id"].(string); ok {
			out.TLS.Reality.ShortID = sid
		}
	}
}

// ============================================================================
// Transport 解析 (通用)
// ============================================================================

func parseTransportConfig(config map[string]interface{}, out *SBOutbound) {
	network, _ := config["network"].(string)
	if network == "" {
		return
	}

	switch network {
	case "ws":
		out.Transport = &SBTransport{Type: "ws"}
		if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
			if path, ok := wsOpts["path"].(string); ok {
				// 处理 early data
				if strings.Contains(path, "?ed=") {
					parts := strings.SplitN(path, "?ed=", 2)
					out.Transport.Path = parts[0]
					if len(parts) > 1 {
						if ed, err := strconv.Atoi(parts[1]); err == nil {
							out.Transport.MaxEarlyData = ed
							out.Transport.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
						}
					}
				} else {
					out.Transport.Path = path
				}
			}
			if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
				out.Transport.Headers = make(map[string]string)
				for k, v := range headers {
					if s, ok := v.(string); ok {
						out.Transport.Headers[k] = s
					}
				}
			}
		}

	case "grpc":
		out.Transport = &SBTransport{Type: "grpc"}
		if grpcOpts, ok := config["grpc-opts"].(map[string]interface{}); ok {
			if sn, ok := grpcOpts["grpc-service-name"].(string); ok {
				out.Transport.ServiceName = sn
			}
		}

	case "h2", "http":
		out.Transport = &SBTransport{Type: "http"}
		if h2Opts, ok := config["h2-opts"].(map[string]interface{}); ok {
			if host, ok := h2Opts["host"].([]interface{}); ok && len(host) > 0 {
				hosts := []string{}
				for _, h := range host {
					if s, ok := h.(string); ok {
						hosts = append(hosts, s)
					}
				}
				if len(hosts) == 1 {
					out.Transport.Host = hosts[0]
				} else {
					out.Transport.Host = hosts
				}
			}
			if path, ok := h2Opts["path"].(string); ok {
				out.Transport.Path = path
			}
		}
	}
}

// ============================================================================
// URI 解析器 (用于直接解析分享链接)
// ============================================================================

// ParseShareLink 解析分享链接
func ParseShareLink(link string) (*SBOutbound, error) {
	link = strings.TrimSpace(link)

	if strings.HasPrefix(link, "vmess://") {
		return parseVMessURI(link)
	}
	if strings.HasPrefix(link, "vless://") {
		return parseVLESSURI(link)
	}
	if strings.HasPrefix(link, "ss://") {
		return parseShadowsocksURI(link)
	}
	if strings.HasPrefix(link, "trojan://") {
		return parseTrojanURI(link)
	}
	if strings.HasPrefix(link, "hysteria2://") || strings.HasPrefix(link, "hy2://") {
		return parseHysteria2URI(link)
	}
	if strings.HasPrefix(link, "tuic://") {
		return parseTUICURI(link)
	}

	return nil, fmt.Errorf("unsupported share link format")
}

// parseVMessURI 解析 vmess:// URI
func parseVMessURI(uri string) (*SBOutbound, error) {
	data := strings.TrimPrefix(uri, "vmess://")

	// Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		// 尝试 URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			// 尝试 RawStdEncoding
			decoded, err = base64.RawStdEncoding.DecodeString(data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode vmess URI: %w", err)
			}
		}
	}

	var vmessConfig map[string]interface{}
	if err := json.Unmarshal(decoded, &vmessConfig); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	name, _ := vmessConfig["ps"].(string)
	if name == "" {
		name = "VMess"
	}

	server, _ := vmessConfig["add"].(string)
	var port int
	if p, ok := vmessConfig["port"].(float64); ok {
		port = int(p)
	} else if p, ok := vmessConfig["port"].(string); ok {
		port, _ = strconv.Atoi(p)
	}

	uuid, _ := vmessConfig["id"].(string)

	var alterId int
	if aid, ok := vmessConfig["aid"].(float64); ok {
		alterId = int(aid)
	} else if aid, ok := vmessConfig["aid"].(string); ok {
		alterId, _ = strconv.Atoi(aid)
	}

	security, _ := vmessConfig["scy"].(string)
	if security == "" {
		security = "auto"
	}

	out := &SBOutbound{
		Tag:            name,
		Type:           "vmess",
		Server:         server,
		ServerPort:     port,
		UUID:           uuid,
		AlterId:        alterId,
		Security:       security,
		PacketEncoding: "xudp",
	}

	// TLS
	if tls, ok := vmessConfig["tls"].(string); ok && tls != "" && tls != "none" {
		out.TLS = &SBTLS{
			Enabled:  true,
			Insecure: true,
		}
		if sni, ok := vmessConfig["sni"].(string); ok {
			out.TLS.ServerName = sni
		} else if host, ok := vmessConfig["host"].(string); ok {
			out.TLS.ServerName = host
		}
		if fp, ok := vmessConfig["fp"].(string); ok && fp != "" {
			out.TLS.UTLS = &SBUTLS{Enabled: true, Fingerprint: fp}
		}
	}

	// Transport
	if net, ok := vmessConfig["net"].(string); ok {
		switch net {
		case "ws":
			out.Transport = &SBTransport{Type: "ws"}
			if path, ok := vmessConfig["path"].(string); ok {
				out.Transport.Path = path
			}
			if host, ok := vmessConfig["host"].(string); ok && host != "" {
				out.Transport.Headers = map[string]string{"Host": host}
			}
		case "grpc":
			out.Transport = &SBTransport{
				Type:        "grpc",
				ServiceName: vmessConfig["path"].(string),
			}
		case "h2", "http":
			out.Transport = &SBTransport{Type: "http"}
			if path, ok := vmessConfig["path"].(string); ok {
				out.Transport.Path = path
			}
		}
	}

	return out, nil
}

// parseVLESSURI 解析 vless:// URI
func parseVLESSURI(uri string) (*SBOutbound, error) {
	// vless://uuid@server:port?params#name
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	uuid := u.User.Username()
	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = "VLESS"
	}

	params := u.Query()

	out := &SBOutbound{
		Tag:            name,
		Type:           "vless",
		Server:         server,
		ServerPort:     port,
		UUID:           uuid,
		PacketEncoding: "xudp",
	}

	// Flow
	if flow := params.Get("flow"); flow != "" {
		out.Flow = flow
	}

	// TLS
	security := params.Get("security")
	if security == "tls" || security == "reality" {
		out.TLS = &SBTLS{
			Enabled:  true,
			Insecure: true,
		}
		if sni := params.Get("sni"); sni != "" {
			out.TLS.ServerName = sni
		}
		if fp := params.Get("fp"); fp != "" {
			out.TLS.UTLS = &SBUTLS{Enabled: true, Fingerprint: fp}
		}

		// Reality
		if security == "reality" {
			out.TLS.Reality = &SBReality{
				Enabled:   true,
				PublicKey: params.Get("pbk"),
				ShortID:   params.Get("sid"),
			}
		}
	}

	// Transport
	if netType := params.Get("type"); netType != "" {
		switch netType {
		case "ws":
			out.Transport = &SBTransport{
				Type: "ws",
				Path: params.Get("path"),
			}
			if host := params.Get("host"); host != "" {
				out.Transport.Headers = map[string]string{"Host": host}
			}
		case "grpc":
			out.Transport = &SBTransport{
				Type:        "grpc",
				ServiceName: params.Get("serviceName"),
			}
		case "h2":
			out.Transport = &SBTransport{
				Type: "http",
				Path: params.Get("path"),
			}
		}
	}

	return out, nil
}

// parseShadowsocksURI 解析 ss:// URI
func parseShadowsocksURI(uri string) (*SBOutbound, error) {
	// ss://method:password@server:port#name
	// or ss://base64@server:port#name

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	name := u.Fragment
	if name == "" {
		name = "Shadowsocks"
	}

	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	var method, password string

	if u.User != nil {
		// 可能是 base64 编码
		userInfo := u.User.String()
		if decoded, err := base64.StdEncoding.DecodeString(userInfo); err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				method = parts[0]
				password = parts[1]
			}
		} else if decoded, err := base64.RawURLEncoding.DecodeString(userInfo); err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				method = parts[0]
				password = parts[1]
			}
		} else {
			method = u.User.Username()
			password, _ = u.User.Password()
		}
	}

	return &SBOutbound{
		Tag:        name,
		Type:       "shadowsocks",
		Server:     server,
		ServerPort: port,
		Method:     method,
		Password:   password,
	}, nil
}

// parseTrojanURI 解析 trojan:// URI
func parseTrojanURI(uri string) (*SBOutbound, error) {
	// trojan://password@server:port?params#name
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	password := u.User.Username()
	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = "Trojan"
	}

	params := u.Query()

	out := &SBOutbound{
		Tag:        name,
		Type:       "trojan",
		Server:     server,
		ServerPort: port,
		Password:   password,
		TLS: &SBTLS{
			Enabled:  true,
			Insecure: true,
		},
	}

	if sni := params.Get("sni"); sni != "" {
		out.TLS.ServerName = sni
	}
	if fp := params.Get("fp"); fp != "" {
		out.TLS.UTLS = &SBUTLS{Enabled: true, Fingerprint: fp}
	}

	// Transport
	if netType := params.Get("type"); netType != "" {
		switch netType {
		case "ws":
			out.Transport = &SBTransport{
				Type: "ws",
				Path: params.Get("path"),
			}
			if host := params.Get("host"); host != "" {
				out.Transport.Headers = map[string]string{"Host": host}
			}
		case "grpc":
			out.Transport = &SBTransport{
				Type:        "grpc",
				ServiceName: params.Get("serviceName"),
			}
		}
	}

	return out, nil
}

// parseHysteria2URI 解析 hysteria2:// 或 hy2:// URI
func parseHysteria2URI(uri string) (*SBOutbound, error) {
	// hysteria2://password@server:port?params#name
	uri = strings.Replace(uri, "hy2://", "hysteria2://", 1)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	password := u.User.Username()
	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = "Hysteria2"
	}

	params := u.Query()

	out := &SBOutbound{
		Tag:        name,
		Type:       "hysteria2",
		Server:     server,
		ServerPort: port,
		Password:   password,
		TLS: &SBTLS{
			Enabled:  true,
			Insecure: true,
		},
	}

	if sni := params.Get("sni"); sni != "" {
		out.TLS.ServerName = sni
	}
	if insecure := params.Get("insecure"); insecure == "0" {
		out.TLS.Insecure = false
	}

	// Obfs
	if obfs := params.Get("obfs"); obfs != "" {
		out.Obfs = &SBObfs{
			Type:     obfs,
			Password: params.Get("obfs-password"),
		}
	}

	return out, nil
}

// parseTUICURI 解析 tuic:// URI
func parseTUICURI(uri string) (*SBOutbound, error) {
	// tuic://uuid:password@server:port?params#name
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	uuid := u.User.Username()
	password, _ := u.User.Password()
	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = "TUIC"
	}

	params := u.Query()

	out := &SBOutbound{
		Tag:               name,
		Type:              "tuic",
		Server:            server,
		ServerPort:        port,
		UUID:              uuid,
		Password:          password,
		CongestionControl: params.Get("congestion_control"),
		UDPRelayMode:      params.Get("udp_relay_mode"),
		TLS: &SBTLS{
			Enabled:  true,
			Insecure: true,
		},
	}

	if sni := params.Get("sni"); sni != "" {
		out.TLS.ServerName = sni
	}
	if alpn := params.Get("alpn"); alpn != "" {
		out.TLS.ALPN = strings.Split(alpn, ",")
	}

	return out, nil
}

// ============================================================================
// 节点过滤器
// ============================================================================

// FilterNodesByKeywords 按关键字过滤节点
func FilterNodesByKeywords(nodes []SBOutbound, keywords []string, action string) []SBOutbound {
	if len(keywords) == 0 {
		return nodes
	}

	// 构建正则
	pattern := strings.Join(keywords, "|")
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nodes
	}

	var result []SBOutbound
	for _, node := range nodes {
		matched := re.MatchString(node.Tag)

		if action == "include" && matched {
			result = append(result, node)
		} else if action == "exclude" && !matched {
			result = append(result, node)
		}
	}

	return result
}

// GetNodeTags 获取节点标签列表
func GetNodeTags(nodes []SBOutbound) []string {
	tags := make([]string, len(nodes))
	for i, node := range nodes {
		tags[i] = node.Tag
	}
	return tags
}
