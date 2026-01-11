package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 认证处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/config", h.GetConfig)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.PUT("/enabled", h.SetEnabled)
		auth.PUT("/username", h.UpdateUsername)
		auth.PUT("/password", h.UpdatePassword)
		auth.PUT("/avatar", h.UpdateAvatar)
		auth.GET("/check", h.CheckAuth)
	}
}

// AuthMiddleware 认证中间件
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未启用认证，直接放行
		if !h.service.IsEnabled() {
			c.Next()
			return
		}

		// 登录和配置检查接口不需要认证
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/auth/login") ||
			strings.HasSuffix(path, "/auth/config") ||
			strings.HasSuffix(path, "/auth/check") {
			c.Next()
			return
		}

		// 从请求头获取令牌
		token := c.GetHeader("Authorization")
		if token == "" {
			// 尝试从 cookie 获取
			token, _ = c.Cookie("p-box-token")
		}

		// 移除 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" || !h.service.ValidateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权，请先登录",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetConfig 获取认证配置
func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetConfig())
}

// CheckAuth 检查认证状态
func (h *Handler) CheckAuth(c *gin.Context) {
	if !h.service.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"enabled":       false,
			"authenticated": true,
		})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("p-box-token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	c.JSON(http.StatusOK, gin.H{
		"enabled":       true,
		"authenticated": h.service.ValidateToken(token),
	})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 设置 cookie
	c.SetCookie("p-box-token", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": req.Username,
		"avatar":   h.service.GetConfig()["avatar"],
	})
}

// Logout 登出
func (h *Handler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("p-box-token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	h.service.Logout(token)

	// 清除 cookie
	c.SetCookie("p-box-token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "已登出"})
}

// SetEnabled 设置是否启用认证
func (h *Handler) SetEnabled(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.service.SetEnabled(req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

// UpdateUsername 更新用户名
func (h *Handler) UpdateUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.service.UpdateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户名已更新"})
}

// UpdatePassword 更新密码
func (h *Handler) UpdatePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.service.UpdatePassword(req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已更新，请重新登录"})
}

// UpdateAvatar 更新头像
func (h *Handler) UpdateAvatar(c *gin.Context) {
	var req struct {
		Avatar string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.service.UpdateAvatar(req.Avatar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像已更新"})
}
