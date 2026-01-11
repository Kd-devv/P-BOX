package subscription

import (
	"errors"
	"net/url"
	"strings"
)

// SSHConfig SSH 配置结构
type SSHConfig struct {
	Username             string   `json:"user"`
	Password             string   `json:"password,omitempty"`
	PrivateKey           string   `json:"private_key,omitempty"`
	PrivateKeyPassphrase string   `json:"private_key_passphrase,omitempty"`
	HostKey              []string `json:"host_key,omitempty"`
	HostKeyAlgorithms    []string `json:"host_key_algorithms,omitempty"`
	ClientVersion        string   `json:"client_version,omitempty"`
}

// ParseSSHURL 解析 SSH 链接
// 格式: ssh://username:password@server:port?params#name
func ParseSSHURL(sshURL string) (*ProxyNode, error) {
	// 移除协议前缀
	urlStr := strings.TrimPrefix(sshURL, "ssh://")

	// 解析 URL
	u, err := url.Parse("ssh://" + urlStr)
	if err != nil {
		return nil, errors.New("ssh URL 格式错误")
	}

	// 提取服务器和端口
	server := u.Hostname()
	port := ParseInt(u.Port(), 22)

	// 提取节点名称
	name := u.Fragment
	if name == "" {
		name = "SSH"
	}
	name, _ = url.QueryUnescape(name)

	// 提取用户名和密码
	username := u.User.Username()
	password, _ := u.User.Password()

	// 解析查询参数
	query := u.Query()

	// 解析 host_key
	hostKey := []string{}
	if hk := query.Get("host_key"); hk != "" {
		hostKey = strings.Split(hk, ",")
	}

	// 解析 host_key_algorithms
	hostKeyAlgorithms := []string{}
	if hka := query.Get("host_key_algorithms"); hka != "" {
		hostKeyAlgorithms = strings.Split(hka, ",")
	}

	// 构建配置
	config := SSHConfig{
		Username:             username,
		Password:             password,
		PrivateKey:           query.Get("private_key"),
		PrivateKeyPassphrase: query.Get("private_key_passphrase"),
		HostKey:              hostKey,
		HostKeyAlgorithms:    hostKeyAlgorithms,
		ClientVersion:        query.Get("client_version"),
	}

	// 转换为 JSON 字符串
	configJSON, err := ToJSONString(config)
	if err != nil {
		return nil, errors.New("ssh 配置序列化失败")
	}

	node := &ProxyNode{
		Name:       name,
		Type:       "ssh",
		Server:     server,
		ServerPort: port,
		Config:     configJSON,
	}

	return node, nil
}

// ToJSON 实现 NodeConfig 接口
func (c *SSHConfig) ToJSON() (string, error) {
	return ToJSONString(c)
}
