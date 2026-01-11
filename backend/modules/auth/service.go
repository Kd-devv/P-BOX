package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用认证
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码哈希
	Avatar   string `json:"avatar"`   // 头像 (base64 或 URL)
}

// Session 会话
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Service 认证服务
type Service struct {
	mu       sync.RWMutex
	config   AuthConfig
	sessions map[string]*Session
	dataDir  string
}

// NewService 创建认证服务
func NewService(dataDir string) *Service {
	s := &Service{
		sessions: make(map[string]*Session),
		dataDir:  dataDir,
	}
	s.loadConfig()
	return s
}

// 配置文件路径
func (s *Service) configPath() string {
	return filepath.Join(s.dataDir, "auth.json")
}

// loadConfig 加载配置
func (s *Service) loadConfig() {
	data, err := os.ReadFile(s.configPath())
	if err != nil {
		// 使用默认配置
		s.config = AuthConfig{
			Enabled:  false,
			Username: "admin",
			Password: s.hashPassword("admin123"),
			Avatar:   "",
		}
		s.saveConfig()
		return
	}
	json.Unmarshal(data, &s.config)
}

// saveConfig 保存配置
func (s *Service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(), data, 0644)
}

// hashPassword 密码哈希
func (s *Service) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password + "p-box-salt"))
	return hex.EncodeToString(hash[:])
}

// generateToken 生成令牌
func (s *Service) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// IsEnabled 是否启用认证
func (s *Service) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Enabled
}

// GetConfig 获取配置（不含密码）
func (s *Service) GetConfig() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"enabled":  s.config.Enabled,
		"username": s.config.Username,
		"avatar":   s.config.Avatar,
	}
}

// SetEnabled 设置是否启用认证
func (s *Service) SetEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Enabled = enabled
	return s.saveConfig()
}

// Login 登录
func (s *Service) Login(username, password string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if username != s.config.Username {
		return "", fmt.Errorf("用户名或密码错误")
	}

	if s.hashPassword(password) != s.config.Password {
		return "", fmt.Errorf("用户名或密码错误")
	}

	// 生成会话
	token := s.generateToken()
	s.sessions[token] = &Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时过期
	}

	return token, nil
}

// Logout 登出
func (s *Service) Logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

// ValidateToken 验证令牌
func (s *Service) ValidateToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return false
	}

	return true
}

// UpdateUsername 更新用户名
func (s *Service) UpdateUsername(newUsername string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newUsername == "" {
		return fmt.Errorf("用户名不能为空")
	}

	s.config.Username = newUsername
	return s.saveConfig()
}

// UpdatePassword 更新密码
func (s *Service) UpdatePassword(oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hashPassword(oldPassword) != s.config.Password {
		return fmt.Errorf("原密码错误")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	s.config.Password = s.hashPassword(newPassword)
	// 清除所有会话，要求重新登录
	s.sessions = make(map[string]*Session)
	return s.saveConfig()
}

// UpdateAvatar 更新头像
func (s *Service) UpdateAvatar(avatar string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Avatar = avatar
	return s.saveConfig()
}
