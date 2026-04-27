package kstypes

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AttestationClaims 是 ks-admin 签发给 ks-client 的实例身份证明 JWT 的 Claims
//
// JSON 字段名与 spec 一致；att_ver 当前固定为 1
type AttestationClaims struct {
	jwt.RegisteredClaims
	InstanceID   string `json:"instance_id"`
	E2EPublicKey string `json:"e2e_public_key"`
	OrgName      string `json:"org_name"`
	InstanceName string `json:"instance_name"`
	AttVer       int    `json:"att_ver"`
}

// SignAttestation 用 Ed25519 私钥签发 Attestation JWT
//
// kid 是签发方密钥 ID（写入 JWT header.kid），ks-client 端比对内置常量用；不能为空
// ttl 决定 ExpiresAt（当前时间 + ttl）
//
// 函数会强制覆写 RegisteredClaims 的 Issuer/Audience/Subject/IssuedAt/ExpiresAt
// 与 AttVer；调用方传入的这些字段会被忽略
func SignAttestation(claims AttestationClaims, privatePEM []byte, kid string, ttl time.Duration) (string, error) {
	if kid == "" {
		return "", fmt.Errorf("kid is required")
	}

	privKey, err := ParseEd25519PrivateKeyPEM(privatePEM)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	now := time.Now().UTC()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.InstanceID,
		Issuer:    "ks-admin",
		Audience:  jwt.ClaimStrings{"ks-client"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	claims.AttVer = 1

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["typ"] = "ATT+JWT"
	token.Header["kid"] = kid
	return token.SignedString(privKey)
}
