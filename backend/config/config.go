package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	DataDir  string         `yaml:"data_dir"`
	Core     CoreConfig     `yaml:"core"`
	Proxy    ProxyConfig    `yaml:"proxy"`
	Log      LogConfig      `yaml:"log"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// CoreConfig 代理核心配置
type CoreConfig struct {
	Type      string `yaml:"type"` // mihomo 或 singbox
	APIPort   int    `yaml:"api_port"`
	APISecret string `yaml:"api_secret"`
}

// ProxyConfig 代理设置
type ProxyConfig struct {
	MixedPort int    `yaml:"mixed_port"`
	SocksPort int    `yaml:"socks_port"`
	AllowLan  bool   `yaml:"allow_lan"`
	IPv6      bool   `yaml:"ipv6"`
	Mode      string `yaml:"mode"` // rule, global, direct
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string `yaml:"level"` // debug, info, warn, error
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// IsDevMode 检测是否为开发模式
// 开发模式：通过环境变量 DEV_MODE=1 或 go run 运行
func IsDevMode() bool {
	// 优先检查环境变量
	if os.Getenv("DEV_MODE") == "1" || os.Getenv("DEV_MODE") == "true" {
		return true
	}

	exe, err := os.Executable()
	if err != nil {
		return true
	}
	// go run 时可执行文件在临时目录
	dir := filepath.Dir(exe)
	return filepath.Base(dir) == "exe" || // Windows: go-build.../exe
		filepath.Dir(dir) == os.TempDir() || // macOS/Linux 临时目录
		strings.Contains(dir, "go-build") // go-build 目录
}

// GetExecutableDir 获取可执行文件所在目录（绝对路径）
// 开发模式使用当前工作目录，生产模式使用可执行文件目录
func GetExecutableDir() string {
	// 开发模式：使用当前工作目录
	if IsDevMode() {
		wd, err := os.Getwd()
		if err == nil {
			return wd
		}
	}

	// 生产模式：使用可执行文件目录
	exe, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exe)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	baseDir := GetExecutableDir()
	dataDir := filepath.Join(baseDir, "data")

	return &Config{
		Server: ServerConfig{
			Port: 8383,
			Host: "0.0.0.0",
		},
		DataDir: dataDir,
		Core: CoreConfig{
			Type:    "mihomo",
			APIPort: 9090,
		},
		Proxy: ProxyConfig{
			MixedPort: 7890,
			SocksPort: 7891,
			AllowLan:  true, // 远程访问需要允许局域网
			IPv6:      false,
			Mode:      "rule",
		},
		Log: LogConfig{
			Level:   "info",
			File:    filepath.Join(dataDir, "logs", "p-box.log"),
			Console: true,
		},
		Security: SecurityConfig{
			Enabled:  false,
			Username: "admin",
			Password: "admin123",
		},
	}
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 如果配置文件存在，加载它
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// 确保所有路径都是绝对路径
	cfg.ensureAbsolutePaths()

	return cfg, nil
}

// ensureAbsolutePaths 确保所有路径都是绝对路径
func (c *Config) ensureAbsolutePaths() {
	baseDir := GetExecutableDir()

	// DataDir
	if !filepath.IsAbs(c.DataDir) {
		c.DataDir = filepath.Join(baseDir, c.DataDir)
	}

	// Log.File
	if c.Log.File != "" && !filepath.IsAbs(c.Log.File) {
		c.Log.File = filepath.Join(baseDir, c.Log.File)
	}
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
