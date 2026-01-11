package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"p-box/backend/modules/system"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

type ProxyMode string

const (
	ModeRule   ProxyMode = "rule"
	ModeGlobal ProxyMode = "global"
	ModeDirect ProxyMode = "direct"
)

type ProxyStatus struct {
	Running         bool      `json:"running"`
	CoreType        string    `json:"coreType"`
	CoreVersion     string    `json:"coreVersion"`
	Mode            ProxyMode `json:"mode"`
	MixedPort       int       `json:"mixedPort"`
	SocksPort       int       `json:"socksPort"`
	AllowLan        bool      `json:"allowLan"`
	TunEnabled      bool      `json:"tunEnabled"`
	TransparentMode string    `json:"transparentMode"` // off, tun, tproxy, redirect
	StartTime       time.Time `json:"startTime,omitempty"`
	Uptime          int64     `json:"uptime"`
	ConfigPath      string    `json:"configPath,omitempty"`
	ApiAddress      string    `json:"apiAddress,omitempty"`
}

type ProxyConfig struct {
	MixedPort          int    `json:"mixedPort" yaml:"mixed-port"`
	SocksPort          int    `json:"socksPort" yaml:"socks-port"`
	RedirPort          int    `json:"redirPort" yaml:"redir-port"`   // REDIRECT 端口
	TProxyPort         int    `json:"tproxyPort" yaml:"tproxy-port"` // TPROXY 端口
	AllowLan           bool   `json:"allowLan" yaml:"allow-lan"`
	IPv6               bool   `json:"ipv6" yaml:"ipv6"`
	Mode               string `json:"mode" yaml:"mode"`
	LogLevel           string `json:"logLevel" yaml:"log-level"`
	ExternalController string `json:"externalController" yaml:"external-controller"`
	TunEnabled         bool   `json:"tunEnabled" yaml:"tun-enabled"`
	TunStack           string `json:"tunStack" yaml:"tun-stack"`               // system, gvisor, mixed
	TransparentMode    string `json:"transparentMode" yaml:"transparent-mode"` // off, tun, tproxy, redirect
	AutoStart          bool   `json:"autoStart" yaml:"auto-start"`             // 开机自动启动
	AutoStartDelay     int    `json:"autoStartDelay" yaml:"auto-start-delay"`  // 自动启动延迟（秒）
}

// NodeProvider 节点提供者接口
type NodeProvider func() []ProxyNode

// SettingsProvider 设置提供者接口（获取代理设置）
type SettingsProvider func() *ProxySettings

type Service struct {
	dataDir          string
	coreType         string
	config           *ProxyConfig
	configGenerator  *ConfigGenerator
	singboxGenerator *SingboxGenerator
	configTemplate   *ConfigTemplate
	process          *exec.Cmd
	running          bool
	startTime        time.Time
	configPath       string
	mu               sync.RWMutex

	// 节点提供者（从节点管理模块获取过滤后的节点）
	nodeProvider NodeProvider

	// 设置提供者（从设置模块获取代理设置）
	settingsProvider SettingsProvider

	// 日志收集
	logs  []string
	logMu sync.RWMutex

	// 启动回调（用于通知其他模块 VPN 已启动）
	onStartCallback func()
}

func NewService(dataDir string) *Service {
	// 根据平台选择默认透明代理模式
	defaultTransparentMode := "off"
	defaultTunEnabled := false
	if runtime.GOOS == "linux" {
		// Linux 支持 TUN 模式（需要 root 权限）
		defaultTransparentMode = "tun"
		defaultTunEnabled = true
	}
	// macOS/Windows 默认使用系统代理模式，不启用 TUN

	s := &Service{
		dataDir:  dataDir,
		coreType: "mihomo",
		config: &ProxyConfig{
			MixedPort:          7890,
			SocksPort:          7891,
			RedirPort:          7892,
			TProxyPort:         7893,
			AllowLan:           true,
			IPv6:               false,
			Mode:               "rule",
			LogLevel:           "info",
			ExternalController: "127.0.0.1:9090",
			TunEnabled:         defaultTunEnabled,
			TunStack:           "mixed",
			TransparentMode:    defaultTransparentMode,
			AutoStart:          false,
			AutoStartDelay:     15, // 默认延迟 15 秒
		},
		configGenerator:  NewConfigGenerator(dataDir),
		singboxGenerator: NewSingboxGenerator(dataDir),
		configTemplate:   GetDefaultConfigTemplate(),
	}
	s.loadConfig()
	s.loadConfigTemplate()
	return s
}

// AutoStartIfEnabled 如果开启了自动启动，则在延迟后启动代理
func (s *Service) AutoStartIfEnabled() {
	if !s.config.AutoStart {
		return
	}

	delay := s.config.AutoStartDelay
	if delay < 0 {
		delay = 0
	}

	fmt.Printf("⏳ 自动启动已开启，将在 %d 秒后启动代理...\n", delay)

	go func() {
		time.Sleep(time.Duration(delay) * time.Second)

		s.mu.RLock()
		if s.running {
			s.mu.RUnlock()
			fmt.Println("✓ 代理已在运行，跳过自动启动")
			return
		}
		s.mu.RUnlock()

		fmt.Println("🚀 开始自动启动代理...")
		if err := s.Start(); err != nil {
			fmt.Printf("❌ 自动启动代理失败: %v\n", err)
		} else {
			fmt.Println("✓ 代理自动启动成功")
		}
	}()
}

func (s *Service) loadConfig() {
	configFile := filepath.Join(s.dataDir, "proxy_settings.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	// 保存默认值
	defaults := *s.config

	// 加载配置
	json.Unmarshal(data, s.config)

	// 对于零值字段，恢复默认值
	if s.config.MixedPort == 0 {
		s.config.MixedPort = defaults.MixedPort
	}
	if s.config.SocksPort == 0 {
		s.config.SocksPort = defaults.SocksPort
	}
	if s.config.RedirPort == 0 {
		s.config.RedirPort = defaults.RedirPort
	}
	if s.config.TProxyPort == 0 {
		s.config.TProxyPort = defaults.TProxyPort
	}
	if s.config.Mode == "" {
		s.config.Mode = defaults.Mode
	}
	if s.config.LogLevel == "" {
		s.config.LogLevel = defaults.LogLevel
	}
	if s.config.ExternalController == "" {
		s.config.ExternalController = defaults.ExternalController
	}
	if s.config.TunStack == "" {
		s.config.TunStack = defaults.TunStack
	}
	if s.config.TransparentMode == "" {
		s.config.TransparentMode = defaults.TransparentMode
	}

	// macOS/Windows 上强制使用系统代理模式（TUN 需要 root 权限）
	if runtime.GOOS != "linux" && s.config.TransparentMode == "tun" {
		fmt.Println("⚠️ 检测到非 Linux 系统，TUN 模式需要 root 权限，自动切换为系统代理模式")
		s.config.TransparentMode = "off"
		s.config.TunEnabled = false
	}
	if s.config.AutoStartDelay == 0 {
		s.config.AutoStartDelay = defaults.AutoStartDelay
	}
}

func (s *Service) saveConfig() error {
	configFile := filepath.Join(s.dataDir, "proxy_settings.json")
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

func (s *Service) GetStatus() *ProxyStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &ProxyStatus{
		Running:         s.running,
		CoreType:        s.coreType,
		Mode:            ProxyMode(s.config.Mode),
		MixedPort:       s.config.MixedPort,
		SocksPort:       s.config.SocksPort,
		AllowLan:        s.config.AllowLan,
		TunEnabled:      s.config.TunEnabled,
		TransparentMode: s.config.TransparentMode,
		ConfigPath:      s.configPath,
		ApiAddress:      s.config.ExternalController,
	}

	if s.running {
		status.StartTime = s.startTime
		status.Uptime = int64(time.Since(s.startTime).Seconds())
	}

	return status
}

func (s *Service) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("proxy is already running")
	}

	corePath := s.findCorePath()
	if corePath == "" {
		s.mu.Unlock()
		return fmt.Errorf("核心文件未找到，请先下载核心")
	}
	s.mu.Unlock() // 释放锁再调用 regenerateConfig

	// 每次启动都重新生成配置（确保配置是最新的）
	configPath, err := s.regenerateConfig()
	if err != nil {
		// 如果重新生成失败，尝试使用已有配置
		if s.coreType == "singbox" {
			configPath = filepath.Join(s.dataDir, "configs", "singbox-config.json")
		} else {
			configPath = filepath.Join(s.dataDir, "configs", "config.yaml")
		}
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("配置文件未找到，请先生成配置")
		}
		fmt.Printf("⚠️ 重新生成配置失败，使用已有配置: %v\n", err)
	}

	s.mu.Lock()         // 重新获取锁
	defer s.mu.Unlock() // 确保释放

	// 再次检查是否已经运行（防止并发启动）
	if s.running {
		return fmt.Errorf("proxy is already running")
	}

	// 创建运行时目录
	runtimeDir := filepath.Join(s.dataDir, "runtime")
	os.MkdirAll(runtimeDir, 0755)

	// 检查并处理系统 DNS 服务（TUN 模式需要 53 端口）
	if s.config.TunEnabled || s.config.TransparentMode == "tun" {
		s.prepareSystemForTUN()
	}

	// 构建命令 - 根据核心类型使用不同参数
	// Mihomo: -d <workdir> -f <config>
	// Sing-Box: run -D <workdir> -c <config>
	if s.coreType == "singbox" {
		s.process = exec.Command(corePath, "run", "-D", s.dataDir, "-c", configPath)
		// 启用已弃用的特殊出站（direct），代理组需要引用"直连"
		s.process.Env = append(os.Environ(), "ENABLE_DEPRECATED_SPECIAL_OUTBOUNDS=true")
	} else {
		s.process = exec.Command(corePath, "-d", s.dataDir, "-f", configPath)
	}
	s.process.Dir = s.dataDir

	// 创建管道捕获输出
	stdout, _ := s.process.StdoutPipe()
	stderr, _ := s.process.StderrPipe()

	if err := s.process.Start(); err != nil {
		return fmt.Errorf("启动核心失败: %w", err)
	}

	// 启动日志收集
	go s.collectLogs(stdout)
	go s.collectLogs(stderr)

	s.running = true
	s.startTime = time.Now()
	s.configPath = configPath

	// 监控进程
	go func() {
		s.process.Wait()
		s.mu.Lock()
		s.running = false
		s.process = nil
		s.mu.Unlock()
	}()

	// 根据透明代理模式自动设置系统代理（macOS/Windows）
	if s.config.TransparentMode == "off" {
		fmt.Println("🔧 检测到系统代理模式，自动设置系统代理...")
		if err := system.SetSystemProxy("127.0.0.1", s.config.MixedPort); err != nil {
			fmt.Printf("⚠️  设置系统代理失败: %v\n", err)
		} else {
			fmt.Println("✅ 系统代理已自动启用")
		}

		// 配置所有浏览器使用系统代理（备份用户原有设置）
		go s.configureAllBrowsers()
	}

	// 调用启动回调（通知其他模块 VPN 已启动）
	if s.onStartCallback != nil {
		s.onStartCallback()
	}

	return nil
}

// configureAllBrowsers 配置所有浏览器使用系统代理
func (s *Service) configureAllBrowsers() {
	// 设置备份路径
	system.SetBrowserBackupPath(s.dataDir)

	fmt.Println("🌐 正在配置浏览器使用系统代理...")
	if err := system.ConfigureAllBrowsersProxy(); err != nil {
		fmt.Printf("⚠️  配置浏览器失败: %v\n", err)
	}
}

func (s *Service) Stop() error {
	s.mu.Lock()

	if !s.running {
		s.mu.Unlock()
		return nil
	}

	wasTunEnabled := s.config.TunEnabled || s.config.TransparentMode == "tun"

	if s.process != nil && s.process.Process != nil {
		if err := s.process.Process.Kill(); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("failed to stop core: %w", err)
		}
		s.process.Wait()
	}

	s.running = false
	s.process = nil
	s.mu.Unlock()

	// 恢复系统环境（在锁外执行）
	if wasTunEnabled {
		s.restoreSystemAfterTUN()
	}

	// 清除系统代理设置（macOS/Windows）
	if err := system.ClearSystemProxy(); err != nil {
		fmt.Printf("⚠️ 清除系统代理失败: %v\n", err)
	} else {
		fmt.Println("✓ 系统代理已清除")
	}

	// 恢复浏览器代理设置（恢复用户原有配置）
	if err := system.RestoreAllBrowsersProxy(); err != nil {
		fmt.Printf("⚠️ 恢复浏览器设置失败: %v\n", err)
	}

	return nil
}

func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

// collectLogs 收集日志输出
func (s *Service) collectLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		s.addLog(line)
		// 同时输出到控制台
		fmt.Println(line)
	}
}

// addLog 添加日志
func (s *Service) addLog(line string) {
	s.logMu.Lock()
	defer s.logMu.Unlock()

	s.logs = append(s.logs, line)
	// 保留最近 1000 条日志
	if len(s.logs) > 1000 {
		s.logs = s.logs[len(s.logs)-1000:]
	}
}

// GetLogs 获取日志
func (s *Service) GetLogs(limit int) []string {
	s.logMu.RLock()
	defer s.logMu.RUnlock()

	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}

	start := len(s.logs) - limit
	if start < 0 {
		start = 0
	}

	result := make([]string, limit)
	copy(result, s.logs[start:])
	return result
}

// ClearLogs 清除日志
func (s *Service) ClearLogs() {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	s.logs = nil
}

// SetNodeProvider 设置节点提供者
func (s *Service) SetNodeProvider(provider NodeProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeProvider = provider
}

// SetSettingsProvider 设置代理设置提供者
func (s *Service) SetSettingsProvider(provider SettingsProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settingsProvider = provider
}

// SetOnStartCallback 设置启动回调（VPN 启动后调用）
func (s *Service) SetOnStartCallback(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStartCallback = callback
}

// RegenerateConfig 从节点管理模块获取过滤后的节点并生成配置（公开方法）
func (s *Service) RegenerateConfig() (string, error) {
	return s.regenerateConfig()
}

// regenerateConfig 从节点管理模块获取过滤后的节点并生成配置
// 注意：调用此方法时不能持有 s.mu 锁
func (s *Service) regenerateConfig() (string, error) {
	provider := s.nodeProvider // nodeProvider 在初始化后不会改变，无需加锁

	if provider == nil {
		return "", fmt.Errorf("节点提供者未设置")
	}

	allNodes := provider()
	if len(allNodes) == 0 {
		return "", fmt.Errorf("没有可用节点")
	}

	fmt.Printf("🔄 重新生成配置，共 %d 个节点\n", len(allNodes))
	return s.GenerateConfig(allNodes)
}

// GetConfigContent 读取生成的 config.yaml 文件内容
func (s *Service) GetConfigContent() (string, error) {
	configPath := filepath.Join(s.dataDir, "configs", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("配置文件不存在: %w", err)
	}
	return string(data), nil
}

func (s *Service) SetMode(mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch mode {
	case "rule", "global", "direct":
		s.config.Mode = mode
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}

	return nil
}

// SetTunEnabled 设置 TUN 模式开关 (兼容旧接口)
func (s *Service) SetTunEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.TunEnabled = enabled
	if enabled {
		s.config.TransparentMode = "tun"
	} else {
		s.config.TransparentMode = "off"
	}
	return nil
}

// SetTransparentMode 设置透明代理模式
// mode: off (关闭), tun (TUN模式), tproxy (TPROXY), redirect (REDIRECT)
func (s *Service) SetTransparentMode(mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch mode {
	case "off", "tun", "tproxy", "redirect":
		s.config.TransparentMode = mode
		s.config.TunEnabled = (mode == "tun")
		return s.saveConfig()
	default:
		return fmt.Errorf("invalid transparent mode: %s", mode)
	}
}

func (s *Service) GetConfig() *ProxyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Service) UpdateConfig(config *ProxyConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return s.saveConfig()
}

// PatchConfig 部分更新配置（只更新传入的字段）
func (s *Service) PatchConfig(updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 根据传入的字段更新配置
	if v, ok := updates["mixedPort"]; ok {
		if val, ok := v.(float64); ok {
			s.config.MixedPort = int(val)
		}
	}
	if v, ok := updates["socksPort"]; ok {
		if val, ok := v.(float64); ok {
			s.config.SocksPort = int(val)
		}
	}
	if v, ok := updates["redirPort"]; ok {
		if val, ok := v.(float64); ok {
			s.config.RedirPort = int(val)
		}
	}
	if v, ok := updates["tproxyPort"]; ok {
		if val, ok := v.(float64); ok {
			s.config.TProxyPort = int(val)
		}
	}
	if v, ok := updates["allowLan"]; ok {
		if val, ok := v.(bool); ok {
			s.config.AllowLan = val
		}
	}
	if v, ok := updates["ipv6"]; ok {
		if val, ok := v.(bool); ok {
			s.config.IPv6 = val
		}
	}
	if v, ok := updates["mode"]; ok {
		if val, ok := v.(string); ok {
			s.config.Mode = val
		}
	}
	if v, ok := updates["logLevel"]; ok {
		if val, ok := v.(string); ok {
			s.config.LogLevel = val
		}
	}
	if v, ok := updates["externalController"]; ok {
		if val, ok := v.(string); ok {
			s.config.ExternalController = val
		}
	}
	if v, ok := updates["tunEnabled"]; ok {
		if val, ok := v.(bool); ok {
			s.config.TunEnabled = val
		}
	}
	if v, ok := updates["tunStack"]; ok {
		if val, ok := v.(string); ok {
			s.config.TunStack = val
		}
	}
	if v, ok := updates["transparentMode"]; ok {
		if val, ok := v.(string); ok {
			s.config.TransparentMode = val
		}
	}
	if v, ok := updates["autoStart"]; ok {
		if val, ok := v.(bool); ok {
			s.config.AutoStart = val
		}
	}
	if v, ok := updates["autoStartDelay"]; ok {
		if val, ok := v.(float64); ok {
			s.config.AutoStartDelay = int(val)
		}
	}

	return s.saveConfig()
}

func (s *Service) findCorePath() string {
	coresDir := filepath.Join(s.dataDir, "cores")
	arch := runtime.GOARCH
	goos := runtime.GOOS

	// 精确匹配
	var binName string
	if s.coreType == "singbox" {
		binName = fmt.Sprintf("sing-box-%s-%s", goos, arch)
	} else {
		binName = fmt.Sprintf("mihomo-%s-%s", goos, arch)
	}

	if goos == "windows" {
		binName += ".exe"
	}

	exactPath := filepath.Join(coresDir, binName)
	if _, err := os.Stat(exactPath); err == nil {
		return exactPath
	}

	// 模糊匹配
	patterns := []string{
		filepath.Join(coresDir, "mihomo*"),
		filepath.Join(coresDir, "sing-box*"),
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return matches[0]
		}
	}

	return ""
}

// GenerateConfig 生成配置文件
func (s *Service) GenerateConfig(nodes []ProxyNode) (string, error) {
	// 根据透明代理模式设置
	enableTUN := s.config.TransparentMode == "tun"
	enableTProxy := s.config.TransparentMode == "tproxy" || s.config.TransparentMode == "redirect"

	options := ConfigGeneratorOptions{
		MixedPort:          s.config.MixedPort,
		AllowLan:           s.config.AllowLan,
		Mode:               s.config.Mode,
		LogLevel:           s.config.LogLevel,
		IPv6:               s.config.IPv6,
		ExternalController: s.config.ExternalController,
		EnableDNS:          true,
		EnhancedMode:       "fake-ip",
		EnableTUN:          enableTUN,
		EnableTProxy:       enableTProxy,
		TProxyPort:         s.config.TProxyPort,
		Template:           s.configTemplate, // 使用配置模板
	}

	// 从代理设置获取优化配置
	if s.settingsProvider != nil {
		settings := s.settingsProvider()
		if settings != nil {
			// 性能优化
			options.UnifiedDelay = settings.UnifiedDelay
			options.TCPConcurrent = settings.TCPConcurrent
			options.FindProcessMode = settings.FindProcessMode
			options.GlobalClientFingerprint = settings.GlobalClientFingerprint
			options.KeepAliveInterval = settings.KeepAliveInterval
			options.KeepAliveIdle = settings.KeepAliveIdle
			options.DisableKeepAlive = settings.DisableKeepAlive

			// GEO 数据
			options.GeodataMode = settings.GeodataMode
			options.GeodataLoader = settings.GeodataLoader
			options.GeositeMatcher = settings.GeositeMatcher
			options.GeoAutoUpdate = settings.GeoAutoUpdate
			options.GeoUpdateInterval = settings.GeoUpdateInterval
			options.GlobalUA = settings.GlobalUA
			options.ETagSupport = settings.ETagSupport

			// TUN 设置
			options.TUNSettings = &settings.TUN
		}
	}

	var configPath string

	if s.coreType == "singbox" {
		// 生成 sing-box 1.12+ 配置
		sbOpts := SingBoxGeneratorOptions{
			Mode:                     "system",
			FakeIP:                   options.EnhancedMode == "fake-ip",
			MixedPort:                options.MixedPort,
			LogLevel:                 options.LogLevel,
			Sniff:                    true,
			SniffOverrideDestination: true,
		}
		// TUN 模式设置
		if options.EnableTUN {
			sbOpts.Mode = "tun"
			if options.TUNSettings != nil {
				sbOpts.TUNStack = options.TUNSettings.Stack
				sbOpts.TUNMTU = options.TUNSettings.MTU
				sbOpts.StrictRoute = options.TUNSettings.StrictRoute
				sbOpts.AutoRedirect = options.TUNSettings.AutoRedirect
			}
		}
		// Clash API
		if options.ExternalController != "" {
			sbOpts.ClashAPIAddr = options.ExternalController
		} else {
			sbOpts.ClashAPIAddr = "127.0.0.1:9090"
		}

		config, err := s.singboxGenerator.GenerateConfigV112(nodes, sbOpts)
		if err != nil {
			return "", err
		}
		path, err := s.singboxGenerator.SaveConfigV112(config, "singbox-config.json")
		if err != nil {
			return "", err
		}
		configPath = path
	} else {
		// 生成 Mihomo/Clash 配置
		config, err := s.configGenerator.GenerateConfig(nodes, options)
		if err != nil {
			return "", err
		}
		path, err := s.configGenerator.SaveConfig(config, "config.yaml")
		if err != nil {
			return "", err
		}
		configPath = path
	}

	s.configPath = configPath
	return configPath, nil
}

// SetCoreType 设置核心类型
func (s *Service) SetCoreType(coreType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coreType = coreType
}

// GetCoreType 获取核心类型
func (s *Service) GetCoreType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coreType
}

// GetConfigGenerator 获取配置生成器
func (s *Service) GetConfigGenerator() *ConfigGenerator {
	return s.configGenerator
}

// loadConfigTemplate 加载配置模板
func (s *Service) loadConfigTemplate() {
	templateFile := filepath.Join(s.dataDir, "config_template.json")
	data, err := os.ReadFile(templateFile)
	if err != nil {
		// 文件不存在，使用默认模板
		return
	}
	var template ConfigTemplate
	if err := json.Unmarshal(data, &template); err == nil {
		s.configTemplate = &template
		// 自动修复旧的英文名称
		s.fixLegacyProxyNames()
		// 自动合并新的默认代理组
		s.mergeDefaultProxyGroups()
	}
}

// fixLegacyProxyNames 修复旧的英文代理名称为中文
func (s *Service) fixLegacyProxyNames() {
	if s.configTemplate == nil {
		return
	}

	// 旧名称到新名称的映射
	nameMap := map[string]string{
		"auto":   "自动选择",
		"direct": "DIRECT",
		"proxy":  "节点选择",
	}

	changed := false
	for i := range s.configTemplate.ProxyGroups {
		group := &s.configTemplate.ProxyGroups[i]
		for j, proxy := range group.Proxies {
			if newName, ok := nameMap[proxy]; ok {
				group.Proxies[j] = newName
				changed = true
			}
		}
	}

	if changed {
		fmt.Println("✓ 自动修复旧的代理组名称引用")
		s.saveConfigTemplate()
	}
}

// mergeDefaultProxyGroups 合并默认代理组（自动添加新的代理组，不覆盖已有的）
func (s *Service) mergeDefaultProxyGroups() {
	if s.configTemplate == nil {
		return
	}

	defaultGroups := GetDefaultProxyGroups()
	existingNames := make(map[string]bool)

	// 记录已存在的代理组名称
	for _, g := range s.configTemplate.ProxyGroups {
		existingNames[g.Name] = true
	}

	// 添加缺失的默认代理组
	var newGroups []ProxyGroupTemplate
	for _, dg := range defaultGroups {
		if !existingNames[dg.Name] {
			dg.Enabled = true // 确保启用
			newGroups = append(newGroups, dg)
			fmt.Printf("✓ 自动添加新代理组: %s\n", dg.Name)
		}
	}

	if len(newGroups) > 0 {
		// 将新组插入到合适位置（按默认顺序）
		s.configTemplate.ProxyGroups = s.insertGroupsInOrder(s.configTemplate.ProxyGroups, newGroups, defaultGroups)
		s.saveConfigTemplate()
	}
}

// insertGroupsInOrder 按默认顺序插入新代理组
func (s *Service) insertGroupsInOrder(existing, newGroups, defaultOrder []ProxyGroupTemplate) []ProxyGroupTemplate {
	// 创建默认顺序映射
	orderMap := make(map[string]int)
	for i, g := range defaultOrder {
		orderMap[g.Name] = i
	}

	// 合并所有组
	all := append(existing, newGroups...)

	// 按默认顺序排序
	sort.Slice(all, func(i, j int) bool {
		oi, oki := orderMap[all[i].Name]
		oj, okj := orderMap[all[j].Name]
		if !oki {
			oi = 999
		}
		if !okj {
			oj = 999
		}
		return oi < oj
	})

	return all
}

// saveConfigTemplate 保存配置模板
func (s *Service) saveConfigTemplate() error {
	templateFile := filepath.Join(s.dataDir, "config_template.json")
	data, err := json.MarshalIndent(s.configTemplate, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(templateFile, data, 0644)
}

// GetConfigTemplate 获取配置模板
func (s *Service) GetConfigTemplate() *ConfigTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configTemplate
}

// UpdateProxyGroups 更新代理组
func (s *Service) UpdateProxyGroups(groups []ProxyGroupTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configTemplate.ProxyGroups = groups
	return s.saveConfigTemplate()
}

// UpdateRules 更新规则
func (s *Service) UpdateRules(rules []RuleTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configTemplate.Rules = rules
	return s.saveConfigTemplate()
}

// UpdateRuleProviders 更新规则提供者
func (s *Service) UpdateRuleProviders(providers []RuleProviderTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configTemplate.RuleProviders = providers
	return s.saveConfigTemplate()
}

// ResetConfigTemplate 重置配置模板（只重置代理组，保留用户自定义规则）
func (s *Service) ResetConfigTemplate() {
	s.mu.Lock()
	defer s.mu.Unlock()

	defaultTemplate := GetDefaultConfigTemplate()

	// 提取用户自定义的规则（非默认规则）
	var customRules []RuleTemplate
	defaultRulePayloads := make(map[string]bool)
	for _, r := range defaultTemplate.Rules {
		key := r.Type + ":" + r.Payload
		defaultRulePayloads[key] = true
	}

	// 保留用户添加的自定义规则
	if s.configTemplate != nil {
		for _, r := range s.configTemplate.Rules {
			key := r.Type + ":" + r.Payload
			if !defaultRulePayloads[key] {
				customRules = append(customRules, r)
			}
		}
	}

	// 使用默认模板
	s.configTemplate = defaultTemplate

	// 将用户自定义规则插入到 MATCH 规则之前
	if len(customRules) > 0 {
		var newRules []RuleTemplate
		for _, r := range s.configTemplate.Rules {
			if r.Type == "MATCH" {
				// 在 MATCH 之前插入自定义规则
				newRules = append(newRules, customRules...)
			}
			newRules = append(newRules, r)
		}
		s.configTemplate.Rules = newRules
	}

	s.saveConfigTemplate()
}

// prepareSystemForTUN 准备系统环境以启用 TUN 模式
// 主要处理：1. 释放 53 端口（停止占用的服务）
//
//  2. 设置 IP 转发
func (s *Service) prepareSystemForTUN() {
	// 检查是否为 Linux
	if runtime.GOOS != "linux" {
		return
	}

	// 1. 检查并释放 53 端口
	s.releasePort53()

	// 2. 启用 IP 转发
	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()
	exec.Command("sysctl", "-w", "net.ipv6.conf.all.forwarding=1").Run()
	s.addLog("已启用 IP 转发")
}

// releasePort53 释放 53 端口
func (s *Service) releasePort53() {
	// 检查 53 端口是否被占用
	if !s.isPortInUse(53) {
		s.addLog("53 端口未被占用，无需处理")
		return
	}

	s.addLog("检测到 53 端口被占用，正在释放...")

	// 方法 1: 停止 systemd-resolved（最常见的占用者）
	if s.isServiceActive("systemd-resolved") {
		s.addLog("检测到 systemd-resolved 服务，正在停止...")
		exec.Command("systemctl", "stop", "systemd-resolved").Run()
		exec.Command("systemctl", "disable", "systemd-resolved").Run()

		// 备份并修改 resolv.conf
		if _, err := os.Stat("/etc/resolv.conf.bak"); os.IsNotExist(err) {
			exec.Command("cp", "/etc/resolv.conf", "/etc/resolv.conf.bak").Run()
		}
		// 删除符号链接并创建新文件，指向 Mihomo 的 DNS
		os.Remove("/etc/resolv.conf")
		os.WriteFile("/etc/resolv.conf", []byte("nameserver 127.0.0.1\n"), 0644)
		s.addLog("已停止 systemd-resolved 并配置 DNS 指向 Mihomo")
	}

	// 方法 2: 停止 dnsmasq（另一个常见的 DNS 服务）
	if s.isServiceActive("dnsmasq") {
		s.addLog("检测到 dnsmasq 服务，正在停止...")
		exec.Command("systemctl", "stop", "dnsmasq").Run()
		s.addLog("已停止 dnsmasq")
	}

	// 方法 3: 使用 fuser 强制杀死占用 53 端口的进程
	if s.isPortInUse(53) {
		s.addLog("尝试使用 fuser 释放 53 端口...")
		exec.Command("fuser", "-k", "53/udp").Run()
		exec.Command("fuser", "-k", "53/tcp").Run()
		time.Sleep(time.Millisecond * 500)
	}

	// 最终检查
	if s.isPortInUse(53) {
		s.addLog("警告：53 端口可能仍被占用，TUN 模式可能无法正常工作")
	} else {
		s.addLog("53 端口已成功释放")
	}
}

// isPortInUse 检查端口是否被占用
func (s *Service) isPortInUse(port int) bool {
	// 尝试监听 UDP
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.ListenPacket("udp", addr)
	if err != nil {
		return true // 端口被占用
	}
	listener.Close()

	// 尝试监听 TCP
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return true // 端口被占用
	}
	tcpListener.Close()

	return false
}

// isServiceActive 检查系统服务是否活动
func (s *Service) isServiceActive(serviceName string) bool {
	output, err := exec.Command("systemctl", "is-active", serviceName).Output()
	if err != nil {
		return false
	}
	return string(output) == "active\n"
}

// restoreSystemAfterTUN 恢复系统环境
func (s *Service) restoreSystemAfterTUN() {
	if runtime.GOOS != "linux" {
		return
	}

	// 恢复 resolv.conf
	if _, err := os.Stat("/etc/resolv.conf.bak"); err == nil {
		exec.Command("cp", "/etc/resolv.conf.bak", "/etc/resolv.conf").Run()
		s.addLog("已恢复 resolv.conf")
	}

	// 重新启动 systemd-resolved
	exec.Command("systemctl", "start", "systemd-resolved").Run()
	s.addLog("已重新启动 systemd-resolved")
}

// ============================================================================
// Sing-Box 相关方法
// ============================================================================

// GetAllNodes 获取所有节点
func (s *Service) GetAllNodes() ([]ProxyNode, error) {
	s.mu.RLock()
	provider := s.nodeProvider
	s.mu.RUnlock()

	if provider == nil {
		return nil, fmt.Errorf("节点提供者未设置")
	}

	nodes := provider()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点")
	}

	return nodes, nil
}

// GetSingBoxConfigContent 读取 Sing-Box 配置文件内容
func (s *Service) GetSingBoxConfigContent() (string, error) {
	configPath := filepath.Join(s.dataDir, "configs", "singbox-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("Sing-Box 配置文件不存在: %w", err)
	}
	return string(data), nil
}

// GetSingBoxTemplate 获取 Sing-Box 模板配置
func (s *Service) GetSingBoxTemplate() *SingBoxTemplate {
	return LoadSingBoxTemplate(s.dataDir)
}

// UpdateSingBoxTemplate 更新 Sing-Box 模板配置
func (s *Service) UpdateSingBoxTemplate(template *SingBoxTemplate) error {
	return SaveSingBoxTemplate(s.dataDir, template)
}

// ResetSingBoxTemplate 重置 Sing-Box 模板为默认值
func (s *Service) ResetSingBoxTemplate() {
	template := GetDefaultSingBoxTemplate()
	SaveSingBoxTemplate(s.dataDir, template)
}
