package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
