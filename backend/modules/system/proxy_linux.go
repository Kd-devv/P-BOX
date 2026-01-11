//go:build linux

package system

import "fmt"

// SetSystemProxy 设置系统代理 (Linux 通常使用 TUN 模式)
func SetSystemProxy(host string, port int) error {
	// Linux 桌面环境多样，通常建议使用 TUN 模式
	// 这里只是占位，不实际设置
	return nil
}

// ClearSystemProxy 清除系统代理
func ClearSystemProxy() error {
	return nil
}

// GetSystemProxyStatus 获取系统代理状态
func GetSystemProxyStatus() (bool, string, int, error) {
	return false, "", 0, fmt.Errorf("not supported on Linux, use TUN mode instead")
}

// BrowserInfo 浏览器信息
type BrowserInfo struct {
	Name            string `json:"name"`
	BundleID        string `json:"bundleId"`
	Path            string `json:"path"`
	FollowsSystem   bool   `json:"followsSystem"`
	ProxyConfigured bool   `json:"proxyConfigured"`
}

// GetInstalledBrowsers 获取已安装的浏览器列表 (Linux 使用 TUN 模式，不需要)
func GetInstalledBrowsers() []BrowserInfo {
	return nil
}

// ConfigureFirefoxProxy 配置 Firefox 使用系统代理 (Linux 使用 TUN 模式)
func ConfigureFirefoxProxy() error {
	return fmt.Errorf("Linux 建议使用 TUN 模式，无需配置浏览器代理")
}

// ClearFirefoxProxy 清除 Firefox 代理配置
func ClearFirefoxProxy() error {
	return nil
}

// SetBrowserBackupPath 设置备份路径 (Linux 使用 TUN 模式)
func SetBrowserBackupPath(dataDir string) {}

// ConfigureAllBrowsersProxy 配置所有浏览器 (Linux 使用 TUN 模式，不需要)
func ConfigureAllBrowsersProxy() error {
	return nil
}

// RestoreAllBrowsersProxy 恢复所有浏览器 (Linux 使用 TUN 模式)
func RestoreAllBrowsersProxy() error {
	return nil
}
