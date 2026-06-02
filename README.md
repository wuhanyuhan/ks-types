# ks-types

Keystone（KS）平台的共享 Go 类型库：统一错误码、Ed25519 JWT、Manifest 解析、权限注册表，以及开箱即用的 Gin 中间件。

被 Keystone 平台的各云服务引用，保证跨服务的类型与契约一致。

## 特性

- **统一错误码**：`BizError` + 分段常量（`40xxx`/`50xxx`），方便前端映射与日志聚合。
- **Ed25519 JWT**：提供实例 JWT（`InstanceClaims`）与开发者 JWT（`DeveloperClaims`）的签发/校验，算法锁定 `EdDSA`。
- **Manifest 解析**：`AppSpec` 同时带 `yaml` 和 `json` tag，内建 `Validate()`。
- **权限注册表**：`PermissionRegistry` 支持动态注册维度、未知维度告警、非法 level 报错、高风险权限检测。
- **Gin 中间件**：`ginmw.InstanceJWTMiddleware` 读取 `Authorization: Bearer`，支持可选的吊销回调。
- **Attestation JWT (ATT+JWT)**：平台签发给客户端的实例身份证明，独立于 Instance JWT，专供局域网发现场景做"实例合法性校验"。`SignAttestation` / `VerifyAttestation` 使用与 Instance JWT 同一对 Ed25519 密钥，但 `aud` 锁死为 `"ks-client"`、`typ` 为 `"ATT+JWT"`、强制 `kid` header，与 Instance JWT 不可互换误用。
- **Widgets 协议（widgets-protocol-v1）**：MCP tool 结果的 widget 渲染契约——绑定类型 + 5 个 MVP widget 数据 schema + `ks://` / `ui://` URI 解析 + postMessage 常量；经 tygo 派生 TypeScript 类型供前端消费（详见 [`docs/widgets-protocol-v1.md`](docs/widgets-protocol-v1.md)）。

## 安装

```bash
go get github.com/wuhanyuhan/ks-types
```

要求 Go `1.26` 及以上（见 `go.mod`）。

### TypeScript 类型分发（widgets-protocol-v1+）

ks-types 同时以 npm 包形态发布 tygo 派生的 TypeScript 类型，供前端消费方使用。

```jsonc
// package.json
"@wuhanyuhan/ks-types": "github:wuhanyuhan/ks-types#v0.31.0"
```

```typescript
import { WidgetListActionsV1, UIResource, PMMethodAppReady } from '@wuhanyuhan/ks-types'
```

- 类型层（`dist/widgets.d.ts`）由 `make types-gen` 从 Go 源码自动派生；
- 运行时层（`dist/index.js`）手动镜像 `widgets.d.ts` 顶层 `export const` 字面量（postmessage 方法名 / sandbox flag / WidgetURIScheme）；新增 const 时需同步更新本文件。

## 快速开始

### 1. 签发与校验实例 JWT

```go
import (
    "time"
    kstypes "github.com/wuhanyuhan/ks-types"
)

privPEM, _ := os.ReadFile("instance_priv.pem")
pubPEM,  _ := os.ReadFile("instance_pub.pem")

token, err := kstypes.SignInstanceJWT(kstypes.InstanceClaims{
    InstanceID: "inst-001",
    Name:       "demo",
    Group:      "default",
}, privPEM, 2*time.Hour)
if err != nil { /* ... */ }

claims, err := kstypes.VerifyInstanceJWT(token, pubPEM)
if err != nil { /* ... */ }
fmt.Println(claims.InstanceID, claims.Name)
```

### 2. 解析并校验应用 Manifest

```go
data, _ := os.ReadFile("manifest.yaml")
m, err := kstypes.ParseAppSpec(data)
if err != nil { /* ... */ }
if err := m.Validate(); err != nil { /* ... */ }
```

### 3. 权限注册表与高风险检测

```go
reg := kstypes.DefaultPermissionRegistry()

warnings, err := reg.Validate(m.Permissions)
if err != nil {
    // level 非法
    return err
}
for _, w := range warnings {
    log.Printf("warn: %s - %s", w.Dimension, w.Message)
}

highRisk := reg.HighRiskPermissions(m.Permissions, 5) // 阈值 > 5
if len(highRisk) > 0 {
    log.Printf("需要人工审核: %v", highRisk)
}
```

自定义维度：

```go
reg.Register("billing", kstypes.PermissionDimension{
    DisplayName: "计费接口",
    Levels:      []string{"none", "read", "write"},
    RiskWeight:  6,
})
```

### 4. Gin 中间件

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/wuhanyuhan/ks-types/ginmw"
)

pubPEM, _ := os.ReadFile("instance_pub.pem")

r := gin.Default()
r.Use(ginmw.InstanceJWTMiddleware(pubPEM, func(id string) bool {
    return revokedCache.Has(id) // 可选吊销检查，传 nil 可跳过
}))

r.GET("/me", func(c *gin.Context) {
    info := ginmw.GetInstanceInfo(c)
    c.JSON(200, gin.H{"instance": info.InstanceID, "name": info.Name})
})
```

### 5. 业务错误码

```go
if user == nil {
    return kstypes.ErrNotFound
}
// 在 HTTP 层统一映射
c.JSON(http.StatusNotFound, gin.H{
    "code":    kstypes.ErrNotFound.Code,
    "message": kstypes.ErrNotFound.Message,
})
```

完整错误码列表见 [`errors.go`](errors.go)。

## Auth: app 鉴权模式

顶层 `auth.mode` 段声明 app 入站 `/mcp` 端点的鉴权模式（`AuthSpec`，见 `manifest.go`）。三种合法值：

| 值 | 语义 |
|----|-----|
| `none` | /mcp 端点不做鉴权，依赖网络边界（内网 + keystone 是唯一调用方） |
| `keystone_jwks` | 通过 keystone `/.well-known/jwks.json` 验证调用者 JWT（推荐，strict-by-default） |
| `static_bearer` | 比对静态 Bearer（由平台侧在调用方连接配置注入 auth_headers） |

默认值：留空时由 `AuthSpec.EffectiveMode` 按 secure-by-default 派生——有入站端点（`runtime.mode=container/process`）→ `keystone_jwks`，无入站端点（`agent` / `skill`）→ `none`。`AuthMode.Default()` 只做"空→none"的朴素归一，新代码用 `EffectiveMode`。

```yaml
type: app
auth:
  mode: keystone_jwks
```

### 生态消费者

- **ks-devkit SDK (ksapp)**: `ksapp.WithKeystoneAuth()` 按 manifest 的 `auth.mode`
  挂载 JWKSVerifier；strict-by-default（`auth.mode=keystone_jwks` 且
  `KEYSTONE_JWKS_URL` 为空时启动 panic）
- **squad 运行时框架**: bootstrap 默认启用同等行为
- **keystone**: MCP proxy 按平台侧记录的 `auth.mode` 决定是否为调用
  动态签发 JWT 并注入 Authorization header

详见平台鉴权约定文档。

## 目录结构

按公开面分组（每个源文件均配同名 `_test.go`）：

**认证 / JWT**

- `jwt.go` — Ed25519 PEM 加载/解析
- `instance_claims.go` — 实例 JWT（`InstanceClaims`）
- `developer_claims.go` — 开发者 JWT（`DeveloperClaims`）
- `attestation_claims.go` — Attestation JWT（`ATT+JWT`，局域网发现场景）
- `ginmw/` — Gin 中间件（`InstanceJWTMiddleware`）

**应用 Manifest 与契约**

- `manifest.go` — `AppSpec` 顶层结构 + `Validate`
- `apptypes.go` — `AppType` / `PricingType` 枚举
- `install.go` — `InstallSpec` 安装规格 + 配置项类型
- `permissions.go` — `PermissionRegistry` 权限注册表
- `config_schema.go` — config schema 与 `AppPackageSignature` 应用包签名
- `standalone_fallback.go` — standalone（非托管）模式本地资源 fallback
- `task_template.go` — manifest `task_templates` 段任务模板
- `localized.go` — `LocalizedString` i18n 字段类型
- `meta.go` — `/meta` 服务能力发现端点类型
- `attachment.go` — MCP 工具结果附件签名约定
- `llm_intent.go` — LLM 意图词表（tier / capability / reasoning）
- `compliance.go` — 决策模式语义枚举

**错误与响应**

- `errors.go` — `BizError` 与分段错误码
- `result.go` — `Result` / `PageResult` / `ListResult` 通用响应

**Widgets 协议（widgets-protocol-v1，tygo 派生 TS）**

- `widgets.go` — widget binding 类型 + `UIResource`
- `widgets_data.go` — 5 个 MVP widget 数据 schema
- `widgets_registry.go` — 共享 widget schema 注册表
- `widget_uri.go` — WidgetURI 解析（`ks://` / `ui://`）
- `widget_postmessage.go` — postMessage 方法名 + sandbox flag 常量

**内部协议类型（keystone↔squad 运行时编排，非对外开发者契约）**

- `a2a_task.go` / `a2a_skill.go` / `a2a_security.go` / `agent_card.go` — A2A 协议类型
- `decision_gate.go` / `deliverable.go` / `expert_activity.go` — squad 过程编排与交付物类型

**测试 fixture**

- `testdata/` — manifest 样例（valid / invalid / i18n）、`delivery/` 交付物 JSON

## 开发

```bash
# 运行全部测试
go test ./...

# 覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 整理依赖
go mod tidy
```

> `testdata/test_*.pem` 仅用于单元测试，**不要**在任何真实环境使用。

## 贡献

- 本库被多个服务引用，请保持 API 稳定；破坏性改动请在 commit 中标注 `BREAKING`。
- 新增错误码需沿用 `errors.go` 中的分段前缀。
- 代码注释与错误消息统一使用中文。
- 面向 AI 编码助手的协作约定见 [`AGENTS.md`](AGENTS.md)（`CLAUDE.md` 为其软链接）。
