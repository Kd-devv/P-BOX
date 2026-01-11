package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(dataDir string) *Handler {
	h := &Handler{
		service: NewService(dataDir),
	}
	// 启动后延迟 5 秒自动检测版本和下载核心
	h.service.Initialize(5)
	return h
}

// GetService 获取服务实例
func (h *Handler) GetService() *Service {
	return h.service
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/status", h.GetStatus)
	r.GET("/versions", h.GetVersions)
	r.POST("/versions/refresh", h.RefreshVersions)
	r.GET("/platform", h.GetPlatformInfo)
	r.POST("/switch", h.SwitchCore)
	r.POST("/download/:core", h.DownloadCore)
	r.GET("/download/:core/progress", h.GetDownloadProgress)
}

func (h *Handler) GetStatus(c *gin.Context) {
	status := h.service.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    status,
	})
}

func (h *Handler) GetVersions(c *gin.Context) {
	versions, err := h.service.GetLatestVersions()
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
		"data":    versions,
	})
}

func (h *Handler) SwitchCore(c *gin.Context) {
	var req struct {
		CoreType string `json:"coreType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SwitchCore(req.CoreType); err != nil {
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

func (h *Handler) DownloadCore(c *gin.Context) {
	coreType := c.Param("core")
	go h.service.DownloadCore(coreType)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "download started",
	})
}

func (h *Handler) GetDownloadProgress(c *gin.Context) {
	coreType := c.Param("core")
	progress := h.service.GetDownloadProgress(coreType)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    progress,
	})
}

// RefreshVersions 手动刷新版本信息
func (h *Handler) RefreshVersions(c *gin.Context) {
	versions, err := h.service.RefreshVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "版本信息已刷新并保存",
		"data":    versions,
	})
}

// GetPlatformInfo 获取平台信息
func (h *Handler) GetPlatformInfo(c *gin.Context) {
	info := h.service.GetPlatformInfo()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    info,
	})
}
