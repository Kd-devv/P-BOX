package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"p-box/backend/config"
	"p-box/backend/middleware"
	"p-box/backend/modules/auth"
	"p-box/backend/modules/core"
	"p-box/backend/modules/node"
	"p-box/backend/modules/proxy"
	"p-box/backend/modules/ruleset"
	"p-box/backend/modules/speedtest"
	"p-box/backend/modules/subscription"
	"p-box/backend/modules/system"
	"p-box/backend/modules/wireguard"
	"p-box/backend/websocket"
)

// 版本信息 (由 main.go 设置)
var (
	Version   = "2.0.3"
	BuildTime = "unknown"
)

// Server HTTP 服务器
type Server struct {
	config       *config.Config
	router       *gin.Engine
	httpServer   *http.Server
	wsHub        *websocket.Hub
	proxyHandler *proxy.Handler
	authHandler  *auth.Handler
}

// New 创建服务器实例
func New(cfg *config.Config) *Server {
	// 设置 gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	wsHub := websocket.NewHub()

	s := &Server{
		config: cfg,
		router: router,
		wsHub:  wsHub,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// 恢复中间件
	s.router.Use(gin.Recovery())

	// 日志中间件
	s.router.Use(middleware.Logger())

	// CORS 中间件
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件服务 (前端)
	s.router.Static("/assets", "./frontend/assets")
	s.router.StaticFile("/", "./frontend/index.html")
	s.router.StaticFile("/favicon.ico", "./frontend/favicon.ico")
	// PNG 图标文件
	s.router.StaticFile("/p-box-logo.png", "./frontend/p-box-logo.png")
	s.router.StaticFile("/favicon-16.png", "./frontend/favicon-16.png")
	s.router.StaticFile("/favicon-32.png", "./frontend/favicon-32.png")
	s.router.StaticFile("/apple-touch-icon.png", "./frontend/apple-touch-icon.png")
	// SVG 文件（兼容）
	s.router.StaticFile("/p-box-logo.svg", "./frontend/p-box-logo.svg")
	s.router.StaticFile("/favicon.svg", "./frontend/favicon.svg")

	// 健康检查
	s.router.GET("/api/health", s.healthCheck)

	// 认证模块
	authService := auth.NewService(s.config.DataDir)
	s.authHandler = auth.NewHandler(authService)

	// API 路由组
	api := s.router.Group("/api")

	// 认证路由（不需要认证中间件）
	s.authHandler.RegisterRoutes(api)

	// 应用认证中间件
	api.Use(s.authHandler.AuthMiddleware())

	{
		// 系统信息
		api.GET("/system/info", s.systemInfo)

		// 代理模块
		s.proxyHandler = proxy.NewHandler(s.config.DataDir)
		s.proxyHandler.RegisterRoutes(api.Group("/proxy"))

		// 代理设置模块
		settingsHandler := proxy.NewSettingsHandler(s.config.DataDir)
		settingsHandler.RegisterRoutes(api.Group("/proxy"))
		// 设置代理服务引用，用于同步 autoStart 等设置
		settingsHandler.SetProxyService(s.proxyHandler.GetService())

		// 设置代理设置提供者（让 proxy service 能获取优化配置）
		s.proxyHandler.GetService().SetSettingsProvider(func() *proxy.ProxySettings {
			return settingsHandler.GetCurrentSettings()
		})

		// 检查自动启动
		s.proxyHandler.GetService().AutoStartIfEnabled()

		// 核心模块
		coreHandler := core.NewHandler(s.config.DataDir)
		coreHandler.RegisterRoutes(api.Group("/core"))

		// 设置核心切换回调，同步更新 proxy 模块的核心类型
		coreHandler.GetService().SetOnCoreSwitch(func(coreType string) {
			s.proxyHandler.GetService().SetCoreType(coreType)
			fmt.Printf("🔄 核心已切换为: %s\n", coreType)
		})

		// 初始化时同步核心类型
		s.proxyHandler.GetService().SetCoreType(coreHandler.GetService().GetCurrentCore())

		// 订阅模块
		subHandler := subscription.NewHandler(s.config.DataDir)
		subHandler.RegisterRoutes(api.Group("/subscriptions"))

		// 节点模块
		nodeHandler := node.NewHandler(s.config.DataDir, subHandler.GetService())
		nodeHandler.RegisterRoutes(api.Group("/nodes"))

		// 设置节点提供者（让 proxy service 能获取过滤后的节点）
		s.proxyHandler.GetService().SetNodeProvider(func() []proxy.ProxyNode {
			nodes := nodeHandler.GetService().ListAll()
			result := make([]proxy.ProxyNode, 0, len(nodes))
			for _, n := range nodes {
				result = append(result, proxy.ProxyNode{
					Name:       n.Name,
					Type:       n.Type,
					Server:     n.Server,
					ServerPort: n.ServerPort,
					Config:     n.Config,
					IsManual:   n.IsManual,
				})
			}
			return result
		})

		// 系统管理模块
		systemHandler := system.NewHandler(s.config.DataDir)
		systemHandler.RegisterRoutes(api.Group("/system"))

		// 规则集模块 (Mihomo)
		rulesetService := ruleset.NewService(s.config.DataDir)
		rulesetHandler := ruleset.NewHandler(rulesetService)
		rulesetHandler.RegisterRoutes(api)

		// Sing-Box 规则集模块
		proxy.SetSingBoxRulesetDir(s.config.DataDir)
		proxy.RegisterSingBoxRulesetRoutes(api)

		// 测速模块
		speedtestHandler := speedtest.NewHandler()
		speedtestHandler.RegisterRoutes(api.Group("/speedtest"))

		// WireGuard 模块
		wgService := wireguard.NewService(s.config.DataDir)
		wgHandler := wireguard.NewHandler(wgService)
		wgHandler.RegisterRoutes(api)

		// 监听 VPN 启动事件，延迟 5 秒后启动 WireGuard
		s.proxyHandler.GetService().SetOnStartCallback(func() {
			go func() {
				time.Sleep(5 * time.Second)
				wgService.AutoStartIfEnabled()
			}()
		})
	}

	// WebSocket 路由
	ws := s.router.Group("/ws")
	{
		ws.GET("/traffic", s.wsHub.HandleTraffic)
		ws.GET("/logs", s.wsHub.HandleLogs)
		ws.GET("/connections", s.wsHub.HandleConnections)
	}

	// 前端路由 fallback (SPA)
	s.router.NoRoute(func(c *gin.Context) {
		c.File("./frontend/index.html")
	})
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"status":  "healthy",
			"version": "0.1.0",
		},
	})
}

// systemInfo 系统信息
func (s *Server) systemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"name":      "P-BOX",
			"version":   Version,
			"buildTime": BuildTime,
		},
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动 WebSocket Hub
	go s.wsHub.Run()

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	// 先停止代理核心
	if s.proxyHandler != nil {
		fmt.Println("正在停止代理核心...")
		if err := s.proxyHandler.GetService().Stop(); err != nil {
			fmt.Printf("停止代理核心失败: %v\n", err)
		} else {
			fmt.Println("代理核心已停止")
		}
	}

	// 再关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}
}
