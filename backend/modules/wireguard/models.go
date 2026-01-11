package wireguard

import (
	"time"
)

// WireGuardServer WireGuard 服务端配置
type WireGuardServer struct {
	ID          string            `json:"id"`          // 唯一标识
	Name        string            `json:"name"`        // 服务名称
	Tag         string            `json:"tag"`         // 接口名（如 wg0）
	Enabled     bool              `json:"enabled"`     // 是否启用
	AutoStart   bool              `json:"auto_start"`  // 开机启动
	Endpoint    string            `json:"endpoint"`    // 公网地址/域名（用于客户端配置）
	ListenPort  int               `json:"listen_port"` // 监听端口
	PrivateKey  string            `json:"private_key"` // 服务器私钥
	PublicKey   string            `json:"public_key"`  // 服务器公钥
	Address     string            `json:"address"`     // 服务器地址（如 10.0.1.1/24）
	MTU         int               `json:"mtu"`         // MTU 设置
	DNS         string            `json:"dns"`         // DNS 设置
	Description string            `json:"description"` // 描述
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Clients     []WireGuardClient `json:"clients"` // 客户端列表
}

// WireGuardClient WireGuard 客户端（Peer）配置
type WireGuardClient struct {
	ID           string    `json:"id"`            // 唯一标识
	Name         string    `json:"name"`          // 客户端名称
	PrivateKey   string    `json:"private_key"`   // 客户端私钥
	PublicKey    string    `json:"public_key"`    // 客户端公钥
	PresharedKey string    `json:"preshared_key"` // 预共享密钥（可选）
	AllowedIPs   string    `json:"allowed_ips"`   // 分配的 IP（如 10.0.1.2/32）
	DNS          string    `json:"dns"`           // 客户端 DNS
	Enabled      bool      `json:"enabled"`       // 是否启用
	Description  string    `json:"description"`   // 描述
	CreatedAt    time.Time `json:"created_at"`
}

// WireGuardConfig 存储配置
type WireGuardConfig struct {
	Servers []WireGuardServer `json:"servers"`
}

// WireGuardKeyPair 密钥对
type WireGuardKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}
