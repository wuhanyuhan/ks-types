// Package kstypes
// config_schema.go — Spec A（MCP 配置 Schema 协议）v0.6.0 新增类型
package kstypes

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// ConfigSchemaResponse 是 MCP /config-schema 端点的响应 data 字段。
type ConfigSchemaResponse struct {
	Schema   map[string]any `json:"schema"`
	UISchema map[string]any `json:"ui_schema"`
	Version  string         `json:"version"`
}

// ConfigPubkeyResponse 是 MCP /config-pubkey 端点的响应 data 字段。
type ConfigPubkeyResponse struct {
	Pubkey      string `json:"pubkey"`      // base64-std 32 bytes
	Fingerprint string `json:"fingerprint"` // 格式见 spec-v1 §4.2
	Algorithm   string `json:"algorithm"`   // "x25519-ecdh-aes256gcm-v1"
	CreatedAt   string `json:"created_at"`  // RFC 3339 UTC
}

// EncryptedConfigPayload 是 POST /ks-config/save 与 /validate 的 request body。
type EncryptedConfigPayload struct {
	Algorithm       string         `json:"algorithm"`        // "x25519-ecdh-aes256gcm-v1"
	EphemeralPubkey string         `json:"ephemeral_pubkey"` // base64-std 32 bytes
	Nonce           string         `json:"nonce"`            // base64-std 12 bytes
	AADFields       map[string]any `json:"aad_fields"`       // mcp_server_id / config_version / fingerprint
	AADCanonical    string         `json:"aad_canonical"`    // base64-std of spec-v1 §2.1 bytes
	Ciphertext      string         `json:"ciphertext"`       // base64-std ciphertext+tag
	IdempotencyKey  string         `json:"idempotency_key"`  // uuid-v4
}

// ConfigApplyResult 是 POST /ks-config/save 成功响应 data 字段。
type ConfigApplyResult struct {
	AppliedAt string `json:"applied_at"` // RFC 3339 UTC
	Version   uint64 `json:"version"`
}

// AADCanonicalBytes 按 spec-v1 §2.1 规范拼接 canonical AAD 字节串。
// 三语言 SDK 与 Keystone 前端互通，输出必须字节级一致。
//
// 编码格式：
//
//	[uint16 big-endian: len(mcpServerID UTF-8)]
//	[mcpServerID UTF-8 bytes]
//	[uint64 big-endian: configVersion]
//	[uint16 big-endian: len(fingerprint UTF-8)]
//	[fingerprint UTF-8 bytes]
func AADCanonicalBytes(mcpServerID string, configVersion uint64, fingerprint string) []byte {
	var buf bytes.Buffer
	idBytes := []byte(mcpServerID)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(idBytes)))
	buf.Write(idBytes)
	_ = binary.Write(&buf, binary.BigEndian, configVersion)
	fpBytes := []byte(fingerprint)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(fpBytes)))
	buf.Write(fpBytes)
	return buf.Bytes()
}

// Fingerprint 按 spec-v1 §4.1 规范计算 X25519 公钥指纹。
// 格式："ab12:cd34:ef56:7890:1234:5678:9abc:def0"（8 段 × 4 hex = 32 hex chars = 前 16 字节）
//
// 算法：SHA-256(pubkey) 取前 16 字节，hex 编码后每 4 字符插入 ':'。
//
// pubkey 必须是 32 字节（X25519 公钥长度），否则 panic。
func Fingerprint(pubkey []byte) string {
	if len(pubkey) != 32 {
		panic(fmt.Sprintf("kstypes: Fingerprint pubkey 必须是 32 字节，收到 %d 字节", len(pubkey)))
	}
	h := sha256.Sum256(pubkey)
	hexStr := hex.EncodeToString(h[:16])
	parts := make([]string, 0, 8)
	for i := 0; i < 32; i += 4 {
		parts = append(parts, hexStr[i:i+4])
	}
	return strings.Join(parts, ":")
}
