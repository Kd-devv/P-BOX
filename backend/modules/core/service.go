package core

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type CoreType string

const (
	CoreTypeMihomo  CoreType = "mihomo"
	CoreTypeSingbox CoreType = "singbox"
)

// CDN 镜像地址
const (
	MihomoCDNBase  = "https://ghfast.top/https://github.com/MetaCubeX/mihomo/releases/download"
	SingboxCDNBase = "https://ghfast.top/https://github.com/SagerNet/sing-box/releases/download"
)

type CoreStatus struct {
	CurrentCore CoreType         `json:"currentCore"`
	Cores       map[string]*Core `json:"cores"`
}

type Core struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	LatestVersion string `json:"latestVersion"`
	Installed     bool   `json:"installed"`
	Path          string `json:"path"`
}

type DownloadProgress struct {
	Downloading bool    `json:"downloading"`
	Progress    float64 `json:"progress"`
	Speed       int64   `json:"speed"`
	Error       string  `json:"error,omitempty"`
}

type Service struct {
	dataDir          string
	currentCore      CoreType
	cores            map[string]*Core
	downloadProgress map[string]*DownloadProgress
	mu               sync.RWMutex
	onCoreSwitch     func(coreType string) // 核心切换回调
}

// 持久化状态
type SavedCoreStatus struct {
	CurrentCore    string            `json:"currentCore"`
	Versions       map[string]string `json:"versions"`
	LatestVersions map[string]string `json:"latestVersions"`
	LastChecked    time.Time         `json:"lastChecked"`
}

func NewService(dataDir string) *Service {
	s := &Service{
		dataDir:          dataDir,
		currentCore:      CoreTypeMihomo,
		cores:            make(map[string]*Core),
		downloadProgress: make(map[string]*DownloadProgress),
	}

	s.cores["mihomo"] = &Core{
		Name:      "Mihomo",
		Installed: false,
		Path:      filepath.Join(dataDir, "cores", "mihomo"),
	}
	s.cores["singbox"] = &Core{
		Name:      "sing-box",
		Installed: false,
		Path:      filepath.Join(dataDir, "cores", "sing-box"),
	}

	s.loadSavedStatus()
	s.checkInstalledCores()
	return s
}

func (s *Service) loadSavedStatus() {
	filePath := filepath.Join(s.dataDir, "core_status.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var saved SavedCoreStatus
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}

	if saved.CurrentCore != "" {
		s.currentCore = CoreType(saved.CurrentCore)
	}

	for name, version := range saved.Versions {
		if core, ok := s.cores[name]; ok {
			core.Version = version
			if version != "" {
				core.Installed = true
			}
		}
	}

	// 加载保存的最新版本信息
	for name, latestVersion := range saved.LatestVersions {
		if core, ok := s.cores[name]; ok {
			core.LatestVersion = latestVersion
		}
	}
}

func (s *Service) saveStatus() error {
	s.mu.RLock()
	saved := SavedCoreStatus{
		CurrentCore:    string(s.currentCore),
		Versions:       make(map[string]string),
		LatestVersions: make(map[string]string),
		LastChecked:    time.Now(),
	}
	for name, core := range s.cores {
		if core.Installed {
			saved.Versions[name] = core.Version
		}
		if core.LatestVersion != "" {
			saved.LatestVersions[name] = core.LatestVersion
		}
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.dataDir, "core_status.json")
	return os.WriteFile(filePath, data, 0644)
}

func (s *Service) checkInstalledCores() {
	for name, core := range s.cores {
		binPath := s.getCoreBinaryPath(name)
		if _, err := os.Stat(binPath); err == nil {
			core.Installed = true
			core.Version = s.getCoreVersion(name)
		}
	}
}

func (s *Service) getCoreBinaryPath(coreType string) string {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	var binName string
	switch coreType {
	case "mihomo":
		binName = fmt.Sprintf("mihomo-%s-%s", goos, arch)
	case "singbox":
		binName = fmt.Sprintf("sing-box-%s-%s", goos, arch)
	}

	return filepath.Join(s.dataDir, "cores", binName)
}

func (s *Service) getCoreVersion(coreType string) string {
	binPath := s.getCoreBinaryPath(coreType)

	// 执行核心获取版本
	var cmd *exec.Cmd
	switch coreType {
	case "mihomo":
		cmd = exec.Command(binPath, "-v")
	case "singbox":
		cmd = exec.Command(binPath, "version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		// 如果有保存的版本，使用保存的
		if core, ok := s.cores[coreType]; ok && core.Version != "" {
			return core.Version
		}
		return "unknown"
	}

	// 解析版本号
	outputStr := string(output)
	version := s.parseVersionFromOutput(coreType, outputStr)
	if version != "" {
		return version
	}

	return "unknown"
}

// parseVersionFromOutput 从输出中解析版本号
func (s *Service) parseVersionFromOutput(coreType, output string) string {
	lines := strings.Split(output, "\n")

	switch coreType {
	case "mihomo":
		// Mihomo v1.18.10 darwin arm64 with go1.23.2
		for _, line := range lines {
			if strings.Contains(line, "Mihomo") || strings.Contains(line, "mihomo") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "v") || strings.HasPrefix(part, "V") {
						return strings.TrimPrefix(strings.TrimPrefix(part, "v"), "V")
					}
				}
			}
		}
	case "singbox":
		// sing-box version 1.10.5
		for _, line := range lines {
			if strings.Contains(line, "version") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[len(parts)-1]
				}
			}
			// 或者直接输出版本号
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, " ") {
				return line
			}
		}
	}

	return ""
}

func (s *Service) GetStatus() *CoreStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &CoreStatus{
		CurrentCore: s.currentCore,
		Cores:       s.cores,
	}
}

func (s *Service) GetLatestVersions() (map[string]string, error) {
	versions := make(map[string]string)

	mihomoVersion, err := s.fetchMihomoLatestVersion()
	if err == nil {
		versions["mihomo"] = mihomoVersion
		s.mu.Lock()
		s.cores["mihomo"].LatestVersion = mihomoVersion
		s.mu.Unlock()
	}

	singboxVersion, err := s.fetchSingboxLatestVersion()
	if err == nil {
		versions["singbox"] = singboxVersion
		s.mu.Lock()
		s.cores["singbox"].LatestVersion = singboxVersion
		s.mu.Unlock()
	}

	return versions, nil
}

func (s *Service) fetchMihomoLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/MetaCubeX/mihomo/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// 去掉版本号前的 v 前缀
	version := release.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}
	return version, nil
}

func (s *Service) fetchSingboxLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/SagerNet/sing-box/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// 去掉版本号前的 v 前缀
	version := release.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}
	return version, nil
}

func (s *Service) SwitchCore(coreType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	core, ok := s.cores[coreType]
	if !ok {
		return fmt.Errorf("unknown core type: %s", coreType)
	}

	if !core.Installed {
		return fmt.Errorf("core %s is not installed", coreType)
	}

	s.currentCore = CoreType(coreType)

	// 通知 proxy 模块切换核心
	if s.onCoreSwitch != nil {
		s.onCoreSwitch(coreType)
	}

	// 持久化保存
	go s.saveStatus()

	return nil
}

// SetOnCoreSwitch 设置核心切换回调
func (s *Service) SetOnCoreSwitch(callback func(coreType string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onCoreSwitch = callback
}

// GetCurrentCore 获取当前核心类型
func (s *Service) GetCurrentCore() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return string(s.currentCore)
}

func (s *Service) DownloadCore(coreType string) error {
	s.mu.Lock()
	s.downloadProgress[coreType] = &DownloadProgress{Downloading: true}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.downloadProgress[coreType].Downloading = false
		s.mu.Unlock()
	}()

	// 获取 CDN 和官方下载地址
	cdnURL, officialURL, err := s.getCoreDownloadURLs(coreType)
	if err != nil {
		s.mu.Lock()
		s.downloadProgress[coreType].Error = err.Error()
		s.mu.Unlock()
		return err
	}

	// 尝试 CDN 下载
	fmt.Printf("📦 尝试从 CDN 下载 %s: %s\n", coreType, cdnURL)
	err = s.downloadFromURL(coreType, cdnURL)
	if err != nil {
		fmt.Printf("⚠️ CDN 下载失败: %v，尝试官方地址...\n", err)
		// 回退到官方地址
		fmt.Printf("📦 尝试从官方下载 %s: %s\n", coreType, officialURL)
		err = s.downloadFromURL(coreType, officialURL)
		if err != nil {
			s.mu.Lock()
			s.downloadProgress[coreType].Error = err.Error()
			s.mu.Unlock()
			return fmt.Errorf("下载失败: %v", err)
		}
	}

	fmt.Printf("✅ %s 下载完成\n", coreType)
	return nil
}

// downloadFromURL 从指定 URL 下载核心
func (s *Service) downloadFromURL(coreType, downloadURL string) error {
	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	os.MkdirAll(filepath.Join(s.dataDir, "cores"), 0755)

	// 下载到临时文件
	tmpFile := filepath.Join(s.dataDir, "cores", "download.tmp")
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	totalSize := resp.ContentLength
	written := int64(0)

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			written += int64(n)

			s.mu.Lock()
			if totalSize > 0 {
				s.downloadProgress[coreType].Progress = float64(written) / float64(totalSize) * 100
			}
			s.mu.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			out.Close()
			os.Remove(tmpFile)
			return err
		}
	}
	out.Close()

	// 解压文件
	binPath := s.getCoreBinaryPath(coreType)
	if err := s.extractCore(tmpFile, binPath, coreType); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("解压失败: %v", err)
	}
	os.Remove(tmpFile)

	// 设置执行权限
	os.Chmod(binPath, 0755)

	s.mu.Lock()
	s.cores[coreType].Installed = true
	s.cores[coreType].Version = s.cores[coreType].LatestVersion
	s.mu.Unlock()

	// 持久化保存
	s.saveStatus()

	return nil
}

// extractCore 解压核心文件
func (s *Service) extractCore(archivePath, destPath, coreType string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建 gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("gzip open failed: %v", err)
	}
	defer gzr.Close()

	// Mihomo 是单文件 .gz，sing-box 是 .tar.gz
	if coreType == "mihomo" {
		// 直接解压 gzip
		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, gzr)
		return err
	}

	// sing-box: tar.gz 格式
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 查找可执行文件
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "sing-box") {
			outFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tr)
			return err
		}
	}

	return fmt.Errorf("executable not found in archive")
}

// getCoreDownloadURLs 获取下载 URL（CDN 优先，官方备用）
func (s *Service) getCoreDownloadURLs(coreType string) (cdnURL, officialURL string, err error) {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	s.mu.RLock()
	version := s.cores[coreType].LatestVersion
	s.mu.RUnlock()

	if version == "" {
		return "", "", fmt.Errorf("version not found, please check latest version first")
	}

	// 转换架构名称
	archName := arch
	if arch == "amd64" {
		archName = "amd64"
	} else if arch == "arm64" {
		archName = "arm64"
	}

	// 转换系统名称
	osName := goos
	if goos == "darwin" {
		osName = "darwin"
	}

	switch coreType {
	case "mihomo":
		// mihomo releases 格式: mihomo-darwin-arm64-v1.18.10.gz
		filename := fmt.Sprintf("mihomo-%s-%s-v%s.gz", osName, archName, version)
		cdnURL = fmt.Sprintf("%s/v%s/%s", MihomoCDNBase, version, filename)
		officialURL = fmt.Sprintf("https://github.com/MetaCubeX/mihomo/releases/download/v%s/%s", version, filename)
		return cdnURL, officialURL, nil

	case "singbox":
		// sing-box releases 格式: sing-box-1.10.5-darwin-arm64.tar.gz
		filename := fmt.Sprintf("sing-box-%s-%s-%s.tar.gz", version, osName, archName)
		cdnURL = fmt.Sprintf("%s/v%s/%s", SingboxCDNBase, version, filename)
		officialURL = fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/v%s/%s", version, filename)
		return cdnURL, officialURL, nil
	}

	return "", "", fmt.Errorf("unknown core type")
}

func (s *Service) GetDownloadProgress(coreType string) *DownloadProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if progress, ok := s.downloadProgress[coreType]; ok {
		return progress
	}
	return &DownloadProgress{}
}

// Initialize 启动时自动初始化（延迟执行）
// delaySeconds: 启动后延迟多少秒执行检测
func (s *Service) Initialize(delaySeconds int) {
	go func() {
		// 延迟执行
		time.Sleep(time.Duration(delaySeconds) * time.Second)
		fmt.Printf("🔍 开始自动检测核心版本...\n")

		// 1. 检测最新版本
		s.GetLatestVersions()

		// 2. 检查是否需要自动下载 mihomo 核心
		s.mu.RLock()
		mihomoInstalled := s.cores["mihomo"].Installed
		mihomoLatestVersion := s.cores["mihomo"].LatestVersion
		s.mu.RUnlock()

		if !mihomoInstalled && mihomoLatestVersion != "" {
			fmt.Printf("📦 检测到未安装 mihomo 核心，开始自动下载...\n")
			fmt.Printf("   平台: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			if err := s.DownloadCore("mihomo"); err != nil {
				fmt.Printf("❌ 自动下载 mihomo 失败: %v\n", err)
			} else {
				fmt.Printf("✅ mihomo 核心自动下载完成\n")
			}
		}

		// 保存状态
		s.saveStatus()
		fmt.Printf("✅ 核心版本检测完成\n")
	}()
}

// RefreshVersions 手动刷新版本信息（前端点击刷新时调用）
func (s *Service) RefreshVersions() (map[string]string, error) {
	fmt.Printf("🔄 手动刷新核心版本信息...\n")

	versions, err := s.GetLatestVersions()
	if err != nil {
		return nil, err
	}

	// 保存到文件
	s.saveStatus()

	fmt.Printf("✅ 版本信息已更新并保存\n")
	return versions, nil
}

// GetPlatformInfo 获取当前平台信息
func (s *Service) GetPlatformInfo() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}
