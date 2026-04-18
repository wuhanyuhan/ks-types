# Changelog

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
