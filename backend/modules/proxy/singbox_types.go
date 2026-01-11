package proxy

// ============================================================================
// Sing-Box 1.12+ 配置类型定义
// ============================================================================

// SingBoxConfig sing-box 主配置结构
type SingBoxConfig struct {
	Log          *SBLog          `json:"log,omitempty"`
	Experimental *SBExperimental `json:"experimental,omitempty"`
	DNS          *SBDNS          `json:"dns,omitempty"`
	Inbounds     []SBInbound     `json:"inbounds"`
	Outbounds    []SBOutbound    `json:"outbounds"`
	Route        *SBRoute        `json:"route,omitempty"`
}

// ============================================================================
// Log 配置
// ============================================================================

type SBLog struct {
	Level     string `json:"level,omitempty"`     // trace, debug, info, warn, error, fatal, panic
	Timestamp bool   `json:"timestamp,omitempty"` // 是否显示时间戳
	Output    string `json:"output,omitempty"`    // 日志输出路径
}

// ============================================================================
// Experimental 配置
// ============================================================================

type SBExperimental struct {
	ClashAPI  *SBClashAPI  `json:"clash_api,omitempty"`
	CacheFile *SBCacheFile `json:"cache_file,omitempty"`
}

type SBClashAPI struct {
	ExternalController string `json:"external_controller,omitempty"`
	Secret             string `json:"secret,omitempty"`
	DefaultMode        string `json:"default_mode,omitempty"` // rule, global, direct
}

type SBCacheFile struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Path        string `json:"path,omitempty"`
	StoreFakeIP bool   `json:"store_fakeip,omitempty"`
	StoreRDRC   bool   `json:"store_rdrc,omitempty"`
	RDRCTimeout string `json:"rdrc_timeout,omitempty"` // 默认 7d
}

// ============================================================================
// DNS 配置
// ============================================================================

type SBDNS struct {
	Servers          []SBDNSServer `json:"servers,omitempty"`
	Rules            []SBDNSRule   `json:"rules,omitempty"`
	Final            string        `json:"final,omitempty"`
	Strategy         string        `json:"strategy,omitempty"` // prefer_ipv4, prefer_ipv6, ipv4_only, ipv6_only
	IndependentCache bool          `json:"independent_cache,omitempty"`
	ReverseMapping   bool          `json:"reverse_mapping,omitempty"`
	Fakeip           *SBFakeIP     `json:"fakeip,omitempty"`
}

type SBDNSServer struct {
	Tag          string `json:"tag"`
	Type         string `json:"type,omitempty"`        // udp, tcp, tls, https, quic, h3, fakeip
	Server       string `json:"server,omitempty"`      // DNS 服务器地址
	ServerPort   int    `json:"server_port,omitempty"` // DNS 服务器端口 (sing-box 1.12+)
	Detour       string `json:"detour,omitempty"`      // 已弃用，保留兼容
	ClientSubnet string `json:"client_subnet,omitempty"`
	// FakeIP 专用
	Inet4Range string `json:"inet4_range,omitempty"`
	Inet6Range string `json:"inet6_range,omitempty"`
}

type SBDNSRule struct {
	// 匹配条件
	Inbound      []string    `json:"inbound,omitempty"`
	QueryType    []string    `json:"query_type,omitempty"`
	ClashMode    string      `json:"clash_mode,omitempty"`
	RuleSet      interface{} `json:"rule_set,omitempty"` // string 或 []string
	Domain       []string    `json:"domain,omitempty"`
	DomainSuffix []string    `json:"domain_suffix,omitempty"`
	Outbound     string      `json:"outbound,omitempty"`

	// 逻辑规则
	Type   string      `json:"type,omitempty"` // logical
	Mode   string      `json:"mode,omitempty"` // and, or
	Rules  []SBDNSRule `json:"rules,omitempty"`
	Invert bool        `json:"invert,omitempty"`

	// 动作
	Server string `json:"server,omitempty"`
}

type SBFakeIP struct {
	Enabled    bool   `json:"enabled,omitempty"`
	Inet4Range string `json:"inet4_range,omitempty"`
	Inet6Range string `json:"inet6_range,omitempty"`
}

// ============================================================================
// Inbound 配置
// ============================================================================

type SBInbound struct {
	Tag  string `json:"tag"`
	Type string `json:"type"` // tun, mixed, http, socks, direct

	// 通用字段
	Listen     string `json:"listen,omitempty"`
	ListenPort int    `json:"listen_port,omitempty"`

	// TUN 专用
	Address      []string    `json:"address,omitempty"`
	MTU          int         `json:"mtu,omitempty"`
	AutoRoute    bool        `json:"auto_route,omitempty"`
	AutoRedirect bool        `json:"auto_redirect,omitempty"` // Linux 性能优化，使用 nftables
	StrictRoute  bool        `json:"strict_route,omitempty"`
	Stack        string      `json:"stack,omitempty"` // system, gvisor, mixed
	UDPTimeout   string      `json:"udp_timeout,omitempty"`
	Platform     *SBPlatform `json:"platform,omitempty"`

	// Sniff
	Sniff                    bool `json:"sniff,omitempty"`
	SniffOverrideDestination bool `json:"sniff_override_destination,omitempty"`
}

type SBPlatform struct {
	HTTPProxy *SBHTTPProxy `json:"http_proxy,omitempty"`
}

type SBHTTPProxy struct {
	Enabled    bool   `json:"enabled"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
}

// ============================================================================
// Outbound 配置
// ============================================================================

type SBOutbound struct {
	Tag  string `json:"tag"`
	Type string `json:"type"` // selector, urltest, direct, block, dns, vmess, vless, shadowsocks, trojan, hysteria2, etc.

	// ===== 代理组专用 =====
	Outbounds                 []string `json:"outbounds,omitempty"`
	Default                   string   `json:"default,omitempty"`
	InterruptExistConnections bool     `json:"interrupt_exist_connections,omitempty"`

	// URLTest 专用
	URL       string `json:"url,omitempty"`
	Interval  string `json:"interval,omitempty"`
	Tolerance int    `json:"tolerance,omitempty"`

	// ===== 节点通用字段 =====
	Server     string `json:"server,omitempty"`
	ServerPort int    `json:"server_port,omitempty"`

	// ===== VMess =====
	UUID           string `json:"uuid,omitempty"`
	Security       string `json:"security,omitempty"`
	AlterId        int    `json:"alter_id,omitempty"`
	PacketEncoding string `json:"packet_encoding,omitempty"` // xudp, packetaddr

	// ===== VLESS =====
	Flow string `json:"flow,omitempty"` // xtls-rprx-vision

	// ===== Shadowsocks =====
	Method     string `json:"method,omitempty"`
	Password   string `json:"password,omitempty"`
	Plugin     string `json:"plugin,omitempty"`
	PluginOpts string `json:"plugin_opts,omitempty"`

	// ===== Trojan =====
	// Password 复用上面的

	// ===== Hysteria2 =====
	UpMbps   int     `json:"up_mbps,omitempty"`
	DownMbps int     `json:"down_mbps,omitempty"`
	Obfs     *SBObfs `json:"obfs,omitempty"`

	// ===== TUIC =====
	CongestionControl string `json:"congestion_control,omitempty"`
	UDPRelayMode      string `json:"udp_relay_mode,omitempty"`
	ZeroRTTHandshake  bool   `json:"zero_rtt_handshake,omitempty"`
	Heartbeat         string `json:"heartbeat,omitempty"`

	// ===== AnyTLS =====
	IdleSessionCheckInterval string `json:"idle_session_check_interval,omitempty"`
	IdleSessionTimeout       string `json:"idle_session_timeout,omitempty"`
	MinIdleSession           int    `json:"min_idle_session,omitempty"`

	// ===== WireGuard =====
	PrivateKey    string   `json:"private_key,omitempty"`
	PeerPublicKey string   `json:"peer_public_key,omitempty"`
	PreSharedKey  string   `json:"pre_shared_key,omitempty"`
	LocalAddress  []string `json:"local_address,omitempty"`
	Reserved      []int    `json:"reserved,omitempty"`

	// ===== 通用传输层 =====
	TLS       *SBTLS       `json:"tls,omitempty"`
	Transport *SBTransport `json:"transport,omitempty"`
	Multiplex *SBMultiplex `json:"multiplex,omitempty"`

	// ===== Dial 性能优化字段 =====
	TCPFastOpen  bool `json:"tcp_fast_open,omitempty"`
	TCPMultiPath bool `json:"tcp_multi_path,omitempty"`
	UDPFragment  bool `json:"udp_fragment,omitempty"`
}

type SBObfs struct {
	Type     string `json:"type,omitempty"` // salamander
	Password string `json:"password,omitempty"`
}

type SBTLS struct {
	Enabled    bool       `json:"enabled,omitempty"`
	ServerName string     `json:"server_name,omitempty"`
	Insecure   bool       `json:"insecure,omitempty"`
	ALPN       []string   `json:"alpn,omitempty"`
	MinVersion string     `json:"min_version,omitempty"`
	MaxVersion string     `json:"max_version,omitempty"`
	UTLS       *SBUTLS    `json:"utls,omitempty"`
	Reality    *SBReality `json:"reality,omitempty"`
}

type SBUTLS struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"` // chrome, firefox, safari, ios, android, edge, 360, qq, random, randomized
}

type SBReality struct {
	Enabled   bool   `json:"enabled,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type SBTransport struct {
	Type        string            `json:"type,omitempty"` // http, ws, quic, grpc, httpupgrade
	Host        interface{}       `json:"host,omitempty"` // string 或 []string
	Path        string            `json:"path,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Method      string            `json:"method,omitempty"`
	ServiceName string            `json:"service_name,omitempty"` // gRPC

	// WebSocket
	MaxEarlyData        int    `json:"max_early_data,omitempty"`
	EarlyDataHeaderName string `json:"early_data_header_name,omitempty"`
}

type SBMultiplex struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Protocol       string `json:"protocol,omitempty"` // smux, yamux, h2mux
	MaxConnections int    `json:"max_connections,omitempty"`
	MinStreams     int    `json:"min_streams,omitempty"`
	MaxStreams     int    `json:"max_streams,omitempty"`
	Padding        bool   `json:"padding,omitempty"`
}

// ============================================================================
// Route 配置
// ============================================================================

type SBRoute struct {
	Rules                 []SBRouteRule     `json:"rules,omitempty"`
	RuleSet               []SBRuleSet       `json:"rule_set,omitempty"`
	Final                 string            `json:"final,omitempty"`
	AutoDetectInterface   bool              `json:"auto_detect_interface,omitempty"`
	DefaultInterface      string            `json:"default_interface,omitempty"`
	DefaultDomainResolver *SBDomainResolver `json:"default_domain_resolver,omitempty"`
}

type SBDomainResolver struct {
	Server string `json:"server,omitempty"`
}

type SBRouteRule struct {
	// 匹配条件
	Inbound       []string    `json:"inbound,omitempty"`
	Protocol      interface{} `json:"protocol,omitempty"` // string 或 []string
	Port          interface{} `json:"port,omitempty"`     // int 或 []int
	PortRange     []string    `json:"port_range,omitempty"`
	Domain        []string    `json:"domain,omitempty"`
	DomainSuffix  []string    `json:"domain_suffix,omitempty"`
	DomainKeyword []string    `json:"domain_keyword,omitempty"`
	DomainRegex   []string    `json:"domain_regex,omitempty"`
	IPIsPrivate   bool        `json:"ip_is_private,omitempty"`
	IPCIDR        []string    `json:"ip_cidr,omitempty"`
	SourceIPCIDR  []string    `json:"source_ip_cidr,omitempty"`
	ClashMode     string      `json:"clash_mode,omitempty"`
	RuleSet       interface{} `json:"rule_set,omitempty"` // string 或 []string

	// 逻辑规则
	Type   string        `json:"type,omitempty"` // logical
	Mode   string        `json:"mode,omitempty"` // and, or
	Rules  []SBRouteRule `json:"rules,omitempty"`
	Invert bool          `json:"invert,omitempty"`

	// 动作
	Action   string `json:"action,omitempty"`   // route, reject, hijack-dns, sniff
	Outbound string `json:"outbound,omitempty"` // 当 action 为 route 时使用
}

type SBRuleSet struct {
	Tag            string `json:"tag"`
	Type           string `json:"type"`   // remote, local
	Format         string `json:"format"` // binary, source
	URL            string `json:"url,omitempty"`
	Path           string `json:"path,omitempty"`
	DownloadDetour string `json:"download_detour,omitempty"`
	UpdateInterval string `json:"update_interval,omitempty"`
}

// ============================================================================
// 节点过滤器 (用于代理组)
// ============================================================================

type SBOutboundFilter struct {
	Action   string   `json:"action"` // include, exclude
	Keywords []string `json:"keywords"`
	For      []string `json:"for,omitempty"` // 仅对指定组生效
}

// ============================================================================
// 配置生成选项
// ============================================================================

type SingBoxGeneratorOptions struct {
	// 模式
	Mode   string `json:"mode"`   // tun, system
	FakeIP bool   `json:"fakeip"` // 启用 FakeIP

	// 端口
	MixedPort int `json:"mixedPort"`
	HTTPPort  int `json:"httpPort"`
	SocksPort int `json:"socksPort"`

	// API
	ClashAPIAddr   string `json:"clashApiAddr"`
	ClashAPISecret string `json:"clashApiSecret"`

	// TUN 设置
	TUNStack     string `json:"tunStack"` // system, gvisor, mixed
	TUNMTU       int    `json:"tunMtu"`
	AutoRedirect bool   `json:"autoRedirect"` // Linux nftables 性能优化
	StrictRoute  bool   `json:"strictRoute"`  // 严格路由

	// DNS
	DNSStrategy string `json:"dnsStrategy"` // prefer_ipv4, prefer_ipv6, ipv4_only

	// 性能优化
	TCPFastOpen              bool `json:"tcpFastOpen"`
	TCPMultiPath             bool `json:"tcpMultiPath"`
	UDPFragment              bool `json:"udpFragment"`
	Sniff                    bool `json:"sniff"`
	SniffOverrideDestination bool `json:"sniffOverrideDestination"`

	// 日志
	LogLevel string `json:"logLevel"`
}
