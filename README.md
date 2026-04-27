# ks-types

Keystone（KS）平台的共享 Go 类型库：统一错误码、Ed25519 JWT、Manifest 解析、权限注册表，以及开箱即用的 Gin 中间件。

被 `ks-hub` / `ks-admin` 等服务引用，保证跨服务的类型与契约一致。

## 特性

- **统一错误码**：`BizError` + 分段常量（`40xxx`/`50xxx`），方便前端映射与日志聚合。
- **Ed25519 JWT**：提供实例 JWT（`InstanceClaims`）与开发者 JWT（`DeveloperClaims`）的签发/校验，算法锁定 `EdDSA`。
- **Manifest 解析**：`AppSpec` 同时带 `yaml` 和 `json` tag，内建 `Validate()`。
- **权限注册表**：`PermissionRegistry` 支持动态注册维度、未知维度告警、非法 level 报错、高风险权限检测。
- **Gin 中间件**：`ginmw.InstanceJWTMiddleware` 读取 `Authorization: Bearer`，支持可选的吊销回调。
- **Attestation JWT (ATT+JWT)**：ks-admin 签发给 ks-client 的实例身份证明，独立于 Instance JWT，专供局域网发现场景做"实例合法性校验"。`SignAttestation` / `VerifyAttestation` 使用与 Instance JWT 同一对 Ed25519 密钥，但 `aud` 锁死为 `"ks-client"`、`typ` 为 `"ATT+JWT"`、强制 `kid` header，与 Instance JWT 不可互换误用。

## 安装

```bash
go get github.com/wuhanyuhan/ks-types
```

要求 Go `1.26` 及以上（见 `go.mod`）。

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

## AuthMode: MCP Service 鉴权模式

从 v0.4.0 起，`mount.service.auth_mode` 声明 MCP service 的 /mcp 端点
鉴权模式。三种合法值：

| 值 | 语义 |
|----|-----|
| `none` | /mcp 端点不做鉴权，依赖网络边界（内网 + keystone 是唯一调用方） |
| `keystone_jwks` | 通过 keystone /.well-known/jwks.json 验证 RS256 JWT（推荐） |
| `static_bearer` | 比对静态 Bearer（调用方在 keystone 侧注入 auth_headers） |

默认值：空字符串（YAML 中省略）等价于 `none`，通过 `AuthMode.Default()` 归一。

### Extension mount 用例

`type: extension` 的应用也可以声明 `auth_mode`（自 v0.4.1 起）：

```yaml
type: extension
mount:
  extension:
    mcp_server_name: my-ext
    transport_type: streamable_http
    endpoint: "http://localhost:9991/mcp"
    auth_mode: keystone_jwks
```

### 生态消费者

- **ks-devkit SDK (ksapp)**: `ksapp.WithKeystoneAuth()` 按 manifest 的 auth_mode
  挂载 JWKSVerifier；strict-by-default（auth_mode=keystone_jwks 且
  `KEYSTONE_JWKS_URL` 为空时启动 panic）
- **ks-squad-framework**: bootstrap 默认启用同等行为
- **keystone**: MCP proxy 按 `t_mcp_servers.auth_mode` 决定是否为调用
  动态签发 JWT 并注入 Authorization header

详见 `docs/ecosystem/standards/service-auth-convention.md`（keystone 仓库）。

## 目录结构

```
.
├── apptypes.go            # AppType / PricingType 枚举
├── errors.go              # BizError 与错误码
├── jwt.go                 # Ed25519 PEM 加载/解析
├── instance_claims.go     # 实例 JWT
├── developer_claims.go    # 开发者 JWT
├── manifest.go            # Manifest 结构体与校验
├── install.go             # InstallSpec 安装规格
├── result.go              # Result / PageResult / ListResult 通用响应
├── permissions.go         # 权限注册表
├── ginmw/                 # Gin 中间件
└── testdata/              # 测试用 PEM 与 manifest 样例
```

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
