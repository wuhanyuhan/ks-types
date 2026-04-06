package kstypes

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// InstanceClaims 实例 JWT Claims
// JSON 字段名与 spec 一致: sub(Subject), name, group
type InstanceClaims struct {
	jwt.RegisteredClaims
	InstanceID string `json:"-"`     // 程序内使用，映射到 RegisteredClaims.Subject
	Name       string `json:"name"`  // 实例名称
	Group      string `json:"group"` // 实例分组
}

// SignInstanceJWT 用 Ed25519 私钥签发实例 JWT
func SignInstanceJWT(claims InstanceClaims, privatePEM []byte, ttl time.Duration) (string, error) {
	privKey, err := ParseEd25519PrivateKeyPEM(privatePEM)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	now := time.Now().UTC()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.InstanceID,
		Issuer:    "ks-admin",
		Audience:  jwt.ClaimStrings{"ks-hub", "ks-admin"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(privKey)
}

// VerifyInstanceJWT 用 Ed25519 公钥验证实例 JWT
func VerifyInstanceJWT(tokenString string, publicPEM []byte) (*InstanceClaims, error) {
	pubKey, err := ParseEd25519PublicKeyPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	claims := &InstanceClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	// 从 Subject 回填 InstanceID
	claims.InstanceID = claims.Subject
	return claims, nil
}
