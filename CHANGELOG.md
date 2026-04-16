# Changelog

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
