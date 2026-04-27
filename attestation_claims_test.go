package kstypes

import (
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
