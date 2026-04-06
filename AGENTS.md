# AGENTS.md

本文件为 AI 编码助手（Claude Code、Codex、Cursor、Aider 等）提供 `ks-types` 仓库的上下文与协作约定。

## 项目概述

`ks-types` 是 **Keystone（KS）平台** 的共享类型库（Go module: `github.com/wuhanyuhan/ks-types`），在 `ks-hub`、`ks-admin` 等服务之间复用以下内容：

- 应用类型与定价枚举（`AppType`、`PricingType`）
- 统一业务错误码（`BizError`）
- Ed25519 密钥加载与 PEM 解析
- **实例 JWT**（`InstanceClaims`）与 **开发者 JWT**（`DeveloperClaims`）的签发/校验
- 应用 `manifest.yaml` 结构体（`ManifestSpec`）与校验
- 权限维度注册表（`PermissionRegistry`）与高风险权限检测
- Gin 中间件 `ginmw.InstanceJWTMiddleware`（含吊销检查钩子）

本项目是纯 **library**，不包含可执行程序，也不持有任何数据库或外部依赖。

## 目录结构

```
.
├── apptypes.go            # AppType / PricingType 枚举
├── errors.go              # BizError 与预定义错误码
├── jwt.go                 # Ed25519 PEM 加载/解析
├── instance_claims.go     # 实例 JWT 签发/校验
├── developer_claims.go    # 开发者 JWT 签发/校验
├── manifest.go            # ManifestSpec + YAML 解析 + Validate
├── permissions.go         # PermissionRegistry + DefaultPermissionRegistry
├── ginmw/
│   └── instance_jwt.go    # Gin 中间件：InstanceJWTMiddleware
├── testdata/
│   ├── test_private.pem   # 测试专用 Ed25519 私钥
│   ├── test_public.pem    # 测试专用 Ed25519 公钥
│   ├── valid_manifest.yaml
│   └── invalid_manifest.yaml
├── go.mod                 # module github.com/wuhanyuhan/ks-types
└── *_test.go              # 每个源文件对应一个测试文件
```

根包的 Go package name 为 `kstypes`；`ginmw/` 子包的 package name 为 `ginmw`。

## 关键设计约定

### 1. 错误码
- 所有业务错误都是 `*BizError`（`errors.go`），结构 `{Code int, Message string}`。
- 错误码按 HTTP 状态码前缀分段：`400xx` 通用、`401xx` 认证、`403xx` 权限、`404xx/409xx` 应用、`503xx` 存储、`500xx` 服务器内部。
- 新增错误时请沿用现有分段，不要新造前缀。

### 2. JWT
- 算法**固定**为 `EdDSA`（Ed25519）。不要引入 HS256/RS256 等其他算法。
- `InstanceClaims.InstanceID` 通过 `RegisteredClaims.Subject` 往返序列化，`json:"-"` 标签防止重复输出。
- Issuer/Audience 在签发函数内部写死：
  - Instance JWT: `iss=ks-admin`, `aud=[ks-hub, ks-admin]`
  - Developer JWT: `iss=ks-hub`, `aud=[ks-hub]`
- 公私钥必须是 PKCS#8 + PKIX 格式的 PEM。

### 3. Manifest
- 字段同时带 `yaml` 和 `json` tag，保证 YAML 解析与 JSON 序列化都可用。
- `Validate()` 只校验必填字段与枚举合法性；与权限注册表的交叉校验由调用方通过 `PermissionRegistry.Validate` 完成。

### 4. 权限
- 权限维度通过 `PermissionRegistry.Register` 动态注册。`DefaultPermissionRegistry()` 预置了 `network / llm / filesystem / user_context` 四项。
- 未知维度返回 `warnings` 而非 `error`（需要人工审核），非法 `level` 直接返回 `error`。
- `RiskWeight` 用于高风险检测；阈值由调用方决定。

### 5. Gin 中间件
- `InstanceJWTMiddleware` 从 `Authorization: Bearer <jwt>` 中取 token，验证后用常量 key `instance_info` 写入 Gin context。
- `isRevoked` 回调是**可选**的（传 nil 跳过）。这样库本身不依赖任何缓存/数据库实现。
- 通过 `ginmw.GetInstanceInfo(c)` 从 handler 中取出 `*kstypes.InstanceClaims`。

### 6. 注释与命名
- 代码注释使用**中文**（与团队习惯一致），公开 API 的 doc comment 也是中文。
- 导出符号遵循 Go 惯例（大写开头、`GoDoc` 风格首词）。

## 开发工作流

### 运行测试
```bash
go test ./...
```

### 覆盖率
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```
`coverage.out` / `coverage.html` 已在 `.gitignore` 中，**不要**提交到仓库。

### 依赖管理
```bash
go mod tidy
```
本项目仅允许引入必要的依赖。当前运行时依赖只有三个：
- `github.com/gin-gonic/gin`
- `github.com/golang-jwt/jwt/v5`
- `gopkg.in/yaml.v3`

新增依赖需要谨慎评估，优先用标准库。

### 测试密钥
`testdata/test_private.pem` 与 `testdata/test_public.pem` **仅用于单元测试**，不得用于任何生产或共享环境。不要在这些文件之外生成/提交密钥。

## 给 AI Agent 的提示

- **改动前先读源文件**。每个文件都很短（大多 < 120 行），完整读取再修改比基于片段推断更安全。
- **先写测试**。每个源文件都配有 `_test.go`；新功能或 bug 修复都应补充断言，避免破坏既有调用方。
- **保持 API 稳定**。本库被多个服务引用，导出符号的破坏性改动需要在 commit message 中明确标注 `BREAKING`。
- **错误要有码**。新增错误请在 `errors.go` 中定义 `*BizError`，不要在业务代码里散落 `fmt.Errorf`。
- **禁止**在根目录引入 `main.go` 或 CLI。如果确实需要示例程序，放到 `examples/` 子目录并加 build tag。
- **中文注释**。代码注释、错误 message、doc comment 保持中文风格一致。

## 常见任务清单

| 任务 | 涉及文件 |
|------|----------|
| 新增业务错误码 | `errors.go` + 对应分段 |
| 新增权限维度 | `permissions.go` 的 `DefaultPermissionRegistry()` |
| 扩展 Manifest 字段 | `manifest.go` + `testdata/valid_manifest.yaml` + `manifest_test.go` |
| 调整 JWT Claims | `instance_claims.go` / `developer_claims.go` + 同名 `_test.go` |
| 新增 Gin 中间件 | `ginmw/` 下新文件，包名 `ginmw` |

## 相关位置

- 模块路径：`github.com/wuhanyuhan/ks-types`
- 默认分支：`master`
- Go 版本：`1.26.1`（见 `go.mod`）
