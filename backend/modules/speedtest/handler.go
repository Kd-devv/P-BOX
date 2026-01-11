package speedtest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

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

// Handler 测速处理器
type Handler struct {
	history []SpeedTestResult
	mu      sync.RWMutex
}

// SpeedTestResult 测速结果
type SpeedTestResult struct {
	ID            int64   `json:"id"`
	Ping          float64 `json:"ping"`
	DownloadSpeed float64 `json:"downloadSpeed"`
	UploadSpeed   float64 `json:"uploadSpeed"`
	Source        string  `json:"source"`
	Threads       int     `json:"threads"`
	Timestamp     string  `json:"timestamp"`
}

// NewHandler 创建处理器
func NewHandler() *Handler {
	return &Handler{
		history: make([]SpeedTestResult, 0),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/ws", h.HandleSpeedTestWS) // WebSocket 实时测速
	r.POST("/start", h.StartSpeedTest)
	r.GET("/history", h.GetHistory)
	r.GET("/sources", h.GetSources)
	r.DELETE("/history/:id", h.DeleteHistory)
	r.DELETE("/history", h.ClearHistory)
}

// HandleSpeedTestWS 处理测速 WebSocket 连接
func (h *Handler) HandleSpeedTestWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[SpeedTest WS] 升级连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Println("[SpeedTest WS] 客户端已连接")

	// 等待客户端发送测速请求
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[SpeedTest WS] 读取消息失败: %v", err)
			return
		}

		var req struct {
			Action          string `json:"action"`
			Source          string `json:"source"`
			DownloadThreads int    `json:"downloadThreads"`
			UploadThreads   int    `json:"uploadThreads"`
		}
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}

		if req.Action == "start" {
			h.runSpeedTestWithProgress(conn, req.Source, req.DownloadThreads, req.UploadThreads)
		}
	}
}

// runSpeedTestWithProgress 运行测速并实时推送进度
func (h *Handler) runSpeedTestWithProgress(conn *websocket.Conn, source string, downloadThreads, uploadThreads int) {
	if downloadThreads <= 0 {
		downloadThreads = 10
	}
	if uploadThreads <= 0 {
		uploadThreads = 3
	}
	if source == "" {
		source = "cloudflare"
	}

	// 创建进度 channel
	progressChan := make(chan SpeedtestProgress, 100)
	defer close(progressChan)

	// 发送进度的 goroutine
	go func() {
		for progress := range progressChan {
			data, _ := json.Marshal(map[string]interface{}{
				"type":     "progress",
				"phase":    progress.Type,
				"progress": progress.Progress,
				"value":    progress.Value,
				"unit":     progress.Unit,
			})
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}()

	// 运行测速
	ctx := context.Background()
	result, err := h.SpeedtestWithProgress(ctx, progressChan, source, downloadThreads, uploadThreads)

	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type":    "error",
			"message": err.Error(),
		})
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	// 保存历史
	h.mu.Lock()
	h.history = append([]SpeedTestResult{*result}, h.history...)
	if len(h.history) > 20 {
		h.history = h.history[:20]
	}
	h.mu.Unlock()

	// 发送完成结果
	data, _ := json.Marshal(map[string]interface{}{
		"type":   "complete",
		"result": result,
	})
	conn.WriteMessage(websocket.TextMessage, data)
}

// StartSpeedTest 开始测速
func (h *Handler) StartSpeedTest(c *gin.Context) {
	var req struct {
		Source          string `json:"source"`
		DownloadThreads int    `json:"downloadThreads"`
		UploadThreads   int    `json:"uploadThreads"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Source = "cloudflare"
		req.DownloadThreads = 10
		req.UploadThreads = 3
	}

	if req.DownloadThreads <= 0 {
		req.DownloadThreads = 10
	}
	if req.UploadThreads <= 0 {
		req.UploadThreads = 3
	}

	// 使用 SimpleSpeedtest 进行测速
	ctx := context.Background()
	result, err := h.SimpleSpeedtest(ctx, req.Source, req.DownloadThreads, req.UploadThreads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	// 保存历史
	h.mu.Lock()
	h.history = append([]SpeedTestResult{*result}, h.history...)
	if len(h.history) > 20 {
		h.history = h.history[:20]
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// GetSources 获取测速源列表
func (h *Handler) GetSources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": GetSpeedTestSources(),
	})
}

// GetHistory 获取历史记录
func (h *Handler) GetHistory(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": h.history,
	})
}

// DeleteHistory 删除单条历史
func (h *Handler) DeleteHistory(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	json.Unmarshal([]byte(idStr), &id)

	h.mu.Lock()
	for i, item := range h.history {
		if item.ID == id {
			h.history = append(h.history[:i], h.history[i+1:]...)
			break
		}
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已删除",
	})
}

// ClearHistory 清空历史
func (h *Handler) ClearHistory(c *gin.Context) {
	h.mu.Lock()
	h.history = make([]SpeedTestResult, 0)
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已清空",
	})
}
