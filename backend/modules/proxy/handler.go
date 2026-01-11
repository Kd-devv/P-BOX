package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(dataDir string) *Handler {
	return &Handler{
		service: NewService(dataDir),
	}
}

// GetService 获取服务实例
func (h *Handler) GetService() *Service {
	return h.service
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/status", h.GetStatus)
	r.POST("/start", h.Start)
	r.POST("/stop", h.Stop)
	r.POST("/restart", h.Restart)
	r.PUT("/mode", h.SetMode)
	r.PUT("/tun", h.SetTunMode)
	r.PUT("/transparent", h.SetTransparentMode) // 透明代理模式切换
	r.GET("/config", h.GetConfig)
	r.PUT("/config", h.UpdateConfig)
	r.POST("/generate", h.GenerateConfig)
	r.GET("/config/preview", h.GetConfigPreview)
	r.GET("/logs", h.GetLogs)

	// 配置模板管理
	r.GET("/template", h.GetConfigTemplate)
	r.PUT("/template/groups", h.UpdateProxyGroups)
	r.PUT("/template/rules", h.UpdateRules)
	r.PUT("/template/providers", h.UpdateRuleProviders)
	r.POST("/template/reset", h.ResetTemplate)

	// Sing-Box 配置生成
	r.POST("/singbox/generate", h.GenerateSingBoxConfig)
	r.GET("/singbox/preview", h.GetSingBoxConfigPreview)
	r.GET("/singbox/download", h.DownloadSingBoxConfig)

	// Sing-Box 模板管理
	r.GET("/singbox/template", h.GetSingBoxTemplate)
	r.PUT("/singbox/template", h.UpdateSingBoxTemplate)
	r.POST("/singbox/template/reset", h.ResetSingBoxTemplate)

	// Mihomo API 代理 (避免 CORS 问题)
	r.GET("/mihomo/proxies", h.ProxyMihomoGetProxies)
	r.GET("/mihomo/proxies/:name", h.ProxyMihomoGetProxy)
	r.PUT("/mihomo/proxies/:name", h.ProxyMihomoSelectProxy)
	r.GET("/mihomo/proxies/:name/delay", h.ProxyMihomoTestDelay)
}

func (h *Handler) GetStatus(c *gin.Context) {
	status := h.service.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    status,
	})
}

func (h *Handler) Start(c *gin.Context) {
	if err := h.service.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *Handler) Stop(c *gin.Context) {
	if err := h.service.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *Handler) Restart(c *gin.Context) {
	if err := h.service.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *Handler) SetMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SetMode(req.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *Handler) SetTunMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SetTunEnabled(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// SetTransparentMode 设置透明代理模式
// mode: off (关闭), tun (TUN模式), tproxy (TPROXY透明代理), redirect (REDIRECT重定向)
func (h *Handler) SetTransparentMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SetTransparentMode(req.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// 返回模式说明
	modeDesc := map[string]string{
		"off":      "已关闭透明代理",
		"tun":      "TUN 模式已开启，需要 root 权限",
		"tproxy":   "TPROXY 模式已开启，需配置 iptables",
		"redirect": "REDIRECT 模式已开启，需配置 iptables",
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": modeDesc[req.Mode],
		"data": gin.H{
			"mode": req.Mode,
		},
	})
}

func (h *Handler) GetConfig(c *gin.Context) {
	config := h.service.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	// 使用 map 接收部分更新
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.PatchConfig(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *Handler) GenerateConfig(c *gin.Context) {
	var req struct {
		Nodes []ProxyNode `json:"nodes"`
	}
	// 允许空 body，此时自动获取节点
	c.ShouldBindJSON(&req)

	var configPath string
	var err error

	if len(req.Nodes) == 0 {
		// 没有传节点，调用 regenerateConfig 自动获取所有节点
		configPath, err = h.service.RegenerateConfig()
	} else {
		// 使用传入的节点
		configPath, err = h.service.GenerateConfig(req.Nodes)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"configPath": configPath,
		},
	})
}

// GetConfigPreview 获取生成的 config.yaml 内容用于预览
func (h *Handler) GetConfigPreview(c *gin.Context) {
	content, err := h.service.GetConfigContent()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"content": "// 配置文件未生成，请先点击「生成配置」按钮",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"content": content,
		},
	})
}

func (h *Handler) GetLogs(c *gin.Context) {
	// 获取参数
	limitStr := c.DefaultQuery("limit", "200")
	level := c.DefaultQuery("level", "all") // all, info, warn, error

	limit := 200
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	logs := h.service.GetLogs(limit)

	// 根据级别过滤
	var filteredLogs []string
	for _, log := range logs {
		switch level {
		case "error":
			if strings.Contains(log, "ERR") || strings.Contains(log, "FATA") || strings.Contains(log, "error") {
				filteredLogs = append(filteredLogs, log)
			}
		case "warn":
			if strings.Contains(log, "WARN") || strings.Contains(log, "warning") {
				filteredLogs = append(filteredLogs, log)
			}
		case "info":
			if strings.Contains(log, "INFO") || strings.Contains(log, "info") {
				filteredLogs = append(filteredLogs, log)
			}
		default:
			filteredLogs = append(filteredLogs, log)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    filteredLogs,
	})
}

// GetConfigTemplate 获取配置模板
func (h *Handler) GetConfigTemplate(c *gin.Context) {
	template := h.service.GetConfigTemplate()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    template,
	})
}

// UpdateProxyGroups 更新代理组
func (h *Handler) UpdateProxyGroups(c *gin.Context) {
	var groups []ProxyGroupTemplate
	if err := c.ShouldBindJSON(&groups); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateProxyGroups(groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateRules 更新规则
func (h *Handler) UpdateRules(c *gin.Context) {
	var rules []RuleTemplate
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateRules(rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateRuleProviders 更新规则提供者
func (h *Handler) UpdateRuleProviders(c *gin.Context) {
	var providers []RuleProviderTemplate
	if err := c.ShouldBindJSON(&providers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateRuleProviders(providers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ResetTemplate 重置配置模板为默认值
func (h *Handler) ResetTemplate(c *gin.Context) {
	h.service.ResetConfigTemplate()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ========== Mihomo API 代理 (避免 CORS 问题) ==========

// ProxyMihomoGetProxies 代理获取所有代理组
func (h *Handler) ProxyMihomoGetProxies(c *gin.Context) {
	apiAddr := h.service.GetConfig().ExternalController
	if apiAddr == "" {
		apiAddr = "127.0.0.1:9090"
	}

	resp, err := http.Get("http://" + apiAddr + "/proxies")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    1,
			"message": "Mihomo API 不可用: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// ProxyMihomoGetProxy 代理获取单个代理组
func (h *Handler) ProxyMihomoGetProxy(c *gin.Context) {
	name := c.Param("name")
	apiAddr := h.service.GetConfig().ExternalController
	if apiAddr == "" {
		apiAddr = "127.0.0.1:9090"
	}

	resp, err := http.Get("http://" + apiAddr + "/proxies/" + name)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    1,
			"message": "Mihomo API 不可用: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// ProxyMihomoSelectProxy 代理切换节点
func (h *Handler) ProxyMihomoSelectProxy(c *gin.Context) {
	name := c.Param("name")
	apiAddr := h.service.GetConfig().ExternalController
	if apiAddr == "" {
		apiAddr = "127.0.0.1:9090"
	}

	body, _ := io.ReadAll(c.Request.Body)
	req, _ := http.NewRequest("PUT", "http://"+apiAddr+"/proxies/"+name, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    1,
			"message": "Mihomo API 不可用: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

// ProxyMihomoTestDelay 代理测试节点延迟
func (h *Handler) ProxyMihomoTestDelay(c *gin.Context) {
	name := c.Param("name")
	url := c.Query("url")
	timeout := c.Query("timeout")

	if url == "" {
		url = "http://www.gstatic.com/generate_204"
	}
	if timeout == "" {
		timeout = "5000"
	}

	apiAddr := h.service.GetConfig().ExternalController
	if apiAddr == "" {
		apiAddr = "127.0.0.1:9090"
	}

	targetURL := fmt.Sprintf("http://%s/proxies/%s/delay?url=%s&timeout=%s", apiAddr, name, url, timeout)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    1,
			"message": "Mihomo API 不可用: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// ========== Sing-Box 1.12+ 配置生成 ==========

// GenerateSingBoxConfig 生成 Sing-Box 1.12+ 配置
func (h *Handler) GenerateSingBoxConfig(c *gin.Context) {
	var req struct {
		Mode           string `json:"mode"`           // tun, system
		FakeIP         bool   `json:"fakeip"`         // 启用 FakeIP
		MixedPort      int    `json:"mixedPort"`      // 混合代理端口
		HTTPPort       int    `json:"httpPort"`       // HTTP 代理端口
		SocksPort      int    `json:"socksPort"`      // SOCKS5 代理端口
		ClashAPIAddr   string `json:"clashApiAddr"`   // Clash API 地址
		ClashAPISecret string `json:"clashApiSecret"` // Clash API 密钥
		TUNStack       string `json:"tunStack"`       // TUN 栈类型
		TUNMTU         int    `json:"tunMtu"`         // TUN MTU
		DNSStrategy    string `json:"dnsStrategy"`    // DNS 策略
		LogLevel       string `json:"logLevel"`       // 日志级别
		// 性能优化
		AutoRedirect             bool `json:"autoRedirect"`             // Linux nftables
		StrictRoute              bool `json:"strictRoute"`              // 严格路由
		TCPFastOpen              bool `json:"tcpFastOpen"`              // TCP Fast Open
		TCPMultiPath             bool `json:"tcpMultiPath"`             // TCP Multi Path
		UDPFragment              bool `json:"udpFragment"`              // UDP 分片
		Sniff                    bool `json:"sniff"`                    // 流量嗅探
		SniffOverrideDestination bool `json:"sniffOverrideDestination"` // 覆盖目标地址
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用默认值
		req.Mode = "tun"
		req.MixedPort = 7890
	}

	// 构建选项
	opts := SingBoxGeneratorOptions{
		Mode:                     req.Mode,
		FakeIP:                   req.FakeIP,
		MixedPort:                req.MixedPort,
		HTTPPort:                 req.HTTPPort,
		SocksPort:                req.SocksPort,
		ClashAPIAddr:             req.ClashAPIAddr,
		ClashAPISecret:           req.ClashAPISecret,
		TUNStack:                 req.TUNStack,
		TUNMTU:                   req.TUNMTU,
		DNSStrategy:              req.DNSStrategy,
		LogLevel:                 req.LogLevel,
		AutoRedirect:             req.AutoRedirect,
		StrictRoute:              req.StrictRoute,
		TCPFastOpen:              req.TCPFastOpen,
		TCPMultiPath:             req.TCPMultiPath,
		UDPFragment:              req.UDPFragment,
		Sniff:                    req.Sniff,
		SniffOverrideDestination: req.SniffOverrideDestination,
	}

	// 获取所有节点
	nodes, err := h.service.GetAllNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "获取节点失败: " + err.Error(),
		})
		return
	}

	// 生成配置
	generator := NewSingboxGenerator(h.service.dataDir)
	config, err := generator.GenerateConfigV112(nodes, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "生成配置失败: " + err.Error(),
		})
		return
	}

	// 保存配置
	filePath, err := generator.SaveConfigV112(config, "singbox-config")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "保存配置失败: " + err.Error(),
		})
		return
	}

	// 使用 sing-box check 验证配置
	singboxPath := filepath.Join(h.service.dataDir, "cores", "sing-box")
	if _, err := os.Stat(singboxPath); err == nil {
		// sing-box 存在，进行配置验证
		checkCmd := exec.Command(singboxPath, "check", "-c", filePath)
		output, checkErr := checkCmd.CombinedOutput()
		if checkErr != nil {
			// 验证失败，返回错误信息
			errorMsg := string(output)
			if errorMsg == "" {
				errorMsg = checkErr.Error()
			}
			c.JSON(http.StatusOK, gin.H{
				"code":    2, // 使用 code 2 表示配置验证失败
				"message": "配置验证失败",
				"data": gin.H{
					"configPath":      filePath,
					"nodeCount":       len(nodes),
					"mode":            opts.Mode,
					"validationError": errorMsg,
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"configPath": filePath,
			"nodeCount":  len(nodes),
			"mode":       opts.Mode,
		},
	})
}

// GetSingBoxConfigPreview 获取 Sing-Box 配置预览
func (h *Handler) GetSingBoxConfigPreview(c *gin.Context) {
	content, err := h.service.GetSingBoxConfigContent()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"content": "// Sing-Box 配置文件未生成，请先点击「生成配置」按钮",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"content": content,
		},
	})
}

// DownloadSingBoxConfig 下载 Sing-Box 配置文件
func (h *Handler) DownloadSingBoxConfig(c *gin.Context) {
	content, err := h.service.GetSingBoxConfigContent()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1,
			"message": "配置文件不存在",
		})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=singbox-config.json")
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, content)
}

// GetSingBoxTemplate 获取 Sing-Box 模板配置
func (h *Handler) GetSingBoxTemplate(c *gin.Context) {
	template := h.service.GetSingBoxTemplate()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    template,
	})
}

// UpdateSingBoxTemplate 更新 Sing-Box 模板配置
func (h *Handler) UpdateSingBoxTemplate(c *gin.Context) {
	var template SingBoxTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateSingBoxTemplate(&template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "保存失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ResetSingBoxTemplate 重置 Sing-Box 模板为默认值
func (h *Handler) ResetSingBoxTemplate(c *gin.Context) {
	h.service.ResetSingBoxTemplate()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.service.GetSingBoxTemplate(),
	})
}
