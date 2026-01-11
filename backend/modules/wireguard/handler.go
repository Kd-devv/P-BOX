package wireguard

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler WireGuard HTTP 处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	wg := r.Group("/wireguard")
	{
		wg.GET("/status", h.GetSystemStatus)
		wg.GET("/default-dns", h.GetDefaultDNS)
		wg.POST("/install", h.Install)
		wg.GET("/servers", h.GetServers)
		wg.POST("/servers", h.CreateServer)
		wg.GET("/servers/:id", h.GetServer)
		wg.PUT("/servers/:id", h.UpdateServer)
		wg.DELETE("/servers/:id", h.DeleteServer)
		wg.POST("/servers/:id/apply", h.ApplyConfig)
		wg.POST("/servers/:id/stop", h.StopServer)
		wg.GET("/servers/:id/status", h.GetServerStatus)

		// 客户端
		wg.POST("/servers/:id/clients", h.AddClient)
		wg.PUT("/servers/:id/clients/:clientId", h.UpdateClient)
		wg.DELETE("/servers/:id/clients/:clientId", h.DeleteClient)
		wg.GET("/servers/:id/clients/:clientId/config", h.GetClientConfig)
	}
}

// GetSystemStatus 获取系统状态
func (h *Handler) GetSystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"linux":     IsLinux(),
			"installed": h.service.CheckInstalled(),
		},
	})
}

// GetDefaultDNS 获取默认 DNS（本机内网 IP）
func (h *Handler) GetDefaultDNS(c *gin.Context) {
	localIP := getLocalIP()
	dns := []string{}
	if localIP != "" {
		dns = append(dns, localIP)
	}
	// 添加备用 DNS
	dns = append(dns, "1.1.1.1")

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"dns":      dns,
			"local_ip": localIP,
		},
	})
}

// getLocalIP 获取本机内网 IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				// 优先返回常见内网段
				if strings.HasPrefix(ip, "192.168.") ||
					strings.HasPrefix(ip, "10.") ||
					strings.HasPrefix(ip, "172.") {
					return ip
				}
			}
		}
	}
	return ""
}

// Install 安装 WireGuard
func (h *Handler) Install(c *gin.Context) {
	if !IsLinux() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "仅支持 Linux 系统"})
		return
	}
	if h.service.CheckInstalled() {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "WireGuard 已安装"})
		return
	}
	if err := h.service.InstallWireGuard(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "安装成功"})
}

// GetServers 获取所有服务器
func (h *Handler) GetServers(c *gin.Context) {
	servers := h.service.GetServers()
	// 更新运行状态
	for i := range servers {
		running, _ := h.service.GetStatus(servers[i].Tag)
		servers[i].Enabled = running
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": servers})
}

// GetServer 获取单个服务器
func (h *Handler) GetServer(c *gin.Context) {
	id := c.Param("id")
	server, err := h.service.GetServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": err.Error()})
		return
	}
	running, _ := h.service.GetStatus(server.Tag)
	server.Enabled = running
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": server})
}

// CreateServer 创建服务器
func (h *Handler) CreateServer(c *gin.Context) {
	if !IsLinux() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "WireGuard 服务仅支持 Linux"})
		return
	}

	var server WireGuardServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.CreateServer(&server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": server, "message": "创建成功"})
}

// DeleteServer 删除服务器
func (h *Handler) DeleteServer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteServer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ApplyConfig 应用配置并启动
func (h *Handler) ApplyConfig(c *gin.Context) {
	if !IsLinux() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "WireGuard 服务仅支持 Linux"})
		return
	}

	id := c.Param("id")
	if err := h.service.ApplyConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "启动成功"})
}

// StopServer 停止服务器
func (h *Handler) StopServer(c *gin.Context) {
	id := c.Param("id")
	server, err := h.service.GetServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": err.Error()})
		return
	}
	if err := h.service.StopInterface(server.Tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "停止成功"})
}

// GetServerStatus 获取服务器状态
func (h *Handler) GetServerStatus(c *gin.Context) {
	id := c.Param("id")
	server, err := h.service.GetServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": err.Error()})
		return
	}
	running, output := h.service.GetStatus(server.Tag)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"running": running,
			"output":  output,
		},
	})
}

// AddClient 添加客户端
func (h *Handler) AddClient(c *gin.Context) {
	serverID := c.Param("id")
	var client WireGuardClient
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	if err := h.service.AddClient(serverID, &client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": client, "message": "添加成功"})
}

// DeleteClient 删除客户端
func (h *Handler) DeleteClient(c *gin.Context) {
	serverID := c.Param("id")
	clientID := c.Param("clientId")

	if err := h.service.DeleteClient(serverID, clientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// GetClientConfig 获取客户端配置
func (h *Handler) GetClientConfig(c *gin.Context) {
	serverID := c.Param("id")
	clientID := c.Param("clientId")
	endpoint := c.Query("endpoint")

	config, err := h.service.GenerateClientConfig(serverID, clientID, endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": config})
}

// UpdateServer 更新服务器配置
func (h *Handler) UpdateServer(c *gin.Context) {
	serverID := c.Param("id")
	var req struct {
		Name        string `json:"name"`
		Endpoint    string `json:"endpoint"`
		AutoStart   bool   `json:"auto_start"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
		return
	}

	server, err := h.service.GetServer(serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "服务器不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		server.Name = req.Name
	}
	server.Endpoint = req.Endpoint
	server.AutoStart = req.AutoStart
	server.Description = req.Description

	if err := h.service.UpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": server})
}

// UpdateClient 更新客户端配置
func (h *Handler) UpdateClient(c *gin.Context) {
	serverID := c.Param("id")
	clientID := c.Param("clientId")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
		return
	}

	client, err := h.service.UpdateClient(serverID, clientID, req.Name, req.Description, req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": client})
}
