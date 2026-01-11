package subscription

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// VLESSConfig VLESS配置结构
type VLESSConfig struct {
	UUID           string                 `json:"uuid"`
	Flow           string                 `json:"flow,omitempty"`
	Encryption     string                 `json:"-"` // 🔥 不序列化到JSON，sing-box 不支持此字段
	Network        string                 `json:"network,omitempty"`
	PacketEncoding string                 `json:"packet_encoding,omitempty"`
	TLS            *TLSConfig             `json:"tls,omitempty"`
	Reality        *RealityConfig         `json:"reality,omitempty"`
	Transport      map[string]interface{} `json:"transport,omitempty"`
	ServerName     string                 `json:"server_name,omitempty"`
}

// RealityConfig Reality配置
type RealityConfig struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key"`
	ShortID   string `json:"short_id,omitempty"`
	SpiderX   string `json:"spider_x,omitempty"`
}

// ParseVLESSURL 解析VLESS链接
// 格式: vless://uuid@server:port?params#name
// 完整对齐 Python 实现和 Swift 实现
func ParseVLESSURL(vlessURL string) (*ProxyNode, error) {
	if !strings.HasPrefix(vlessURL, "vless://") {
		return nil, errors.New("无效的VLESS URL")
	}

	// 移除协议前缀
	urlStr := strings.TrimPrefix(vlessURL, "vless://")

	// 分离UUID和服务器部分
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("vless URL格式错误")
	}

	uuid := parts[0]
	remaining := parts[1]

	// 分离服务器和参数
	serverAndParams := strings.SplitN(remaining, "?", 2)
	serverPart := serverAndParams[0]
	queryString := ""
	if len(serverAndParams) == 2 {
		queryString = serverAndParams[1]
	}

	// 提取节点名称
	name := "VLESS"
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
		return nil, errors.New("vless服务器地址格式错误")
	}

	server := serverParts[0]
	port := ParseInt(serverParts[1], 443)

	// 解析查询参数
	params := ParseQueryParams(queryString)

	// 🔥 完整实现：flow 和 encryption 处理
	// VLESS 标准协议中 encryption 永远是 "none"
	// 某些实现（如 Xray）可能在 URL 中包含 Post-Quantum 密钥:
	// encryption=mlkem768x25519plus.native.0rtt.xxx
	//
	// ❌ sing-box 不支持 encryption 字段中的 Post-Quantum 密钥
	// ✅ 我们只关注 flow 字段，不将 encryption 映射到 flow
	
	encryption := params["encryption"]
	if encryption == "" {
		encryption = "none"
	}

	flowValue := params["flow"]
	
	// 🆕 过滤掉 "none" 值和空值，sing-box 不支持 flow: "none"
	if flowValue == "none" || flowValue == "" {
		flowValue = ""
	}

	// 构建配置
	config := VLESSConfig{
		UUID:       uuid,
		Encryption: encryption,
		Flow:       flowValue,
		Network:    params["type"],
	}

	// 🆕 packet_encoding 仅在没有 flow 时设置为 xudp
	if config.Flow == "" {
		config.PacketEncoding = "xudp"
	}

	// 🔧 完整实现传输层配置 (对齐 Python 实现)
	network := params["type"]
	if network == "" {
		// 🆕 支持 Shadowrocket 的 obfs 参数
		network = params["obfs"]
	}

	// 🆕 将 "raw" 视为 "tcp"（无传输层）
	// 某些客户端（如 v2rayN）使用 type=raw 表示纯 TCP 连接
	// sing-box 不认识 "raw" 类型，需要转换
	if network == "raw" || network == "none" {
		network = "tcp"
	}

	if network != "" && network != "tcp" {
		transport := make(map[string]interface{})
		transport["type"] = network

		switch network {
		case "ws", "websocket":
			// 🆕 支持 Early Data (正则匹配 ?ed=数字)
			pathValue := params["path"]
			if pathValue == "" {
				pathValue = "/"
			}
			decodedPath, err := url.QueryUnescape(pathValue)
			if err != nil {
				decodedPath = pathValue
			}

			// 🆕 正则匹配 ?ed=数字
			earlyDataRe := regexp.MustCompile(`\?ed=(\d+)$`)
			matches := earlyDataRe.FindStringSubmatch(decodedPath)
			if len(matches) > 0 {
				earlyDataSize, _ := strconv.Atoi(matches[1])
				decodedPath = earlyDataRe.ReplaceAllString(decodedPath, "")
				transport["path"] = decodedPath
				transport["early_data_header_name"] = "Sec-WebSocket-Protocol"
				transport["max_early_data"] = earlyDataSize
			} else {
				transport["path"] = decodedPath
			}

			// 🆕 Host 头优先级: host > obfsParam > peer > sni (Shadowrocket 兼容)
			hostValue := params["host"]
			if hostValue == "" {
				hostValue = params["obfsParam"]
			}
			if hostValue == "" {
				hostValue = params["peer"]
			}
			if hostValue != "" && hostValue != "None" {
				headers := map[string]interface{}{"Host": hostValue}
				transport["headers"] = headers
			}

		case "grpc":
			// gRPC service_name (直接在 transport 层级)
			if serviceName := params["serviceName"]; serviceName != "" {
				transport["service_name"] = serviceName
			}

		case "http", "h2":
			// 🔧 HTTP/H2 的 host 和 path 直接放在 transport 层级
			transport["type"] = "http" // 统一使用 "http"
			if host := params["host"]; host != "" {
				transport["host"] = []string{host}
			}
			if path := params["path"]; path != "" {
				// 🆕 去除查询参数
				pathParts := strings.Split(path, "?")
				transport["path"] = pathParts[0]
			}

		case "quic":
			// QUIC 类型
			if key := params["key"]; key != "" {
				transport["key"] = key
			}
		}

		config.Transport = transport
	}

	// 🔧 完整实现 TLS/Reality 配置 (对齐 Python 实现)
	security := params["security"]
	// 🆕 也检查 tls=1 参数 (Python 实现)
	if security == "" && params["tls"] == "1" {
		security = "tls"
	}

	// 🆕 检查是否有 Reality 参数 (pbk)
	hasReality := params["pbk"] != ""
	if hasReality {
		security = "reality"
	}

	hasTLS := security == "tls" || (security != "" && security != "none" && security != "None")
	hasRealityOrTLS := hasTLS || hasReality

	if hasRealityOrTLS {
		// 🆕 SNI 优先级: sni > peer > host (Python 实现)
		sni := params["sni"]
		if sni == "" {
			sni = params["peer"]
		}
		// 过滤 "None"
		if sni == "None" {
			sni = ""
		}

		// 🆕 insecure 支持
		insecure := params["allowInsecure"] == "1" || params["insecure"] == "1"

		// 🆕 ALPN 支持
		var alpn []string
		if alpnStr := params["alpn"]; alpnStr != "" {
			alpn = strings.Split(alpnStr, ",")
		}

		// 构建 TLS 配置
		tlsConfig := &TLSConfig{
			Enabled:    true,
			ServerName: sni,
			Insecure:   insecure,
			ALPN:       alpn,
		}

		// 🆕 uTLS 指纹支持 (TLS 协议)
		if security == "tls" {
			if fp := params["fp"]; fp != "" {
				tlsConfig.UTLS = &UTLSConfig{
					Enabled:     true,
					Fingerprint: fp,
				}
			}
		}

		// 🆕 Reality 配置 (Reality 协议)
		if hasReality {
			// Reality 也支持 fingerprint (uTLS)
			if fp := params["fp"]; fp != "" {
				tlsConfig.UTLS = &UTLSConfig{
					Enabled:     true,
					Fingerprint: fp,
				}
			}

			shortID := params["sid"]
			// 过滤 "None" 和 "null"
			if shortID == "None" || shortID == "null" || shortID == "" {
				shortID = ""
			}

			spiderX := params["spx"]
			if spiderX != "" {
				if decoded, err := url.QueryUnescape(spiderX); err == nil {
					spiderX = decoded
				}
				if spiderX == "None" || spiderX == "null" {
					spiderX = ""
				}
			}

			realityConfig := &RealityConfig{
				Enabled:   true,
				PublicKey: params["pbk"],
				ShortID:   shortID,
				SpiderX:   spiderX,
			}

			config.Reality = realityConfig
		}

		config.TLS = tlsConfig
		config.ServerName = sni
	}

	// 转换为JSON字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("vless配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "vless",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现NodeConfig接口
func (c *VLESSConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
