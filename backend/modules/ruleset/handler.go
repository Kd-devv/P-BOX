package ruleset

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 规则集 API 处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/ruleset")
	{
		group.GET("/geo", h.GetGeoFiles)
		group.GET("/providers", h.GetProviderFiles)
		group.GET("/config", h.GetConfig)
		group.PUT("/config", h.SetConfig)
		group.POST("/update", h.UpdateAll)
		group.GET("/status", h.GetStatus)
	}
}

// GetGeoFiles 获取 GEO 数据文件列表
func (h *Handler) GetGeoFiles(c *gin.Context) {
	files := h.service.GetGeoFiles()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    files,
	})
}

// GetProviderFiles 获取规则提供者文件列表
func (h *Handler) GetProviderFiles(c *gin.Context) {
	files := h.service.GetRuleProviderFiles()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    files,
	})
}

// GetConfig 获取配置
func (h *Handler) GetConfig(c *gin.Context) {
	config := h.service.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// SetConfig 设置配置
func (h *Handler) SetConfig(c *gin.Context) {
	var config RuleSetConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.SetConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "保存配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateAll 更新所有规则文件
func (h *Handler) UpdateAll(c *gin.Context) {
	if h.service.IsUpdating() {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "正在更新中，请稍后再试",
		})
		return
	}

	// 解析请求参数
	var req struct {
		GitHubProxy string `json:"githubProxy"`
	}
	c.ShouldBindJSON(&req)

	// 异步更新
	go func() {
		h.service.DownloadAllFilesWithProxy(req.GitHubProxy)
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "开始更新规则文件",
	})
}

// GetStatus 获取更新状态
func (h *Handler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"updating":   h.service.IsUpdating(),
			"lastUpdate": h.service.GetConfig().LastUpdate,
		},
	})
}
