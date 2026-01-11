package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemConfig 系统配置
type SystemConfig struct {
	AutoStart    bool `json:"autoStart"`    // 开机自启
	IPForward    bool `json:"ipForward"`    // IP 转发
	BBREnabled   bool `json:"bbrEnabled"`   // BBR 拥塞控制
	TUNOptimized bool `json:"tunOptimized"` // TUN 优化
}

// SystemResources 系统资源信息
type SystemResources struct {
	OS            string  `json:"os"`            // 操作系统
	Platform      string  `json:"platform"`      // 发行版
	Kernel        string  `json:"kernel"`        // 内核版本
	Arch          string  `json:"arch"`          // 架构
	CPUModel      string  `json:"cpuModel"`      // CPU 型号
	CPUCores      int     `json:"cpuCores"`      // CPU 核心数
	CPUUsage      float64 `json:"cpuUsage"`      // CPU 使用率
	MemoryTotal   uint64  `json:"memoryTotal"`   // 总内存 (bytes)
	MemoryUsed    uint64  `json:"memoryUsed"`    // 已用内存 (bytes)
	MemoryPercent float64 `json:"memoryPercent"` // 内存使用率
	DiskTotal     uint64  `json:"diskTotal"`     // 总磁盘 (bytes)
	DiskUsed      uint64  `json:"diskUsed"`      // 已用磁盘 (bytes)
	DiskPercent   float64 `json:"diskPercent"`   // 磁盘使用率
	Uptime        int64   `json:"uptime"`        // 运行时间 (秒)
}

// parseVmStatValue 解析 vm_stat 输出行的数值
// 格式: "Pages active:                            675445."
func parseVmStatValue(line string) uint64 {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return 0
	}
	valStr := strings.TrimSpace(parts[1])
	valStr = strings.TrimSuffix(valStr, ".")
	val, _ := strconv.ParseUint(valStr, 10, 64)
	return val
}

// Service 系统服务
type Service struct {
	dataDir    string
	binaryPath string
}

// NewService 创建系统服务
func NewService(dataDir string) *Service {
	// 获取当前执行文件路径
	execPath, _ := os.Executable()
	return &Service{
		dataDir:    dataDir,
		binaryPath: execPath,
	}
}

// GetConfig 获取系统配置
func (s *Service) GetConfig() *SystemConfig {
	return &SystemConfig{
		AutoStart:    s.IsAutoStartEnabled(),
		IPForward:    s.IsIPForwardEnabled(),
		BBREnabled:   s.IsBBREnabled(),
		TUNOptimized: s.IsTUNOptimized(),
	}
}

// GetResources 获取系统资源信息 (支持 Linux, macOS, Windows)
func (s *Service) GetResources() *SystemResources {
	res := &SystemResources{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCores: runtime.NumCPU(),
	}

	switch runtime.GOOS {
	case "linux":
		s.getLinuxResources(res)
	case "darwin":
		s.getMacOSResources(res)
	case "windows":
		s.getWindowsResources(res)
	}

	return res
}

// getLinuxResources 获取 Linux 系统信息
func (s *Service) getLinuxResources(res *SystemResources) {
	// 内核版本
	if data, err := os.ReadFile("/proc/version"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			res.Kernel = parts[2]
		}
	} else if out, err := exec.Command("uname", "-r").Output(); err == nil {
		res.Kernel = strings.TrimSpace(string(out))
	}

	// 发行版信息
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				res.Platform = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}
	if res.Platform == "" {
		res.Platform = "Linux"
	}

	// CPU 型号
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "Model") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					res.CPUModel = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}
	if res.CPUModel == "" {
		if out, err := exec.Command("lscpu").Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "Model name:") {
					res.CPUModel = strings.TrimSpace(strings.TrimPrefix(line, "Model name:"))
					break
				}
			}
		}
	}
	if res.CPUModel == "" {
		res.CPUModel = runtime.GOARCH
	}

	// 内存信息
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		var total, available uint64
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					val, _ := strconv.ParseUint(fields[1], 10, 64)
					total = val * 1024
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					val, _ := strconv.ParseUint(fields[1], 10, 64)
					available = val * 1024
				}
			}
		}
		res.MemoryTotal = total
		res.MemoryUsed = total - available
		if total > 0 {
			res.MemoryPercent = float64(res.MemoryUsed) / float64(total) * 100
		}
	}

	// 磁盘信息
	diskInfo := GetDiskInfo("/")
	res.DiskTotal = diskInfo.Total
	res.DiskUsed = diskInfo.Used
	res.DiskPercent = diskInfo.Percent

	// 运行时间
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			uptime, _ := strconv.ParseFloat(fields[0], 64)
			res.Uptime = int64(uptime)
		}
	}
}

// getMacOSResources 获取 macOS 系统信息
func (s *Service) getMacOSResources(res *SystemResources) {
	res.Platform = "macOS"

	// 获取 macOS 版本
	if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		res.Platform = "macOS " + strings.TrimSpace(string(out))
	}

	// 内核版本
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		res.Kernel = strings.TrimSpace(string(out))
	}

	// CPU 型号 - 智能获取，支持 Intel 和 Apple Silicon (M1/M2/M3/M4/M5+)
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
		res.CPUModel = strings.TrimSpace(string(out))
	} else {
		// 尝试从 system_profiler 获取芯片名称 (适用于所有 Apple Silicon)
		if out, err := exec.Command("system_profiler", "SPHardwareDataType").Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Chip:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						res.CPUModel = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
		// 如果还是没获取到，使用架构信息
		if res.CPUModel == "" {
			res.CPUModel = "Apple Silicon (" + res.Arch + ")"
		}
	}

	// 内存信息
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if total, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
			res.MemoryTotal = total
			// 获取页面大小
			pageSize := uint64(16384) // macOS 默认
			if psOut, err := exec.Command("sysctl", "-n", "hw.pagesize").Output(); err == nil {
				if ps, err := strconv.ParseUint(strings.TrimSpace(string(psOut)), 10, 64); err == nil {
					pageSize = ps
				}
			}
			// 获取已用内存 (通过 vm_stat)
			if vmOut, err := exec.Command("vm_stat").Output(); err == nil {
				var pagesActive, pagesWired, pagesCompressed, pagesInactive uint64
				scanner := bufio.NewScanner(strings.NewReader(string(vmOut)))
				for scanner.Scan() {
					line := scanner.Text()
					// 解析格式: "Pages active:                            675445."
					if strings.HasPrefix(line, "Pages active:") {
						pagesActive = parseVmStatValue(line)
					} else if strings.HasPrefix(line, "Pages wired down:") {
						pagesWired = parseVmStatValue(line)
					} else if strings.HasPrefix(line, "Pages occupied by compressor:") {
						pagesCompressed = parseVmStatValue(line)
					} else if strings.HasPrefix(line, "Pages inactive:") {
						pagesInactive = parseVmStatValue(line)
					}
				}
				// 已用内存 = active + wired + compressed (不包括 inactive)
				res.MemoryUsed = (pagesActive + pagesWired + pagesCompressed) * pageSize
				if res.MemoryTotal > 0 {
					res.MemoryPercent = float64(res.MemoryUsed) / float64(res.MemoryTotal) * 100
				}
				_ = pagesInactive // 避免未使用警告
			}
		}
	}

	// 磁盘信息
	diskInfo := GetDiskInfo("/")
	res.DiskTotal = diskInfo.Total
	res.DiskUsed = diskInfo.Used
	res.DiskPercent = diskInfo.Percent

	// 运行时间
	if out, err := exec.Command("sysctl", "-n", "kern.boottime").Output(); err == nil {
		// 格式: { sec = 1234567890, usec = 0 }
		s := string(out)
		if idx := strings.Index(s, "sec = "); idx != -1 {
			s = s[idx+6:]
			if idx = strings.Index(s, ","); idx != -1 {
				if bootTime, err := strconv.ParseInt(s[:idx], 10, 64); err == nil {
					res.Uptime = int64(time.Now().Unix() - bootTime)
				}
			}
		}
	}
}

// getWindowsResources 获取 Windows 系统信息
func (s *Service) getWindowsResources(res *SystemResources) {
	res.Platform = "Windows"

	// 获取 Windows 版本
	if out, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
		res.Platform = strings.TrimSpace(string(out))
	}

	// CPU 型号
	if out, err := exec.Command("wmic", "cpu", "get", "name").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "Name" {
				res.CPUModel = line
				break
			}
		}
	}

	// 内存信息
	if out, err := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "TotalPhysicalMemory" {
				if total, err := strconv.ParseUint(line, 10, 64); err == nil {
					res.MemoryTotal = total
				}
				break
			}
		}
	}
	if out, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "FreePhysicalMemory" {
				if free, err := strconv.ParseUint(line, 10, 64); err == nil {
					freeBytes := free * 1024
					if res.MemoryTotal > freeBytes {
						res.MemoryUsed = res.MemoryTotal - freeBytes
					}
				}
				break
			}
		}
	}
	if res.MemoryTotal > 0 {
		res.MemoryPercent = float64(res.MemoryUsed) / float64(res.MemoryTotal) * 100
	}

	// 磁盘信息 (C: 盘)
	if out, err := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'", "get", "Size,FreeSpace").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				free, _ := strconv.ParseUint(fields[0], 10, 64)
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				if total > 0 {
					res.DiskTotal = total
					res.DiskUsed = total - free
					res.DiskPercent = float64(res.DiskUsed) / float64(total) * 100
				}
			}
		}
	}

	// 运行时间
	if out, err := exec.Command("wmic", "os", "get", "LastBootUpTime").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "LastBootUpTime" && len(line) >= 14 {
				// 格式: 20231215123456.000000+480
				year, _ := strconv.Atoi(line[0:4])
				month, _ := strconv.Atoi(line[4:6])
				day, _ := strconv.Atoi(line[6:8])
				hour, _ := strconv.Atoi(line[8:10])
				min, _ := strconv.Atoi(line[10:12])
				sec, _ := strconv.Atoi(line[12:14])
				bootTime := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
				res.Uptime = int64(time.Since(bootTime).Seconds())
				break
			}
		}
	}
}

// SetAutoStart 设置开机自启
func (s *Service) SetAutoStart(enabled bool) error {
	if enabled {
		return s.enableAutoStart()
	}
	return s.disableAutoStart()
}

// IsAutoStartEnabled 检查是否已开启自启
func (s *Service) IsAutoStartEnabled() bool {
	// 检查 systemd 服务是否存在且启用
	cmd := exec.Command("systemctl", "is-enabled", "p-box")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) == "enabled"
}

// enableAutoStart 启用开机自启
func (s *Service) enableAutoStart() error {
	// 生成 systemd 服务文件
	serviceContent := fmt.Sprintf(`[Unit]
Description=P-BOX Proxy Gateway
After=network.target

[Service]
Type=simple
ExecStart=%s
WorkingDirectory=%s
Restart=always
RestartSec=5
LimitNOFILE=1048576

# 安全配置
NoNewPrivileges=false
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE CAP_NET_RAW

[Install]
WantedBy=multi-user.target
`, s.binaryPath, filepath.Dir(s.binaryPath))

	servicePath := "/etc/systemd/system/p-box.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("写入服务文件失败: %v", err)
	}

	// 重新加载 systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("重新加载 systemd 失败: %v", err)
	}

	// 启用服务
	if err := exec.Command("systemctl", "enable", "p-box").Run(); err != nil {
		return fmt.Errorf("启用服务失败: %v", err)
	}

	return nil
}

// disableAutoStart 禁用开机自启
func (s *Service) disableAutoStart() error {
	exec.Command("systemctl", "disable", "p-box").Run()
	exec.Command("systemctl", "stop", "p-box").Run()
	os.Remove("/etc/systemd/system/p-box.service")
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}

// SetIPForward 设置 IP 转发
func (s *Service) SetIPForward(enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}

	// 设置 IPv4 转发
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte(value), 0644); err != nil {
		return fmt.Errorf("设置 IPv4 转发失败: %v", err)
	}

	// 设置 IPv6 转发
	os.WriteFile("/proc/sys/net/ipv6/conf/all/forwarding", []byte(value), 0644)

	// 持久化配置
	return s.persistSysctl("net.ipv4.ip_forward", value, "net.ipv6.conf.all.forwarding", value)
}

// IsIPForwardEnabled 检查 IP 转发是否开启
func (s *Service) IsIPForwardEnabled() bool {
	data, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "1"
}

// SetBBR 设置 BBR 拥塞控制
func (s *Service) SetBBR(enabled bool) error {
	if enabled {
		// 设置 fq 队列调度
		if err := os.WriteFile("/proc/sys/net/core/default_qdisc", []byte("fq"), 0644); err != nil {
			return fmt.Errorf("设置队列调度失败: %v", err)
		}
		// 设置 BBR
		if err := os.WriteFile("/proc/sys/net/ipv4/tcp_congestion_control", []byte("bbr"), 0644); err != nil {
			return fmt.Errorf("设置 BBR 失败: %v", err)
		}
		return s.persistSysctl("net.core.default_qdisc", "fq", "net.ipv4.tcp_congestion_control", "bbr")
	} else {
		os.WriteFile("/proc/sys/net/core/default_qdisc", []byte("pfifo_fast"), 0644)
		os.WriteFile("/proc/sys/net/ipv4/tcp_congestion_control", []byte("cubic"), 0644)
		return s.persistSysctl("net.core.default_qdisc", "pfifo_fast", "net.ipv4.tcp_congestion_control", "cubic")
	}
}

// IsBBREnabled 检查 BBR 是否开启
func (s *Service) IsBBREnabled() bool {
	data, err := os.ReadFile("/proc/sys/net/ipv4/tcp_congestion_control")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "bbr"
}

// SetTUNOptimize 设置 TUN 模式优化
func (s *Service) SetTUNOptimize(enabled bool) error {
	if enabled {
		// 优化设置
		optimizations := map[string]string{
			"/proc/sys/net/core/rmem_max":                  "16777216",
			"/proc/sys/net/core/wmem_max":                  "16777216",
			"/proc/sys/net/ipv4/tcp_rmem":                  "4096 87380 16777216",
			"/proc/sys/net/ipv4/tcp_wmem":                  "4096 65536 16777216",
			"/proc/sys/net/ipv4/tcp_mtu_probing":           "1",
			"/proc/sys/net/ipv4/tcp_fastopen":              "3",
			"/proc/sys/net/ipv4/tcp_slow_start_after_idle": "0",
		}
		for path, value := range optimizations {
			os.WriteFile(path, []byte(value), 0644)
		}
		return s.persistSysctl(
			"net.core.rmem_max", "16777216",
			"net.core.wmem_max", "16777216",
			"net.ipv4.tcp_mtu_probing", "1",
			"net.ipv4.tcp_fastopen", "3",
			"net.ipv4.tcp_slow_start_after_idle", "0",
		)
	}
	return nil
}

// IsTUNOptimized 检查 TUN 优化是否开启
func (s *Service) IsTUNOptimized() bool {
	data, err := os.ReadFile("/proc/sys/net/ipv4/tcp_fastopen")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "3"
}

// persistSysctl 持久化 sysctl 配置
func (s *Service) persistSysctl(keyValues ...string) error {
	configPath := "/etc/sysctl.d/99-p-box.conf"

	// 读取现有配置
	existingContent := ""
	if data, err := os.ReadFile(configPath); err == nil {
		existingContent = string(data)
	}

	// 构建新配置
	lines := make(map[string]string)
	for _, line := range strings.Split(existingContent, "\n") {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				lines[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// 更新配置
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			lines[keyValues[i]] = keyValues[i+1]
		}
	}

	// 写入文件
	var content strings.Builder
	content.WriteString("# P-BOX 系统优化配置\n")
	for key, value := range lines {
		content.WriteString(fmt.Sprintf("%s = %s\n", key, value))
	}

	if err := os.WriteFile(configPath, []byte(content.String()), 0644); err != nil {
		return err
	}

	// 应用配置
	exec.Command("sysctl", "-p", configPath).Run()
	return nil
}

// ApplyAllOptimizations 一键应用所有优化
func (s *Service) ApplyAllOptimizations() error {
	if err := s.SetIPForward(true); err != nil {
		return err
	}
	if err := s.SetBBR(true); err != nil {
		return err
	}
	if err := s.SetTUNOptimize(true); err != nil {
		return err
	}
	return nil
}
