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

// VerifyAttestation 用 Ed25519 公钥验证 Attestation JWT
//
// expectedKid 非空时，校验 JWT header.kid 必须与之相等；为空时跳过 kid 校验
//
// 校验项（任一失败返回 error）：
//   - 签名 (Ed25519) 有效
//   - header.alg == "EdDSA"
//   - header.typ == "ATT+JWT"
//   - header.kid == expectedKid（仅当 expectedKid 非空）
//   - claims.Issuer == "ks-admin"
//   - claims.Audience 包含 "ks-client"
//   - claims.AttVer == 1
//   - claims.ExpiresAt > now（jwt 库自动校验）
func VerifyAttestation(tokenString string, publicPEM []byte, expectedKid string) (*AttestationClaims, error) {
	pubKey, err := ParseEd25519PublicKeyPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	claims := &AttestationClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		typ, _ := t.Header["typ"].(string)
		if typ != "ATT+JWT" {
			return nil, fmt.Errorf("unexpected typ: %q", typ)
		}
		if expectedKid != "" {
			kid, _ := t.Header["kid"].(string)
			if kid != expectedKid {
				return nil, fmt.Errorf("kid mismatch: got %q, want %q", kid, expectedKid)
			}
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	if claims.Issuer != "ks-admin" {
		return nil, fmt.Errorf("iss mismatch: got %q", claims.Issuer)
	}
	if !claimsAudienceContains(claims.Audience, "ks-client") {
		return nil, fmt.Errorf("aud mismatch: got %v, want ks-client", claims.Audience)
	}
	if claims.AttVer != 1 {
		return nil, fmt.Errorf("unsupported att_ver: %d", claims.AttVer)
	}

	claims.InstanceID = claims.Subject
	return claims, nil
}

func claimsAudienceContains(aud jwt.ClaimStrings, want string) bool {
	for _, a := range aud {
		if a == want {
			return true
		}
	}
	return false
}
