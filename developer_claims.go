package kstypes

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DeveloperClaims 开发者 JWT Claims
type DeveloperClaims struct {
	jwt.RegisteredClaims
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	TokenVersion int    `json:"token_version"`
}

// SignDeveloperJWT 用 Ed25519 私钥签发开发者 JWT
func SignDeveloperJWT(claims DeveloperClaims, privatePEM []byte, ttl time.Duration) (string, error) {
	privKey, err := ParseEd25519PrivateKeyPEM(privatePEM)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	now := time.Now().UTC()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", claims.UserID),
		Issuer:    "ks-hub",
		Audience:  jwt.ClaimStrings{"ks-hub"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(privKey)
}

// VerifyDeveloperJWT 用 Ed25519 公钥验证开发者 JWT
func VerifyDeveloperJWT(tokenString string, publicPEM []byte) (*DeveloperClaims, error) {
	pubKey, err := ParseEd25519PublicKeyPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	claims := &DeveloperClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	return claims, nil
}
