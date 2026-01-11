package proxy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// CoreManager 核心管理器
type CoreManager struct {
	dataDir    string
	coreType   string // mihomo or singbox
	process    *exec.Cmd
	configPath string
	running    bool
	logs       []string
	logMu      sync.RWMutex
	mu         sync.Mutex
	onLog      func(string)
	stopChan   chan struct{}
}

// CoreStatus 核心状态
type CoreStatus struct {
	Running    bool      `json:"running"`
	CoreType   string    `json:"coreType"`
	ConfigPath string    `json:"configPath"`
	StartTime  time.Time `json:"startTime,omitempty"`
	PID        int       `json:"pid,omitempty"`
}

func NewCoreManager(dataDir string) *CoreManager {
	return &CoreManager{
		dataDir:  dataDir,
		coreType: "mihomo",
		logs:     make([]string, 0, 1000),
	}
}

// SetLogCallback 设置日志回调
func (m *CoreManager) SetLogCallback(callback func(string)) {
	m.onLog = callback
}

// GetCoreBinaryPath 获取核心二进制文件路径
func (m *CoreManager) GetCoreBinaryPath() string {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	var binName string
	switch m.coreType {
	case "mihomo":
		binName = fmt.Sprintf("mihomo-%s-%s", goos, arch)
	case "singbox":
		binName = fmt.Sprintf("sing-box-%s-%s", goos, arch)
	default:
		binName = fmt.Sprintf("mihomo-%s-%s", goos, arch)
	}

	if goos == "windows" {
		binName += ".exe"
	}

	return filepath.Join(m.dataDir, "cores", binName)
}

// SetCoreType 设置核心类型
func (m *CoreManager) SetCoreType(coreType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coreType = coreType
}

// Start 启动核心
func (m *CoreManager) Start(configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("core is already running")
	}

	binPath := m.GetCoreBinaryPath()
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("core binary not found: %s", binPath)
	}

	// 确保配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	m.configPath = configPath

	// 创建工作目录
	workDir := filepath.Join(m.dataDir, "runtime")
	os.MkdirAll(workDir, 0755)

	// 构建命令
	var cmd *exec.Cmd
	if m.coreType == "singbox" {
		cmd = exec.Command(binPath, "run", "-c", configPath)
	} else {
		cmd = exec.Command(binPath, "-d", workDir, "-f", configPath)
	}

	cmd.Dir = workDir

	// 获取输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start core: %v", err)
	}

	m.process = cmd
	m.running = true
	m.stopChan = make(chan struct{})

	// 读取日志
	go m.readLogs(stdout)
	go m.readLogs(stderr)

	// 监控进程
	go m.watchProcess()

	m.addLog(fmt.Sprintf("[INFO] %s started with config: %s", m.coreType, configPath))

	return nil
}

// Stop 停止核心
func (m *CoreManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running || m.process == nil {
		return nil
	}

	close(m.stopChan)

	// 发送终止信号
	if err := m.process.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill core process: %v", err)
	}

	m.process.Wait()
	m.running = false
	m.process = nil
	m.configPath = ""

	m.addLog(fmt.Sprintf("[INFO] %s stopped", m.coreType))

	return nil
}

// Restart 重启核心
func (m *CoreManager) Restart() error {
	configPath := m.configPath
	if configPath == "" {
		return fmt.Errorf("no config path set")
	}

	if err := m.Stop(); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	return m.Start(configPath)
}

// Status 获取状态
func (m *CoreManager) Status() *CoreStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := &CoreStatus{
		Running:    m.running,
		CoreType:   m.coreType,
		ConfigPath: m.configPath,
	}

	if m.running && m.process != nil && m.process.Process != nil {
		status.PID = m.process.Process.Pid
	}

	return status
}

// IsRunning 检查是否运行中
func (m *CoreManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// GetLogs 获取日志
func (m *CoreManager) GetLogs(limit int) []string {
	m.logMu.RLock()
	defer m.logMu.RUnlock()

	if limit <= 0 || limit > len(m.logs) {
		limit = len(m.logs)
	}

	start := len(m.logs) - limit
	if start < 0 {
		start = 0
	}

	result := make([]string, limit)
	copy(result, m.logs[start:])
	return result
}

// ClearLogs 清除日志
func (m *CoreManager) ClearLogs() {
	m.logMu.Lock()
	defer m.logMu.Unlock()
	m.logs = make([]string, 0, 1000)
}

func (m *CoreManager) readLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		m.addLog(line)
	}
}

func (m *CoreManager) addLog(line string) {
	m.logMu.Lock()

	// 限制日志数量
	if len(m.logs) >= 1000 {
		m.logs = m.logs[100:]
	}
	m.logs = append(m.logs, line)

	m.logMu.Unlock()

	// 回调
	if m.onLog != nil {
		m.onLog(line)
	}
}

func (m *CoreManager) watchProcess() {
	if m.process == nil {
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- m.process.Wait()
	}()

	select {
	case <-m.stopChan:
		return
	case err := <-done:
		m.mu.Lock()
		m.running = false
		m.process = nil
		m.mu.Unlock()

		if err != nil {
			m.addLog(fmt.Sprintf("[ERROR] Core process exited with error: %v", err))
		} else {
			m.addLog("[INFO] Core process exited normally")
		}
	}
}

// ReloadConfig 重新加载配置（热重载）
func (m *CoreManager) ReloadConfig() error {
	if !m.running || m.process == nil {
		return fmt.Errorf("core is not running")
	}

	// Mihomo 支持 SIGHUP 热重载
	if m.coreType == "mihomo" {
		// 发送 SIGHUP 信号（仅 Unix）
		if runtime.GOOS != "windows" {
			if err := m.process.Process.Signal(syscall.SIGHUP); err != nil {
				return fmt.Errorf("failed to send reload signal: %v", err)
			}
			m.addLog("[INFO] Configuration reload signal sent")
			return nil
		}
	}

	// 其他情况需要重启
	return m.Restart()
}
