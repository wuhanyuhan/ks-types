package kstypes

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"
)

func loadTestKeys(t *testing.T) (priv, pub []byte) {
	t.Helper()
	var err error
	priv, err = os.ReadFile("testdata/test_private.pem")
	if err != nil {
		t.Fatalf("read private key: %v", err)
	}
	pub, err = os.ReadFile("testdata/test_public.pem")
	if err != nil {
		t.Fatalf("read public key: %v", err)
	}
	return
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

	token, err := SignInstanceJWT(claims, priv, 90*24*time.Hour)
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
	token, err := SignInstanceJWT(claims, priv, -1*time.Hour)
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
	token, err := SignInstanceJWT(claims, priv, time.Hour)
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
	token, err := SignInstanceJWT(claims, priv, time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := VerifyInstanceJWT(token, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	aud := got.Audience
	if len(aud) != 2 || aud[0] != "ks-hub" || aud[1] != "ks-admin" {
		t.Errorf("audience: got %v", aud)
	}
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
