package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub WebSocket 连接管理中心 (代理模式)
type Hub struct{}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{}
}

// Run 运行 Hub (保留接口兼容)
func (h *Hub) Run() {
	// 代理模式不需要运行循环
}

// HandleTraffic 处理流量 WebSocket (代理到 Mihomo)
func (h *Hub) HandleTraffic(c *gin.Context) {
	h.proxyMihomoWebSocket(c, "/traffic")
}

// HandleLogs 处理日志 WebSocket (代理到 Mihomo)
func (h *Hub) HandleLogs(c *gin.Context) {
	h.proxyMihomoWebSocket(c, "/logs")
}

// HandleConnections 处理连接 WebSocket (代理到 Mihomo)
func (h *Hub) HandleConnections(c *gin.Context) {
	h.proxyMihomoWebSocket(c, "/connections")
}

// proxyMihomoWebSocket 代理 Mihomo WebSocket
func (h *Hub) proxyMihomoWebSocket(c *gin.Context, path string) {
	log.Printf("[WebSocket] 收到代理请求: %s", path)

	// 升级前端连接为 WebSocket
	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] 升级连接失败: %v", err)
		return
	}
	defer clientConn.Close()

	// 连接到 Mihomo WebSocket
	// Mihomo API 默认监听 127.0.0.1:9090
	mihomoURL := "ws://127.0.0.1:9090" + path

	// 转发所有查询参数 (token, level 等)
	queryString := c.Request.URL.RawQuery
	if queryString != "" {
		mihomoURL += "?" + queryString
	}

	log.Printf("[WebSocket] 连接 Mihomo: %s", mihomoURL)
	mihomoConn, _, err := websocket.DefaultDialer.Dial(mihomoURL, nil)
	if err != nil {
		log.Printf("[WebSocket] 连接 Mihomo 失败: %v", err)
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"无法连接到 Mihomo: `+err.Error()+`"}`))
		return
	}
	defer mihomoConn.Close()
	log.Printf("[WebSocket] 已连接 Mihomo, 开始转发")

	// 双向转发
	done := make(chan struct{})

	// Mihomo -> 前端
	go func() {
		defer close(done)
		for {
			msgType, msg, err := mihomoConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}()

	// 前端 -> Mihomo
	go func() {
		for {
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
			if err := mihomoConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}()

	<-done
}
