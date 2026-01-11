package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 规则集存储目录（相对于数据目录）
const SingBoxRulesetSubDir = "singbox/ruleset"

// 实际规则集存储目录（运行时设置）
var singBoxRulesetDir string

// 官方规则仓库地址
const (
	// SagerNet 官方 GEO 数据库
	OfficialGeoIPURL   = "https://github.com/SagerNet/sing-geoip/releases/latest/download/geoip.db"
	OfficialGeoSiteURL = "https://github.com/SagerNet/sing-geosite/releases/latest/download/geosite.db"
	// SagerNet 官方 Rule-Set 基础 URL
	OfficialRuleSetBaseURL = "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set"
)

// 默认 GitHub 代理列表 (2025年可用)
var defaultGitHubProxies = []string{
	"",                     // 直连（无代理）
	"https://ghfast.top",   // ghfast - 稳定
	"https://ghproxy.link", // ghproxy.link
	"https://gh-proxy.com", // gh-proxy.com
	"https://ghps.cc",      // ghps.cc
}

// SetSingBoxRulesetDir 设置规则集存储目录
func SetSingBoxRulesetDir(dataDir string) {
	singBoxRulesetDir = filepath.Join(dataDir, SingBoxRulesetSubDir)
	rulesetConfigPath = filepath.Join(dataDir, "singbox-ruleset-config.json")
	loadRuleSetConfig()
}

// GetSingBoxRulesetDir 获取规则集存储目录
func GetSingBoxRulesetDir() string {
	return singBoxRulesetDir
}

// RuleSetFileInfo 规则集文件信息
type RuleSetFileInfo struct {
	Tag       string `json:"tag"`
	Type      string `json:"type"`   // local, remote
	Format    string `json:"format"` // binary, source
	URL       string `json:"url,omitempty"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updatedAt"`
	Status    string `json:"status"` // pending, downloading, completed, failed
	Exists    bool   `json:"exists"` // 本地文件是否存在
}

// GeoResourceInfo GEO 资源信息
type GeoResourceInfo struct {
	Name           string `json:"name"`
	Type           string `json:"type"` // geoip, geosite
	URL            string `json:"url"`
	Path           string `json:"path"`
	DownloadDetour string `json:"downloadDetour,omitempty"`
	UpdateInterval string `json:"updateInterval,omitempty"`
	Size           int64  `json:"size"`
	UpdatedAt      string `json:"updatedAt"`
	Status         string `json:"status"`
	Exists         bool   `json:"exists"`
}

// RuleSetConfig 规则集配置
type RuleSetConfig struct {
	AutoUpdate     bool     `json:"autoUpdate"`
	UpdateInterval int      `json:"updateInterval"` // 天
	LastUpdate     string   `json:"lastUpdate"`
	GitHubProxy    string   `json:"githubProxy"`   // 当前使用的 GitHub 代理
	GitHubProxies  []string `json:"githubProxies"` // 可用的 GitHub 代理列表
	CustomProxies  []string `json:"customProxies"` // 用户自定义代理
	RulesetDir     string   `json:"rulesetDir"`    // 规则集存储目录
}

// 配置文件路径
var rulesetConfigPath string

// 当前配置
var currentConfig *RuleSetConfig

// 下载状态跟踪
var (
	downloadStatus     = make(map[string]string) // tag -> status
	downloadStatusLock sync.RWMutex
)

// InitSingBoxRulesetDir 初始化规则集目录
func InitSingBoxRulesetDir() error {
	if err := os.MkdirAll(singBoxRulesetDir, 0755); err != nil {
		return fmt.Errorf("创建规则集目录失败: %w", err)
	}
	return nil
}

// RegisterSingBoxRulesetRoutes 注册 Sing-Box 规则集路由
func RegisterSingBoxRulesetRoutes(r *gin.RouterGroup) {
	// 初始化目录
	if err := InitSingBoxRulesetDir(); err != nil {
		fmt.Printf("警告: %v\n", err)
	}

	singbox := r.Group("/singbox/ruleset")
	{
		singbox.GET("/geo", handleGetGeoResources)
		singbox.GET("/rules", handleGetRuleSets)
		singbox.GET("/config", handleGetRuleSetConfig)
		singbox.POST("/config", handleSaveRuleSetConfig)
		singbox.GET("/status", handleGetDownloadStatus)
		singbox.POST("/update", handleUpdateRuleSets)
		singbox.POST("/download", handleDownloadRuleSet)
	}
}

// handleGetGeoResources 获取 GEO 资源列表
func handleGetGeoResources(c *gin.Context) {
	geoResources := []GeoResourceInfo{
		{
			Name:           "geoip.db",
			Type:           "geoip",
			URL:            "https://github.com/SagerNet/sing-geoip/releases/latest/download/geoip.db",
			Path:           filepath.Join(singBoxRulesetDir, "geoip.db"),
			DownloadDetour: "direct",
			UpdateInterval: "7d",
		},
		{
			Name:           "geosite.db",
			Type:           "geosite",
			URL:            "https://github.com/SagerNet/sing-geosite/releases/latest/download/geosite.db",
			Path:           filepath.Join(singBoxRulesetDir, "geosite.db"),
			DownloadDetour: "direct",
			UpdateInterval: "7d",
		},
	}

	// 检查文件状态
	for i := range geoResources {
		checkFileStatus(&geoResources[i].Path, &geoResources[i].Size, &geoResources[i].UpdatedAt, &geoResources[i].Status, &geoResources[i].Exists)
	}

	c.JSON(http.StatusOK, gin.H{"data": geoResources})
}

// handleGetRuleSets 获取规则集列表
func handleGetRuleSets(c *gin.Context) {
	// 使用 SagerNet 官方 Rule-Set 仓库
	// https://github.com/SagerNet/sing-geosite (rule-set 分支)
	// https://github.com/SagerNet/sing-geoip (rule-set 分支)
	baseURL := OfficialRuleSetBaseURL // https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set
	geoipBaseURL := "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set"

	defaultRuleSets := []RuleSetFileInfo{
		// 广告拦截
		{Tag: "geosite-category-ads-all", Type: "local", Format: "binary", URL: baseURL + "/geosite-category-ads-all.srs"},
		// AI 服务
		{Tag: "geosite-openai", Type: "local", Format: "binary", URL: baseURL + "/geosite-openai.srs"},
		{Tag: "geosite-anthropic", Type: "local", Format: "binary", URL: baseURL + "/geosite-anthropic.srs"},
		{Tag: "geosite-google-gemini", Type: "local", Format: "binary", URL: baseURL + "/geosite-google-gemini.srs"},
		{Tag: "geosite-cursor", Type: "local", Format: "binary", URL: baseURL + "/geosite-cursor.srs"},
		{Tag: "geosite-category-ai-!cn", Type: "local", Format: "binary", URL: baseURL + "/geosite-category-ai-!cn.srs"},
		// 游戏平台
		{Tag: "geosite-steam", Type: "local", Format: "binary", URL: baseURL + "/geosite-steam.srs"},
		{Tag: "geosite-epicgames", Type: "local", Format: "binary", URL: baseURL + "/geosite-epicgames.srs"},
		// 流媒体
		{Tag: "geosite-netflix", Type: "local", Format: "binary", URL: baseURL + "/geosite-netflix.srs"},
		{Tag: "geosite-disney", Type: "local", Format: "binary", URL: baseURL + "/geosite-disney.srs"},
		{Tag: "geosite-youtube", Type: "local", Format: "binary", URL: baseURL + "/geosite-youtube.srs"},
		{Tag: "geosite-spotify", Type: "local", Format: "binary", URL: baseURL + "/geosite-spotify.srs"},
		// 社交媒体
		{Tag: "geosite-twitter", Type: "local", Format: "binary", URL: baseURL + "/geosite-twitter.srs"},
		{Tag: "geosite-facebook", Type: "local", Format: "binary", URL: baseURL + "/geosite-facebook.srs"},
		{Tag: "geosite-instagram", Type: "local", Format: "binary", URL: baseURL + "/geosite-instagram.srs"},
		// 海外聊天
		{Tag: "geosite-telegram", Type: "local", Format: "binary", URL: baseURL + "/geosite-telegram.srs"},
		{Tag: "geosite-whatsapp", Type: "local", Format: "binary", URL: baseURL + "/geosite-whatsapp.srs"},
		{Tag: "geosite-discord", Type: "local", Format: "binary", URL: baseURL + "/geosite-discord.srs"},
		// Google
		{Tag: "geosite-google", Type: "local", Format: "binary", URL: baseURL + "/geosite-google.srs"},
		// 开发者
		{Tag: "geosite-github", Type: "local", Format: "binary", URL: baseURL + "/geosite-github.srs"},
		// Microsoft
		{Tag: "geosite-microsoft", Type: "local", Format: "binary", URL: baseURL + "/geosite-microsoft.srs"},
		// Apple
		{Tag: "geosite-apple", Type: "local", Format: "binary", URL: baseURL + "/geosite-apple.srs"},
		// 中国直连
		{Tag: "geosite-bilibili", Type: "local", Format: "binary", URL: baseURL + "/geosite-bilibili.srs"},
		{Tag: "geosite-iqiyi", Type: "local", Format: "binary", URL: baseURL + "/geosite-iqiyi.srs"},
		{Tag: "geosite-alibaba", Type: "local", Format: "binary", URL: baseURL + "/geosite-alibaba.srs"},
		{Tag: "geosite-cn", Type: "local", Format: "binary", URL: baseURL + "/geosite-cn.srs"},
		{Tag: "geoip-cn", Type: "local", Format: "binary", URL: geoipBaseURL + "/geoip-cn.srs"},
		// 其他海外
		{Tag: "geosite-geolocation-!cn", Type: "local", Format: "binary", URL: baseURL + "/geosite-geolocation-!cn.srs"},
	}

	// 设置路径并检查文件状态
	for i := range defaultRuleSets {
		defaultRuleSets[i].Path = filepath.Join(singBoxRulesetDir, defaultRuleSets[i].Tag+".srs")
		checkFileStatus(&defaultRuleSets[i].Path, &defaultRuleSets[i].Size, &defaultRuleSets[i].UpdatedAt, &defaultRuleSets[i].Status, &defaultRuleSets[i].Exists)
	}

	c.JSON(http.StatusOK, gin.H{"data": defaultRuleSets})
}

// checkFileStatus 检查文件状态
func checkFileStatus(path *string, size *int64, updatedAt *string, status *string, exists *bool) {
	info, err := os.Stat(*path)
	if err == nil {
		*exists = true
		*size = info.Size()
		*updatedAt = info.ModTime().Format(time.RFC3339)
		*status = "completed"
	} else {
		*exists = false
		*size = 0
		*updatedAt = ""
		*status = "pending"
	}

	// 检查是否正在下载
	downloadStatusLock.RLock()
	if s, ok := downloadStatus[*path]; ok {
		*status = s
	}
	downloadStatusLock.RUnlock()
}

// handleGetRuleSetConfig 获取规则集配置
func handleGetRuleSetConfig(c *gin.Context) {
	loadRuleSetConfig()
	currentConfig.RulesetDir = singBoxRulesetDir
	currentConfig.GitHubProxies = defaultGitHubProxies
	c.JSON(http.StatusOK, gin.H{"data": currentConfig})
}

// handleSaveRuleSetConfig 保存规则集配置
func handleSaveRuleSetConfig(c *gin.Context) {
	var config RuleSetConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存配置
	currentConfig.AutoUpdate = config.AutoUpdate
	currentConfig.UpdateInterval = config.UpdateInterval
	currentConfig.GitHubProxy = config.GitHubProxy
	currentConfig.CustomProxies = config.CustomProxies

	if err := saveRuleSetConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// loadRuleSetConfig 加载配置
func loadRuleSetConfig() {
	if currentConfig == nil {
		currentConfig = &RuleSetConfig{
			AutoUpdate:     true,
			UpdateInterval: 1,
			GitHubProxy:    "",
			GitHubProxies:  defaultGitHubProxies,
			CustomProxies:  []string{},
		}
	}

	if rulesetConfigPath == "" {
		return
	}

	data, err := os.ReadFile(rulesetConfigPath)
	if err != nil {
		return // 文件不存在，使用默认配置
	}

	var config RuleSetConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	currentConfig.AutoUpdate = config.AutoUpdate
	currentConfig.UpdateInterval = config.UpdateInterval
	currentConfig.GitHubProxy = config.GitHubProxy
	currentConfig.CustomProxies = config.CustomProxies
	currentConfig.LastUpdate = config.LastUpdate
}

// saveRuleSetConfig 保存配置
func saveRuleSetConfig() error {
	if rulesetConfigPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(rulesetConfigPath, data, 0644)
}

// handleGetDownloadStatus 获取下载状态
func handleGetDownloadStatus(c *gin.Context) {
	downloadStatusLock.RLock()
	defer downloadStatusLock.RUnlock()
	c.JSON(http.StatusOK, gin.H{"data": downloadStatus})
}

// handleUpdateRuleSets 更新所有规则集
func handleUpdateRuleSets(c *gin.Context) {
	var req struct {
		GitHubProxy string `json:"githubProxy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保目录存在
	if err := InitSingBoxRulesetDir(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取所有规则集和 GEO 资源
	baseURL := OfficialRuleSetBaseURL
	geoipBaseURL := "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set"

	// GEO 资源
	geoResources := []struct{ url, path string }{
		{OfficialGeoIPURL, filepath.Join(singBoxRulesetDir, "geoip.db")},
		{OfficialGeoSiteURL, filepath.Join(singBoxRulesetDir, "geosite.db")},
	}

	// 规则集
	ruleSets := []struct{ tag, url string }{
		{"geosite-category-ads-all", baseURL + "/geosite-category-ads-all.srs"},
		{"geosite-openai", baseURL + "/geosite-openai.srs"},
		{"geosite-anthropic", baseURL + "/geosite-anthropic.srs"},
		{"geosite-google-gemini", baseURL + "/geosite-google-gemini.srs"},
		{"geosite-cursor", baseURL + "/geosite-cursor.srs"},
		{"geosite-category-ai-!cn", baseURL + "/geosite-category-ai-!cn.srs"},
		{"geosite-steam", baseURL + "/geosite-steam.srs"},
		{"geosite-epicgames", baseURL + "/geosite-epicgames.srs"},
		{"geosite-netflix", baseURL + "/geosite-netflix.srs"},
		{"geosite-disney", baseURL + "/geosite-disney.srs"},
		{"geosite-youtube", baseURL + "/geosite-youtube.srs"},
		{"geosite-spotify", baseURL + "/geosite-spotify.srs"},
		{"geosite-twitter", baseURL + "/geosite-twitter.srs"},
		{"geosite-facebook", baseURL + "/geosite-facebook.srs"},
		{"geosite-instagram", baseURL + "/geosite-instagram.srs"},
		{"geosite-telegram", baseURL + "/geosite-telegram.srs"},
		{"geosite-whatsapp", baseURL + "/geosite-whatsapp.srs"},
		{"geosite-discord", baseURL + "/geosite-discord.srs"},
		{"geosite-google", baseURL + "/geosite-google.srs"},
		{"geosite-github", baseURL + "/geosite-github.srs"},
		{"geosite-microsoft", baseURL + "/geosite-microsoft.srs"},
		{"geosite-apple", baseURL + "/geosite-apple.srs"},
		{"geosite-bilibili", baseURL + "/geosite-bilibili.srs"},
		{"geosite-iqiyi", baseURL + "/geosite-iqiyi.srs"},
		{"geosite-alibaba", baseURL + "/geosite-alibaba.srs"},
		{"geosite-cn", baseURL + "/geosite-cn.srs"},
		{"geoip-cn", geoipBaseURL + "/geoip-cn.srs"},
		{"geosite-geolocation-!cn", baseURL + "/geosite-geolocation-!cn.srs"},
	}

	// 构建下载任务列表
	type downloadTask struct {
		url  string
		path string
	}
	var tasks []downloadTask

	// GEO 资源
	for _, geo := range geoResources {
		tasks = append(tasks, downloadTask{
			url:  applyGitHubProxy(geo.url, req.GitHubProxy),
			path: geo.path,
		})
	}
	// 规则集
	for _, rs := range ruleSets {
		tasks = append(tasks, downloadTask{
			url:  applyGitHubProxy(rs.url, req.GitHubProxy),
			path: filepath.Join(singBoxRulesetDir, rs.tag+".srs"),
		})
	}

	// 使用 5 个并发线程下载
	const maxConcurrent = 5
	go func() {
		taskChan := make(chan downloadTask, len(tasks))
		var wg sync.WaitGroup

		// 启动 worker
		for i := 0; i < maxConcurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range taskChan {
					downloadFile(task.url, task.path)
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
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "开始下载规则集 (5线程并发)", "total": len(tasks)})
}

// handleDownloadRuleSet 下载单个规则集
func handleDownloadRuleSet(c *gin.Context) {
	var req struct {
		Tag         string `json:"tag"`
		IsGeo       bool   `json:"isGeo"`
		GitHubProxy string `json:"githubProxy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保目录存在
	if err := InitSingBoxRulesetDir(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var url, path string
	if req.IsGeo {
		// GEO 资源
		if req.Tag == "geoip.db" {
			url = OfficialGeoIPURL
			path = filepath.Join(singBoxRulesetDir, "geoip.db")
		} else if req.Tag == "geosite.db" {
			url = OfficialGeoSiteURL
			path = filepath.Join(singBoxRulesetDir, "geosite.db")
		}
	} else {
		// 规则集
		baseURL := OfficialRuleSetBaseURL
		if req.Tag == "geoip-cn" {
			url = "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-cn.srs"
		} else {
			url = baseURL + "/" + req.Tag + ".srs"
		}
		path = filepath.Join(singBoxRulesetDir, req.Tag+".srs")
	}

	if url == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则集"})
		return
	}

	url = applyGitHubProxy(url, req.GitHubProxy)
	go downloadFile(url, path)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "开始下载"})
}

// applyGitHubProxy 应用 GitHub 代理
func applyGitHubProxy(url, proxy string) string {
	if proxy == "" {
		return url
	}
	// 代理格式: https://proxy.com/https://github.com/...
	return proxy + "/" + url
}

// downloadFile 下载文件
func downloadFile(url, path string) error {
	// 更新状态为下载中
	downloadStatusLock.Lock()
	downloadStatus[path] = "downloading"
	downloadStatusLock.Unlock()

	defer func() {
		// 延迟清理状态
		time.Sleep(2 * time.Second)
		downloadStatusLock.Lock()
		delete(downloadStatus, path)
		downloadStatusLock.Unlock()
	}()

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建 HTTP 客户端
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpPath := path + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("创建文件失败: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpPath)
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 重命名为最终文件
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		downloadStatusLock.Lock()
		downloadStatus[path] = "failed"
		downloadStatusLock.Unlock()
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	downloadStatusLock.Lock()
	downloadStatus[path] = "completed"
	downloadStatusLock.Unlock()

	return nil
}

// CheckRuleSetExists 检查规则集文件是否存在
func CheckRuleSetExists(tag string) bool {
	path := filepath.Join(singBoxRulesetDir, tag+".srs")
	_, err := os.Stat(path)
	return err == nil
}

// GetRuleSetPath 获取规则集路径（如果本地存在则返回本地路径）
func GetRuleSetPath(tag, remoteURL string) (path string, isLocal bool) {
	localPath := filepath.Join(singBoxRulesetDir, tag+".srs")
	if _, err := os.Stat(localPath); err == nil {
		return localPath, true
	}
	return remoteURL, false
}
