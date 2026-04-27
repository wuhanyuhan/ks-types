package kstypes

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignAttestation_StructureAndHeaders(t *testing.T) {
	priv, _ := loadTestKeys(t)

	claims := AttestationClaims{
		InstanceID:   "inst_signtest",
		E2EPublicKey: "test-pubkey",
		OrgName:      "测试公司",
		InstanceName: "测试实例",
	}

	token, err := SignAttestation(claims, priv, "ks-admin-2026", 1*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	// JWT 标准三段
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 segments, got %d (%s)", len(parts), token)
	}

	// 解析 header 校验 alg / typ / kid
	header := decodeJSONSegment(t, parts[0])
	if header["alg"] != "EdDSA" {
		t.Errorf("alg: got %v, want EdDSA", header["alg"])
	}
	if header["typ"] != "ATT+JWT" {
		t.Errorf("typ: got %v, want ATT+JWT", header["typ"])
	}
	if header["kid"] != "ks-admin-2026" {
		t.Errorf("kid: got %v, want ks-admin-2026", header["kid"])
	}
}

// decodeJSONSegment base64url decode + JSON unmarshal
func decodeJSONSegment(t *testing.T, segment string) map[string]any {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	return m
}

func TestVerifyAttestation_RoundTrip(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := AttestationClaims{
		InstanceID:   "inst_round",
		E2EPublicKey: "round-pubkey",
		OrgName:      "宇寒",
		InstanceName: "总部",
	}
	token, err := SignAttestation(claims, priv, "ks-admin-2026", 1*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := VerifyAttestation(token, pub, "ks-admin-2026")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if got.InstanceID != "inst_round" {
		t.Errorf("instance_id: got %q", got.InstanceID)
	}
	if got.E2EPublicKey != "round-pubkey" {
		t.Errorf("e2e_public_key: got %q", got.E2EPublicKey)
	}
	if got.OrgName != "宇寒" {
		t.Errorf("org_name: got %q", got.OrgName)
	}
	if got.InstanceName != "总部" {
		t.Errorf("instance_name: got %q", got.InstanceName)
	}
	if got.AttVer != 1 {
		t.Errorf("att_ver: got %d, want 1", got.AttVer)
	}
	if got.Issuer != "ks-admin" {
		t.Errorf("iss: got %q", got.Issuer)
	}
	if len(got.Audience) != 1 || got.Audience[0] != "ks-client" {
		t.Errorf("aud: got %v", got.Audience)
	}
	if got.Subject != "inst_round" {
		t.Errorf("sub: got %q", got.Subject)
	}
}

func TestVerifyAttestation_EmptyExpectedKidSkipsKidCheck(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := AttestationClaims{InstanceID: "inst_nokid"}
	token, err := SignAttestation(claims, priv, "any-kid", 1*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// expectedKid 为空字符串时跳过 kid 校验
	got, err := VerifyAttestation(token, pub, "")
	if err != nil {
		t.Fatalf("verify with empty kid: %v", err)
	}
	if got.InstanceID != "inst_nokid" {
		t.Errorf("instance_id: got %q", got.InstanceID)
	}
}

func TestAttestationClaims_JSONFieldNames(t *testing.T) {
	claims := AttestationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "inst_abc",
			Issuer:    "ks-admin",
			Audience:  jwt.ClaimStrings{"ks-client"},
			IssuedAt:  jwt.NewNumericDate(time.Unix(1714000000, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(1721776000, 0)),
		},
		InstanceID:   "inst_abc",
		E2EPublicKey: "AAAAB3NzaC1yc2E=",
		OrgName:      "宇寒科技",
		InstanceName: "武汉总部",
		AttVer:       1,
	}

	b, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)

	for _, want := range []string{
		`"instance_id":"inst_abc"`,
		`"e2e_public_key":"AAAAB3NzaC1yc2E="`,
		`"org_name":"宇寒科技"`,
		`"instance_name":"武汉总部"`,
		`"att_ver":1`,
		`"sub":"inst_abc"`,
		`"iss":"ks-admin"`,
		`"aud":["ks-client"]`, // jwt.ClaimStrings 始终序列化为 JSON 数组
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in JSON: %s", want, s)
		}
	}
}

func TestVerifyAttestation_Expired(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := AttestationClaims{InstanceID: "inst_expired"}
	token, err := SignAttestation(claims, priv, "ks-admin-2026", -1*time.Hour) // 已经过期
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := VerifyAttestation(token, pub, "ks-admin-2026"); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyAttestation_InvalidSignature(t *testing.T) {
	priv, _ := loadTestKeys(t)

	claims := AttestationClaims{InstanceID: "inst_badsig"}
	token, err := SignAttestation(claims, priv, "ks-admin-2026", 1*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	otherPub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	otherPubPEM := marshalPublicKeyPEM(otherPub)

	if _, err := VerifyAttestation(token, otherPubPEM, "ks-admin-2026"); err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestVerifyAttestation_MalformedTokens(t *testing.T) {
	_, pub := loadTestKeys(t)

	cases := []struct {
		name  string
		token string
	}{
		{"空字符串", ""},
		{"随机垃圾", "not-a-jwt-at-all"},
		{"只有 header", "eyJhbGciOiJFZERTQSIsInR5cCI6IkFUVCtKV1QifQ"},
		{"两段（缺 signature）", "eyJhbGciOiJFZERTQSIsInR5cCI6IkFUVCtKV1QifQ.eyJzdWIiOiJ0ZXN0In0"},
		{"三段但 signature 损坏", "eyJhbGciOiJFZERTQSIsInR5cCI6IkFUVCtKV1QifQ.eyJzdWIiOiJ0ZXN0In0.AAAA_invalid_sig"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := VerifyAttestation(c.token, pub, "ks-admin-2026"); err == nil {
				t.Errorf("expected error for malformed token %q", c.name)
			}
		})
	}
}

func TestVerifyAttestation_KidMismatch(t *testing.T) {
	priv, pub := loadTestKeys(t)

	claims := AttestationClaims{InstanceID: "inst_kid"}
	token, err := SignAttestation(claims, priv, "ks-admin-2026", 1*time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := VerifyAttestation(token, pub, "ks-admin-2099"); err == nil {
		t.Error("expected error for kid mismatch")
	}
}

// signCraftedAttestation 用 Ed25519 私钥手工构造一份 attestation token，
// 允许测试覆盖 SignAttestation 不会产生但 verify 必须挡住的篡改场景。
//
// headerOverrides / claimsOverrides 会原样合入 header / payload，nil 则不覆写。
func signCraftedAttestation(t *testing.T, privPEM []byte, headerOverrides, claimsOverrides map[string]any) string {
	t.Helper()
	priv, err := ParseEd25519PrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("parse priv: %v", err)
	}
	now := time.Now().UTC()

	claims := jwt.MapClaims{
		"iss":            "ks-admin",
		"sub":            "inst_crafted",
		"aud":            []string{"ks-client"},
		"iat":            now.Unix(),
		"exp":            now.Add(time.Hour).Unix(),
		"instance_id":    "inst_crafted",
		"e2e_public_key": "crafted-pubkey",
		"org_name":       "测试",
		"instance_name":  "篡改",
		"att_ver":        1,
	}
	for k, v := range claimsOverrides {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["typ"] = "ATT+JWT"
	token.Header["kid"] = "ks-admin-2026"
	for k, v := range headerOverrides {
		token.Header[k] = v
	}

	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign crafted: %v", err)
	}
	return signed
}

func TestVerifyAttestation_TypMismatch(t *testing.T) {
	priv, pub := loadTestKeys(t)

	token := signCraftedAttestation(t, priv, map[string]any{"typ": "JWT"}, nil)

	if _, err := VerifyAttestation(token, pub, "ks-admin-2026"); err == nil {
		t.Error("expected error for typ mismatch")
	}
}

func TestVerifyAttestation_AudMismatch(t *testing.T) {
	priv, pub := loadTestKeys(t)

	cases := []struct {
		name string
		aud  any
	}{
		{"aud=ks-relay", []string{"ks-relay"}},
		{"aud=ks-hub", []string{"ks-hub"}},
		{"aud 为空数组", []string{}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token := signCraftedAttestation(t, priv, nil, map[string]any{"aud": c.aud})
			if _, err := VerifyAttestation(token, pub, "ks-admin-2026"); err == nil {
				t.Errorf("expected error for %s", c.name)
			}
		})
	}
}
