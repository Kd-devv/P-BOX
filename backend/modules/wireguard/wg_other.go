//go:build !linux

package wireguard

import "fmt"

// GetTunInterface 非 Linux 返回空
func GetTunInterface() string {
	return ""
}

// GenerateWGConfig 非 Linux 不支持
func (s *Service) GenerateWGConfig(serverID string) (string, error) {
	return "", fmt.Errorf("WireGuard 服务仅支持 Linux")
}

// ApplyConfig 非 Linux 不支持
func (s *Service) ApplyConfig(serverID string) error {
	return fmt.Errorf("WireGuard 服务仅支持 Linux")
}

// StopInterface 非 Linux 不支持
func (s *Service) StopInterface(tag string) error {
	return fmt.Errorf("WireGuard 服务仅支持 Linux")
}

// GetStatus 非 Linux 返回 false
func (s *Service) GetStatus(tag string) (bool, string) {
	return false, ""
}

// GenerateClientConfig 非 Linux 不支持
func (s *Service) GenerateClientConfig(serverID, clientID, endpoint string) (string, error) {
	return "", fmt.Errorf("WireGuard 服务仅支持 Linux")
}

// InstallWireGuard 非 Linux 不支持
func (s *Service) InstallWireGuard() error {
	return fmt.Errorf("WireGuard 自动安装仅支持 Linux")
}

// ForceCleanupInterface 非 Linux 不支持
func (s *Service) ForceCleanupInterface(tag string) {
	// 非 Linux 系统不执行任何操作
}
