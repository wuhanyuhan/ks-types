# Attestation JWT Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 ks-types 仓新增 `AttestationClaims` + `SignAttestation` / `VerifyAttestation`，为 ks-admin → ks-client 信任链提供 Ed25519 签发与验证基础设施。

**Architecture:** 复用既有 `jwt.go` 里的 `ParseEd25519PrivateKeyPEM` / `ParseEd25519PublicKeyPEM`、既有 `testdata/test_*.pem` fixture、既有 `golang-jwt/v5` 库。新增的 `attestation_claims.go` 与现有 `instance_claims.go` 风格一致——签发函数固定 `iss=ks-admin` / `aud=["ks-client"]` / `typ=ATT+JWT` 这些不变量；验签函数同步校验。新增 `kid` 作为强制 header 字段（`SignAttestation` 必填参数），便于后续 client 端做"内置 kid"比对。

**Tech Stack:** Go 1.26.1, golang-jwt/v5 v5.3.1（已有依赖）, crypto/ed25519（标准库）

**Spec:** `~/projects/yuhan/ks-client/docs/superpowers/specs/2026-04-27-mdns-attestation-trust-anchor-design.md` §4 + §5.3

**Out of scope（其它仓 plan 处理）：**
- ks-admin 端的 `/v1/license/activate` 响应扩展、`/v1/instances/renew-attestation` 端点、`t_instance` schema 变更
- keystone 端的 well-known 字段扩展、启动闸门、续签 goroutine
- ks-client 端的内置信任根常量、AttestationVerifier、DiscoveryService 改造

---

## File Structure

| 文件 | 动作 | 职责 |
|------|------|------|
| `attestation_claims.go` | 新建 | `AttestationClaims` struct + `SignAttestation` + `VerifyAttestation` |
| `attestation_claims_test.go` | 新建 | 单测：合法 sign/verify、过期、签名错、kid mismatch、typ/aud/att_ver 篡改、malformed |
| `README.md` | 修改 | 在"特性"段落补充 Attestation JWT 说明 |
| `CHANGELOG.md` | 修改 | 新增一条 `feat: AttestationClaims + SignAttestation / VerifyAttestation` |

**复用（不新建）：**
- `jwt.go` 的 `ParseEd25519PrivateKeyPEM` / `ParseEd25519PublicKeyPEM`
- `testdata/test_private.pem` + `testdata/test_public.pem`（已存在）
- `instance_claims_test.go` 里的 `loadTestKeys` helper（attestation 测试在同包内，可直接调用）

---

## API Contract（实现前先锁定）

```go
// AttestationClaims 是 ks-admin 签发给 ks-client 的实例身份证明 JWT 的 Claims
//
// JSON 字段名与 spec §4.3 一致；att_ver 当前固定为 1
type AttestationClaims struct {
    jwt.RegisteredClaims  // iss="ks-admin", aud=["ks-client"], sub=instance_id, iat, exp
    InstanceID    string `json:"instance_id"`
    E2EPublicKey  string `json:"e2e_public_key"`
    OrgName       string `json:"org_name"`
    InstanceName  string `json:"instance_name"`
    AttVer        int    `json:"att_ver"`
}

// SignAttestation 用 Ed25519 私钥签发 Attestation JWT
//
// kid 是签发方密钥 ID（写入 JWT header），ks-client 端比对内置常量用；不能为空
// ttl 决定 ExpiresAt（当前时间 + ttl）
//
// 函数会强制覆写 RegisteredClaims 的 Issuer/Audience/Subject/IssuedAt/ExpiresAt
// 与 AttVer，调用方传入的这些字段会被忽略
func SignAttestation(claims AttestationClaims, privatePEM []byte, kid string, ttl time.Duration) (string, error)

// VerifyAttestation 用 Ed25519 公钥验证 Attestation JWT
//
// expectedKid 非空时，校验 JWT header.kid 必须与之相等；为空时跳过 kid 校验
// 校验项（任一失败返回 error）：
//   - 签名 (Ed25519) 有效
//   - header.alg == "EdDSA"
//   - header.typ == "ATT+JWT"
//   - header.kid == expectedKid（仅当 expectedKid 非空）
//   - claims.Issuer == "ks-admin"
//   - claims.Audience 包含 "ks-client"
//   - claims.AttVer == 1
//   - claims.ExpiresAt > now（jwt 库自动校验）
func VerifyAttestation(tokenString string, publicPEM []byte, expectedKid string) (*AttestationClaims, error)
```

---

## Task 1: 定义 `AttestationClaims` struct + JSON tag 单测

**Files:**
- Create: `attestation_claims.go`
- Create: `attestation_claims_test.go`

- [ ] **Step 1: Write the failing test**

写到 `attestation_claims_test.go`：

```go
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
		`"aud":["ks-client"]`, // jwt.ClaimStrings 始终序列化为 JSON 数组（jwt/v5 默认 MarshalSingleStringAsArray=true）
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in JSON: %s", want, s)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestAttestationClaims_JSONFieldNames -v
```
Expected: FAIL with `undefined: AttestationClaims`

- [ ] **Step 3: Write minimal implementation**

写到 `attestation_claims.go`：

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestAttestationClaims_JSONFieldNames -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims.go attestation_claims_test.go && git commit -m "feat(attestation): 新增 AttestationClaims struct + JSON tag 单测"
```

---

## Task 2: 实现 `SignAttestation` 基础签发

**Files:**
- Modify: `attestation_claims.go`
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the failing test**

追加到 `attestation_claims_test.go`：

```go
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
```

记得在文件顶部 import 加 `"encoding/base64"`。

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestSignAttestation_StructureAndHeaders -v
```
Expected: FAIL with `undefined: SignAttestation`

- [ ] **Step 3: Write minimal implementation**

追加到 `attestation_claims.go`：

```go
import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestSignAttestation_StructureAndHeaders -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims.go attestation_claims_test.go && git commit -m "feat(attestation): 实现 SignAttestation 基础签发（typ/kid/aud=ks-client）"
```

---

## Task 3: 实现 `VerifyAttestation` 主路径（合法 round-trip）

**Files:**
- Modify: `attestation_claims.go`
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the failing test**

追加到 `attestation_claims_test.go`：

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestVerifyAttestation_RoundTrip -v
```
Expected: FAIL with `undefined: VerifyAttestation`

- [ ] **Step 3: Write minimal implementation**

追加到 `attestation_claims.go`：

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestVerifyAttestation_RoundTrip|TestVerifyAttestation_EmptyExpectedKidSkipsKidCheck" -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims.go attestation_claims_test.go && git commit -m "feat(attestation): 实现 VerifyAttestation 主路径 + InstanceID 回填"
```

---

## Task 4: 失败路径——过期、签名错、malformed

**Files:**
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the failing test**

追加：

```go
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
```

需要确认 import 包含 `crypto/ed25519`（`marshalPublicKeyPEM` helper 来自 `instance_claims_test.go`，同包内可直接调用）。

- [ ] **Step 2: Run tests to verify they pass directly**

由于这些都是合法 sign 但失败场景（过期 / 错公钥 / malformed），实现已经能 work（jwt 库自动处理过期与签名校验）。直接跑：

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestVerifyAttestation_Expired|TestVerifyAttestation_InvalidSignature|TestVerifyAttestation_MalformedTokens" -v
```
Expected: PASS（如果某个 sub-case fail，再看是否需要细化错误处理）

- [ ] **Step 3: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims_test.go && git commit -m "test(attestation): 覆盖 VerifyAttestation 过期/签名错/malformed 失败路径"
```

---

## Task 5: 失败路径——`kid mismatch`

**Files:**
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the test**

追加：

```go
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
```

- [ ] **Step 2: Run test to verify behavior**

Task 3 的实现已经包含 kid 校验逻辑，应该直接 PASS：
```bash
cd ~/projects/yuhan/ks-types && go test -run TestVerifyAttestation_KidMismatch -v
```
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims_test.go && git commit -m "test(attestation): 覆盖 VerifyAttestation kid mismatch 场景"
```

---

## Task 6: 失败路径——篡改 `typ` header

**Files:**
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the failing test（构造 typ != ATT+JWT 的合法签名 token）**

追加：

```go
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
```

- [ ] **Step 2: Run test to verify it passes**

Task 3 的实现里已经校验 `typ != "ATT+JWT"` 时报错，应该直接 PASS：
```bash
cd ~/projects/yuhan/ks-types && go test -run TestVerifyAttestation_TypMismatch -v
```
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims_test.go && git commit -m "test(attestation): 覆盖 typ 篡改场景 + 加 signCraftedAttestation 测试 helper"
```

---

## Task 7: 失败路径——篡改 `aud`

**Files:**
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the test**

追加：

```go
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
```

- [ ] **Step 2: Run test**

```bash
cd ~/projects/yuhan/ks-types && go test -run TestVerifyAttestation_AudMismatch -v
```
Expected: PASS（Task 3 的 `claimsAudienceContains` 处理了这点）

- [ ] **Step 3: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims_test.go && git commit -m "test(attestation): 覆盖 aud 篡改场景"
```

---

## Task 8: 失败路径——篡改 `att_ver` + `iss`

**Files:**
- Modify: `attestation_claims_test.go`

- [ ] **Step 1: Write the test**

追加：

```go
func TestVerifyAttestation_AttVerUnsupported(t *testing.T) {
	priv, pub := loadTestKeys(t)

	cases := []struct {
		name   string
		attVer any
	}{
		{"att_ver=2 (未来版本)", 2},
		{"att_ver=0 (零值)", 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token := signCraftedAttestation(t, priv, nil, map[string]any{"att_ver": c.attVer})
			if _, err := VerifyAttestation(token, pub, "ks-admin-2026"); err == nil {
				t.Errorf("expected error for %s", c.name)
			}
		})
	}
}

func TestVerifyAttestation_IssMismatch(t *testing.T) {
	priv, pub := loadTestKeys(t)

	token := signCraftedAttestation(t, priv, nil, map[string]any{"iss": "imposter"})
	if _, err := VerifyAttestation(token, pub, "ks-admin-2026"); err == nil {
		t.Error("expected error for iss mismatch")
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestVerifyAttestation_AttVerUnsupported|TestVerifyAttestation_IssMismatch" -v
```
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add attestation_claims_test.go && git commit -m "test(attestation): 覆盖 att_ver / iss 篡改场景"
```

---

## Task 9: 全量回归 + README + CHANGELOG

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: 跑全量测试，确认无回归**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v
```
Expected: 所有测试 PASS（包括既有 InstanceJWT、Manifest、Permissions 等）

- [ ] **Step 2: 更新 README.md**

在"特性"段落后追加一条（参照已有"Ed25519 JWT"那段的风格）：

```markdown
- **Attestation JWT (ATT+JWT)**：ks-admin 签发给 ks-client 的实例身份证明，独立于 Instance JWT，专供局域网发现场景做"实例合法性校验"。`SignAttestation` / `VerifyAttestation` 使用与 Instance JWT 同一对 Ed25519 密钥，但 `aud` 锁死为 `"ks-client"`、`typ` 为 `"ATT+JWT"`、强制 `kid` header，与 Instance JWT 不可互换误用。
```

- [ ] **Step 3: 更新 CHANGELOG.md**

在最顶部新增一条（参照已有 entry 风格）：

```markdown
## [Unreleased]

### Added
- `AttestationClaims` struct + `SignAttestation` / `VerifyAttestation`：ks-admin 签发给 ks-client 的实例身份证明 JWT，typ=ATT+JWT，aud=ks-client，强制 kid header。仅用于 ks-client 端验证 LAN 内发现的 keystone 实例是否由 ks-admin 合法签发。
```

（如果 CHANGELOG 已经有 `[Unreleased]` section，把上面的 `### Added` 项追加进去。）

- [ ] **Step 4: Commit**

```bash
cd ~/projects/yuhan/ks-types && git add README.md CHANGELOG.md && git commit -m "docs(attestation): README 与 CHANGELOG 同步 AttestationClaims 新增能力"
```

---

## 完成判定

- [ ] `attestation_claims.go` 与 `attestation_claims_test.go` 文件存在
- [ ] `go test ./...` 全部 PASS（包含新增 12 个 test 函数 + 10 个子用例）
- [ ] `go vet ./...` 无 warning
- [ ] README "特性" 段落与 CHANGELOG `[Unreleased]` 都已记录新能力
- [ ] git log 显示 9 次新增 commit（task 1-9 各一次）

完成后，下一步是 **ks-admin 仓的 implementation plan**——参照 spec §5（签发端点 / `t_instance` schema 变更 / 后台 UI）。
