//go:build windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetSystemProxy 设置系统代理 (Windows)
func SetSystemProxy(host string, port int) error {
	proxyServer := fmt.Sprintf("%s:%d", host, port)

	// 启用代理
	cmd := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f")
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置代理服务器
	cmd = exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyServer", "/t", "REG_SZ", "/d", proxyServer, "/f")
	return cmd.Run()
}

// ClearSystemProxy 清除系统代理
func ClearSystemProxy() error {
	cmd := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f")
	return cmd.Run()
}

// GetSystemProxyStatus 获取系统代理状态
func GetSystemProxyStatus() (bool, string, int, error) {
	cmd := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable")
	output, err := cmd.Output()
	if err != nil {
		return false, "", 0, err
	}

	// 简单检查是否包含 0x1
	enabled := string(output)
	return enabled != "" && (enabled[len(enabled)-2] == '1'), "", 0, nil
}

// BrowserInfo 浏览器信息
type BrowserInfo struct {
	Name            string `json:"name"`
	BundleID        string `json:"bundleId"`
	Path            string `json:"path"`
	FollowsSystem   bool   `json:"followsSystem"`
	ProxyConfigured bool   `json:"proxyConfigured"`
}

// GetInstalledBrowsers 获取已安装的浏览器列表 (Windows)
func GetInstalledBrowsers() []BrowserInfo {
	// Windows 上大多数浏览器跟随系统代理
	browsers := []BrowserInfo{}

	// 检查常见浏览器
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	localAppData := os.Getenv("LOCALAPPDATA")

	browserPaths := []struct {
		name          string
		paths         []string
		followsSystem bool
	}{
		{"Google Chrome", []string{
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
		}, true},
		{"Microsoft Edge", []string{
			filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"),
		}, true},
		{"Firefox", []string{
			filepath.Join(programFiles, "Mozilla Firefox", "firefox.exe"),
			filepath.Join(programFilesX86, "Mozilla Firefox", "firefox.exe"),
		}, false},
		{"Brave", []string{
			filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
		}, true},
	}

	for _, browser := range browserPaths {
		for _, path := range browser.paths {
			if _, err := os.Stat(path); err == nil {
				info := BrowserInfo{
					Name:          browser.name,
					Path:          path,
					FollowsSystem: browser.followsSystem,
				}
				if !browser.followsSystem && strings.Contains(strings.ToLower(browser.name), "firefox") {
					info.ProxyConfigured = isFirefoxProxyConfiguredWindows()
				}
				browsers = append(browsers, info)
				break
			}
		}
	}

	return browsers
}

// isFirefoxProxyConfiguredWindows 检查 Windows 上 Firefox 是否已配置系统代理
func isFirefoxProxyConfiguredWindows() bool {
	appData := os.Getenv("APPDATA")
	profilesDir := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	profiles, err := os.ReadDir(profilesDir)
	if err != nil {
		return false
	}

	for _, profile := range profiles {
		if !profile.IsDir() {
			continue
		}
		userJS := filepath.Join(profilesDir, profile.Name(), "user.js")
		content, err := os.ReadFile(userJS)
		if err != nil {
			continue
		}
		if strings.Contains(string(content), `"network.proxy.type", 5`) {
			return true
		}
	}
	return false
}

// ConfigureFirefoxProxy 配置 Firefox 使用系统代理 (Windows)
func ConfigureFirefoxProxy() error {
	appData := os.Getenv("APPDATA")
	profilesDir := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	profiles, err := os.ReadDir(profilesDir)
	if err != nil {
		return fmt.Errorf("未找到 Firefox 配置目录: %v", err)
	}

	configured := 0
	for _, profile := range profiles {
		if !profile.IsDir() {
			continue
		}
		if !strings.Contains(profile.Name(), "default") {
			continue
		}

		userJS := filepath.Join(profilesDir, profile.Name(), "user.js")
		proxyConfig := `// P-BOX Auto Configuration - Firefox System Proxy
user_pref("network.proxy.type", 5);
`
		existingContent, _ := os.ReadFile(userJS)
		if strings.Contains(string(existingContent), `"network.proxy.type", 5`) {
			configured++
			continue
		}

		newContent := proxyConfig + string(existingContent)
		if err := os.WriteFile(userJS, []byte(newContent), 0644); err != nil {
			continue
		}
		configured++
	}

	if configured == 0 {
		return fmt.Errorf("未找到 Firefox 配置文件")
	}
	return nil
}

// ClearFirefoxProxy 清除 Firefox 代理配置 (Windows)
func ClearFirefoxProxy() error {
	appData := os.Getenv("APPDATA")
	profilesDir := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	profiles, err := os.ReadDir(profilesDir)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		if !profile.IsDir() {
			continue
		}
		userJS := filepath.Join(profilesDir, profile.Name(), "user.js")
		content, err := os.ReadFile(userJS)
		if err != nil {
			continue
		}
		newContent := strings.Replace(string(content), `// P-BOX Auto Configuration - Firefox System Proxy
user_pref("network.proxy.type", 5);
`, "", 1)
		if newContent != string(content) {
			os.WriteFile(userJS, []byte(newContent), 0644)
		}
	}
	return nil
}

// SetBrowserBackupPath 设置备份路径 (Windows)
func SetBrowserBackupPath(dataDir string) {}

// ConfigureAllBrowsersProxy 配置所有浏览器使用系统代理 (Windows)
func ConfigureAllBrowsersProxy() error {
	// Windows 上 Chrome/Edge 默认跟随系统代理
	// 只需配置 Firefox
	ConfigureFirefoxProxy()
	return nil
}

// RestoreAllBrowsersProxy 恢复所有浏览器代理设置 (Windows)
func RestoreAllBrowsersProxy() error {
	ClearFirefoxProxy()
	return nil
}
