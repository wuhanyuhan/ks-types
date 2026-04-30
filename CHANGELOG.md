# Changelog

## [Unreleased]

### Added

- `AttestationClaims` struct + `SignAttestation` / `VerifyAttestation`：ks-admin 签发给 ks-client 的实例身份证明 JWT，typ=ATT+JWT，aud=ks-client，强制 kid header。仅用于 ks-client 端验证 LAN 内发现的 keystone 实例是否由 ks-admin 合法签发。

## [v0.8.0] - 2026-04-30

### Added

- `ginmw.RequireAudience(svc string) Option`：`InstanceJWTMiddleware` 新增 functional option，强制要求 JWT 的 `aud` 包含指定服务名；不传 option 时保持默认放行（向后兼容）。
- `ginmw.Option` 类型 + 中间件签名扩展：`InstanceJWTMiddleware(publicPEM, isRevoked, opts ...Option)`，旧调用方无需改动。

### Changed

- `SignInstanceJWT` 的 `Audience` 由 `["ks-hub", "ks-admin"]` 扩展为 `["ks-admin", "ks-hub", "ks-relay", "ks-llm-gateway"]`，覆盖生态全部 4 个云服务。

### Rationale

为统一实例鉴权改造（spec `2026-04-30-unified-instance-auth-and-shared-redis-design.md`）的 Phase A 协议层前置。让 keystone 一份 `instance_jwt` 通用于 ks-admin / ks-hub / ks-relay / ks-llm-gateway 四端；中间件 RequireAudience 默认放行，由各服务在 Phase B/C 改造完成后显式启用，避免存量 token 因 aud 列表不匹配被立即拒收。

### Breaking Changes

无。新增字段值（aud 增加 2 个元素）与新 option 均向后兼容：旧调用方 `InstanceJWTMiddleware(pub, isRevoked)` 无需改动；旧 token（aud 仅 2 项）只要不启用 `RequireAudience`，仍可被中间件接受。

### Ecosystem

消费方：`ks-admin` / `ks-llm-gateway` / `keystone` / `ks-hub` 四仓 `go.mod` 同步升级到 `v0.8.0`。

## [v0.6.0] - 2026-04-19

### Added

- 新增 `config_schema.go`（Spec A §8 端点契约）：
  - `ConfigSchemaResponse` — MCP `/config-schema` 端点响应 data 字段
  - `ConfigPubkeyResponse` — MCP `/config-pubkey` 端点响应 data 字段
  - `EncryptedConfigPayload` — `POST /ks-config/save` 与 `/validate` 的 request body
  - `ConfigApplyResult` — `POST /ks-config/save` 成功响应 data 字段
  - `AADCanonicalBytes(mcpID string, version uint64, fingerprint string) []byte` — spec-v1 §2.1 canonical AAD 字节序列化 helper
  - `Fingerprint(pubkey []byte) string` — spec-v1 §4.1 X25519 公钥指纹 helper

### Tests

- `config_schema_test.go`：内联 12 条 `aad_canonical` testvectors（来自 ks-devkit conformance/config-schema/testvectors.json）+ 4 条 `Fingerprint` 断言，全量字节级对齐验证

### Ecosystem

Spec A（MCP 配置 Schema + E2E 加密协议）的共享契约前置。消费方：`ks-devkit/sdk/go`（go.mod replace 过渡）、`ks-devkit/sdk/typescript`（types.ts 镜像 v0.3.0）、`ks-devkit/sdk/python`（pydantic 镜像 config_schema.py）、`ks-squad-framework`（被动升级）。

### Breaking Changes

无。所有变更为新增类型和函数；旧消费者无需修改。

## [v0.5.0] - 2026-04-24

### Added

- `MetaResponse.Nav` (`*MetaNavDecl`) — MCP 自声明左侧菜单项（label / icon / category / order / open_mode / entry_path / required_perms）
- `MetaResponse.Permissions` (`[]MetaPermissionDecl`) — 权限码目录数组
- `MetaResponse.ConfigMode` — `schema` / `iframe` / `none` 枚举（与既有 `ConfigUI` 字段并存，详见 meta.go 注释中"v0.5.0 共存约定"）
- `MetaResponse.ProtocolVersion` — SemVer MAJOR.MINOR，MVP `1.0`
- `MetaResponse.ConfigStatus` — `unconfigured` / `via_frontend` / `via_cli` / `mixed`（由 Spec A §6.4 引入）

### Ecosystem

Spec B（MCP 前端鉴权与导航统一）的前置契约。Go 消费方直接 `go get -u github.com/wuhanyuhan/ks-types@v0.5.0`；Python / TS / squad-framework 各自手写镜像同步（见 keystone Spec B 子 plan Task 1.b/1.c/1.d）。

### Breaking Changes

无。所有变更为 `omitempty` 可选字段；旧消费者解析未含新字段的响应行为完全不变。

## [v0.4.1] - 2026-04-17

### Added
- `ExtensionMountSpec.AuthMode` 字段（对齐 `ServiceMountSpec.AuthMode`）
- `Validate()` 加 extension mount auth_mode 合法性校验

### Rationale
ks-mcp-image-gen 是 type=extension 且需启用 keystone_jwks 鉴权的首个场景；
原先只有 service mount 有 auth_mode 字段，extension 无法声明。

## [0.4.0] - 2026-04-17

### Added

- `AuthMode` 枚举（`apptypes.go`）：`none` / `keystone_jwks` / `static_bearer`
- `ServiceMountSpec.AuthMode` 字段（`manifest.go`）——声明 MCP service /mcp 端点鉴权模式
- `Validate()` 增强：非法 auth_mode 值在解析校验阶段拒绝
- `MetaResponse` / `ConfigUIInfo` / `ToolInfo` 契约类型（`meta.go`）——`/meta` 端点响应结构
- `AuthMode.Default()` 辅助：空字符串归一为 `AuthModeNone`

### Rationale

为 MCP 生态统一鉴权协议提供单一真相源。消费者（ks-devkit SDK、
ks-squad-framework、keystone）读同一 schema 实现一致行为。

### Breaking Changes

无。所有变更为向前兼容的新增字段。
