# 共享类型提取设计文档

> 日期：2026-04-07
> 范围：ks-types + keystone + ks-hub + ks-admin + ks-devkit

## 1. 背景与目标

ks-types 是 Keystone 生态的共享类型库，当前已覆盖 JWT、Manifest、Permission、AppType 等协议层类型。但在实际使用中，多个消费方存在以下问题：

- **BizError 重复定义**：keystone 和 ks-admin 各自定义了 BizError struct，而非使用 ks-types
- **Result 响应结构重复**：ks-hub 和 ks-admin 有完全相同的 Result/PageResult/ListResult 定义
- **枚举不完整**：ManifestSpec.Protection 是裸 string，keystone 本地定义了 ProtectionLevel typed enum；RuntimeMode 虽在 ks-types 中定义但 ManifestSpec.Runtime.Mode 仍为 string
- **常量重复**：keystone marketplace 本地重复定义了 RuntimeMode 常量

**目标**：消除跨项目的类型重复，让 ks-types 成为唯一事实来源。

## 2. 设计决策记录

| 决策 | 选项 | 结论 | 理由 |
|------|------|------|------|
| BizError 增强程度 | A) 最小 B) 适度 C) 完全 | **B** | Newf + Is 是通用需求，Gin ErrorHandler 各项目差异大不适合统一 |
| Result 是否提取 | A) 提取 B) 不提取 | **A** | ks-hub/ks-admin 定义完全相同，是事实标准 |
| ProtectionLevel 类型化 | A) typed enum B) 保持 string | **A** | 与 RuntimeMode/AppType/PricingType 风格一致 |
| 专有错误码归属 | A) 留本地 B) 全搬 ks-types | **A** | ks-types 只存跨项目共用码，各项目用 alias+扩展模式 |
| 实施方案 | A) 原地增强 B) 拆子包 C) v2 | **A** | 代码量小（~600行），根包足够，v0.x 无兼容包袱 |
| 兼容性考虑 | 需要/不需要 | **不需要** | 项目未上线 |

## 3. ks-types 变更

### 3.1 errors.go — BizError 增强

**变更内容**：

1. BizError struct 加 json tags：
```go
type BizError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

2. 新增 Newf 工厂函数：
```go
func Newf(code int, format string, args ...any) *BizError {
    return &BizError{Code: code, Message: fmt.Sprintf(format, args...)}
}
```

3. 新增 Is 方法（满足 errors.Is 协议）：
```go
func (e *BizError) Is(target error) bool {
    t, ok := target.(*BizError)
    return ok && e.Code == t.Code
}
```

4. 现有 18 个错误码常量不变，不新增专有码。

### 3.2 result.go — 新建

```go
type Result struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
}

func OK(data any) Result {
    return Result{Code: 0, Message: "ok", Data: data}
}

func Fail(code int, message string) Result {
    return Result{Code: code, Message: message}
}

func FailErr(err *BizError) Result {
    return Result{Code: err.Code, Message: err.Message}
}

type PageResult[T any] struct {
    Items []T `json:"items"`
    Total int `json:"total"`
    Page  int `json:"page"`
    Size  int `json:"size"`
}

type ListResult[T any] struct {
    Items []T `json:"items"`
}
```

### 3.3 apptypes.go — 新增 ProtectionLevel + RuntimeMode.Valid 调整

新增 ProtectionLevel 枚举：
```go
type ProtectionLevel string

const (
    ProtectionNone         ProtectionLevel = "none"
    ProtectionPreinstalled ProtectionLevel = "preinstalled"
    ProtectionProtected    ProtectionLevel = "protected"
    ProtectionSystem       ProtectionLevel = "system"
)

func (p ProtectionLevel) Valid() bool {
    return p == "" || validProtectionLevels[p]
}
```

RuntimeMode.Valid() 支持空值：
```go
func (m RuntimeMode) Valid() bool { return m == "" || validRuntimeModes[m] }
```

### 3.4 manifest.go — 字段类型化 + Validate 简化

字段类型变更：
- `ManifestSpec.Protection`: `string` → `ProtectionLevel`
- `RuntimeSpec.Mode`: `string` → `RuntimeMode`

Validate() 简化：
- Protection 校验：内联 map → `m.Protection.Valid()`
- RuntimeMode 校验：`RuntimeMode(m.Runtime.Mode).Valid()` → `m.Runtime.Mode.Valid()`

### 3.5 测试文件

| 文件 | 动作 |
|------|------|
| `result_test.go` | 新建：覆盖 OK/Fail/FailErr/PageResult/ListResult |
| `errors_test.go` | 补充：Newf 构造、Is 比较（同码/异码/非 BizError） |
| `apptypes_test.go` | 补充：ProtectionLevel 各值 + 空值 Valid() |
| `manifest_test.go` | 适配字段类型变更（赋值改用常量） |

## 4. ks-hub 变更

### 4.1 internal/core/result.go — 改为 re-export

删除本地 Result/PageResult/ListResult/OK/Fail 定义，改为：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

// 响应结构 re-export
type Result = kstypes.Result
type PageResult[T any] = kstypes.PageResult[T]
type ListResult[T any] = kstypes.ListResult[T]

var (
    OK   = kstypes.OK
    Fail = kstypes.Fail
)
```

这样 ks-hub 内部所有 `core.OK()`/`core.Fail()`/`core.Result` 的调用代码零改动。

### 4.2 internal/core/errors.go — 补充新符号

现有 alias 模式不变，补充 re-export：

```go
var Newf = kstypes.Newf
var FailErr = kstypes.FailErr
```

### 4.3 go.mod

ks-types replace 指向本地路径已存在，无需改动。

## 5. ks-admin 变更

### 5.1 internal/core/errors.go — 重写

删除本地 BizError struct 和 NewBizError/NewBizErrorPair 工厂函数。改为：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

type BizError = kstypes.BizError

// re-export 通用错误码
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

var Newf = kstypes.Newf
```

### 5.2 internal/core/result.go — 改为 re-export

与 ks-hub 同理：

```go
package core

import kstypes "github.com/wuhanyuhan/ks-types"

type Result = kstypes.Result
type PageResult[T any] = kstypes.PageResult[T]
type ListResult[T any] = kstypes.ListResult[T]

var (
    OK   = kstypes.OK
    Fail = kstypes.Fail
)
```

### 5.3 全项目编译适配

ks-admin 内部引用 `core.BizError`/`core.ErrXxx`/`core.OK`/`core.Fail` 的代码不需要改动（alias 透明）。需要检查：
- `NewBizError(code, msg)` 调用改为 `kstypes.Newf(code, msg)` 或 `&core.BizError{Code: code, Message: msg}`
- `NewBizErrorPair(code, msg)` 调用同理替换

### 5.4 go.mod

ks-types replace 指向本地路径已存在，无需改动。

## 6. keystone 变更

### 6.1 internal/core/errors/errors.go — 重写

删除本地 BizError struct、Error()、New()、Newf()、Is()、httpStatusFromCode()。改为：

```go
package errors

import kstypes "github.com/wuhanyuhan/ks-types"

type BizError = kstypes.BizError

var Newf = kstypes.Newf

// re-export 通用错误码
var (
    ErrNotFound     = kstypes.ErrNotFound
    ErrDuplicate    = kstypes.ErrDuplicate
    ErrTokenExpired = kstypes.ErrTokenExpired
    ErrTokenInvalid = kstypes.ErrTokenInvalid
    ErrForbidden    = kstypes.ErrForbidden
    // ... 按需 re-export
)

// keystone 专有错误码（保留所有 60+ 个）
var (
    ErrInvalidCredentials = &BizError{Code: 40101, Message: "邮箱或密码错误"}
    // ... 其余专有码保持不变
)
```

ErrorHandler() Gin 中间件保留在本地（各项目实现不同）。`httpStatusFromCode()` 改为调用 `BizError.HTTPStatus()`（ks-types 已提供）。

### 6.2 internal/marketplace/manifest.go — 删除重复常量

删除本地定义的：
- `ProtectionLevel` 类型及 `ProtectionSystem`/`ProtectionProtected`/`ProtectionPreinstalled`/`ProtectionNone` 常量
- `RuntimeModeNone`/`RuntimeModeProcess`/`RuntimeModeContainer` 常量

改为使用 `kstypes.ProtectionSystem`、`kstypes.RuntimeModeNone` 等。

`AppManifest` 嵌入 `kstypes.ManifestSpec` 的模式不变。

### 6.3 全项目编译适配

- 搜索所有 `errors.New(`、`errors.Newf(` 调用确保签名兼容
- 搜索 `ProtectionSystem`/`RuntimeModeNone` 等常量引用，加 `kstypes.` 前缀或通过本地 re-export
- 搜索 `errors.Is(err, target)` 确保 BizError.Is() 行为一致

### 6.4 go.mod

ks-types replace 指向本地路径已存在，无需改动。

## 7. ks-devkit 变更

### 7.1 cli/go.mod

ks-types replace 指向本地路径已存在，无需改动。

### 7.2 其他

无改动。ks-devkit CLI 的 `APIResponse`（使用 `json.RawMessage`）保持不动，与服务端 Result 需求不同。

## 8. 验证策略

每个项目改完后执行：

```bash
cd ~/projects/yuhan/<project> && go build ./... && go test ./...
```

验证顺序：ks-types → ks-hub → ks-admin → keystone → ks-devkit（按依赖链顺序）。

## 9. 文件变更矩阵

| 项目 | 文件 | 动作 |
|------|------|------|
| **ks-types** | `errors.go` | 改：json tags + Newf + Is |
| | `result.go` | 新建 |
| | `apptypes.go` | 改：新增 ProtectionLevel + RuntimeMode.Valid 调整 |
| | `manifest.go` | 改：字段类型化 + Validate 简化 |
| | `result_test.go` | 新建 |
| | `errors_test.go` | 改：补充测试 |
| | `apptypes_test.go` | 改：补充测试 |
| | `manifest_test.go` | 改：适配 |
| **ks-hub** | `internal/core/result.go` | 重写为 re-export |
| | `internal/core/errors.go` | 改：补充 Newf/FailErr re-export |
| **ks-admin** | `internal/core/errors.go` | 重写为 alias + re-export |
| | `internal/core/result.go` | 重写为 re-export |
| **keystone** | `internal/core/errors/errors.go` | 重写为 alias + re-export + 保留专有码 |
| | `internal/marketplace/manifest.go` | 改：删除 ProtectionLevel/RuntimeMode 本地定义 |
| **ks-devkit** | `cli/go.mod` | 版本对齐（如需要） |
