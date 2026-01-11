package node

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"p-box/backend/modules/subscription"

	"github.com/google/uuid"
)

type Node struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Server         string `json:"server"`
	ServerPort     int    `json:"serverPort"`
	SubscriptionID string `json:"subscriptionId,omitempty"` // 来源订阅
	IsManual       bool   `json:"isManual"`                 // 手动添加
	Enabled        bool   `json:"enabled"`
	Delay          int    `json:"delay"`    // 延迟 ms, 0=超时, -1=未测试
	LastTest       int64  `json:"lastTest"` // 上次测速时间戳
	Config         string `json:"config"`   // JSON格式的完整配置
	ShareURL       string `json:"shareUrl"` // 分享链接
}

type Service struct {
	dataDir     string
	manualNodes map[string]*Node
	delayCache  map[string]int // 节点延迟缓存
	subService  *subscription.Service
	mu          sync.RWMutex
}

func NewService(dataDir string, subService *subscription.Service) *Service {
	s := &Service{
		dataDir:     dataDir,
		manualNodes: make(map[string]*Node),
		delayCache:  make(map[string]int),
		subService:  subService,
	}
	s.loadManualNodes()
	s.loadDelayCache()
	return s
}

func (s *Service) loadDelayCache() {
	filePath := filepath.Join(s.dataDir, "delay_cache.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.delayCache)
}

func (s *Service) saveDelayCache() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.delayCache, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	filePath := filepath.Join(s.dataDir, "delay_cache.json")
	return os.WriteFile(filePath, data, 0644)
}

// SaveDelay 保存节点延迟
func (s *Service) SaveDelay(nodeID string, delay int) {
	s.mu.Lock()
	s.delayCache[nodeID] = delay
	s.mu.Unlock()
	s.saveDelayCache()
}

// SaveDelayBatch 批量保存延迟
func (s *Service) SaveDelayBatch(results map[string]int) {
	s.mu.Lock()
	for id, delay := range results {
		s.delayCache[id] = delay
	}
	s.mu.Unlock()
	s.saveDelayCache()
}

// GetDelay 获取节点延迟
func (s *Service) GetDelay(nodeID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if delay, ok := s.delayCache[nodeID]; ok {
		return delay
	}
	return -1
}

func (s *Service) loadManualNodes() {
	filePath := filepath.Join(s.dataDir, "manual_nodes.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	var nodes []*Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return
	}
	for _, node := range nodes {
		s.manualNodes[node.ID] = node
	}
}

func (s *Service) saveManualNodes() error {
	s.mu.RLock()
	nodes := make([]*Node, 0, len(s.manualNodes))
	for _, node := range s.manualNodes {
		nodes = append(nodes, node)
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	filePath := filepath.Join(s.dataDir, "manual_nodes.json")
	return os.WriteFile(filePath, data, 0644)
}

// ListAll 获取所有节点（订阅+手动）
func (s *Service) ListAll() []*Node {
	nodes := make([]*Node, 0)

	// 1. 获取所有订阅的节点
	subs := s.subService.List()
	for _, sub := range subs {
		subNodes, err := s.subService.GetNodes(sub.ID)
		if err != nil {
			continue
		}
		for _, sn := range subNodes {
			if sn.IsFiltered {
				continue // 跳过被过滤的节点
			}
			nodeID := fmt.Sprintf("%s_%s", sub.ID, sn.Name)
			node := &Node{
				ID:             nodeID,
				Name:           sn.Name,
				Type:           sn.Type,
				Server:         sn.Server,
				ServerPort:     sn.ServerPort,
				SubscriptionID: sub.ID,
				IsManual:       false,
				Enabled:        sn.Enabled,
				Delay:          s.GetDelay(nodeID), // 使用缓存的延迟
				Config:         sn.Config,
				ShareURL:       sn.ShareURL,
			}
			nodes = append(nodes, node)
		}
	}

	// 2. 添加手动节点
	s.mu.RLock()
	for _, node := range s.manualNodes {
		// 更新手动节点的延迟
		node.Delay = s.GetDelay(node.ID)
		nodes = append(nodes, node)
	}
	s.mu.RUnlock()

	return nodes
}

// AddManual 手动添加节点
func (s *Service) AddManual(name, nodeType, server string, port int, config string) (*Node, error) {
	node := &Node{
		ID:         uuid.New().String(),
		Name:       name,
		Type:       nodeType,
		Server:     server,
		ServerPort: port,
		IsManual:   true,
		Enabled:    true,
		Delay:      -1,
		Config:     config,
	}

	s.mu.Lock()
	s.manualNodes[node.ID] = node
	s.mu.Unlock()

	return node, s.saveManualNodes()
}

// AddManualAdvanced 高级手动添加节点（支持完整配置）
func (s *Service) AddManualAdvanced(name, nodeType, server string, port int, config map[string]interface{}) (*Node, error) {
	// 将 config map 转换为 JSON 字符串
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("配置序列化失败: %w", err)
	}

	node := &Node{
		ID:         uuid.New().String(),
		Name:       name,
		Type:       nodeType,
		Server:     server,
		ServerPort: port,
		IsManual:   true,
		Enabled:    true,
		Delay:      -1,
		Config:     string(configJSON),
	}

	// 生成分享链接
	shareURL := s.generateShareURLFromConfig(node, config)
	if shareURL != "" {
		node.ShareURL = shareURL
	}

	s.mu.Lock()
	s.manualNodes[node.ID] = node
	s.mu.Unlock()

	return node, s.saveManualNodes()
}

// generateShareURLFromConfig 根据配置生成分享链接
func (s *Service) generateShareURLFromConfig(node *Node, config map[string]interface{}) string {
	// 基于节点类型生成对应的分享链接
	// 这里简化处理，后续可以根据需要完善
	return ""
}

// ImportURL 从URL导入节点
func (s *Service) ImportURL(url string) (*Node, error) {
	// 使用订阅模块的解析器
	proxyNode, err := subscription.ParseURL(url)
	if err != nil {
		return nil, err
	}

	node := &Node{
		ID:         uuid.New().String(),
		Name:       proxyNode.Name,
		Type:       proxyNode.Type,
		Server:     proxyNode.Server,
		ServerPort: proxyNode.ServerPort,
		IsManual:   true,
		Enabled:    true,
		Delay:      -1,
		Config:     proxyNode.Config,
		ShareURL:   url,
	}

	s.mu.Lock()
	s.manualNodes[node.ID] = node
	s.mu.Unlock()

	return node, s.saveManualNodes()
}

// DeleteManual 删除手动节点
func (s *Service) DeleteManual(id string) error {
	s.mu.Lock()
	delete(s.manualNodes, id)
	s.mu.Unlock()
	return s.saveManualNodes()
}

// TestDelay 测试单个节点延迟 (TCP连接测试)
func (s *Service) TestDelay(server string, port int, timeout time.Duration) int {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", server, port), timeout)
	if err != nil {
		return 0 // 超时或连接失败
	}
	conn.Close()
	return int(time.Since(start).Milliseconds())
}

// TestDelayBatch 批量测试延迟
func (s *Service) TestDelayBatch(nodeIDs []string, timeout time.Duration) map[string]int {
	results := make(map[string]int)
	var wg sync.WaitGroup
	var mu sync.Mutex

	nodes := s.ListAll()
	nodeMap := make(map[string]*Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	// 限制并发数
	sem := make(chan struct{}, 20)

	for _, id := range nodeIDs {
		node, ok := nodeMap[id]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			delay := s.TestDelay(n.Server, n.ServerPort, timeout)
			mu.Lock()
			results[n.ID] = delay
			mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// GetShareURL 获取节点分享链接
func (s *Service) GetShareURL(id string) (string, error) {
	// 先检查手动节点
	s.mu.RLock()
	if node, ok := s.manualNodes[id]; ok {
		s.mu.RUnlock()
		if node.ShareURL != "" {
			return node.ShareURL, nil
		}
		return "", fmt.Errorf("该节点没有分享链接")
	}
	s.mu.RUnlock()

	// 检查订阅节点 - 从所有节点列表中查找
	nodes := s.ListAll()
	for _, node := range nodes {
		if node.ID == id {
			if node.ShareURL != "" {
				return node.ShareURL, nil
			}
			// 尝试从 Config 生成分享链接
			if node.Config != "" {
				return s.generateShareURL(node)
			}
			return "", fmt.Errorf("该节点不支持导出")
		}
	}

	return "", fmt.Errorf("节点不存在")
}

// generateShareURL 根据节点配置生成分享链接
func (s *Service) generateShareURL(node *Node) (string, error) {
	// 这里需要根据节点类型和配置生成对应的分享链接
	// 暂时返回错误，后续可以根据 Config JSON 生成
	return "", fmt.Errorf("该节点不支持导出")
}
