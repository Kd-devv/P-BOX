package subscription

import (
	"net/http"

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
	r.GET("", h.List)
	r.GET("/:id", h.Get)
	r.GET("/:id/nodes", h.GetNodes)
	r.POST("", h.Add)
	r.PUT("/:id", h.UpdateConfig)
	r.DELETE("/:id", h.Delete)
	r.POST("/:id/update", h.Update)
	r.POST("/update-all", h.UpdateAll)
}

func (h *Handler) List(c *gin.Context) {
	subs := h.service.List()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    subs,
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.service.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    sub,
	})
}

func (h *Handler) GetNodes(c *gin.Context) {
	id := c.Param("id")
	nodes, err := h.service.GetNodes(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nodes,
	})
}

func (h *Handler) Add(c *gin.Context) {
	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	sub, err := h.service.Add(&req)
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
		"data":    sub,
	})
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	id := c.Param("id")
	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateConfig(id, &req); err != nil {
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

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
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

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Update(id); err != nil {
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

func (h *Handler) UpdateAll(c *gin.Context) {
	if err := h.service.UpdateAll(); err != nil {
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
