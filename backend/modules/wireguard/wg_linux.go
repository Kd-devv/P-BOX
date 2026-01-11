//go:build linux

package wireguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GetTunInterface 获取 Mihomo TUN 接口
func GetTunInterface() string {
	// 方法1: 查找 tun 类型接口
	cmd := exec.Command("sh", "-c", "ip link show type tun | grep -v 'wg-' | grep -o '^[0-9]*: [^:]*' | awk '{print $2}' | head -1")
	output, _ := cmd.Output()
	if len(output) > 0 {
		name := strings.TrimSpace(strings.TrimSuffix(string(output), ":"))
		if name != "" {
			fmt.Printf("🔍 检测到 TUN 接口: %s\n", name)
			return name
		}
	}

	// 方法2: 从路由表查找
	cmd = exec.Command("sh", "-c", "ip route | grep -o 'dev tun[^ ]*' | awk '{print $2}' | grep -v 'wg-' | head -1")
	output, _ = cmd.Output()
	if len(output) > 0 {
		name := strings.TrimSpace(string(output))
		if name != "" {
			fmt.Printf("🔍 从路由表检测到 TUN 接口: %s\n", name)
			return name
		}
	}

	fmt.Println("⚠️  未检测到 TUN 接口，使用默认名称: tun0")
	return "tun0"
}

func getLocalNetworks() []string {
	cmd := exec.Command("sh", "-c", "ip route | grep -v default | grep -E '^[0-9]' | awk '{print $1}' | grep -E '^(10\\.|172\\.(1[6-9]|2[0-9]|3[01])\\.|192\\.168\\.)'")
	output, _ := cmd.Output()
	if len(output) > 0 {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		var nets []string
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				nets = append(nets, l)
			}
		}
		if len(nets) > 0 {
			return nets
		}
	}
	return []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "169.254.0.0/16"}
}

// isMihomoAutoRouteEnabled 检查 Mihomo 配置中是否启用了 auto_route
func (s *Service) isMihomoAutoRouteEnabled() bool {
	// 读取 mihomo 配置文件
	configPath := filepath.Join(s.dataDir, "configs", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false // 默认 false（推荐 auto_route: false）
	}

	// 简单检测是否包含 auto-route: true
	content := string(data)
	if strings.Contains(content, "auto-route: true") || strings.Contains(content, "auto_route: true") {
		return true
	}
	return false
}

// GenerateWGConfig 生成 WireGuard 配置文件内容
func (s *Service) GenerateWGConfig(serverID string) (string, error) {
	server, err := s.GetServer(serverID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	tunIf := GetTunInterface()
	ifaceName := server.Tag
	autoRouteEnabled := s.isMihomoAutoRouteEnabled()

	// [Interface] 部分
	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", server.PrivateKey))
	b.WriteString(fmt.Sprintf("Address = %s\n", server.Address))
	b.WriteString(fmt.Sprintf("ListenPort = %d\n", server.ListenPort))
	if server.MTU > 0 {
		b.WriteString(fmt.Sprintf("MTU = %d\n", server.MTU))
	}

	// ========================================
	// 流量导入 Mihomo TUN（策略路由）
	// ========================================
	if autoRouteEnabled {
		b.WriteString("\n# ========================================\n")
		b.WriteString("# 🚀 强制流量走 Mihomo TUN（auto_route: true 模式）\n")
		b.WriteString(fmt.Sprintf("# TUN 接口: %s (自动检测)\n", tunIf))
		b.WriteString("# ========================================\n\n")
	} else {
		b.WriteString("\n# ========================================\n")
		b.WriteString("# 🚀 强制流量走 Mihomo TUN（auto_route: false 模式）\n")
		b.WriteString(fmt.Sprintf("# TUN 接口: %s (自动检测)\n", tunIf))
		b.WriteString("# ========================================\n\n")
	}

	// 1. 内核参数
	b.WriteString("# 1. 内核参数\n")
	b.WriteString("PostUp = sysctl -w net.ipv4.ip_forward=1\n")
	b.WriteString("PostUp = sysctl -w net.ipv6.conf.all.forwarding=1\n")
	b.WriteString("PostUp = sysctl -w net.ipv4.conf.all.rp_filter=0\n")
	b.WriteString("PostUp = sysctl -w net.ipv4.conf.default.rp_filter=0\n\n")

	localNetworks := getLocalNetworks()

	if autoRouteEnabled {
		// auto_route: true 模式 - 使用 nftables + fwmark 0x66
		b.WriteString("# 2. 给 WG 转发流量打标（使用 nftables）\n")
		b.WriteString("PostUp = nft add table inet mangle 2>/dev/null || true\n")
		b.WriteString("PostUp = nft add chain inet mangle prerouting '{ type filter hook prerouting priority -150; }' 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostUp = nft add rule inet mangle prerouting iifname \"%s\" meta mark set 0x66\n\n", ifaceName))

		b.WriteString("# 3. 表100默认走 TUN\n")
		b.WriteString(fmt.Sprintf("PostUp = ip route replace default dev %s table 100 2>/dev/null || true\n\n", tunIf))

		b.WriteString("# 4. 内网直连优先（fwmark 0x66）\n")
		for i, network := range localNetworks {
			b.WriteString(fmt.Sprintf("PostUp = ip rule add fwmark 0x66 to %s lookup main priority %d 2>/dev/null || true\n", network, 9990+i))
		}
		b.WriteString("\n")

		b.WriteString("# 5. 外网走表100（fwmark 0x66）\n")
		b.WriteString("PostUp = ip rule add fwmark 0x66 lookup 100 priority 10000 2>/dev/null || true\n\n")

		b.WriteString("# 6. WG 握手直连豁免（防止被 auto_route 接管）\n")
		b.WriteString("PostUp = nft add chain inet mangle output '{ type route hook output priority -150; }' 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostUp = nft add rule inet mangle output udp dport %d meta mark set 0x1\n", server.ListenPort))
		b.WriteString("PostUp = ip rule add fwmark 0x1 lookup main priority 50 2>/dev/null || true\n\n")
	} else {
		// auto_route: false 模式 - 使用 iptables + fwmark 0x30（推荐）
		b.WriteString("# 2. 打标（mangle PREROUTING）\n")
		b.WriteString(fmt.Sprintf("PostUp = iptables -t mangle -A PREROUTING -i %s -j MARK --set-mark 0x30\n", ifaceName))
		b.WriteString(fmt.Sprintf("PostUp = ip6tables -t mangle -A PREROUTING -i %s -j MARK --set-mark 0x30\n\n", ifaceName))

		b.WriteString("# 3. 表100默认走 TUN\n")
		b.WriteString(fmt.Sprintf("PostUp = ip route add default dev %s table 100 2>/dev/null || true\n\n", tunIf))

		b.WriteString("# 4. 内网直连优先\n")
		for i, network := range localNetworks {
			b.WriteString(fmt.Sprintf("PostUp = ip rule add fwmark 0x30 to %s lookup main priority %d 2>/dev/null || true\n", network, 9990+i))
		}
		b.WriteString("\n")

		b.WriteString("# 5. 外网走表100\n")
		b.WriteString("PostUp = ip rule add fwmark 0x30 table 100 priority 10000 2>/dev/null || true\n\n")
	}

	// 6. DNS 劫持（REDIRECT 到 53）
	b.WriteString("# 6. DNS 劫持（REDIRECT 到 53）\n")
	b.WriteString(fmt.Sprintf("PostUp = iptables -t nat -A PREROUTING -i %s -p udp --dport 53 -j REDIRECT --to-ports 53\n", ifaceName))
	b.WriteString(fmt.Sprintf("PostUp = iptables -t nat -A PREROUTING -i %s -p tcp --dport 53 -j REDIRECT --to-ports 53\n\n", ifaceName))

	// 7. FORWARD 放行（有状态）
	b.WriteString("# 7. FORWARD 放行（有状态）\n")
	b.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %s -m conntrack --ctstate NEW,ESTABLISHED,RELATED -j ACCEPT\n", ifaceName))
	b.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -o %s -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT\n\n", ifaceName))

	// IPv6 同步
	if autoRouteEnabled {
		b.WriteString("# IPv6 同步（auto_route: true 模式）\n")
		b.WriteString("PostUp = ip -6 rule add fwmark 0x66 table 100 priority 10000 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostUp = ip -6 route add default dev %s table 100 2>/dev/null || true\n\n", tunIf))
	} else {
		b.WriteString("# IPv6 同步（auto_route: false 模式）\n")
		b.WriteString("PostUp = ip -6 rule add fwmark 0x30 table 100 priority 10000 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostUp = ip -6 route add default dev %s table 100 2>/dev/null || true\n\n", tunIf))
	}

	// ========================================
	// PostDown 清理规则
	// ========================================
	b.WriteString("# ========================================\n")
	b.WriteString("# 🧹 停止时清理所有规则\n")
	b.WriteString("# ========================================\n\n")

	if autoRouteEnabled {
		// auto_route: true 模式清理
		b.WriteString("# 清理 nftables 规则\n")
		b.WriteString(fmt.Sprintf("PostDown = nft delete rule inet mangle prerouting iifname \"%s\" meta mark set 0x66 2>/dev/null || true\n", ifaceName))
		b.WriteString(fmt.Sprintf("PostDown = nft delete rule inet mangle output udp dport %d meta mark set 0x1 2>/dev/null || true\n", server.ListenPort))
		b.WriteString("PostDown = ip rule del fwmark 0x66 lookup 100 priority 10000 2>/dev/null || true\n")
		b.WriteString("PostDown = ip rule del fwmark 0x1 lookup main priority 50 2>/dev/null || true\n")
		for i, network := range localNetworks {
			b.WriteString(fmt.Sprintf("PostDown = ip rule del fwmark 0x66 to %s lookup main priority %d 2>/dev/null || true\n", network, 9990+i))
		}
		b.WriteString(fmt.Sprintf("PostDown = ip route del default dev %s table 100 2>/dev/null || true\n", tunIf))
		b.WriteString("PostDown = ip -6 rule del fwmark 0x66 table 100 priority 10000 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostDown = ip -6 route del default dev %s table 100 2>/dev/null || true\n", tunIf))
	} else {
		// auto_route: false 模式清理
		b.WriteString(fmt.Sprintf("PostDown = iptables -t mangle -D PREROUTING -i %s -j MARK --set-mark 0x30 2>/dev/null || true\n", ifaceName))
		b.WriteString(fmt.Sprintf("PostDown = ip6tables -t mangle -D PREROUTING -i %s -j MARK --set-mark 0x30 2>/dev/null || true\n", ifaceName))
		b.WriteString("PostDown = ip rule del fwmark 0x30 table 100 priority 10000 2>/dev/null || true\n")
		for i, network := range localNetworks {
			b.WriteString(fmt.Sprintf("PostDown = ip rule del fwmark 0x30 to %s lookup main priority %d 2>/dev/null || true\n", network, 9990+i))
		}
		b.WriteString(fmt.Sprintf("PostDown = ip route del default dev %s table 100 2>/dev/null || true\n", tunIf))
		b.WriteString("PostDown = ip -6 rule del fwmark 0x30 table 100 priority 10000 2>/dev/null || true\n")
		b.WriteString(fmt.Sprintf("PostDown = ip -6 route del default dev %s table 100 2>/dev/null || true\n", tunIf))
	}

	// DNS 和 FORWARD 清理（两种模式通用）
	b.WriteString(fmt.Sprintf("PostDown = iptables -t nat -D PREROUTING -i %s -p udp --dport 53 -j REDIRECT --to-ports 53 2>/dev/null || true\n", ifaceName))
	b.WriteString(fmt.Sprintf("PostDown = iptables -t nat -D PREROUTING -i %s -p tcp --dport 53 -j REDIRECT --to-ports 53 2>/dev/null || true\n", ifaceName))
	b.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %s -m conntrack --ctstate NEW,ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || true\n", ifaceName))
	b.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -o %s -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || true\n", ifaceName))

	b.WriteString("\n")

	// [Peer] 部分
	for _, c := range server.Clients {
		if !c.Enabled {
			continue
		}
		b.WriteString(fmt.Sprintf("# %s\n[Peer]\n", c.Name))
		b.WriteString(fmt.Sprintf("PublicKey = %s\n", c.PublicKey))
		if c.PresharedKey != "" {
			b.WriteString(fmt.Sprintf("PresharedKey = %s\n", c.PresharedKey))
		}
		b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", c.AllowedIPs))
	}

	return b.String(), nil
}

// InstallWireGuard 自动安装 WireGuard
func (s *Service) InstallWireGuard() error {
	// 检测包管理器
	if _, err := exec.LookPath("apt"); err == nil {
		// Debian/Ubuntu
		fmt.Println("📦 使用 apt 安装 WireGuard...")
		cmd := exec.Command("apt", "update")
		cmd.Run()
		cmd = exec.Command("apt", "install", "-y", "wireguard", "wireguard-tools")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("apt 安装失败: %v, %s", err, string(out))
		}
		return nil
	}

	if _, err := exec.LookPath("yum"); err == nil {
		// CentOS/RHEL
		fmt.Println("📦 使用 yum 安装 WireGuard...")
		cmd := exec.Command("yum", "install", "-y", "wireguard-tools")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("yum 安装失败: %v, %s", err, string(out))
		}
		return nil
	}

	if _, err := exec.LookPath("dnf"); err == nil {
		// Fedora
		fmt.Println("📦 使用 dnf 安装 WireGuard...")
		cmd := exec.Command("dnf", "install", "-y", "wireguard-tools")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("dnf 安装失败: %v, %s", err, string(out))
		}
		return nil
	}

	return fmt.Errorf("不支持的包管理器，请手动安装: apt/yum/dnf install wireguard-tools")
}

// ForceCleanupInterface 强制清理 WireGuard 接口（启动前调用）
func (s *Service) ForceCleanupInterface(tag string) {
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", tag)

	// 1. wg-quick down（会执行 PostDown 清理规则）
	if _, err := os.Stat(configPath); err == nil {
		exec.Command("wg-quick", "down", configPath).Run()
	}

	// 2. 停止 systemd 服务（如果使用过）
	exec.Command("systemctl", "disable", "--now", "wg-quick@"+tag).Run()

	// 3. 双保险：直接删除接口
	exec.Command("ip", "link", "delete", "dev", tag).Run()

	// 4. 清理策略路由规则
	exec.Command("ip", "rule", "del", "fwmark", "0x30", "table", "100", "priority", "10000").Run()
	exec.Command("ip", "rule", "del", "fwmark", "0x66", "lookup", "100", "priority", "10000").Run()
	exec.Command("ip", "-6", "rule", "del", "fwmark", "0x30", "table", "100", "priority", "10000").Run()
	exec.Command("ip", "-6", "rule", "del", "fwmark", "0x66", "table", "100", "priority", "10000").Run()

	// 5. 清理内网直连规则（遍历 priority 9990-9999）
	for priority := 9990; priority < 10000; priority++ {
		exec.Command("ip", "rule", "del", "priority", fmt.Sprintf("%d", priority)).Run()
	}

	// 6. 清理路由
	tunIf := GetTunInterface()
	exec.Command("ip", "route", "del", "default", "dev", tunIf, "table", "100").Run()
	exec.Command("ip", "-6", "route", "del", "default", "dev", tunIf, "table", "100").Run()
	exec.Command("ip", "route", "del", "default", "table", "100").Run()

	fmt.Printf("✅ 接口 %s 清理完成\n", tag)
}

// ApplyConfig 应用配置并启动
func (s *Service) ApplyConfig(serverID string) error {
	server, err := s.GetServer(serverID)
	if err != nil {
		return err
	}

	// 检查是否已安装
	if !s.CheckInstalled() {
		return fmt.Errorf("WireGuard 未安装，请先点击安装按钮")
	}

	// 强制清理旧接口（避免地址冲突）
	fmt.Printf("🧹 清理旧接口: %s\n", server.Tag)
	s.ForceCleanupInterface(server.Tag)
	time.Sleep(500 * time.Millisecond)

	// 生成配置
	configContent, err := s.GenerateWGConfig(serverID)
	if err != nil {
		return err
	}

	// 写入配置文件
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", server.Tag)
	os.MkdirAll("/etc/wireguard", 0700)
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("写入配置失败: %v", err)
	}

	fmt.Printf("📄 配置文件已写入: %s\n", configPath)

	// 启动
	cmd := exec.Command("wg-quick", "up", configPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动失败: %v, %s", err, string(out))
	}

	// 更新状态
	s.mu.Lock()
	for i := range s.config.Servers {
		if s.config.Servers[i].ID == serverID {
			s.config.Servers[i].Enabled = true
			break
		}
	}
	s.saveConfig()
	s.mu.Unlock()

	fmt.Printf("✅ WireGuard 接口 %s 启动成功\n", server.Tag)
	return nil
}

// StopInterface 停止接口
func (s *Service) StopInterface(tag string) error {
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", tag)
	exec.Command("wg-quick", "down", configPath).Run()
	exec.Command("ip", "link", "delete", "dev", tag).Run()

	// 更新状态
	s.mu.Lock()
	for i := range s.config.Servers {
		if s.config.Servers[i].Tag == tag {
			s.config.Servers[i].Enabled = false
			break
		}
	}
	s.saveConfig()
	s.mu.Unlock()

	return nil
}

// GetStatus 获取接口状态
func (s *Service) GetStatus(tag string) (bool, string) {
	cmd := exec.Command("wg", "show", tag)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}
	return true, string(out)
}

// GenerateClientConfig 生成客户端配置
func (s *Service) GenerateClientConfig(serverID, clientID, endpoint string) (string, error) {
	server, err := s.GetServer(serverID)
	if err != nil {
		return "", err
	}

	var client *WireGuardClient
	for _, c := range server.Clients {
		if c.ID == clientID {
			client = &c
			break
		}
	}
	if client == nil {
		return "", fmt.Errorf("客户端不存在")
	}

	// 优先使用传入的 endpoint，其次使用服务器配置的 endpoint
	actualEndpoint := endpoint
	if actualEndpoint == "" && server.Endpoint != "" {
		actualEndpoint = server.Endpoint
	}
	if actualEndpoint == "" {
		return "", fmt.Errorf("请设置服务器的公网地址/域名")
	}

	var b strings.Builder
	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", client.PrivateKey))
	b.WriteString(fmt.Sprintf("Address = %s\n", client.AllowedIPs))
	if client.DNS != "" {
		b.WriteString(fmt.Sprintf("DNS = %s\n", client.DNS))
	}

	b.WriteString("\n[Peer]\n")
	b.WriteString(fmt.Sprintf("PublicKey = %s\n", server.PublicKey))
	if client.PresharedKey != "" {
		b.WriteString(fmt.Sprintf("PresharedKey = %s\n", client.PresharedKey))
	}
	b.WriteString("AllowedIPs = 0.0.0.0/0, ::/0\n")
	b.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", actualEndpoint, server.ListenPort))
	b.WriteString("PersistentKeepalive = 25\n")

	return b.String(), nil
}
