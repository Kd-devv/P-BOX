package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// GenerateKeyPair 生成 WireGuard 密钥对
func GenerateKeyPair() (*WireGuardKeyPair, error) {
	// 生成 32 字节的随机私钥
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}

	// WireGuard 私钥需要进行 clamp 操作
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// 从私钥生成公钥 (Curve25519)
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("生成公钥失败: %w", err)
	}

	return &WireGuardKeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
	}, nil
}

// GeneratePresharedKey 生成预共享密钥
func GeneratePresharedKey() (string, error) {
	psk := make([]byte, 32)
	if _, err := rand.Read(psk); err != nil {
		return "", fmt.Errorf("生成预共享密钥失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(psk), nil
}

// PublicKeyFromPrivateKey 从私钥计算公钥
func PublicKeyFromPrivateKey(privateKeyBase64 string) (string, error) {
	privateKey, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("私钥解码失败: %w", err)
	}

	if len(privateKey) != 32 {
		return "", fmt.Errorf("私钥长度错误: 期望 32 字节，实际 %d 字节", len(privateKey))
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("计算公钥失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(publicKey), nil
}
