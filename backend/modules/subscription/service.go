package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Subscription struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	URL               string     `json:"url"`
	NodeCount         int        `json:"nodeCount"`         // 总节点数
	FilteredNodeCount int        `json:"filteredNodeCount"` // 过滤后有效节点数
	Traffic           *Traffic   `json:"traffic,omitempty"`
	ExpireTime        *time.Time `json:"expireTime,omitempty"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	// 定时更新
	AutoUpdate     bool `json:"autoUpdate"`
	UpdateInterval int  `json:"updateInterval"` // 秒，默认 86400 (24小时)
	// 关键词过滤
	FilterKeywords []string `json:"filterKeywords,omitempty"` // 过滤关键词
	FilterMode     string   `json:"filterMode"`               // include: 包含, exclude: 排除
	// 自定义请求头
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
	// 更新状态
	LastUpdateStatus string `json:"lastUpdateStatus,omitempty"` // success, failed
	LastError        string `json:"lastError,omitempty"`        // 最后一次错误信息
}

type Traffic struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
}

type SubscriptionNode struct {
	*ProxyNode
	Enabled    bool   `json:"enabled"`
	IsFiltered bool   `json:"isFiltered"` // 被过滤隐藏
	Ping       int    `json:"ping"`       // 延迟 ms
	Country    string `json:"country,omitempty"`
}

type Service struct {
	dataDir       string
	subscriptions map[string]*Subscription
	stopChan      chan struct{}
	mu            sync.RWMutex
}

func NewService(dataDir string) *Service {
	s := &Service{
		dataDir:       dataDir,
		subscriptions: make(map[string]*Subscription),
		stopChan:      make(chan struct{}),
	}
	s.loadSubscriptions()
	go s.startAutoUpdateLoop()
	return s
}

// 定时更新循环
func (s *Service) startAutoUpdateLoop() {
	ticker := time.NewTicker(time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndUpdate()
		case <-s.stopChan:
			return
		}
	}
}

// 检查并更新需要更新的订阅
func (s *Service) checkAndUpdate() {
	s.mu.RLock()
	subs := make([]*Subscription, 0)
	now := time.Now()

	for _, sub := range s.subscriptions {
		if !sub.AutoUpdate {
			continue
		}
		interval := sub.UpdateInterval
		if interval <= 0 {
			interval = 86400 // 默认24小时
		}
		// 检查是否需要更新
		if now.Sub(sub.UpdatedAt) >= time.Duration(interval)*time.Second {
			subs = append(subs, sub)
		}
	}
	s.mu.RUnlock()

	// 更新订阅
	for _, sub := range subs {
		s.updateSubscription(sub)
	}
	if len(subs) > 0 {
		s.saveSubscriptions()
	}
}

// 停止定时更新
func (s *Service) Stop() {
	close(s.stopChan)
}

func (s *Service) loadSubscriptions() {
	filePath := filepath.Join(s.dataDir, "subscriptions.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var subs []*Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return
	}

	for _, sub := range subs {
		s.subscriptions[sub.ID] = sub
	}
}

// saveSubscriptions 保存订阅（调用者必须已持有锁或不需要锁）
func (s *Service) saveSubscriptions() error {
	subs := make([]*Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subs = append(subs, sub)
	}

	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.dataDir, "subscriptions.json")
	return os.WriteFile(filePath, data, 0644)
}

func (s *Service) List() []*Subscription {
	s.mu.RLock()
	subs := make([]*Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subs = append(subs, sub)
	}
	s.mu.RUnlock()

	// 计算过滤后的节点数
	for _, sub := range subs {
		sub.FilteredNodeCount = s.GetFilteredNodeCount(sub.ID)
	}

	return subs
}

func (s *Service) Get(id string) (*Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, ok := s.subscriptions[id]
	if !ok {
		return nil, fmt.Errorf("subscription not found")
	}
	return sub, nil
}

// AddRequest 添加订阅请求
type AddRequest struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	AutoUpdate     bool              `json:"autoUpdate"`
	UpdateInterval int               `json:"updateInterval"`
	FilterKeywords []string          `json:"filterKeywords"`
	FilterMode     string            `json:"filterMode"` // include/exclude
	CustomHeaders  map[string]string `json:"customHeaders"`
}

func (s *Service) Add(req *AddRequest) (*Subscription, error) {
	sub := &Subscription{
		ID:             uuid.New().String(),
		Name:           req.Name,
		URL:            req.URL,
		AutoUpdate:     req.AutoUpdate,
		UpdateInterval: req.UpdateInterval,
		FilterKeywords: req.FilterKeywords,
		FilterMode:     req.FilterMode,
		CustomHeaders:  req.CustomHeaders,
		CreatedAt:      time.Now(),
	}

	if sub.UpdateInterval <= 0 {
		sub.UpdateInterval = 86400 // 默认24小时
	}
	if sub.FilterMode == "" {
		sub.FilterMode = "exclude" // 默认排除模式
	}

	// 获取订阅内容
	if err := s.updateSubscription(sub); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.subscriptions[sub.ID] = sub
	s.mu.Unlock()

	s.saveSubscriptions()
	return sub, nil
}

// UpdateConfig 更新订阅配置
func (s *Service) UpdateConfig(id string, req *AddRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscriptions[id]
	if !ok {
		return fmt.Errorf("subscription not found")
	}

	sub.Name = req.Name
	sub.URL = req.URL
	sub.AutoUpdate = req.AutoUpdate
	sub.UpdateInterval = req.UpdateInterval
	sub.FilterKeywords = req.FilterKeywords
	sub.FilterMode = req.FilterMode
	sub.CustomHeaders = req.CustomHeaders

	return s.saveSubscriptions()
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	delete(s.subscriptions, id)
	s.mu.Unlock()

	// 删除节点文件
	os.Remove(filepath.Join(s.dataDir, "configs", id+".yaml"))
	os.Remove(filepath.Join(s.dataDir, "configs", id+"_nodes.json"))

	return s.saveSubscriptions()
}

// GetNodes 获取订阅的节点列表（带过滤）
func (s *Service) GetNodes(id string) ([]*SubscriptionNode, error) {
	s.mu.RLock()
	sub, ok := s.subscriptions[id]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("subscription not found")
	}

	// 读取节点文件
	nodesPath := filepath.Join(s.dataDir, "configs", id+"_nodes.json")
	data, err := os.ReadFile(nodesPath)
	if err != nil {
		return nil, fmt.Errorf("nodes not found")
	}

	var nodes []*ProxyNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}

	// 应用过滤
	result := make([]*SubscriptionNode, 0, len(nodes))
	for _, node := range nodes {
		sn := &SubscriptionNode{
			ProxyNode: node,
			Enabled:   true,
		}

		// 关键词过滤
		if len(sub.FilterKeywords) > 0 {
			matched := false
			for _, keyword := range sub.FilterKeywords {
				if strings.Contains(strings.ToLower(node.Name), strings.ToLower(keyword)) {
					matched = true
					break
				}
			}

			if sub.FilterMode == "include" {
				// 包含模式：匹配的显示，不匹配的过滤
				sn.IsFiltered = !matched
			} else {
				// 排除模式：匹配的过滤，不匹配的显示
				sn.IsFiltered = matched
			}
		}

		result = append(result, sn)
	}

	return result, nil
}

// GetFilteredNodeCount 获取过滤后的有效节点数
func (s *Service) GetFilteredNodeCount(id string) int {
	nodes, err := s.GetNodes(id)
	if err != nil {
		return 0
	}
	count := 0
	for _, node := range nodes {
		if !node.IsFiltered {
			count++
		}
	}
	return count
}

func (s *Service) Update(id string) error {
	s.mu.RLock()
	sub, ok := s.subscriptions[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("subscription not found")
	}

	if err := s.updateSubscription(sub); err != nil {
		return err
	}

	return s.saveSubscriptions()
}

func (s *Service) UpdateAll() error {
	s.mu.RLock()
	subs := make([]*Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subs = append(subs, sub)
	}
	s.mu.RUnlock()

	for _, sub := range subs {
		s.updateSubscription(sub)
	}

	return s.saveSubscriptions()
}

func (s *Service) updateSubscription(sub *Subscription) error {
	// 辅助函数：设置失败状态
	setFailed := func(errMsg string) {
		sub.LastUpdateStatus = "failed"
		sub.LastError = errMsg
		sub.UpdatedAt = time.Now()
	}

	// 创建请求
	req, err := http.NewRequest("GET", sub.URL, nil)
	if err != nil {
		setFailed(fmt.Sprintf("创建请求失败: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 添加自定义请求头
	req.Header.Set("User-Agent", "P-BOX/1.0")
	for key, value := range sub.CustomHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		setFailed(fmt.Sprintf("请求失败: %v", err))
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	// 解析流量信息
	if info := resp.Header.Get("subscription-userinfo"); info != "" {
		sub.Traffic = parseTrafficInfo(info)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		setFailed(fmt.Sprintf("读取响应失败: %v", err))
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		setFailed(fmt.Sprintf("HTTP 错误: %d", resp.StatusCode))
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 尝试解析订阅内容
	content := string(body)

	// 尝试 Base64 解码
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content)); err == nil {
		content = string(decoded)
	}

	// 解析节点
	var nodes []*ProxyNode

	if strings.Contains(content, "proxies:") {
		// YAML 格式 (Clash 配置) - 保存完整的代理配置
		var config struct {
			Proxies []map[string]interface{} `yaml:"proxies"`
		}
		if err := yaml.Unmarshal([]byte(content), &config); err == nil {
			for _, p := range config.Proxies {
				name, nameOk := p["name"].(string)
				nodeType, typeOk := p["type"].(string)
				server, serverOk := p["server"].(string)

				if !nameOk || !typeOk || !serverOk {
					continue
				}

				// 获取端口
				var port int
				switch v := p["port"].(type) {
				case int:
					port = v
				case float64:
					port = int(v)
				}

				// 将完整配置序列化为 JSON 保存
				configJSON, _ := json.Marshal(p)

				node := &ProxyNode{
					Name:       name,
					Type:       nodeType,
					Server:     server,
					ServerPort: port,
					Config:     string(configJSON), // 保存完整配置
				}
				nodes = append(nodes, node)
			}
		}
	} else {
		// 链接格式 - 使用协议解析器
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			node, err := ParseURL(line)
			if err == nil && node != nil {
				// 保存原始链接用于分享
				node.ShareURL = line
				nodes = append(nodes, node)
			}
		}
	}

	sub.NodeCount = len(nodes)
	sub.UpdatedAt = time.Now()

	// 检查是否解析到节点
	if len(nodes) == 0 {
		setFailed("未解析到任何节点")
		return fmt.Errorf("no nodes found")
	}

	// 更新成功
	sub.LastUpdateStatus = "success"
	sub.LastError = ""

	// 保存订阅内容和节点列表
	configPath := filepath.Join(s.dataDir, "configs", sub.ID+".yaml")
	nodesPath := filepath.Join(s.dataDir, "configs", sub.ID+"_nodes.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// 保存原始内容
	os.WriteFile(configPath, body, 0644)

	// 保存解析后的节点
	if len(nodes) > 0 {
		nodesJSON, _ := json.MarshalIndent(nodes, "", "  ")
		os.WriteFile(nodesPath, nodesJSON, 0644)
	}

	return nil
}

func parseTrafficInfo(info string) *Traffic {
	traffic := &Traffic{}
	parts := strings.Split(info, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		var n int64
		fmt.Sscanf(value, "%d", &n)

		switch key {
		case "upload":
			traffic.Upload = n
		case "download":
			traffic.Download = n
		case "total":
			traffic.Total = n
		}
	}
	return traffic
}
