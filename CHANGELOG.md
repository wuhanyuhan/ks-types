# Changelog

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
