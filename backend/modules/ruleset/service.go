package ruleset

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RuleFile 规则文件定义
type RuleFile struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
	UpdatedAt   string `json:"updatedAt"`
	Status      string `json:"status"` // pending, downloading, completed, failed
}

// RuleSetConfig 规则集配置
type RuleSetConfig struct {
	AutoUpdate     bool     `json:"autoUpdate"`
	UpdateInterval int      `json:"updateInterval"` // 更新间隔（天）
	LastUpdate     string   `json:"lastUpdate"`
	GitHubProxy    string   `json:"githubProxy"`   // 当前使用的代理
	GitHubProxies  []string `json:"githubProxies"` // 默认代理列表
	CustomProxies  []string `json:"customProxies"` // 用户自定义代理
}

// Service 规则集管理服务
type Service struct {
	dataDir    string
	rulesetDir string
	config     *RuleSetConfig
	configPath string
	mu         sync.RWMutex
	updating   bool
}

// NewService 创建规则集服务
func NewService(dataDir string) *Service {
	rulesetDir := filepath.Join(dataDir, "ruleset")
	os.MkdirAll(rulesetDir, 0755)

	s := &Service{
		dataDir:    dataDir,
		rulesetDir: rulesetDir,
		configPath: filepath.Join(dataDir, "ruleset_config.json"),
		config: &RuleSetConfig{
			AutoUpdate:     true,
			UpdateInterval: 1, // 默认1天
		},
	}

	s.loadConfig()
	go s.autoUpdateLoop()

	return s
}

// GetGeoFiles 获取 GEO 数据文件列表
func (s *Service) GetGeoFiles() []RuleFile {
	baseURL := "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release"

	files := []RuleFile{
		{
			Name:        "geoip.dat",
			URL:         baseURL + "/geoip.dat",
			Path:        filepath.Join(s.dataDir, "geoip.dat"),
			Description: "GeoIP 数据库",
		},
		{
			Name:        "geosite.dat",
			URL:         baseURL + "/geosite.dat",
			Path:        filepath.Join(s.dataDir, "geosite.dat"),
			Description: "GeoSite 域名数据库",
		},
		{
			Name:        "country.mmdb",
			URL:         baseURL + "/country.mmdb",
			Path:        filepath.Join(s.dataDir, "country.mmdb"),
			Description: "MaxMind 国家数据库",
		},
		{
			Name:        "GeoLite2-ASN.mmdb",
			URL:         baseURL + "/GeoLite2-ASN.mmdb",
			Path:        filepath.Join(s.dataDir, "GeoLite2-ASN.mmdb"),
			Description: "ASN 自治系统数据库",
		},
	}

	// 更新文件状态
	for i := range files {
		files[i] = s.updateFileStatus(files[i])
	}

	return files
}

// GetRuleProviderFiles 获取规则提供者文件列表
func (s *Service) GetRuleProviderFiles() []RuleFile {
	baseURL := "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@meta/geo"

	files := []RuleFile{
		// 私有网络
		{Name: "private-domain.mrs", URL: baseURL + "/geosite/private.mrs", Path: filepath.Join(s.rulesetDir, "private-domain.mrs"), Description: "私有网络域名"},
		{Name: "private-ip.mrs", URL: baseURL + "/geoip/private.mrs", Path: filepath.Join(s.rulesetDir, "private-ip.mrs"), Description: "私有网络 IP"},
		// AI 平台
		{Name: "ai-domain.mrs", URL: baseURL + "/geosite/category-ai-!cn.mrs", Path: filepath.Join(s.rulesetDir, "ai-domain.mrs"), Description: "AI 服务域名"},
		// Google
		{Name: "google-domain.mrs", URL: baseURL + "/geosite/google.mrs", Path: filepath.Join(s.rulesetDir, "google-domain.mrs"), Description: "Google 域名"},
		{Name: "google-ip.mrs", URL: baseURL + "/geoip/google.mrs", Path: filepath.Join(s.rulesetDir, "google-ip.mrs"), Description: "Google IP"},
		// YouTube
		{Name: "youtube-domain.mrs", URL: baseURL + "/geosite/youtube.mrs", Path: filepath.Join(s.rulesetDir, "youtube-domain.mrs"), Description: "YouTube 域名"},
		// Telegram
		{Name: "telegram-domain.mrs", URL: baseURL + "/geosite/telegram.mrs", Path: filepath.Join(s.rulesetDir, "telegram-domain.mrs"), Description: "Telegram 域名"},
		{Name: "telegram-ip.mrs", URL: baseURL + "/geoip/telegram.mrs", Path: filepath.Join(s.rulesetDir, "telegram-ip.mrs"), Description: "Telegram IP"},
		// Twitter
		{Name: "twitter-domain.mrs", URL: baseURL + "/geosite/twitter.mrs", Path: filepath.Join(s.rulesetDir, "twitter-domain.mrs"), Description: "Twitter 域名"},
		{Name: "twitter-ip.mrs", URL: baseURL + "/geoip/twitter.mrs", Path: filepath.Join(s.rulesetDir, "twitter-ip.mrs"), Description: "Twitter IP"},
		// Facebook
		{Name: "facebook-domain.mrs", URL: baseURL + "/geosite/facebook.mrs", Path: filepath.Join(s.rulesetDir, "facebook-domain.mrs"), Description: "Facebook 域名"},
		{Name: "facebook-ip.mrs", URL: baseURL + "/geoip/facebook.mrs", Path: filepath.Join(s.rulesetDir, "facebook-ip.mrs"), Description: "Facebook IP"},
		// GitHub
		{Name: "github-domain.mrs", URL: baseURL + "/geosite/github.mrs", Path: filepath.Join(s.rulesetDir, "github-domain.mrs"), Description: "GitHub 域名"},
		// Microsoft
		{Name: "microsoft-domain.mrs", URL: baseURL + "/geosite/microsoft.mrs", Path: filepath.Join(s.rulesetDir, "microsoft-domain.mrs"), Description: "Microsoft 域名"},
		// Apple
		{Name: "apple-domain.mrs", URL: baseURL + "/geosite/apple.mrs", Path: filepath.Join(s.rulesetDir, "apple-domain.mrs"), Description: "Apple 域名"},
		{Name: "apple-cn-domain.mrs", URL: baseURL + "/geosite/apple-cn.mrs", Path: filepath.Join(s.rulesetDir, "apple-cn-domain.mrs"), Description: "Apple 中国域名"},
		// 游戏平台
		{Name: "steam-domain.mrs", URL: baseURL + "/geosite/steam.mrs", Path: filepath.Join(s.rulesetDir, "steam-domain.mrs"), Description: "Steam 域名"},
		{Name: "epic-domain.mrs", URL: baseURL + "/geosite/epicgames.mrs", Path: filepath.Join(s.rulesetDir, "epic-domain.mrs"), Description: "Epic Games 域名"},
		// 流媒体
		{Name: "netflix-domain.mrs", URL: baseURL + "/geosite/netflix.mrs", Path: filepath.Join(s.rulesetDir, "netflix-domain.mrs"), Description: "Netflix 域名"},
		{Name: "netflix-ip.mrs", URL: baseURL + "/geoip/netflix.mrs", Path: filepath.Join(s.rulesetDir, "netflix-ip.mrs"), Description: "Netflix IP"},
		{Name: "spotify-domain.mrs", URL: baseURL + "/geosite/spotify.mrs", Path: filepath.Join(s.rulesetDir, "spotify-domain.mrs"), Description: "Spotify 域名"},
		{Name: "tiktok-domain.mrs", URL: baseURL + "/geosite/tiktok.mrs", Path: filepath.Join(s.rulesetDir, "tiktok-domain.mrs"), Description: "TikTok 域名"},
		// 哔哩哔哩
		{Name: "bilibili-domain.mrs", URL: baseURL + "/geosite/bilibili.mrs", Path: filepath.Join(s.rulesetDir, "bilibili-domain.mrs"), Description: "哔哩哔哩域名"},
		// 广告拦截
		{Name: "ads-domain.mrs", URL: baseURL + "/geosite/category-ads-all.mrs", Path: filepath.Join(s.rulesetDir, "ads-domain.mrs"), Description: "广告域名"},
		// 中国大陆
		{Name: "cn-domain.mrs", URL: baseURL + "/geosite/cn.mrs", Path: filepath.Join(s.rulesetDir, "cn-domain.mrs"), Description: "中国大陆域名"},
		{Name: "cn-ip.mrs", URL: baseURL + "/geoip/cn.mrs", Path: filepath.Join(s.rulesetDir, "cn-ip.mrs"), Description: "中国大陆 IP"},
		// 非中国大陆
		{Name: "geolocation-!cn-domain.mrs", URL: baseURL + "/geosite/geolocation-!cn.mrs", Path: filepath.Join(s.rulesetDir, "geolocation-!cn-domain.mrs"), Description: "非中国大陆域名"},
	}

	// 更新文件状态
	for i := range files {
		files[i] = s.updateFileStatus(files[i])
	}

	return files
}

// updateFileStatus 更新文件状态
func (s *Service) updateFileStatus(f RuleFile) RuleFile {
	info, err := os.Stat(f.Path)
	if err == nil {
		f.Size = info.Size()
		f.UpdatedAt = info.ModTime().Format("2006-01-02 15:04:05")
		f.Status = "completed"
	} else {
		f.Status = "pending"
	}
	return f
}

// GetConfig 获取配置
func (s *Service) GetConfig() *RuleSetConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig 设置配置
func (s *Service) SetConfig(config *RuleSetConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return s.saveConfig()
}

// loadConfig 加载配置
func (s *Service) loadConfig() {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, s.config)
}

// saveConfig 保存配置
func (s *Service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

// DownloadFile 下载单个文件
func (s *Service) DownloadFile(url, path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)

	// 创建临时文件
	tmpPath := path + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	// 下载文件
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpPath)
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 重命名临时文件
	out.Close()
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// DownloadAllFiles 下载所有规则文件
func (s *Service) DownloadAllFiles() (int, int, []string) {
	s.mu.Lock()
	if s.updating {
		s.mu.Unlock()
		return 0, 0, []string{"正在更新中，请稍后再试"}
	}
	s.updating = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.updating = false
		s.mu.Unlock()
	}()

	var success, failed int
	var errors []string

	// 下载 GEO 文件
	for _, f := range s.GetGeoFiles() {
		if err := s.DownloadFile(f.URL, f.Path); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", f.Name, err))
		} else {
			success++
		}
	}

	// 下载规则提供者文件
	for _, f := range s.GetRuleProviderFiles() {
		if err := s.DownloadFile(f.URL, f.Path); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", f.Name, err))
		} else {
			success++
		}
	}

	// 更新最后更新时间
	s.mu.Lock()
	s.config.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	s.saveConfig()
	s.mu.Unlock()

	return success, failed, errors
}

// IsUpdating 是否正在更新
func (s *Service) IsUpdating() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.updating
}

// autoUpdateLoop 自动更新循环
func (s *Service) autoUpdateLoop() {
	for {
		time.Sleep(1 * time.Hour) // 每小时检查一次

		s.mu.RLock()
		autoUpdate := s.config.AutoUpdate
		interval := s.config.UpdateInterval
		lastUpdate := s.config.LastUpdate
		s.mu.RUnlock()

		if !autoUpdate {
			continue
		}

		// 检查是否需要更新
		if lastUpdate != "" {
			lastTime, err := time.Parse("2006-01-02 15:04:05", lastUpdate)
			if err == nil {
				if time.Since(lastTime).Hours() < float64(interval*24) {
					continue
				}
			}
		}

		// 执行更新
		s.DownloadAllFiles()
	}
}

// GetRulesetDir 获取规则集目录
func (s *Service) GetRulesetDir() string {
	return s.rulesetDir
}

// applyGitHubProxy 应用 GitHub 代理到 URL
func applyGitHubProxy(url, proxy string) string {
	if proxy == "" || proxy == "__custom__" {
		return url
	}
	// 处理 GitHub 相关的 URL
	// https://raw.githubusercontent.com/xxx -> proxy/https://raw.githubusercontent.com/xxx
	// https://github.com/xxx -> proxy/https://github.com/xxx
	if strings.Contains(url, "github") || strings.Contains(url, "githubusercontent") {
		return strings.TrimSuffix(proxy, "/") + "/" + url
	}
	return url
}

// DownloadAllFilesWithProxy 使用代理下载所有规则文件
func (s *Service) DownloadAllFilesWithProxy(proxy string) (int, int, []string) {
	s.mu.Lock()
	if s.updating {
		s.mu.Unlock()
		return 0, 0, []string{"正在更新中，请稍后再试"}
	}
	s.updating = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.updating = false
		s.mu.Unlock()
	}()

	var success, failed int
	var errors []string

	// 收集所有下载任务
	type downloadTask struct {
		name string
		url  string
		path string
	}
	var tasks []downloadTask

	for _, f := range s.GetGeoFiles() {
		tasks = append(tasks, downloadTask{f.Name, f.URL, f.Path})
	}
	for _, f := range s.GetRuleProviderFiles() {
		tasks = append(tasks, downloadTask{f.Name, f.URL, f.Path})
	}

	// 使用 5 个并发线程下载
	const maxConcurrent = 5
	taskChan := make(chan downloadTask, len(tasks))
	resultChan := make(chan error, len(tasks))
	var wg sync.WaitGroup

	// 启动 worker
	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				url := applyGitHubProxy(task.url, proxy)
				err := s.DownloadFile(url, task.path)
				resultChan <- err
			}
		}()
	}

	// 发送任务
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 等待所有下载完成
	wg.Wait()
	close(resultChan)

	// 统计结果
	for err := range resultChan {
		if err != nil {
			failed++
			errors = append(errors, err.Error())
		} else {
			success++
		}
	}

	// 更新最后更新时间
	s.mu.Lock()
	s.config.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	s.saveConfig()
	s.mu.Unlock()

	return success, failed, errors
}
