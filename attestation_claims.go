package kstypes

import (
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
