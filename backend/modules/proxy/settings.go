package proxy

// AuthUser 代理认证用户
type AuthUser struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Enabled  bool   `json:"enabled" yaml:"enabled"` // 是否启用此账号
}

// ProxySettings 代理核心设置
type ProxySettings struct {
	// === 端口设置 ===
	MixedPortEnabled  bool `json:"mixedPortEnabled" yaml:"mixed-port-enabled"`   // 是否启用混合代理
	MixedPort         int  `json:"mixedPort" yaml:"mixed-port"`                  // 混合代理端口 (HTTP+SOCKS5)
	SocksPortEnabled  bool `json:"socksPortEnabled" yaml:"socks-port-enabled"`   // 是否启用 SOCKS5
	SocksPort         int  `json:"socksPort" yaml:"socks-port"`                  // SOCKS5 端口
	HTTPPortEnabled   bool `json:"httpPortEnabled" yaml:"http-port-enabled"`     // 是否启用 HTTP
	HTTPPort          int  `json:"httpPort" yaml:"port"`                         // HTTP 端口
	RedirPortEnabled  bool `json:"redirPortEnabled" yaml:"redir-port-enabled"`   // 是否启用透明代理
	RedirPort         int  `json:"redirPort" yaml:"redir-port"`                  // 透明代理端口 (Linux)
	TProxyPortEnabled bool `json:"tproxyPortEnabled" yaml:"tproxy-port-enabled"` // 是否启用 TProxy
	TProxyPort        int  `json:"tproxyPort" yaml:"tproxy-port"`                // TProxy 端口 (Linux)

	// === 认证设置 ===
	Authentication []AuthUser `json:"authentication" yaml:"authentication"` // 代理认证用户列表 (启用的账号自动开启认证)

	// === 基础设置 ===
	AllowLan       bool   `json:"allowLan" yaml:"allow-lan"`              // 允许局域网连接
	BindAddress    string `json:"bindAddress" yaml:"bind-address"`        // 绑定地址
	AutoStart      bool   `json:"autoStart" yaml:"auto-start"`            // 开机自动启动代理
	AutoStartDelay int    `json:"autoStartDelay" yaml:"auto-start-delay"` // 开机启动延迟（秒）

	// === 运行模式 ===
	Mode     string `json:"mode" yaml:"mode"`          // rule/global/direct
	LogLevel string `json:"logLevel" yaml:"log-level"` // silent/error/warning/info/debug
	IPv6     bool   `json:"ipv6" yaml:"ipv6"`          // 启用 IPv6

	// === 性能优化 ===
	UnifiedDelay    bool   `json:"unifiedDelay" yaml:"unified-delay"`        // 统一延迟计算
	TCPConcurrent   bool   `json:"tcpConcurrent" yaml:"tcp-concurrent"`      // TCP 并发连接
	FindProcessMode string `json:"findProcessMode" yaml:"find-process-mode"` // 进程匹配模式: always/strict/off

	// === TCP Keep-Alive ===
	KeepAliveInterval int  `json:"keepAliveInterval" yaml:"keep-alive-interval"` // Keep-Alive 间隔 (秒)
	KeepAliveIdle     int  `json:"keepAliveIdle" yaml:"keep-alive-idle"`         // Keep-Alive 空闲时间 (秒)
	DisableKeepAlive  bool `json:"disableKeepAlive" yaml:"disable-keep-alive"`   // 禁用 Keep-Alive

	// === TLS ===
	GlobalClientFingerprint string `json:"globalClientFingerprint" yaml:"global-client-fingerprint"` // TLS 指纹
	SkipCertVerify          bool   `json:"skipCertVerify" yaml:"skip-cert-verify"`                   // 跳过证书验证

	// === GEO 数据 ===
	GeodataMode       bool   `json:"geodataMode" yaml:"geodata-mode"`              // 使用 dat 格式
	GeodataLoader     string `json:"geodataLoader" yaml:"geodata-loader"`          // 加载器: standard/memconservative
	GeositeMatcher    string `json:"geositeMatcher" yaml:"geosite-matcher"`        // 匹配器: hybrid/succinct
	GeoAutoUpdate     bool   `json:"geoAutoUpdate" yaml:"geo-auto-update"`         // 自动更新
	GeoUpdateInterval int    `json:"geoUpdateInterval" yaml:"geo-update-interval"` // 更新间隔 (小时)

	// === 外部资源 ===
	GlobalUA    string `json:"globalUa" yaml:"global-ua"`       // 下载外部资源的 UA
	ETagSupport bool   `json:"etagSupport" yaml:"etag-support"` // ETag 缓存支持

	// === 网络接口 ===
	InterfaceName string `json:"interfaceName" yaml:"interface-name"` // 出站接口
	RoutingMark   int    `json:"routingMark" yaml:"routing-mark"`     // 路由标记 (Linux)

	// === DNS 设置 ===
	DNS DNSSettings `json:"dns" yaml:"dns"`

	// === TUN 设置 ===
	TUN TUNSettings `json:"tun" yaml:"tun"`

	// === 嗅探设置 ===
	Sniffer SnifferSettings `json:"sniffer" yaml:"sniffer"`
}

// DNSSettings DNS 设置
type DNSSettings struct {
	Enable         bool   `json:"enable" yaml:"enable"`
	Listen         string `json:"listen" yaml:"listen"`
	PreferH3       bool   `json:"preferH3" yaml:"prefer-h3"`             // 优先 HTTP/3
	CacheAlgorithm string `json:"cacheAlgorithm" yaml:"cache-algorithm"` // lru/arc
	IPv6           bool   `json:"ipv6" yaml:"ipv6"`
	UseHosts       bool   `json:"useHosts" yaml:"use-hosts"`
	UseSystemHosts bool   `json:"useSystemHosts" yaml:"use-system-hosts"`
	RespectRules   bool   `json:"respectRules" yaml:"respect-rules"` // DNS 查询遵循代理规则

	// Fake-IP 模式
	EnhancedMode     string   `json:"enhancedMode" yaml:"enhanced-mode"` // fake-ip/redir-host
	FakeIPRange      string   `json:"fakeIpRange" yaml:"fake-ip-range"`
	FakeIPRange6     string   `json:"fakeIpRange6" yaml:"fake-ip-range6"`
	FakeIPFilterMode string   `json:"fakeIpFilterMode" yaml:"fake-ip-filter-mode"` // blacklist/whitelist
	FakeIPFilter     []string `json:"fakeIpFilter" yaml:"fake-ip-filter"`

	// DNS 服务器
	DefaultNameserver     []string `json:"defaultNameserver" yaml:"default-nameserver"`
	Nameserver            []string `json:"nameserver" yaml:"nameserver"`
	Fallback              []string `json:"fallback" yaml:"fallback"`
	ProxyServerNameserver []string `json:"proxyServerNameserver" yaml:"proxy-server-nameserver"`
	DirectNameserver      []string `json:"directNameserver" yaml:"direct-nameserver"`
}

// TUNSettings TUN 设置
type TUNSettings struct {
	Enable              bool     `json:"enable" yaml:"enable"`
	Device              string   `json:"device" yaml:"device"`
	Stack               string   `json:"stack" yaml:"stack"` // system/gvisor/mixed
	MTU                 int      `json:"mtu" yaml:"mtu"`
	GSO                 bool     `json:"gso" yaml:"gso"` // Linux GSO
	GSOMaxSize          int      `json:"gsoMaxSize" yaml:"gso-max-size"`
	AutoRoute           bool     `json:"autoRoute" yaml:"auto-route"`
	AutoRedirect        bool     `json:"autoRedirect" yaml:"auto-redirect"` // Linux 自动重定向
	AutoDetectInterface bool     `json:"autoDetectInterface" yaml:"auto-detect-interface"`
	StrictRoute         bool     `json:"strictRoute" yaml:"strict-route"`
	UDPTimeout          int      `json:"udpTimeout" yaml:"udp-timeout"`
	DNSHijack           []string `json:"dnsHijack" yaml:"dns-hijack"`

	// 高级路由
	EndpointIndependentNat bool     `json:"endpointIndependentNat" yaml:"endpoint-independent-nat"`
	RouteAddress           []string `json:"routeAddress" yaml:"route-address"`
	RouteExcludeAddress    []string `json:"routeExcludeAddress" yaml:"route-exclude-address"`

	// Linux iproute2
	Iproute2TableIndex int `json:"iproute2TableIndex" yaml:"iproute2-table-index"`
	Iproute2RuleIndex  int `json:"iproute2RuleIndex" yaml:"iproute2-rule-index"`
}

// SnifferSettings 嗅探设置
type SnifferSettings struct {
	Enable          bool     `json:"enable" yaml:"enable"`
	ForceDNSMapping bool     `json:"forceDnsMapping" yaml:"force-dns-mapping"`
	ParsePureIP     bool     `json:"parsePureIp" yaml:"parse-pure-ip"`
	OverrideDest    bool     `json:"overrideDest" yaml:"override-destination"`
	SniffHTTP       bool     `json:"sniffHttp" yaml:"sniff-http"`
	SniffTLS        bool     `json:"sniffTls" yaml:"sniff-tls"`
	SniffQUIC       bool     `json:"sniffQuic" yaml:"sniff-quic"`
	SkipDomain      []string `json:"skipDomain" yaml:"skip-domain"`
}

// GetDefaultProxySettings 获取默认代理设置 (Linux 网关最优配置)
func GetDefaultProxySettings() *ProxySettings {
	return &ProxySettings{
		// 端口设置 (默认只启用混合端口)
		MixedPortEnabled:  true,
		MixedPort:         7890,
		SocksPortEnabled:  false,
		SocksPort:         7891,
		HTTPPortEnabled:   false,
		HTTPPort:          7892,
		RedirPortEnabled:  false,
		RedirPort:         7893,
		TProxyPortEnabled: false,
		TProxyPort:        7894,

		// 认证 (默认为空，不启用认证)
		Authentication: []AuthUser{},

		// 基础设置
		AllowLan:       true,
		BindAddress:    "*",
		AutoStart:      false,
		AutoStartDelay: 15, // 默认延迟15秒启动

		// 运行模式
		Mode:     "rule",
		LogLevel: "info",
		IPv6:     false,

		// 性能优化
		UnifiedDelay:    true,  // 更准确的延迟测试
		TCPConcurrent:   true,  // 并发连接，使用最快的 IP
		FindProcessMode: "off", // 网关模式关闭进程匹配

		// TCP Keep-Alive
		KeepAliveInterval: 30,
		KeepAliveIdle:     60,
		DisableKeepAlive:  false,

		// TLS
		GlobalClientFingerprint: "chrome",
		SkipCertVerify:          false,

		// GEO 数据
		GeodataMode:       true,       // dat 格式查询更快
		GeodataLoader:     "standard", // 内存充足用 standard
		GeositeMatcher:    "succinct", // 高效匹配器
		GeoAutoUpdate:     true,
		GeoUpdateInterval: 24,

		// 外部资源
		GlobalUA:    "clash.meta",
		ETagSupport: true,

		// 网络接口
		InterfaceName: "",
		RoutingMark:   0,

		// DNS 设置
		DNS: DNSSettings{
			Enable:           true,
			Listen:           "0.0.0.0:1053",
			PreferH3:         true,
			CacheAlgorithm:   "arc",
			IPv6:             false,
			UseHosts:         true,
			UseSystemHosts:   false,
			RespectRules:     true,
			EnhancedMode:     "fake-ip",
			FakeIPRange:      "198.18.0.1/16",
			FakeIPRange6:     "fc00::/18",
			FakeIPFilterMode: "blacklist",
			FakeIPFilter: []string{
				"geosite:cn",
				"geosite:private",
				"*.lan",
				"*.local",
				"*.localhost",
			},
			DefaultNameserver:     []string{"223.5.5.5", "119.29.29.29"},
			Nameserver:            []string{"https://dns.google/dns-query", "https://cloudflare-dns.com/dns-query"},
			Fallback:              []string{"https://dns.google/dns-query", "https://cloudflare-dns.com/dns-query"},
			ProxyServerNameserver: []string{"223.5.5.5", "119.29.29.29"},
			DirectNameserver:      []string{"223.5.5.5", "119.29.29.29"},
		},

		// TUN 设置
		TUN: TUNSettings{
			Enable:                 false, // 默认关闭，需要 root 权限
			Device:                 "mihomo",
			Stack:                  "mixed",
			MTU:                    9000,
			GSO:                    true,
			GSOMaxSize:             65536,
			AutoRoute:              true,
			AutoRedirect:           true,
			AutoDetectInterface:    true,
			StrictRoute:            true,
			UDPTimeout:             300,
			DNSHijack:              []string{"any:53", "tcp://any:53"},
			EndpointIndependentNat: true,
			RouteAddress:           []string{},
			RouteExcludeAddress: []string{
				"192.168.0.0/16",
				"10.0.0.0/8",
				"172.16.0.0/12",
				"127.0.0.0/8",
				"fc00::/7",
				"fe80::/10",
			},
			Iproute2TableIndex: 2022,
			Iproute2RuleIndex:  9000,
		},

		// 嗅探设置
		Sniffer: SnifferSettings{
			Enable:          true,
			ForceDNSMapping: true,
			ParsePureIP:     true,
			OverrideDest:    true,
			SniffHTTP:       true,
			SniffTLS:        true,
			SniffQUIC:       true,
			SkipDomain:      []string{"+.push.apple.com"},
		},
	}
}
