package node

import (
	"net/http"
	"time"

	"p-box/backend/modules/subscription"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(dataDir string, subService *subscription.Service) *Handler {
	return &Handler{
		service: NewService(dataDir, subService),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", h.List)
	r.POST("/import", h.ImportURL)
	r.POST("/manual", h.AddManual)
	r.POST("/manual/advanced", h.AddManualAdvanced)
	r.DELETE("/:id", h.Delete)
	r.POST("/test", h.TestDelay)
	r.POST("/test-batch", h.TestDelayBatch)
	r.GET("/:id/share", h.GetShareURL)
	r.GET("/protocols/:protocol/fields", h.GetProtocolFields)
}

// GetService 获取节点服务
func (h *Handler) GetService() *Service {
	return h.service
}

// List 获取所有节点
func (h *Handler) List(c *gin.Context) {
	nodes := h.service.ListAll()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nodes,
	})
}

// ImportURL 从URL导入节点
func (h *Handler) ImportURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	node, err := h.service.ImportURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "解析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    node,
	})
}

// AddManual 手动添加节点
func (h *Handler) AddManual(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		Server string `json:"server" binding:"required"`
		Port   int    `json:"port" binding:"required"`
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	node, err := h.service.AddManual(req.Name, req.Type, req.Server, req.Port, req.Config)
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
		"data":    node,
	})
}

// Delete 删除手动节点
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteManual(id); err != nil {
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

// TestDelay 测试单个节点延迟
func (h *Handler) TestDelay(c *gin.Context) {
	var req struct {
		NodeID  string `json:"nodeId"`
		Server  string `json:"server" binding:"required"`
		Port    int    `json:"port" binding:"required"`
		Timeout int    `json:"timeout"` // 毫秒
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	delay := h.service.TestDelay(req.Server, req.Port, timeout)

	// 保存延迟到缓存
	if req.NodeID != "" {
		h.service.SaveDelay(req.NodeID, delay)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"delay": delay,
		},
	})
}

// TestDelayBatch 批量测试延迟
func (h *Handler) TestDelayBatch(c *gin.Context) {
	var req struct {
		NodeIDs []string `json:"nodeIds" binding:"required"`
		Timeout int      `json:"timeout"` // 毫秒
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	results := h.service.TestDelayBatch(req.NodeIDs, timeout)

	// 批量保存延迟
	h.service.SaveDelayBatch(results)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    results,
	})
}

// GetShareURL 获取分享链接
func (h *Handler) GetShareURL(c *gin.Context) {
	id := c.Param("id")
	url, err := h.service.GetShareURL(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"url": url,
		},
	})
}

// AddManualAdvanced 高级手动添加节点（支持完整配置）
func (h *Handler) AddManualAdvanced(c *gin.Context) {
	var req struct {
		Name       string                 `json:"name" binding:"required"`
		Type       string                 `json:"type" binding:"required"`
		Server     string                 `json:"server" binding:"required"`
		ServerPort int                    `json:"server_port" binding:"required"`
		Config     map[string]interface{} `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	node, err := h.service.AddManualAdvanced(req.Name, req.Type, req.Server, req.ServerPort, req.Config)
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
		"data":    node,
	})
}

// GetProtocolFields 获取协议字段定义
func (h *Handler) GetProtocolFields(c *gin.Context) {
	protocol := c.Param("protocol")
	fields := GetProtocolFieldDefinitions(protocol)
	if fields == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "不支持的协议类型",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"protocol": protocol,
			"fields":   fields,
		},
	})
}
