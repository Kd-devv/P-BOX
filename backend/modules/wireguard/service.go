package wireguard

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"p-box/backend/config"

	"github.com/google/uuid"
)

// Service WireGuard 服务管理
type Service struct {
	dataDir    string
	configPath string
	config     WireGuardConfig
	mu         sync.RWMutex
}

// NewService 创建服务
func NewService(dataDir string) *Service {
	s := &Service{
		dataDir:    dataDir,
		configPath: filepath.Join(dataDir, "wireguard.json"),
	}
	s.loadConfig()
	return s
}

func (s *Service) loadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.config = WireGuardConfig{Servers: []WireGuardServer{}}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.config)
}

func (s *Service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

// IsLinux 检查是否为 Linux
// 开发模式下跳过限制，便于在 macOS 上调试
func IsLinux() bool {
	if config.IsDevMode() {
		return true // 开发模式跳过限制
	}
	return runtime.GOOS == "linux"
}

// CheckInstalled 检查 WireGuard 是否安装
func (s *Service) CheckInstalled() bool {
	_, err := exec.LookPath("wg")
	return err == nil
}

// GetServers 获取所有服务器
func (s *Service) GetServers() []WireGuardServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Servers
}

// GetServer 获取服务器
func (s *Service) GetServer(id string) (*WireGuardServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == id {
			return &s.config.Servers[i], nil
		}
	}
	return nil, fmt.Errorf("服务器不存在")
}

// CreateServer 创建服务器
func (s *Service) CreateServer(server *WireGuardServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyPair, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	server.ID = uuid.New().String()
	server.PrivateKey = keyPair.PrivateKey
	server.PublicKey = keyPair.PublicKey
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	server.Clients = []WireGuardClient{}
	if server.MTU == 0 {
		server.MTU = 1420
	}
	if server.DNS == "" {
		server.DNS = "1.1.1.1,8.8.8.8"
	}

	s.config.Servers = append(s.config.Servers, *server)
	return s.saveConfig()
}

// DeleteServer 删除服务器
func (s *Service) DeleteServer(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == id {
			if s.config.Servers[i].Enabled {
				s.StopInterface(s.config.Servers[i].Tag)
			}
			s.config.Servers = append(s.config.Servers[:i], s.config.Servers[i+1:]...)
			return s.saveConfig()
		}
	}
	return fmt.Errorf("服务器不存在")
}

// AddClient 添加客户端
func (s *Service) AddClient(serverID string, client *WireGuardClient) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.config.Servers {
		if s.config.Servers[i].ID == serverID {
			server := &s.config.Servers[i]

			// 生成密钥对
			keyPair, err := GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("生成密钥失败: %v", err)
			}

			// 生成预共享密钥（增强安全性）
			psk, err := GeneratePresharedKey()
			if err != nil {
				return fmt.Errorf("生成预共享密钥失败: %v", err)
			}

			client.ID = uuid.New().String()
			client.PrivateKey = keyPair.PrivateKey
			client.PublicKey = keyPair.PublicKey
			client.PresharedKey = psk
			client.Enabled = true
			client.CreatedAt = time.Now()

			// 智能分配 IP（避免冲突）
			if client.AllowedIPs == "" {
				client.AllowedIPs = s.allocateClientIP(server)
			}

			// 继承服务器 DNS
			if client.DNS == "" {
				client.DNS = server.DNS
			}

			server.Clients = append(server.Clients, *client)
			return s.saveConfig()
		}
	}
	return fmt.Errorf("服务器不存在")
}

// allocateClientIP 智能分配客户端 IP（避免冲突）
func (s *Service) allocateClientIP(server *WireGuardServer) string {
	// 解析服务器地址，获取网段前缀
	baseIP := strings.Split(server.Address, "/")[0]
	parts := strings.Split(baseIP, ".")
	if len(parts) != 4 {
		return "10.0.0.2/32" // 兜底
	}

	prefix := fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])

	// 收集已使用的 IP
	usedIPs := make(map[int]bool)
	// 服务器自身 IP
	if serverIP := strings.Split(baseIP, "."); len(serverIP) == 4 {
		if num, err := parseInt(serverIP[3]); err == nil {
			usedIPs[num] = true
		}
	}
	// 已有客户端 IP
	for _, c := range server.Clients {
		ip := strings.Split(c.AllowedIPs, "/")[0]
		ipParts := strings.Split(ip, ".")
		if len(ipParts) == 4 {
			if num, err := parseInt(ipParts[3]); err == nil {
				usedIPs[num] = true
			}
		}
	}

	// 从 2 开始分配（1 通常是网关/服务器）
	for i := 2; i <= 254; i++ {
		if !usedIPs[i] {
			return fmt.Sprintf("%s.%d/32", prefix, i)
		}
	}

	// 地址耗尽，使用随机
	return fmt.Sprintf("%s.%d/32", prefix, len(server.Clients)+2)
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// DeleteClient 删除客户端
func (s *Service) DeleteClient(serverID, clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == serverID {
			for j := range s.config.Servers[i].Clients {
				if s.config.Servers[i].Clients[j].ID == clientID {
					s.config.Servers[i].Clients = append(s.config.Servers[i].Clients[:j], s.config.Servers[i].Clients[j+1:]...)
					return s.saveConfig()
				}
			}
		}
	}
	return fmt.Errorf("不存在")
}

// UpdateServer 更新服务器配置
func (s *Service) UpdateServer(server *WireGuardServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == server.ID {
			server.UpdatedAt = time.Now()
			// 保持原有的密钥和客户端列表
			server.PrivateKey = s.config.Servers[i].PrivateKey
			server.PublicKey = s.config.Servers[i].PublicKey
			server.Clients = s.config.Servers[i].Clients
			server.CreatedAt = s.config.Servers[i].CreatedAt
			s.config.Servers[i] = *server
			return s.saveConfig()
		}
	}
	return fmt.Errorf("服务器不存在")
}

// UpdateClient 更新客户端配置
func (s *Service) UpdateClient(serverID, clientID, name, description string, enabled bool) (*WireGuardClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == serverID {
			for j := range s.config.Servers[i].Clients {
				if s.config.Servers[i].Clients[j].ID == clientID {
					if name != "" {
						s.config.Servers[i].Clients[j].Name = name
					}
					s.config.Servers[i].Clients[j].Description = description
					s.config.Servers[i].Clients[j].Enabled = enabled
					if err := s.saveConfig(); err != nil {
						return nil, err
					}
					client := s.config.Servers[i].Clients[j]
					return &client, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("客户端不存在")
}

// AutoStartIfEnabled 自动启动已启用开机启动的服务器
func (s *Service) AutoStartIfEnabled() {
	if !IsLinux() {
		return // 非 Linux 不支持
	}

	servers := s.GetServers()
	for _, server := range servers {
		if server.AutoStart && server.Enabled {
			fmt.Printf("🔄 自动启动 WireGuard 服务器: %s (%s)\n", server.Name, server.Tag)
			if err := s.ApplyConfig(server.ID); err != nil {
				fmt.Printf("⚠️ WireGuard 自动启动失败: %v\n", err)
			} else {
				fmt.Printf("✅ WireGuard 服务器 %s 已启动\n", server.Name)
			}
		}
	}
}
