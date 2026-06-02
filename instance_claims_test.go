package kstypes

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func loadTestKeys(t *testing.T) (priv, pub []byte) {
	t.Helper()
	return mustTestKeyPEM(t)
}

// marshalPublicKeyPEM 将 Ed25519 公钥编码为 PEM（仅测试用）
func marshalPublicKeyPEM(pub ed25519.PublicKey) []byte {
	b, _ := x509.MarshalPKIXPublicKey(pub)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b})
}

func TestInstanceJWT_SignAndVerify(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := InstanceClaims{
		InstanceID: "inst_123",
		Name:       "客户A-生产",
		Group:      "enterprise",
	}

	token, err := SignInstanceJWT(claims, priv, 90*24*time.Hour, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	got, err := VerifyInstanceJWT(token, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if got.InstanceID != "inst_123" {
		t.Errorf("instance_id: got %q, want %q", got.InstanceID, "inst_123")
	}
	if got.Name != "客户A-生产" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.Group != "enterprise" {
		t.Errorf("group: got %q", got.Group)
	}
	// Subject 应与 InstanceID 一致
	if got.Subject != "inst_123" {
		t.Errorf("subject: got %q", got.Subject)
	}
}

func TestInstanceJWT_Expired(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := InstanceClaims{InstanceID: "inst_expired"}
	token, err := SignInstanceJWT(claims, priv, -1*time.Hour, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	_, err = VerifyInstanceJWT(token, pub)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestInstanceJWT_InvalidSignature(t *testing.T) {
	priv, _ := loadTestKeys(t)

	claims := InstanceClaims{InstanceID: "inst_bad"}
	token, err := SignInstanceJWT(claims, priv, time.Hour, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// 用另一对密钥的公钥验证
	otherPub, _, _ := ed25519.GenerateKey(nil)
	otherPubPEM := marshalPublicKeyPEM(otherPub)

	_, err = VerifyInstanceJWT(token, otherPubPEM)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestInstanceJWT_Audience(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := InstanceClaims{InstanceID: "inst_aud"}
	token, err := SignInstanceJWT(claims, priv, time.Hour, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := VerifyInstanceJWT(token, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	// audience 扩展为 4 项，覆盖所有云服务
	want := []string{"ks-admin", "ks-hub", "ks-relay", "ks-llm-gateway"}
	aud := []string(got.Audience)
	if len(aud) != len(want) {
		t.Fatalf("audience len: got %d %v, want %d %v", len(aud), aud, len(want), want)
	}
	for i, v := range want {
		if aud[i] != v {
			t.Errorf("audience[%d]: got %q, want %q", i, aud[i], v)
		}
	}
}

func TestInstanceClaims_TzRoundTrip(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := InstanceClaims{
		InstanceID: "inst_tz_1",
		Name:       "demo",
		Group:      "default",
		Tz:         "Asia/Shanghai",
	}
	tok, err := SignInstanceJWT(claims, priv, time.Minute, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, err := VerifyInstanceJWT(tok, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Tz != "Asia/Shanghai" {
		t.Fatalf("Tz = %q, want Asia/Shanghai", got.Tz)
	}
}

// Tz="" 时序列化为 omitempty——验签拿到空串，消费方应 fallback。
// 这是 ks-types 与老代码（未升级到此版本的 ks-admin）兼容的关键保证。
func TestInstanceClaims_TzOmittedWhenEmpty(t *testing.T) {
	priv, pub := loadTestKeys(t)

	tok, err := SignInstanceJWT(InstanceClaims{InstanceID: "inst_tz_empty", Name: "x"}, priv, time.Minute, "")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, err := VerifyInstanceJWT(tok, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Tz != "" {
		t.Fatalf("Tz = %q, want empty", got.Tz)
	}
}

// kid 非空 → token header 写入对应值；空串 → header 不出现 kid 字段。
// 下游验签链路（如 ks-relay 走 JWKS）依赖 header.kid 选公钥，
// kid 算法由调用方控制（避免在 ks-types 重复实现），SignInstanceJWT 只负责按入参写入。
func TestSignInstanceJWT_SetsKID(t *testing.T) {
	priv, _ := loadTestKeys(t)

	t.Run("非空 kid 写入 header", func(t *testing.T) {
		token, err := SignInstanceJWT(InstanceClaims{InstanceID: "inst_kid"}, priv, time.Hour, "my-kid-123")
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		parsed, _, err := jwt.NewParser().ParseUnverified(token, &InstanceClaims{})
		if err != nil {
			t.Fatalf("parse unverified: %v", err)
		}
		got, _ := parsed.Header["kid"].(string)
		if got != "my-kid-123" {
			t.Fatalf("header.kid: got %q, want %q", got, "my-kid-123")
		}
	})

	t.Run("空 kid 不写 header", func(t *testing.T) {
		token, err := SignInstanceJWT(InstanceClaims{InstanceID: "inst_no_kid"}, priv, time.Hour, "")
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		parsed, _, err := jwt.NewParser().ParseUnverified(token, &InstanceClaims{})
		if err != nil {
			t.Fatalf("parse unverified: %v", err)
		}
		if _, ok := parsed.Header["kid"]; ok {
			t.Fatalf("header.kid 应不存在，实际 = %v", parsed.Header["kid"])
		}
	})
}

func TestVerifyInstanceJWT_MalformedTokens(t *testing.T) {
	_, pub := loadTestKeys(t)

	cases := []struct {
		name  string
		token string
	}{
		{"空字符串", ""},
		{"随机垃圾", "not-a-jwt-token-at-all"},
		{"只有 header", "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9"},
		{"两段（缺 signature）", "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0"},
		{"三段但 signature 损坏", "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.AAAA_invalid_sig"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := VerifyInstanceJWT(c.token, pub)
			if err == nil {
				t.Errorf("expected error for malformed token %q", c.name)
			}
		})
	}
}
