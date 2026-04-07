package kstypes

import (
	"os"
	"testing"
	"time"
)

func TestDeveloperJWT_SignAndVerify(t *testing.T) {
	priv, _ := os.ReadFile("testdata/test_private.pem")
	pub, _ := os.ReadFile("testdata/test_public.pem")

	claims := DeveloperClaims{
		UserID:      42,
		Email:       "dev@example.com",
		DisplayName: "测试开发者",
	}

	token, err := SignDeveloperJWT(claims, priv, 24*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := VerifyDeveloperJWT(token, pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if got.UserID != 42 {
		t.Errorf("user_id: got %d", got.UserID)
	}
	if got.Email != "dev@example.com" {
		t.Errorf("email: got %q", got.Email)
	}
	if got.Issuer != "ks-hub" {
		t.Errorf("issuer: got %q", got.Issuer)
	}
}

func TestDeveloperJWT_Expired(t *testing.T) {
	priv, _ := os.ReadFile("testdata/test_private.pem")
	pub, _ := os.ReadFile("testdata/test_public.pem")

	claims := DeveloperClaims{UserID: 1}
	token, _ := SignDeveloperJWT(claims, priv, -1*time.Hour)

	_, err := VerifyDeveloperJWT(token, pub)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyDeveloperJWT_MalformedTokens(t *testing.T) {
	pub, _ := os.ReadFile("testdata/test_public.pem")

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
			_, err := VerifyDeveloperJWT(c.token, pub)
			if err == nil {
				t.Errorf("expected error for malformed token %q", c.name)
			}
		})
	}
}
