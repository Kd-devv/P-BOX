//go:build darwin

package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// getNetworkServices 获取所有网络服务
func getNetworkServices() ([]string, error) {
	cmd := exec.Command("networksetup", "-listallnetworkservices")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var services []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过第一行说明和空行
		if line == "" || strings.HasPrefix(line, "An asterisk") {
			continue
		}
		services = append(services, line)
	}
	return services, nil
}

// SetSystemProxy 设置系统代理（参考 flyclash 实现）
func SetSystemProxy(host string, port int) error {
	services, err := getNetworkServices()
	if err != nil {
		return fmt.Errorf("failed to get network services: %v", err)
	}

	portStr := fmt.Sprintf("%d", port)
	fmt.Printf("🔧 设置系统代理: %s:%s\n", host, portStr)

	for _, service := range services {
		// 设置 HTTP 代理（会自动启用）
		if err := exec.Command("networksetup", "-setwebproxy", service, host, portStr).Run(); err != nil {
			fmt.Printf("⚠ %s: HTTP 代理设置失败\n", service)
			continue
		}
		fmt.Printf("✓ %s: HTTP 代理已启用\n", service)

		// 设置 HTTPS 代理（会自动启用）
		if err := exec.Command("networksetup", "-setsecurewebproxy", service, host, portStr).Run(); err != nil {
			fmt.Printf("⚠ %s: HTTPS 代理设置失败\n", service)
		} else {
			fmt.Printf("✓ %s: HTTPS 代理已启用\n", service)
		}

		// 设置 SOCKS 代理（会自动启用）
		if err := exec.Command("networksetup", "-setsocksfirewallproxy", service, host, portStr).Run(); err != nil {
			fmt.Printf("⚠ %s: SOCKS 代理设置失败\n", service)
		} else {
			fmt.Printf("✓ %s: SOCKS 代理已启用\n", service)
		}

		// 设置绕过代理的域名（与 flyclash 一致）
		exec.Command("networksetup", "-setproxybypassdomains", service, "localhost", "127.0.0.1", "::1", "*.local").Run()
		fmt.Printf("✓ %s: 绕过域名已设置\n", service)
	}

	fmt.Println("✅ 系统代理设置完成")
	return nil
}

// ClearSystemProxy 清除系统代理
func ClearSystemProxy() error {
	services, err := getNetworkServices()
	if err != nil {
		return fmt.Errorf("failed to get network services: %v", err)
	}

	for _, service := range services {
		exec.Command("networksetup", "-setwebproxystate", service, "off").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", service, "off").Run()
		exec.Command("networksetup", "-setsocksfirewallproxystate", service, "off").Run()
	}

	return nil
}

// GetSystemProxyStatus 获取系统代理状态
func GetSystemProxyStatus() (bool, string, int, error) {
	services, err := getNetworkServices()
	if err != nil || len(services) == 0 {
		return false, "", 0, err
	}

	// 检查第一个服务的代理状态
	service := services[0]
	cmd := exec.Command("networksetup", "-getwebproxy", service)
	output, err := cmd.Output()
	if err != nil {
		return false, "", 0, err
	}

	lines := strings.Split(string(output), "\n")
	var enabled bool
	var host string
	var port int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Enabled:") {
			enabled = strings.Contains(line, "Yes")
		} else if strings.HasPrefix(line, "Server:") {
			host = strings.TrimPrefix(line, "Server: ")
		} else if strings.HasPrefix(line, "Port:") {
			fmt.Sscanf(strings.TrimPrefix(line, "Port: "), "%d", &port)
		}
	}

	return enabled, host, port, nil
}

// BrowserInfo 浏览器信息
type BrowserInfo struct {
	Name            string `json:"name"`
	BundleID        string `json:"bundleId"`
	Path            string `json:"path"`
	FollowsSystem   bool   `json:"followsSystem"`   // 是否跟随系统代理
	ProxyConfigured bool   `json:"proxyConfigured"` // 是否已配置代理
}

// BrowserProxyBackup 浏览器代理备份
type BrowserProxyBackup struct {
	Chrome  map[string]interface{} `json:"chrome,omitempty"`
	Edge    map[string]interface{} `json:"edge,omitempty"`
	Firefox map[string]interface{} `json:"firefox,omitempty"`
}

var browserBackupPath string

// GetInstalledBrowsers 获取已安装的浏览器列表
func GetInstalledBrowsers() []BrowserInfo {
	// 常见浏览器的 Bundle ID 和名称
	knownBrowsers := []struct {
		name          string
		bundleID      string
		followsSystem bool
	}{
		{"Safari", "com.apple.Safari", true},
		{"Google Chrome", "com.google.Chrome", true},
		{"Microsoft Edge", "com.microsoft.edgemac", true},
		{"Arc", "company.thebrowser.Browser", true},
		{"Brave Browser", "com.brave.Browser", true},
		{"Opera", "com.operasoftware.Opera", true},
		{"Vivaldi", "com.vivaldi.Vivaldi", true},
		{"Firefox", "org.mozilla.firefox", false}, // Firefox 不跟随系统代理
		{"Firefox Developer Edition", "org.mozilla.firefoxdeveloperedition", false},
		{"Firefox Nightly", "org.mozilla.nightly", false},
	}

	var browsers []BrowserInfo

	for _, browser := range knownBrowsers {
		// 使用 mdfind 查找应用
		cmd := exec.Command("mdfind", fmt.Sprintf("kMDItemCFBundleIdentifier == '%s'", browser.bundleID))
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		path := strings.TrimSpace(string(output))
		if path == "" {
			continue
		}

		// 取第一个路径
		paths := strings.Split(path, "\n")
		appPath := paths[0]

		info := BrowserInfo{
			Name:          browser.name,
			BundleID:      browser.bundleID,
			Path:          appPath,
			FollowsSystem: browser.followsSystem,
		}

		// 检查 Firefox 是否已配置系统代理
		if !browser.followsSystem && strings.Contains(browser.bundleID, "firefox") {
			info.ProxyConfigured = isFirefoxProxyConfigured()
		}

		browsers = append(browsers, info)
	}

	return browsers
}

// isFirefoxProxyConfigured 检查 Firefox 是否已配置使用系统代理
func isFirefoxProxyConfigured() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	profilesDir := filepath.Join(homeDir, "Library", "Application Support", "Firefox", "Profiles")
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
		// 检查是否已设置使用系统代理 (network.proxy.type = 5)
		if strings.Contains(string(content), `"network.proxy.type", 5`) {
			return true
		}
	}
	return false
}

// ConfigureFirefoxProxy 配置 Firefox 使用系统代理
func ConfigureFirefoxProxy() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户目录: %v", err)
	}

	profilesDir := filepath.Join(homeDir, "Library", "Application Support", "Firefox", "Profiles")
	profiles, err := os.ReadDir(profilesDir)
	if err != nil {
		return fmt.Errorf("未找到 Firefox 配置目录: %v", err)
	}

	configured := 0
	for _, profile := range profiles {
		if !profile.IsDir() || !strings.HasSuffix(profile.Name(), ".default") && !strings.HasSuffix(profile.Name(), ".default-release") && !strings.Contains(profile.Name(), "default") {
			continue
		}

		profilePath := filepath.Join(profilesDir, profile.Name())
		userJS := filepath.Join(profilePath, "user.js")

		// Firefox 代理配置
		// network.proxy.type:
		//   0 = 直连
		//   1 = 手动配置
		//   2 = PAC
		//   4 = 自动检测
		//   5 = 使用系统代理
		proxyConfig := `// P-BOX Auto Configuration - Firefox System Proxy
// 自动配置 Firefox 使用系统代理
user_pref("network.proxy.type", 5);
`

		// 读取现有内容
		existingContent, _ := os.ReadFile(userJS)

		// 检查是否已经配置
		if strings.Contains(string(existingContent), `"network.proxy.type", 5`) {
			fmt.Printf("✓ Firefox profile %s 已配置系统代理\n", profile.Name())
			configured++
			continue
		}

		// 移除旧的 P-BOX 配置（如果有）
		content := string(existingContent)
		if idx := strings.Index(content, "// P-BOX Auto Configuration"); idx != -1 {
			// 找到配置块的结束位置
			endIdx := strings.Index(content[idx:], "\n\n")
			if endIdx != -1 {
				content = content[:idx] + content[idx+endIdx+2:]
			}
		}

		// 添加新配置
		newContent := proxyConfig + content

		if err := os.WriteFile(userJS, []byte(newContent), 0644); err != nil {
			fmt.Printf("⚠ Firefox profile %s 配置失败: %v\n", profile.Name(), err)
			continue
		}

		fmt.Printf("✓ Firefox profile %s 已配置使用系统代理\n", profile.Name())
		configured++
	}

	if configured == 0 {
		return fmt.Errorf("未找到 Firefox 配置文件")
	}

	fmt.Printf("✅ 已配置 %d 个 Firefox profile 使用系统代理\n", configured)
	fmt.Println("⚠️  请重启 Firefox 使配置生效")
	return nil
}

// ClearFirefoxProxy 清除 Firefox 代理配置（恢复直连）
func ClearFirefoxProxy() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	profilesDir := filepath.Join(homeDir, "Library", "Application Support", "Firefox", "Profiles")
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

		// 移除 P-BOX 配置块
		newContent := string(content)
		if idx := strings.Index(newContent, "// P-BOX Auto Configuration"); idx != -1 {
			// 找到下一个空行或文件结束
			endIdx := strings.Index(newContent[idx:], "\nuser_pref")
			if endIdx == -1 {
				// 没有其他配置，找到配置块结束
				lines := strings.Split(newContent[idx:], "\n")
				endIdx = 0
				for i, line := range lines {
					if !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "user_pref") && strings.TrimSpace(line) != "" {
						break
					}
					if strings.HasPrefix(line, "user_pref") {
						endIdx = len(strings.Join(lines[:i+1], "\n")) + 1
					}
				}
			}
			if endIdx > 0 {
				newContent = newContent[:idx] + newContent[idx+endIdx:]
			}
		}

		// 只在内容有变化时写入
		if newContent != string(content) {
			os.WriteFile(userJS, []byte(strings.TrimSpace(newContent)+"\n"), 0644)
		}
	}

	return nil
}

// SetBrowserBackupPath 设置备份路径
func SetBrowserBackupPath(dataDir string) {
	browserBackupPath = filepath.Join(dataDir, "browser_proxy_backup.json")
}

// ConfigureAllBrowsersProxy 配置所有浏览器使用系统代理（启动代理时调用）
func ConfigureAllBrowsersProxy() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 先备份当前设置
	backupBrowserSettings()

	// 配置 Chrome
	configureChromiumProxy(homeDir, "Google/Chrome", "com.google.Chrome")

	// 配置 Edge
	configureChromiumProxy(homeDir, "Microsoft Edge", "com.microsoft.Edge")

	// 配置 Arc
	configureChromiumProxy(homeDir, "Arc", "company.thebrowser.Browser")

	// 配置 Brave
	configureChromiumProxy(homeDir, "BraveSoftware/Brave-Browser", "com.brave.Browser")

	// 配置 Firefox
	ConfigureFirefoxProxy()

	fmt.Println("✅ 所有浏览器已配置使用系统代理")
	return nil
}

// configureChromiumProxy 配置 Chromium 系浏览器使用系统代理
// 通过强制策略禁用扩展程序的代理控制
func configureChromiumProxy(homeDir, browserName, bundleID string) error {
	// 检查浏览器是否安装
	cmd := exec.Command("mdfind", fmt.Sprintf("kMDItemCFBundleIdentifier == '%s'", bundleID))
	output, _ := cmd.Output()
	if strings.TrimSpace(string(output)) == "" {
		return nil // 未安装，跳过
	}

	// 获取 Chrome 配置目录
	var policyDir string
	switch bundleID {
	case "com.google.Chrome":
		policyDir = filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "policies", "managed")
	case "com.microsoft.Edge":
		policyDir = filepath.Join(homeDir, "Library", "Application Support", "Microsoft Edge", "policies", "managed")
	case "com.brave.Browser":
		policyDir = filepath.Join(homeDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser", "policies", "managed")
	case "company.thebrowser.Browser":
		policyDir = filepath.Join(homeDir, "Library", "Application Support", "Arc", "User Data", "policies", "managed")
	default:
		return nil
	}

	// 方法1: 创建策略目录和文件
	os.MkdirAll(policyDir, 0755)
	policyPath := filepath.Join(policyDir, "proxy_policy.json")
	policyContent := `{
  "ProxyMode": "system",
  "ProxySettings": {
    "ProxyMode": "system"
  }
}`
	os.WriteFile(policyPath, []byte(policyContent), 0644)

	// 方法2: 检测并配置 SwitchyOmega
	if bundleID == "com.google.Chrome" {
		configureSwitchyOmega(homeDir)
	}

	fmt.Printf("✓ %s 代理配置完成\n", browserName)
	return nil
}

// configureSwitchyOmega 配置 SwitchyOmega 切换到系统代理模式
func configureSwitchyOmega(homeDir string) {
	// SwitchyOmega 扩展 ID
	extIDs := []string{
		"pfnededegaaopdmhkdmcofjmoldfiped", // SwitchyOmega 3 / ZeroOmega
		"padekgcemlokbadohgkifijomclgjgif", // Proxy SwitchyOmega (旧版)
	}

	found := false
	for _, extID := range extIDs {
		extDir := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "Local Extension Settings", extID)
		if _, err := os.Stat(extDir); err == nil {
			found = true
			fmt.Printf("✓ 检测到 SwitchyOmega 扩展\n")
			break
		}
	}

	if !found {
		return
	}

	// 提示用户切换 SwitchyOmega 到系统代理
	fmt.Println("")
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  📌 检测到 SwitchyOmega 扩展                              ║")
	fmt.Println("║  请点击 Chrome 工具栏的 SwitchyOmega 图标                 ║")
	fmt.Println("║  然后选择 [系统代理] 选项                                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println("")

	// 打开一个新标签页提示用户
	// 使用 Chrome 的通知而不是打开新页面
}

// RestoreAllBrowsersProxy 恢复所有浏览器代理设置（停止代理时调用）
func RestoreAllBrowsersProxy() error {
	homeDir, _ := os.UserHomeDir()

	// 删除 Chrome 系浏览器的策略文件
	policyPaths := []string{
		filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "policies", "managed", "proxy_policy.json"),
		filepath.Join(homeDir, "Library", "Application Support", "Microsoft Edge", "policies", "managed", "proxy_policy.json"),
		filepath.Join(homeDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser", "policies", "managed", "proxy_policy.json"),
		filepath.Join(homeDir, "Library", "Application Support", "Arc", "User Data", "policies", "managed", "proxy_policy.json"),
	}

	for _, policyPath := range policyPaths {
		if err := os.Remove(policyPath); err == nil {
			fmt.Printf("✓ 已删除策略文件: %s\n", filepath.Base(filepath.Dir(filepath.Dir(policyPath))))
		}
	}

	// 尝试从备份恢复
	restoreBrowserSettings()

	// 清除 Firefox 配置
	ClearFirefoxProxy()

	fmt.Println("✓ 浏览器代理设置已恢复，请重启浏览器")
	return nil
}

// backupBrowserSettings 备份浏览器设置
func backupBrowserSettings() error {
	if browserBackupPath == "" {
		return nil
	}

	backup := BrowserProxyBackup{
		Chrome:  make(map[string]interface{}),
		Edge:    make(map[string]interface{}),
		Firefox: make(map[string]interface{}),
	}

	// 备份 Chrome 设置
	output, err := exec.Command("defaults", "read", "com.google.Chrome", "ProxyMode").Output()
	if err == nil {
		backup.Chrome["ProxyMode"] = strings.TrimSpace(string(output))
	}

	// 备份 Edge 设置
	output, err = exec.Command("defaults", "read", "com.microsoft.Edge", "ProxyMode").Output()
	if err == nil {
		backup.Edge["ProxyMode"] = strings.TrimSpace(string(output))
	}

	// 备份 Brave 设置
	output, err = exec.Command("defaults", "read", "com.brave.Browser", "ProxyMode").Output()
	if err == nil {
		backup.Chrome["BraveProxyMode"] = strings.TrimSpace(string(output))
	}

	// 保存备份
	data, _ := json.MarshalIndent(backup, "", "  ")
	return os.WriteFile(browserBackupPath, data, 0644)
}

// restoreBrowserSettings 恢复浏览器设置
func restoreBrowserSettings() error {
	if browserBackupPath == "" {
		return fmt.Errorf("no backup path")
	}

	data, err := os.ReadFile(browserBackupPath)
	if err != nil {
		return err
	}

	var backup BrowserProxyBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return err
	}

	// 恢复 Chrome 设置
	if mode, ok := backup.Chrome["ProxyMode"].(string); ok && mode != "" {
		exec.Command("defaults", "write", "com.google.Chrome", "ProxyMode", "-string", mode).Run()
	} else {
		exec.Command("defaults", "delete", "com.google.Chrome", "ProxyMode").Run()
	}

	// 恢复 Edge 设置
	if mode, ok := backup.Edge["ProxyMode"].(string); ok && mode != "" {
		exec.Command("defaults", "write", "com.microsoft.Edge", "ProxyMode", "-string", mode).Run()
	} else {
		exec.Command("defaults", "delete", "com.microsoft.Edge", "ProxyMode").Run()
	}

	// 恢复 Brave 设置
	if mode, ok := backup.Chrome["BraveProxyMode"].(string); ok && mode != "" {
		exec.Command("defaults", "write", "com.brave.Browser", "ProxyMode", "-string", mode).Run()
	} else {
		exec.Command("defaults", "delete", "com.brave.Browser", "ProxyMode").Run()
	}

	// 删除备份文件
	os.Remove(browserBackupPath)
	return nil
}
