package system

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
)

// Handler 系统管理 API 处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器
func NewHandler(dataDir string) *Handler {
	return &Handler{
		service: NewService(dataDir),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/config", h.GetConfig)
	r.GET("/resources", h.GetResources)
	r.PUT("/autostart", h.SetAutoStart)
	r.PUT("/ipforward", h.SetIPForward)
	r.PUT("/bbr", h.SetBBR)
	r.PUT("/tunoptimize", h.SetTUNOptimize)
	r.POST("/optimize-all", h.OptimizeAll)
	// 系统代理
	r.POST("/proxy/enable", h.EnableSystemProxy)
	r.POST("/proxy/disable", h.DisableSystemProxy)
	r.GET("/proxy/status", h.GetSystemProxyStatus)
	// 浏览器代理
	r.GET("/browsers", h.GetBrowsers)
	r.POST("/browsers/firefox/configure", h.ConfigureFirefox)
	r.POST("/browsers/firefox/clear", h.ClearFirefox)
	// 出口 IP 信息
	r.GET("/geoip", h.GetGeoIP)
}

// GetResources 获取系统资源信息
func (h *Handler) GetResources(c *gin.Context) {
	resources := h.service.GetResources()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resources,
	})
}

// GetConfig 获取系统配置
func (h *Handler) GetConfig(c *gin.Context) {
	config := h.service.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// SetAutoStart 设置开机自启
func (h *Handler) SetAutoStart(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.SetAutoStart(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	msg := "开机自启已关闭"
	if req.Enabled {
		msg = "开机自启已开启"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

// SetIPForward 设置 IP 转发
func (h *Handler) SetIPForward(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.SetIPForward(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	msg := "IP 转发已关闭"
	if req.Enabled {
		msg = "IP 转发已开启"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

// SetBBR 设置 BBR
func (h *Handler) SetBBR(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.SetBBR(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	msg := "BBR 拥塞控制已关闭"
	if req.Enabled {
		msg = "BBR 拥塞控制已开启"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

// SetTUNOptimize 设置 TUN 优化
func (h *Handler) SetTUNOptimize(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.SetTUNOptimize(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	msg := "TUN 优化已关闭"
	if req.Enabled {
		msg = "TUN 网络优化已开启"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

// OptimizeAll 一键优化
func (h *Handler) OptimizeAll(c *gin.Context) {
	if err := h.service.ApplyAllOptimizations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已开启所有优化：IP转发、BBR、TUN优化",
	})
}

// EnableSystemProxy 启用系统代理
func (h *Handler) EnableSystemProxy(c *gin.Context) {
	var req struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if req.Host == "" {
		req.Host = "127.0.0.1"
	}
	if req.Port == 0 {
		req.Port = 7890
	}

	if err := SetSystemProxy(req.Host, req.Port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "系统代理已启用",
	})
}

// DisableSystemProxy 禁用系统代理
func (h *Handler) DisableSystemProxy(c *gin.Context) {
	if err := ClearSystemProxy(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "系统代理已禁用",
	})
}

// GetSystemProxyStatus 获取系统代理状态
func (h *Handler) GetSystemProxyStatus(c *gin.Context) {
	enabled, host, port, err := GetSystemProxyStatus()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"enabled": false,
				"host":    "",
				"port":    0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"enabled": enabled,
			"host":    host,
			"port":    port,
		},
	})
}

// GeoIPInfo 出口 IP 信息
type GeoIPInfo struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
}

// 常用国家/地区名称到 ISO 3166-1 alpha-2 代码映射
var countryCodeMap = map[string]string{
	"中国": "CN", "香港": "HK", "台湾": "TW", "澳门": "MO",
	"新加坡": "SG", "日本": "JP", "韩国": "KR", "美国": "US",
	"英国": "GB", "德国": "DE", "法国": "FR", "俄罗斯": "RU",
	"加拿大": "CA", "澳大利亚": "AU", "印度": "IN", "巴西": "BR",
	"荷兰": "NL", "瑞士": "CH", "瑞典": "SE", "挪威": "NO",
	"芬兰": "FI", "丹麦": "DK", "意大利": "IT", "西班牙": "ES",
	"葡萄牙": "PT", "波兰": "PL", "土耳其": "TR", "泰国": "TH",
	"越南": "VN", "马来西亚": "MY", "印度尼西亚": "ID", "菲律宾": "PH",
	"阿联酋": "AE", "沙特阿拉伯": "SA", "以色列": "IL", "南非": "ZA",
	"墨西哥": "MX", "阿根廷": "AR", "智利": "CL", "新西兰": "NZ",
	"爱尔兰": "IE", "比利时": "BE", "奥地利": "AT", "捷克": "CZ",
	"匈牙利": "HU", "罗马尼亚": "RO", "乌克兰": "UA", "希腊": "GR",
	"埃及": "EG", "尼日利亚": "NG", "肯尼亚": "KE", "哥伦比亚": "CO",
	"秘鲁": "PE", "委内瑞拉": "VE", "巴基斯坦": "PK", "孟加拉国": "BD",
	"斯里兰卡": "LK", "缅甸": "MM", "柬埔寨": "KH", "老挝": "LA",
}

// getCountryCode 根据国家名获取国家代码
func getCountryCode(country string) string {
	// 直接匹配
	if code, ok := countryCodeMap[country]; ok {
		return code
	}
	// 部分匹配（处理"中国香港"、"中国台湾"等情况）
	for name, code := range countryCodeMap {
		if len(name) > 0 && len(country) > 0 && (contains(country, name) || contains(name, country)) {
			return code
		}
	}
	return ""
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createProxyClient 创建通过 Mihomo 代理的 HTTP 客户端
func createProxyClient() *http.Client {
	// 尝试通过 SOCKS5 代理（端口 7891）
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:7890", nil, proxy.Direct)
	if err != nil {
		// 如果 SOCKS5 失败，尝试 HTTP 代理
		proxyURL, _ := url.Parse("http://127.0.0.1:7890")
		return &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
}

// GetGeoIP 获取出口 IP 信息（通过 Mihomo 代理请求）
func (h *Handler) GetGeoIP(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	// 通过代理请求，获取代理出口的真实 IP
	client := createProxyClient()
	var geoInfo GeoIPInfo

	if lang == "zh" {
		// 中文 API
		resp, err := client.Get("https://api.vore.top/api/IPdata")
		if err == nil {
			defer resp.Body.Close()
			var data struct {
				Code   int `json:"code"`
				IPInfo struct {
					Text string `json:"text"`
				} `json:"ipinfo"`
				IPData struct {
					IP    string `json:"ip"`
					Info1 string `json:"info1"`
					Info2 string `json:"info2"`
					Info3 string `json:"info3"`
				} `json:"ipdata"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil && data.Code == 200 {
				geoInfo.IP = data.IPInfo.Text
				if geoInfo.IP == "" {
					geoInfo.IP = data.IPData.IP
				}
				geoInfo.Country = data.IPData.Info1
				// 根据国家名获取国家代码
				geoInfo.CountryCode = getCountryCode(data.IPData.Info1)
				geoInfo.ISP = data.IPData.Info2
				if geoInfo.ISP == "" {
					geoInfo.ISP = data.IPData.Info3
				}
			}
		}
	} else {
		// 英文 API (使用 ipwho.is)
		resp, err := client.Get("https://ipwho.is/")
		if err == nil {
			defer resp.Body.Close()
			var data struct {
				Success     bool   `json:"success"`
				IP          string `json:"ip"`
				Country     string `json:"country"`
				CountryCode string `json:"country_code"`
				Region      string `json:"region"`
				City        string `json:"city"`
				ISP         string `json:"connection.isp"`
				Org         string `json:"connection.org"`
				Connection  struct {
					ISP string `json:"isp"`
					Org string `json:"org"`
				} `json:"connection"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil && data.Success {
				geoInfo.IP = data.IP
				geoInfo.Country = data.Country
				geoInfo.CountryCode = data.CountryCode
				geoInfo.Region = data.Region
				geoInfo.City = data.City
				geoInfo.ISP = data.Connection.ISP
				if geoInfo.ISP == "" {
					geoInfo.ISP = data.Connection.Org
				}
				geoInfo.Org = data.Connection.Org
			}
		}
	}

	// 如果主 API 失败，尝试备用 API
	if geoInfo.IP == "" {
		resp, err := client.Get("https://ipinfo.io/json")
		if err == nil {
			defer resp.Body.Close()
			var data struct {
				IP      string `json:"ip"`
				Country string `json:"country"`
				Region  string `json:"region"`
				City    string `json:"city"`
				Org     string `json:"org"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil {
				geoInfo.IP = data.IP
				geoInfo.CountryCode = data.Country
				geoInfo.Region = data.Region
				geoInfo.City = data.City
				geoInfo.Org = data.Org
				geoInfo.ISP = data.Org
			}
		}
	}

	if geoInfo.IP == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "获取 IP 信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    geoInfo,
	})
}

// GetBrowsers 获取已安装的浏览器列表
func (h *Handler) GetBrowsers(c *gin.Context) {
	browsers := GetInstalledBrowsers()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    browsers,
	})
}

// ConfigureFirefox 配置 Firefox 使用系统代理
func (h *Handler) ConfigureFirefox(c *gin.Context) {
	if err := ConfigureFirefoxProxy(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Firefox 已配置使用系统代理，请重启浏览器",
	})
}

// ClearFirefox 清除 Firefox 代理配置
func (h *Handler) ClearFirefox(c *gin.Context) {
	if err := ClearFirefoxProxy(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Firefox 代理配置已清除",
	})
}
