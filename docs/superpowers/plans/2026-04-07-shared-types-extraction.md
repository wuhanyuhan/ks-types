# 共享类型提取 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 消除 ks-types 与 keystone/ks-hub/ks-admin/ks-devkit 之间的类型重复定义，让 ks-types 成为 BizError、Result、枚举的唯一事实来源。

**Architecture:** 原地增强 ks-types 根包（`kstypes`），各消费方通过 type alias + re-export 模式接入，保持内部调用代码零改动。

**Tech Stack:** Go 1.26.1, gin, golang-jwt/v5, yaml.v3

**Spec:** `docs/superpowers/specs/2026-04-07-shared-types-extraction-design.md`

---

## File Structure

### ks-types（新建/修改）

| 文件 | 动作 | 职责 |
|------|------|------|
| `errors.go` | 修改 | BizError 加 json tags + Newf + Is |
| `errors_test.go` | 修改 | 补充 Newf/Is 测试 |
| `result.go` | 新建 | Result/OK/Fail/FailErr/PageResult/ListResult |
| `result_test.go` | 新建 | Result 系列测试 |
| `apptypes.go` | 修改 | ProtectionLevel 枚举 + RuntimeMode.Valid 空值支持 |
| `apptypes_test.go` | 修改 | ProtectionLevel 测试 + RuntimeMode 空值测试修复 |
| `manifest.go` | 修改 | Protection→ProtectionLevel, Mode→RuntimeMode, Validate 简化 |
| `manifest_test.go` | 修改 | 适配字段类型变更 |

### ks-hub（修改）

| 文件 | 动作 | 职责 |
|------|------|------|
| `internal/core/result.go` | 重写 | 改为 re-export kstypes |
| `internal/core/errors.go` | 修改 | 补充 Newf/FailErr re-export |

### ks-admin（修改）

| 文件 | 动作 | 职责 |
|------|------|------|
| `internal/core/errors.go` | 重写 | 改为 alias + re-export |
| `internal/core/result.go` | 重写 | 改为 re-export kstypes |

### keystone（修改）

| 文件 | 动作 | 职责 |
|------|------|------|
| `internal/core/errors/errors.go` | 重写 | 改为 alias + re-export + 保留专有码 |
| `internal/marketplace/manifest.go` | 修改 | 删除本地 ProtectionLevel/RuntimeMode 常量 |

---

### Task 1: ks-types — BizError 增强

**Files:**
- Modify: `~/projects/yuhan/ks-types/errors.go`
- Modify: `~/projects/yuhan/ks-types/errors_test.go`

- [ ] **Step 1: 写 Newf 和 Is 的失败测试**

在 `errors_test.go` 末尾追加：

```go
func TestNewf(t *testing.T) {
	err := Newf(40001, "字段 %s 不能为空", "name")
	if err.Code != 40001 {
		t.Errorf("code: got %d", err.Code)
	}
	if err.Message != "字段 name 不能为空" {
		t.Errorf("message: got %q", err.Message)
	}
	if err.Error() != "[40001] 字段 name 不能为空" {
		t.Errorf("Error(): got %q", err.Error())
	}
}

func TestBizErrorIs(t *testing.T) {
	a := &BizError{Code: 40001, Message: "参数校验失败"}
	b := &BizError{Code: 40001, Message: "不同的消息但同码"}
	c := &BizError{Code: 40401, Message: "资源不存在"}

	if !a.Is(b) {
		t.Error("同码 BizError 应匹配")
	}
	if a.Is(c) {
		t.Error("异码 BizError 不应匹配")
	}

	// 非 BizError 类型
	if a.Is(fmt.Errorf("普通错误")) {
		t.Error("非 BizError 不应匹配")
	}
}

func TestBizErrorJSON(t *testing.T) {
	err := &BizError{Code: 40001, Message: "参数校验失败"}
	data, _ := json.Marshal(err)
	got := string(data)
	want := `{"code":40001,"message":"参数校验失败"}`
	if got != want {
		t.Errorf("json: got %s, want %s", got, want)
	}
}
```

同时在文件顶部 import 块补充 `"encoding/json"` 和 `"fmt"`。

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestNewf|TestBizErrorIs|TestBizErrorJSON" -v
```

预期：编译失败，`Newf` 未定义。

- [ ] **Step 3: 实现 BizError 增强**

修改 `errors.go`：

1. BizError struct 加 json tags：
```go
type BizError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
```

2. 在 `HTTPStatus()` 方法后追加：
```go
// Newf 构造带格式化 message 的 BizError
func Newf(code int, format string, args ...any) *BizError {
	return &BizError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Is 判断当前错误是否与目标 BizError 同码（满足 errors.Is 协议）
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	return ok && e.Code == t.Code
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestNewf|TestBizErrorIs|TestBizErrorJSON" -v
```

预期：3 个测试全部 PASS。

- [ ] **Step 5: 运行全量测试，确认无回归**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v
```

预期：全部 PASS。

- [ ] **Step 6: 提交**

```bash
cd ~/projects/yuhan/ks-types && git add errors.go errors_test.go && git commit -m "feat: BizError 增强——json tags + Newf + Is"
```

---

### Task 2: ks-types — Result 类型

**Files:**
- Create: `~/projects/yuhan/ks-types/result.go`
- Create: `~/projects/yuhan/ks-types/result_test.go`

- [ ] **Step 1: 写 Result 测试**

创建 `result_test.go`：

```go
package kstypes

import (
	"encoding/json"
	"testing"
)

func TestOK(t *testing.T) {
	r := OK(map[string]string{"key": "value"})
	if r.Code != 0 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "ok" {
		t.Errorf("message: got %q", r.Message)
	}
}

func TestOK_JSON(t *testing.T) {
	r := OK(nil)
	data, _ := json.Marshal(r)
	got := string(data)
	want := `{"code":0,"message":"ok","data":null}`
	if got != want {
		t.Errorf("json: got %s", got)
	}
}

func TestFail(t *testing.T) {
	r := Fail(40001, "参数校验失败")
	if r.Code != 40001 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "参数校验失败" {
		t.Errorf("message: got %q", r.Message)
	}
	if r.Data != nil {
		t.Errorf("data: got %v, want nil", r.Data)
	}
}

func TestFailErr(t *testing.T) {
	err := &BizError{Code: 40401, Message: "资源不存在"}
	r := FailErr(err)
	if r.Code != 40401 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "资源不存在" {
		t.Errorf("message: got %q", r.Message)
	}
}

func TestPageResult_JSON(t *testing.T) {
	pr := PageResult[string]{
		Items: []string{"a", "b"},
		Total: 10,
		Page:  1,
		Size:  2,
	}
	data, _ := json.Marshal(pr)
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["total"].(float64) != 10 {
		t.Errorf("total: got %v", m["total"])
	}
	if m["page"].(float64) != 1 {
		t.Errorf("page: got %v", m["page"])
	}
}

func TestListResult_JSON(t *testing.T) {
	lr := ListResult[int]{Items: []int{1, 2, 3}}
	data, _ := json.Marshal(lr)
	var m map[string]any
	json.Unmarshal(data, &m)
	items := m["items"].([]any)
	if len(items) != 3 {
		t.Errorf("items len: got %d", len(items))
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestOK|TestFail|TestPageResult|TestListResult" -v
```

预期：编译失败，`OK`/`Fail`/`FailErr`/`PageResult`/`ListResult` 未定义。

- [ ] **Step 3: 创建 result.go**

```go
package kstypes

// Result 统一 HTTP JSON 响应结构
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 成功响应
func OK(data any) Result {
	return Result{Code: 0, Message: "ok", Data: data}
}

// Fail 失败响应
func Fail(code int, message string) Result {
	return Result{Code: code, Message: message}
}

// FailErr 从 BizError 构建失败响应
func FailErr(err *BizError) Result {
	return Result{Code: err.Code, Message: err.Message}
}

// PageResult 分页列表响应
type PageResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

// ListResult 非分页列表响应
type ListResult[T any] struct {
	Items []T `json:"items"`
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestOK|TestFail|TestPageResult|TestListResult" -v
```

预期：6 个测试全部 PASS。

- [ ] **Step 5: 运行全量测试**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v
```

预期：全部 PASS。

- [ ] **Step 6: 提交**

```bash
cd ~/projects/yuhan/ks-types && git add result.go result_test.go && git commit -m "feat: 新增 Result/PageResult/ListResult 统一响应结构"
```

---

### Task 3: ks-types — ProtectionLevel 枚举 + RuntimeMode.Valid 调整

**Files:**
- Modify: `~/projects/yuhan/ks-types/apptypes.go`
- Modify: `~/projects/yuhan/ks-types/apptypes_test.go`

- [ ] **Step 1: 写 ProtectionLevel 测试**

在 `apptypes_test.go` 末尾追加：

```go
func TestProtectionLevelValid(t *testing.T) {
	valid := []ProtectionLevel{
		"", ProtectionNone, ProtectionPreinstalled, ProtectionProtected, ProtectionSystem,
	}
	for _, p := range valid {
		if !p.Valid() {
			t.Errorf("expected %q to be valid", p)
		}
	}
}

func TestProtectionLevelInvalid(t *testing.T) {
	invalid := ProtectionLevel("unknown")
	if invalid.Valid() {
		t.Error("expected unknown ProtectionLevel to be invalid")
	}
}

func TestRuntimeModeEmptyIsValid(t *testing.T) {
	empty := RuntimeMode("")
	if !empty.Valid() {
		t.Error("expected empty RuntimeMode to be valid")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestProtectionLevel|TestRuntimeModeEmpty" -v
```

预期：编译失败，`ProtectionLevel` 未定义。

- [ ] **Step 3: 实现 ProtectionLevel + 调整 RuntimeMode.Valid**

在 `apptypes.go` 末尾追加：

```go
// ProtectionLevel 保护级别
type ProtectionLevel string

const (
	ProtectionNone         ProtectionLevel = "none"
	ProtectionPreinstalled ProtectionLevel = "preinstalled"
	ProtectionProtected    ProtectionLevel = "protected"
	ProtectionSystem       ProtectionLevel = "system"
)

var validProtectionLevels = map[ProtectionLevel]bool{
	ProtectionNone: true, ProtectionPreinstalled: true,
	ProtectionProtected: true, ProtectionSystem: true,
}

// Valid 检查 ProtectionLevel 是否合法；空值视为合法（等同 none）
func (p ProtectionLevel) Valid() bool {
	return p == "" || validProtectionLevels[p]
}
```

修改 `RuntimeMode.Valid()`：

```go
// Valid 检查 RuntimeMode 是否合法；空值视为合法（等同 none）
func (m RuntimeMode) Valid() bool { return m == "" || validRuntimeModes[m] }
```

- [ ] **Step 4: 运行新测试，确认通过**

```bash
cd ~/projects/yuhan/ks-types && go test -run "TestProtectionLevel|TestRuntimeModeEmpty" -v
```

预期：3 个测试全部 PASS。

- [ ] **Step 5: 修复 RuntimeMode 空值测试**

`manifest_test.go` 中 `TestRuntimeMode_Valid` 的空值用例需要从 `false` 改为 `true`：

```go
// 原来：
{RuntimeMode(""), false},
// 改为：
{RuntimeMode(""), true},
```

- [ ] **Step 6: 运行全量测试**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v
```

预期：全部 PASS。

- [ ] **Step 7: 提交**

```bash
cd ~/projects/yuhan/ks-types && git add apptypes.go apptypes_test.go manifest_test.go && git commit -m "feat: 新增 ProtectionLevel 枚举 + RuntimeMode.Valid 支持空值"
```

---

### Task 4: ks-types — manifest.go 字段类型化 + 测试适配

**Files:**
- Modify: `~/projects/yuhan/ks-types/manifest.go`
- Modify: `~/projects/yuhan/ks-types/manifest_test.go`
- Modify: `~/projects/yuhan/ks-types/testdata/valid_manifest.yaml`（可能需要加 protection 字段）

- [ ] **Step 1: 修改 manifest.go 字段类型**

1. `ManifestSpec.Protection` 从 `string` 改为 `ProtectionLevel`：

```go
Protection    ProtectionLevel           `yaml:"protection,omitempty" json:"protection,omitempty"`
```

2. `RuntimeSpec.Mode` 从 `string` 改为 `RuntimeMode`：

```go
Mode           RuntimeMode   `yaml:"mode,omitempty" json:"mode,omitempty"`
```

3. `Validate()` 中 Protection 校验简化——替换内联 map 为方法调用：

```go
// 原来：
validProtections := map[string]bool{
    "": true, "none": true, "preinstalled": true, "protected": true, "system": true,
}
if !validProtections[m.Protection] {
    return fmt.Errorf("manifest: invalid protection %q", m.Protection)
}

// 改为：
if !m.Protection.Valid() {
    return fmt.Errorf("manifest: invalid protection %q", m.Protection)
}
```

4. `Validate()` 中 RuntimeMode 校验简化：

```go
// 原来：
if m.Runtime.Mode != "" && !RuntimeMode(m.Runtime.Mode).Valid() {
    return fmt.Errorf("manifest: invalid runtime mode %q", m.Runtime.Mode)
}

// 改为：
if !m.Runtime.Mode.Valid() {
    return fmt.Errorf("manifest: invalid runtime mode %q", m.Runtime.Mode)
}
```

- [ ] **Step 2: 适配 manifest_test.go**

1. `TestParseManifest_Valid` 中 Runtime.Mode 比较改用常量：

```go
// 原来：
if m.Runtime.Mode != "container" {
// 改为：
if m.Runtime.Mode != RuntimeModeContainer {
```

2. `TestValidateManifest_InvalidRuntimeMode` 中 Mode 赋值改为 RuntimeMode 类型：

```go
// 原来：
Runtime: RuntimeSpec{Mode: "invalid"},
// 改为：
Runtime: RuntimeSpec{Mode: RuntimeMode("invalid")},
```

3. `TestValidateManifest_ValidRuntimeMode` 中 `[]string` 改为 `[]RuntimeMode`：

```go
// 原来：
for _, mode := range []string{"", "none", "process", "container"} {
    m := &ManifestSpec{
        ID: "test", Name: "test", Version: "1.0.0",
        Type:    AppTypeService,
        Runtime: RuntimeSpec{Mode: mode},
    }

// 改为：
for _, mode := range []RuntimeMode{"", RuntimeModeNone, RuntimeModeProcess, RuntimeModeContainer} {
    m := &ManifestSpec{
        ID: "test", Name: "test", Version: "1.0.0",
        Type:    AppTypeService,
        Runtime: RuntimeSpec{Mode: mode},
    }
```

4. `TestValidateManifest_InvalidProtection` 中 Protection 赋值改为 ProtectionLevel：

```go
// 原来：
Protection: "invalid",
// 改为：
Protection: ProtectionLevel("invalid"),
```

5. `TestValidateManifest_ValidProtection` 中 `[]string` 改为 `[]ProtectionLevel`：

```go
// 原来：
for _, p := range []string{"", "none", "preinstalled", "protected", "system"} {
    m := &ManifestSpec{
        Protection: p,
    }

// 改为：
for _, p := range []ProtectionLevel{"", ProtectionNone, ProtectionPreinstalled, ProtectionProtected, ProtectionSystem} {
    m := &ManifestSpec{
        Protection: p,
    }
```

6. `TestParseRuntimeSpec_ProcessMode` 和 `TestParseRuntimeSpec_ContainerMode` 中 Mode 比较改用常量：

```go
// ProcessMode:
// 原来：
if w.Runtime.Mode != "process" {
// 改为：
if w.Runtime.Mode != RuntimeModeProcess {

// ContainerMode:
// 原来：
if w.Runtime.Mode != "container" {
// 改为：
if w.Runtime.Mode != RuntimeModeContainer {
```

7. `TestManifestSpec_RoundTrip` 中 Mode 赋值和比较：

```go
// 赋值：
// 原来：
Mode: "container",
// 改为：
Mode: RuntimeModeContainer,

// 比较：
// 原来：
if parsed.Runtime.Mode != "container" {
// 改为：
if parsed.Runtime.Mode != RuntimeModeContainer {
```

- [ ] **Step 3: 运行全量测试**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v
```

预期：全部 PASS。

- [ ] **Step 4: 提交**

```bash
cd ~/projects/yuhan/ks-types && git add manifest.go manifest_test.go && git commit -m "feat: ManifestSpec.Protection 和 RuntimeSpec.Mode 类型化"
```

---

### Task 5: ks-hub — Result + Errors 迁移

**Files:**
- Modify: `~/projects/yuhan/ks-hub/internal/core/result.go`
- Modify: `~/projects/yuhan/ks-hub/internal/core/errors.go`

- [ ] **Step 1: 重写 result.go 为 re-export**

将 `~/projects/yuhan/ks-hub/internal/core/result.go` 完整替换为：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

// Result 统一 HTTP JSON 响应结构（re-export from ks-types）
type Result = kstypes.Result

// PageResult 分页列表响应
type PageResult[T any] = kstypes.PageResult[T]

// ListResult 非分页列表响应
type ListResult[T any] = kstypes.ListResult[T]

var (
	// OK 成功响应
	OK = kstypes.OK
	// Fail 失败响应
	Fail = kstypes.Fail
	// FailErr 从 BizError 构建失败响应
	FailErr = kstypes.FailErr
)
```

- [ ] **Step 2: errors.go 补充新符号 re-export**

在 `~/projects/yuhan/ks-hub/internal/core/errors.go` 的现有 re-export 块中追加：

```go
var (
	// Newf 构造带格式化 message 的 BizError
	Newf = kstypes.Newf
)
```

- [ ] **Step 3: 编译验证**

```bash
cd ~/projects/yuhan/ks-hub && go build ./...
```

预期：编译成功。如果有编译错误，根据错误信息调整（可能需要在引用 `core.FailErr` 的地方确认导入）。

- [ ] **Step 4: 运行测试**

```bash
cd ~/projects/yuhan/ks-hub && go test ./...
```

预期：全部 PASS。

- [ ] **Step 5: 提交**

```bash
cd ~/projects/yuhan/ks-hub && git add internal/core/result.go internal/core/errors.go && git commit -m "refactor: Result 和 BizError 迁移到 ks-types re-export"
```

---

### Task 6: ks-admin — Errors + Result 迁移

**Files:**
- Modify: `~/projects/yuhan/ks-admin/internal/core/errors.go`
- Modify: `~/projects/yuhan/ks-admin/internal/core/result.go`
- Modify: 可能需要修改引用 `NewBizError`/`NewBizErrorPair` 的文件

- [ ] **Step 1: 搜索 NewBizError/NewBizErrorPair 的使用**

```bash
cd ~/projects/yuhan/ks-admin && grep -rn "NewBizError\|NewBizErrorPair" --include="*.go" .
```

记录所有调用位置，后续替换。

- [ ] **Step 2: 重写 errors.go**

将 `~/projects/yuhan/ks-admin/internal/core/errors.go` 完整替换为：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

// BizError 业务错误（re-export from ks-types）
type BizError = kstypes.BizError

// Newf 构造带格式化 message 的 BizError
var Newf = kstypes.Newf

// 通用错误码（re-export from ks-types）
var (
	ErrInvalidParams   = kstypes.ErrInvalidParams
	ErrNotFound        = kstypes.ErrNotFound
	ErrDuplicate       = kstypes.ErrDuplicate
	ErrTokenExpired    = kstypes.ErrTokenExpired
	ErrTokenInvalid    = kstypes.ErrTokenInvalid
	ErrForbidden       = kstypes.ErrForbidden
	ErrInstanceInvalid = kstypes.ErrInstanceInvalid
	ErrInstanceRevoked = kstypes.ErrInstanceRevoked
	ErrInternalServer  = kstypes.ErrInternalServer
)

// ks-admin 专有错误码
var (
	ErrSSOUnavailable = &BizError{Code: 50201, Message: "SSO 服务不可用"}
)
```

- [ ] **Step 3: 替换 NewBizError/NewBizErrorPair 调用**

根据 Step 1 的搜索结果，将所有调用替换：

- `NewBizError(code, msg)` → `&BizError{Code: code, Message: msg}`
- `NewBizErrorPair(code, msg)` → `&BizError{Code: code, Message: msg}`（如果返回值不同，按实际签名调整）

- [ ] **Step 4: 重写 result.go 为 re-export**

将 `~/projects/yuhan/ks-admin/internal/core/result.go` 完整替换为：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

// Result 统一 HTTP JSON 响应结构（re-export from ks-types）
type Result = kstypes.Result

// PageResult 分页列表响应
type PageResult[T any] = kstypes.PageResult[T]

// ListResult 非分页列表响应
type ListResult[T any] = kstypes.ListResult[T]

var (
	// OK 成功响应
	OK = kstypes.OK
	// Fail 失败响应
	Fail = kstypes.Fail
	// FailErr 从 BizError 构建失败响应
	FailErr = kstypes.FailErr
)
```

- [ ] **Step 5: 编译验证**

```bash
cd ~/projects/yuhan/ks-admin && go build ./...
```

预期：编译成功。如有编译错误（如删除了 `Error()` 方法的引用），根据错误信息逐一修复——`Error()` 现在来自 ks-types 的 BizError，应当自动可用。

- [ ] **Step 6: 运行测试**

```bash
cd ~/projects/yuhan/ks-admin && go test ./...
```

预期：全部 PASS。

- [ ] **Step 7: 提交**

```bash
cd ~/projects/yuhan/ks-admin && git add -A && git commit -m "refactor: BizError 和 Result 迁移到 ks-types re-export"
```

---

### Task 7: keystone — errors.go 迁移

**Files:**
- Modify: `~/projects/yuhan/keystone/internal/core/errors/errors.go`
- Modify: 可能需要修改引用 `errors.New()` 的文件

- [ ] **Step 1: 搜索 errors.New() 的使用**

```bash
cd ~/projects/yuhan/keystone && grep -rn "errors\.New(" --include="*.go" . | grep -v "vendor\|_test.go" | grep "core/errors"
```

注意区分标准库 `errors.New(msg)` 和本地 `errors.New(code, msg)`（两个参数）。记录所有本地两参数调用位置。

- [ ] **Step 2: 重写 errors.go 头部**

将 `~/projects/yuhan/keystone/internal/core/errors/errors.go` 的头部（package 声明到错误码 var 块之前）替换为：

```go
package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	kstypes "github.com/wuhanyuhan/ks-types"
)

// BizError 业务错误（re-export from ks-types）
type BizError = kstypes.BizError

// Newf 构造带格式化 message 的 BizError
var Newf = kstypes.Newf

// New 构造 BizError（保持本地调用兼容）
func New(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}
```

删除原有的：
- `type BizError struct { ... }`
- `func (e *BizError) Error() string { ... }`
- `func New(code int, message string) *BizError { ... }`
- `func Newf(code int, format string, args ...any) *BizError { ... }`
- `func (e *BizError) Is(target error) bool { ... }`

- [ ] **Step 3: Re-export 通用错误码**

在错误码 var 块中，将与 ks-types 重叠的错误码改为 re-export。例如：

```go
// re-export 通用错误码
var (
	ErrNotFound     = kstypes.ErrNotFound
	ErrDuplicate    = kstypes.ErrDuplicate
	ErrTokenExpired = kstypes.ErrTokenExpired
	ErrTokenInvalid = kstypes.ErrTokenInvalid
	ErrForbidden    = kstypes.ErrForbidden
)
```

其余 keystone 专有错误码（ErrInvalidCredentials、ErrLLMConnection 等全部 60+ 个）保持 `&BizError{Code: ..., Message: "..."}` 不变。

- [ ] **Step 4: 更新 ErrorHandler()**

将 `ErrorHandler()` 中的 `httpStatusFromCode(bizErr.Code)` 替换为 `bizErr.HTTPStatus()`（来自 ks-types），删除 `httpStatusFromCode()` 函数。

```go
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		if bizErr, ok := err.(*BizError); ok {
			c.JSON(http.StatusOK, gin.H{
				"code":    bizErr.Code,
				"message": bizErr.Message,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "服务器内部错误",
			})
		}
	}
}
```

注意：keystone 的 ErrorHandler 具体实现可能与上面不完全相同，以实际代码为准，关键是将 `httpStatusFromCode` 调用替换为 `bizErr.HTTPStatus()`，并删除 `httpStatusFromCode` 函数。

- [ ] **Step 5: 编译验证**

```bash
cd ~/projects/yuhan/keystone && go build ./...
```

如有编译错误，逐一修复。常见问题：
- `errors.New(code, msg)` 调用已保留本地 `New()` 函数，应该无问题
- 如果有直接引用 `BizError{Code: ..., Message: ...}` 加了 json tags 但字段名不变，无影响
- `fmt` 包导入可能不再需要（`Error()` 和 `Newf()` 已由 ks-types 提供）

- [ ] **Step 6: 运行测试**

```bash
cd ~/projects/yuhan/keystone && go test ./...
```

预期：全部 PASS。

- [ ] **Step 7: 提交**

```bash
cd ~/projects/yuhan/keystone && git add internal/core/errors/ && git commit -m "refactor: BizError 迁移到 ks-types re-export + 保留专有错误码"
```

---

### Task 8: keystone — marketplace/manifest.go 迁移

**Files:**
- Modify: `~/projects/yuhan/keystone/internal/marketplace/manifest.go`
- Modify: 可能需要修改引用 `ProtectionLevel`/`RuntimeMode` 常量的文件

- [ ] **Step 1: 搜索 ProtectionLevel 和 RuntimeMode 的使用**

```bash
cd ~/projects/yuhan/keystone && grep -rn "ProtectionSystem\|ProtectionProtected\|ProtectionPreinstalled\|ProtectionNone\|ProtectionLevel" --include="*.go" .
cd ~/projects/yuhan/keystone && grep -rn "marketplace\.RuntimeMode" --include="*.go" .
```

记录所有引用位置。

- [ ] **Step 2: 删除本地 ProtectionLevel 定义**

在 `~/projects/yuhan/keystone/internal/marketplace/manifest.go` 中，删除本地 `ProtectionLevel` 类型及其常量：

```go
// 删除以下内容：
type ProtectionLevel string

const (
	ProtectionSystem       ProtectionLevel = "system"
	ProtectionProtected    ProtectionLevel = "protected"
	ProtectionPreinstalled ProtectionLevel = "preinstalled"
	ProtectionNone         ProtectionLevel = "none"
)
```

- [ ] **Step 3: 删除本地 RuntimeMode 常量**

删除本地 `RuntimeMode` 常量（注意不要删除事件类型等其他常量）：

```go
// 删除以下内容：
const (
	RuntimeModeNone      = "none"
	RuntimeModeProcess   = "process"
	RuntimeModeContainer = "container"
)
```

- [ ] **Step 4: 更新引用**

根据 Step 1 的搜索结果，将所有引用改为 `kstypes.ProtectionXxx` 和 `kstypes.RuntimeModeXxx`。

如果 `AppManifest` 结构体有 `Protection ProtectionLevel` 字段，改为 `Protection kstypes.ProtectionLevel`。

确保文件顶部已有 `kstypes "github.com/wuhanyuhan/ks-types"` 导入（已有则无需改动）。

- [ ] **Step 5: 编译验证**

```bash
cd ~/projects/yuhan/keystone && go build ./...
```

如有编译错误，逐一修复——主要是加 `kstypes.` 前缀。

- [ ] **Step 6: 运行测试**

```bash
cd ~/projects/yuhan/keystone && go test ./...
```

预期：全部 PASS。

- [ ] **Step 7: 提交**

```bash
cd ~/projects/yuhan/keystone && git add internal/marketplace/ && git commit -m "refactor: ProtectionLevel/RuntimeMode 迁移到 ks-types 枚举"
```

---

### Task 9: 全项目交叉验证

**Files:** 无新改动

- [ ] **Step 1: ks-types 全量测试**

```bash
cd ~/projects/yuhan/ks-types && go test ./... -v -count=1
```

预期：全部 PASS。

- [ ] **Step 2: ks-hub 全量测试**

```bash
cd ~/projects/yuhan/ks-hub && go test ./... -count=1
```

预期：全部 PASS。

- [ ] **Step 3: ks-admin 全量测试**

```bash
cd ~/projects/yuhan/ks-admin && go test ./... -count=1
```

预期：全部 PASS。

- [ ] **Step 4: keystone 全量测试**

```bash
cd ~/projects/yuhan/keystone && go test ./... -count=1
```

预期：全部 PASS。

- [ ] **Step 5: ks-devkit 全量测试**

```bash
cd ~/projects/yuhan/ks-devkit/cli && go test ./... -count=1
```

预期：全部 PASS（ks-devkit 无代码改动，仅验证 ks-types API 兼容性）。

- [ ] **Step 6: 验证通过后完成**

所有项目测试通过，共享类型提取完成。
