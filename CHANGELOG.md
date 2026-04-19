# Changelog

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
