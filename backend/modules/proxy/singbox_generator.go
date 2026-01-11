package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SingboxConfig sing-box 配置结构
type SingboxConfig struct {
	Log       SingboxLog        `json:"log"`
	DNS       SingboxDNS        `json:"dns"`
	Inbounds  []SingboxInbound  `json:"inbounds"`
	Outbounds []SingboxOutbound `json:"outbounds"`
	Route     SingboxRoute      `json:"route"`
}

type SingboxLog struct {
	Level     string `json:"level"`
	Timestamp bool   `json:"timestamp"`
}

type SingboxDNS struct {
	Servers []SingboxDNSServer `json:"servers"`
	Rules   []SingboxDNSRule   `json:"rules,omitempty"`
}

type SingboxDNSServer struct {
	Tag     string `json:"tag"`
	Address string `json:"address"`
	Detour  string `json:"detour,omitempty"`
}

type SingboxDNSRule struct {
	Domain   []string `json:"domain,omitempty"`
	GeoSite  []string `json:"geosite,omitempty"`
	Server   string   `json:"server"`
	Outbound string   `json:"outbound,omitempty"`
}

type SingboxInbound struct {
	Type                     string   `json:"type"`
	Tag                      string   `json:"tag"`
	Listen                   string   `json:"listen,omitempty"`
	ListenPort               int      `json:"listen_port,omitempty"`
	Sniff                    bool     `json:"sniff,omitempty"`
	SniffOverrideDestination bool     `json:"sniff_override_destination,omitempty"`
	DomainStrategy           string   `json:"domain_strategy,omitempty"`
	InterfaceName            string   `json:"interface_name,omitempty"`
	MTU                      int      `json:"mtu,omitempty"`
	Inet4Address             []string `json:"inet4_address,omitempty"`
	AutoRoute                bool     `json:"auto_route,omitempty"`
	StrictRoute              bool     `json:"strict_route,omitempty"`
}

type SingboxOutbound struct {
	Type       string            `json:"type"`
	Tag        string            `json:"tag"`
	Server     string            `json:"server,omitempty"`
	ServerPort int               `json:"server_port,omitempty"`
	Method     string            `json:"method,omitempty"`
	Password   string            `json:"password,omitempty"`
	Username   string            `json:"username,omitempty"` // SOCKS5/HTTP 认证用户名
	UUID       string            `json:"uuid,omitempty"`
	Security   string            `json:"security,omitempty"`
	AlterId    int               `json:"alter_id,omitempty"`
	Flow       string            `json:"flow,omitempty"` // VLESS flow (xtls-rprx-vision)
	TLS        *SingboxTLS       `json:"tls,omitempty"`
	Transport  *SingboxTransport `json:"transport,omitempty"`
	Obfs       *SingboxObfs      `json:"obfs,omitempty"`      // Hysteria2 混淆
	UpMbps     string            `json:"up_mbps,omitempty"`   // Hysteria2 上行带宽
	DownMbps   string            `json:"down_mbps,omitempty"` // Hysteria2 下行带宽
	// AnyTLS 字段
	IdleSessionCheckInterval string   `json:"idle_session_check_interval,omitempty"`
	IdleSessionTimeout       string   `json:"idle_session_timeout,omitempty"`
	MinIdleSession           int      `json:"min_idle_session,omitempty"`
	Outbounds                []string `json:"outbounds,omitempty"`
	Default                  string   `json:"default,omitempty"`
	URL                      string   `json:"url,omitempty"`
	Interval                 string   `json:"interval,omitempty"`
	InterruptExist           bool     `json:"interrupt_exist_connections,omitempty"`
}

type SingboxObfs struct {
	Type     string `json:"type"`
	Password string `json:"password,omitempty"`
}

type SingboxTLS struct {
	Enabled    bool            `json:"enabled"`
	ServerName string          `json:"server_name,omitempty"`
	Insecure   bool            `json:"insecure,omitempty"`
	ALPN       []string        `json:"alpn,omitempty"`
	UTLS       *SingboxUTLS    `json:"utls,omitempty"`    // UTLS 配置
	Reality    *SingboxReality `json:"reality,omitempty"` // Reality 配置
}

type SingboxUTLS struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"` // chrome, firefox, safari, etc
}

type SingboxReality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type SingboxTransport struct {
	Type        string            `json:"type"`
	Path        string            `json:"path,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ServiceName string            `json:"service_name,omitempty"` // gRPC service name
}

type SingboxRoute struct {
	Rules               []SingboxRouteRule `json:"rules"`
	AutoDetectInterface bool               `json:"auto_detect_interface"`
	FinalOutbound       string             `json:"final,omitempty"`
}

type SingboxRouteRule struct {
	Protocol      []string `json:"protocol,omitempty"`
	Domain        []string `json:"domain,omitempty"`
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	IPCidr        []string `json:"ip_cidr,omitempty"`
	GeoIP         []string `json:"geoip,omitempty"`
	GeoSite       []string `json:"geosite,omitempty"`
	Port          []int    `json:"port,omitempty"`
	Outbound      string   `json:"outbound"`
}

// SingboxGenerator sing-box 配置生成器
type SingboxGenerator struct {
	dataDir string
}

func NewSingboxGenerator(dataDir string) *SingboxGenerator {
	return &SingboxGenerator{dataDir: dataDir}
}

// GenerateConfig 生成 sing-box 配置
func (g *SingboxGenerator) GenerateConfig(nodes []ProxyNode, options ConfigGeneratorOptions) (*SingboxConfig, error) {
	if options.MixedPort == 0 {
		options.MixedPort = 7890
	}

	config := &SingboxConfig{
		Log: SingboxLog{
			Level:     "info",
			Timestamp: true,
		},
	}

	// DNS 配置
	config.DNS = g.generateDNS(options)

	// 入站配置
	config.Inbounds = g.generateInbounds(options)

	// 出站配置
	config.Outbounds = g.generateOutbounds(nodes, options)

	// 路由配置
	config.Route = g.generateRoute(options)

	return config, nil
}

func (g *SingboxGenerator) generateDNS(options ConfigGeneratorOptions) SingboxDNS {
	dns := SingboxDNS{
		Servers: []SingboxDNSServer{
			{Tag: "dns-direct", Address: "https://dns.alidns.com/dns-query", Detour: "direct"},
			{Tag: "dns-proxy", Address: "https://dns.google/dns-query", Detour: "proxy"},
			{Tag: "dns-block", Address: "rcode://success"},
		},
		Rules: []SingboxDNSRule{
			{GeoSite: []string{"cn"}, Server: "dns-direct"},
			{GeoSite: []string{"geolocation-!cn"}, Server: "dns-proxy"},
		},
	}
	return dns
}

func (g *SingboxGenerator) generateInbounds(options ConfigGeneratorOptions) []SingboxInbound {
	inbounds := []SingboxInbound{
		{
			Type: "mixed",
			Tag:  "mixed-in",
			Listen: func() string {
				if options.AllowLan {
					return "0.0.0.0"
				}
				return "127.0.0.1"
			}(),
			ListenPort:               options.MixedPort,
			Sniff:                    true,
			SniffOverrideDestination: false,
		},
	}

	// TUN 入站
	if options.EnableTUN {
		inbounds = append(inbounds, SingboxInbound{
			Type:          "tun",
			Tag:           "tun-in",
			InterfaceName: "utun",
			MTU:           9000,
			Inet4Address:  []string{"172.19.0.1/30"},
			AutoRoute:     true,
			StrictRoute:   true,
			Sniff:         true,
		})
	}

	return inbounds
}

func (g *SingboxGenerator) generateOutbounds(nodes []ProxyNode, options ConfigGeneratorOptions) []SingboxOutbound {
	outbounds := []SingboxOutbound{}
	nodeNames := []string{}

	// 转换代理节点
	for _, node := range nodes {
		outbound := g.convertProxyNode(node)
		if outbound != nil {
			outbounds = append(outbounds, *outbound)
			nodeNames = append(nodeNames, node.Name)
		}
	}

	// 添加选择器
	if len(nodeNames) > 0 {
		outbounds = append(outbounds, SingboxOutbound{
			Type:      "selector",
			Tag:       "proxy",
			Outbounds: append([]string{"auto"}, nodeNames...),
			Default:   "auto",
		})

		outbounds = append(outbounds, SingboxOutbound{
			Type:           "urltest",
			Tag:            "auto",
			Outbounds:      nodeNames,
			URL:            "https://www.gstatic.com/generate_204",
			Interval:       "300s",
			InterruptExist: true,
		})
	}

	// 直连和阻止
	outbounds = append(outbounds,
		SingboxOutbound{Type: "direct", Tag: "direct"},
		SingboxOutbound{Type: "block", Tag: "block"},
		SingboxOutbound{Type: "dns", Tag: "dns-out"},
	)

	return outbounds
}

func (g *SingboxGenerator) convertProxyNode(node ProxyNode) *SingboxOutbound {
	// 优先从完整配置解析
	if node.Config != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(node.Config), &config); err == nil {
			// 传递 node 以获取 server/port
			return g.convertFromConfig(node, config)
		}
	}

	// 没有完整配置，使用基础字段构建
	outbound := &SingboxOutbound{
		Tag:        node.Name,
		Server:     node.Server,
		ServerPort: node.GetPort(),
	}

	nodeType := strings.ToLower(node.Type)

	switch nodeType {
	case "ss", "shadowsocks":
		outbound.Type = "shadowsocks"
	case "vmess":
		outbound.Type = "vmess"
		outbound.Security = "auto"
	case "vless":
		outbound.Type = "vless"
	case "trojan":
		outbound.Type = "trojan"
		// Trojan 默认启用 TLS
		outbound.TLS = &SingboxTLS{
			Enabled:  true,
			Insecure: true,
		}
	case "hysteria2", "hy2":
		outbound.Type = "hysteria2"
	case "anytls":
		outbound.Type = "anytls"
		// AnyTLS 需要启用 TLS
		outbound.TLS = &SingboxTLS{
			Enabled:  true,
			Insecure: true,
		}
	case "socks", "socks5":
		outbound.Type = "socks"
		// 解析配置 JSON
		var socksConfig map[string]interface{}
		if node.Config != "" {
			json.Unmarshal([]byte(node.Config), &socksConfig)
		}
		// SOCKS5 认证
		if socksConfig != nil {
			if username, ok := socksConfig["username"].(string); ok && username != "" {
				outbound.Username = username
			}
			if password, ok := socksConfig["password"].(string); ok && password != "" {
				outbound.Password = password
			}
		}
	case "http":
		outbound.Type = "http"
	default:
		return nil
	}

	return outbound
}

func (g *SingboxGenerator) convertFromConfig(node ProxyNode, config map[string]interface{}) *SingboxOutbound {
	outbound := &SingboxOutbound{
		Tag:        node.Name,
		Server:     node.Server,    // 优先使用 node 的 server
		ServerPort: node.GetPort(), // 优先使用 node 的 port
	}

	// 如果 config 中有 server/port，则使用 config 中的值
	if server, ok := config["server"].(string); ok && server != "" {
		outbound.Server = server
	}
	if port, ok := config["port"].(float64); ok && port > 0 {
		outbound.ServerPort = int(port)
	}

	nodeType := strings.ToLower(node.Type)
	if t, ok := config["type"].(string); ok && t != "" {
		nodeType = strings.ToLower(t)
	}

	switch nodeType {
	case "ss", "shadowsocks":
		outbound.Type = "shadowsocks"
		outbound.Method, _ = config["cipher"].(string)
		outbound.Password, _ = config["password"].(string)

	case "vmess":
		outbound.Type = "vmess"
		outbound.UUID, _ = config["uuid"].(string)
		if alterId, ok := config["alterId"].(float64); ok {
			outbound.AlterId = int(alterId)
		}
		outbound.Security, _ = config["cipher"].(string)
		if outbound.Security == "" {
			outbound.Security = "auto"
		}

		// TLS
		if tls, ok := config["tls"].(bool); ok && tls {
			outbound.TLS = &SingboxTLS{
				Enabled:    true,
				ServerName: getStringOr(config, "sni", outbound.Server),
				Insecure:   getBool(config, "skip-cert-verify"),
			}
		}

		// Transport
		if network, ok := config["network"].(string); ok && network != "tcp" {
			outbound.Transport = &SingboxTransport{Type: network}
			if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
				outbound.Transport.Path, _ = wsOpts["path"].(string)
				if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
					outbound.Transport.Headers = make(map[string]string)
					for k, v := range headers {
						outbound.Transport.Headers[k], _ = v.(string)
					}
				}
			}
		}

	case "trojan":
		outbound.Type = "trojan"
		outbound.Password, _ = config["password"].(string)
		outbound.TLS = &SingboxTLS{
			Enabled:    true,
			ServerName: getStringOr(config, "sni", outbound.Server),
			Insecure:   getBool(config, "skip-cert-verify"),
		}

	case "vless":
		outbound.Type = "vless"
		outbound.UUID, _ = config["uuid"].(string)

		// Flow (xtls-rprx-vision 等)
		if flow, ok := config["flow"].(string); ok {
			outbound.Flow = flow
		}

		// TLS / Reality
		if tls, ok := config["tls"].(bool); ok && tls {
			outbound.TLS = &SingboxTLS{
				Enabled:    true,
				ServerName: getStringOr(config, "sni", outbound.Server),
				Insecure:   getBool(config, "skip-cert-verify"),
			}
			// UTLS Fingerprint
			if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
				outbound.TLS.UTLS = &SingboxUTLS{Enabled: true, Fingerprint: fp}
			}
			// Reality
			if realityOpts, ok := config["reality-opts"].(map[string]interface{}); ok {
				outbound.TLS.Reality = &SingboxReality{
					Enabled:   true,
					PublicKey: getStringOr(realityOpts, "public-key", ""),
					ShortID:   getStringOr(realityOpts, "short-id", ""),
				}
			}
		}

		// Transport (ws, grpc, http 等)
		if network, ok := config["network"].(string); ok && network != "" && network != "tcp" {
			outbound.Transport = &SingboxTransport{Type: network}
			switch network {
			case "ws", "websocket":
				outbound.Transport.Type = "ws"
				if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
					outbound.Transport.Path, _ = wsOpts["path"].(string)
					if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
						outbound.Transport.Headers = make(map[string]string)
						for k, v := range headers {
							outbound.Transport.Headers[k], _ = v.(string)
						}
					}
				}
			case "grpc":
				if grpcOpts, ok := config["grpc-opts"].(map[string]interface{}); ok {
					outbound.Transport.ServiceName, _ = grpcOpts["grpc-service-name"].(string)
				}
			case "http", "h2":
				outbound.Transport.Type = "http"
				if httpOpts, ok := config["http-opts"].(map[string]interface{}); ok {
					outbound.Transport.Path, _ = httpOpts["path"].(string)
				}
			}
		}

	case "hysteria2", "hy2":
		outbound.Type = "hysteria2"
		outbound.Password, _ = config["password"].(string)

		// Obfs (混淆)
		if obfs, ok := config["obfs"].(string); ok && obfs != "" {
			outbound.Obfs = &SingboxObfs{
				Type:     obfs,
				Password: getStringOr(config, "obfs-password", ""),
			}
		}

		// 带宽限制 (可选)
		if up, ok := config["up"].(string); ok {
			outbound.UpMbps = up
		}
		if down, ok := config["down"].(string); ok {
			outbound.DownMbps = down
		}

		// TLS (Hysteria2 必须启用 TLS)
		outbound.TLS = &SingboxTLS{
			Enabled:    true,
			ServerName: getStringOr(config, "sni", outbound.Server),
			Insecure:   getBool(config, "skip-cert-verify"),
		}
		// ALPN
		if alpn, ok := config["alpn"].([]interface{}); ok {
			for _, a := range alpn {
				if s, ok := a.(string); ok {
					outbound.TLS.ALPN = append(outbound.TLS.ALPN, s)
				}
			}
		}
		// UTLS Fingerprint
		if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
			outbound.TLS.UTLS = &SingboxUTLS{Enabled: true, Fingerprint: fp}
		}

	case "anytls":
		outbound.Type = "anytls"
		outbound.Password, _ = config["password"].(string)

		// Session 配置
		if interval, ok := config["idle-session-check-interval"].(float64); ok {
			outbound.IdleSessionCheckInterval = fmt.Sprintf("%ds", int(interval))
		}
		if timeout, ok := config["idle-session-timeout"].(float64); ok {
			outbound.IdleSessionTimeout = fmt.Sprintf("%ds", int(timeout))
		}
		if minSession, ok := config["min-idle-session"].(float64); ok {
			outbound.MinIdleSession = int(minSession)
		}

		// TLS (AnyTLS 必须启用 TLS)
		outbound.TLS = &SingboxTLS{
			Enabled:    true,
			ServerName: getStringOr(config, "sni", outbound.Server),
			Insecure:   getBool(config, "skip-cert-verify"),
		}
		// ALPN
		if alpn, ok := config["alpn"].([]interface{}); ok {
			for _, a := range alpn {
				if s, ok := a.(string); ok {
					outbound.TLS.ALPN = append(outbound.TLS.ALPN, s)
				}
			}
		}
		// UTLS Fingerprint
		if fp, ok := config["client-fingerprint"].(string); ok && fp != "" {
			outbound.TLS.UTLS = &SingboxUTLS{Enabled: true, Fingerprint: fp}
		}
		// Reality 配置
		if realityOpts, ok := config["reality-opts"].(map[string]interface{}); ok {
			outbound.TLS.Reality = &SingboxReality{
				Enabled:   true,
				PublicKey: getStringOr(realityOpts, "public-key", ""),
				ShortID:   getStringOr(realityOpts, "short-id", ""),
			}
		}

	default:
		return nil
	}

	return outbound
}

func (g *SingboxGenerator) generateRoute(options ConfigGeneratorOptions) SingboxRoute {
	route := SingboxRoute{
		AutoDetectInterface: true,
		FinalOutbound:       "proxy",
		Rules: []SingboxRouteRule{
			// DNS 劫持
			{Protocol: []string{"dns"}, Outbound: "dns-out"},
			// 私有地址直连
			{IPCidr: []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, Outbound: "direct"},
			// 中国直连
			{GeoIP: []string{"cn"}, Outbound: "direct"},
			{GeoSite: []string{"cn"}, Outbound: "direct"},
			// 广告拦截
			{GeoSite: []string{"category-ads-all"}, Outbound: "block"},
		},
	}
	return route
}

// SaveConfig 保存配置到文件
func (g *SingboxGenerator) SaveConfig(config *SingboxConfig, filename string) (string, error) {
	configDir := filepath.Join(g.dataDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	filePath := filepath.Join(configDir, filename)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// Helper functions
func getStringOr(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// ============================================================================
// Sing-Box 1.12+ 配置生成
// ============================================================================

// GenerateConfigV112 生成 Sing-Box 1.12+ 配置
func (g *SingboxGenerator) GenerateConfigV112(nodes []ProxyNode, opts SingBoxGeneratorOptions) (*SingBoxConfig, error) {
	// 获取基础模板
	var config *SingBoxConfig
	if opts.Mode == "tun" {
		config = GetSingBoxTUNTemplate(opts)
	} else {
		config = GetSingBoxSystemTemplate(opts)
	}

	// 转换节点为 outbounds，并收集手动节点名称
	nodeOutbounds := make([]SBOutbound, 0, len(nodes))
	manualNodeNames := make([]string, 0)
	for _, node := range nodes {
		outbound, err := ParseNodeToSingBox(node)
		if err != nil {
			continue // 跳过无法解析的节点
		}
		nodeOutbounds = append(nodeOutbounds, *outbound)
		// 收集手动节点名称（与 Mihomo 一致）
		if node.IsManual {
			manualNodeNames = append(manualNodeNames, outbound.Tag)
		}
	}

	// 生成代理组（传入手动节点名称列表）
	proxyGroups := g.generateProxyGroupsV112(nodeOutbounds, manualNodeNames)

	// 组合所有 outbounds
	// 顺序: 代理组 -> 节点 -> 特殊出站(direct/block/dns-out)
	allOutbounds := make([]SBOutbound, 0)
	allOutbounds = append(allOutbounds, proxyGroups...)
	allOutbounds = append(allOutbounds, nodeOutbounds...)
	// 添加内置出站 (Sing-Box 内置 direct/block)
	allOutbounds = append(allOutbounds,
		SBOutbound{Type: "direct", Tag: "direct"},
		SBOutbound{Type: "block", Tag: "block"},
		SBOutbound{Type: "dns", Tag: "dns-out"},
	)

	config.Outbounds = allOutbounds

	// 添加路由规则
	config.Route.Rules = GetDefaultRouteRules()
	config.Route.RuleSet = GetDefaultRuleSets()

	return config, nil
}
func (g *SingboxGenerator) generateProxyGroupsV112(nodes []SBOutbound, manualNodeNames []string) []SBOutbound {
	// 地区过滤关键字 (与 Mihomo 保持一致)
	regionFilters := map[string][]string{
		"HongKong":  {"🇭🇰", "HK", "hk", "香港", "港", "HongKong", "Hong Kong", "HONG KONG", "沪港", "呼港", "中港", "HKT", "HKBN", "HGC", "WTT", "CMI", "穗港", "广港", "京港"},
		"Taiwan":    {"🇨🇳", "🇹🇼", "TW", "tw", "台湾", "台灣", "臺灣", "台北", "台中", "新北", "彰化", "CHT", "HINET", "Taiwan", "TAIWAN"},
		"Japan":     {"🇯🇵", "JP", "jp", "日本", "东京", "東京", "大阪", "埼玉", "京日", "苏日", "沪日", "广日", "上日", "穗日", "川日", "中日", "泉日", "杭日", "深日", "Japan", "JAPAN"},
		"Singapore": {"🇸🇬", "SG", "sg", "新加坡", "狮城", "獅城", "沪新", "京新", "泉新", "穗新", "深新", "杭新", "广新", "廣新", "滬新", "Singapore", "SINGAPORE"},
		"America":   {"🇺🇸", "US", "us", "美国", "美國", "京美", "硅谷", "凤凰城", "洛杉矶", "西雅图", "圣何塞", "芝加哥", "哥伦布", "纽约", "广美", "America", "United States", "USA"},
	}

	// 分类节点
	regionGroups := make(map[string][]string)
	otherNodes := []string{}
	allNodeTags := []string{}

	// 创建手动节点名称集合，用于快速查找
	manualNodeSet := make(map[string]bool)
	for _, name := range manualNodeNames {
		manualNodeSet[name] = true
	}

	for _, node := range nodes {
		allNodeTags = append(allNodeTags, node.Tag)

		// 手动节点不参与地区分类（它们单独在手动节点分组中）
		if manualNodeSet[node.Tag] {
			continue
		}

		matched := false
		for region, keywords := range regionFilters {
			if matchesKeywords(node.Tag, keywords) {
				regionGroups[region] = append(regionGroups[region], node.Tag)
				matched = true
				break
			}
		}
		if !matched {
			otherNodes = append(otherNodes, node.Tag)
		}
	}
	regionGroups["Manual"] = manualNodeNames // 使用传入的手动节点名称列表
	regionGroups["Others"] = otherNodes

	// 获取基础代理组
	groups := GetSingBoxProxyGroups()

	// 地区名称映射（英文 -> 中文）
	regionNameMap := map[string]string{
		"HongKong":  "香港节点",
		"Taiwan":    "台湾节点",
		"Japan":     "日本节点",
		"Singapore": "新加坡节点",
		"America":   "美国节点",
	}

	// 填充节点
	for i := range groups {
		switch groups[i].Tag {
		case "自动选择":
			// 自动测速添加所有节点
			groups[i].Outbounds = allNodeTags
		case "香港节点", "台湾节点", "日本节点", "新加坡节点", "美国节点":
			// 地区组填充对应节点（需要反向查找英文 key）
			var englishKey string
			for k, v := range regionNameMap {
				if v == groups[i].Tag {
					englishKey = k
					break
				}
			}
			if regionNodes, ok := regionGroups[englishKey]; ok && len(regionNodes) > 0 {
				groups[i].Outbounds = regionNodes
			} else {
				// 如果没有节点，添加 节点选择 作为后备
				groups[i].Outbounds = []string{"节点选择"}
			}
		case "手动节点":
			// 手动节点 - 只包含手动添加的节点（URL导入/手动添加，非订阅节点）
			// 如果没有手动节点，留空（后面会过滤掉）
			if len(manualNodeNames) > 0 {
				groups[i].Outbounds = manualNodeNames
			}
			// 没有手动节点时 Outbounds 为空，会被后面过滤掉
		case "其他节点":
			// 其他节点
			if len(otherNodes) > 0 {
				groups[i].Outbounds = otherNodes
			} else {
				groups[i].Outbounds = []string{"节点选择"}
			}
		}
	}

	// 过滤掉没有有效 outbounds 的组
	validGroups := make([]SBOutbound, 0)
	removedTags := make(map[string]bool)
	for _, group := range groups {
		if len(group.Outbounds) > 0 {
			validGroups = append(validGroups, group)
		} else {
			removedTags[group.Tag] = true
		}
	}

	// 从其他分组的 outbounds 中移除已删除的分组引用
	if len(removedTags) > 0 {
		for i := range validGroups {
			filteredOutbounds := make([]string, 0)
			for _, out := range validGroups[i].Outbounds {
				if !removedTags[out] {
					filteredOutbounds = append(filteredOutbounds, out)
				}
			}
			validGroups[i].Outbounds = filteredOutbounds
		}
	}

	return validGroups
}

// matchesKeywords 检查名称是否匹配关键字
func matchesKeywords(name string, keywords []string) bool {
	nameLower := strings.ToLower(name)
	for _, kw := range keywords {
		if strings.Contains(nameLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// SaveConfigV112 保存 1.12+ 配置到文件
func (g *SingboxGenerator) SaveConfigV112(config *SingBoxConfig, filename string) (string, error) {
	configDir := filepath.Join(g.dataDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	filePath := filepath.Join(configDir, filename)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}
