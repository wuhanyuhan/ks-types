package kstypes

// /meta 端点响应契约。
// 所有 MCP service 实现 /meta 端点时应返回此结构的 JSON 序列化。
// keystone 在动态注册时或 ConfigUI 代理时调用此端点发现服务能力。

// MetaResponse 是 MCP service /meta 端点的 JSON 响应结构。
//
// 字段对应 docs/mcp-server-ui-integration.md 中的协议约定。
// 新字段（如 AuthMode）向前扩展；旧消费者看到未知字段应忽略。
type MetaResponse struct {
	// Name 服务名，对应 manifest.AppSpec.ID。
	Name string `json:"name"`
	// Version 服务版本，对应 manifest.AppSpec.Version。
	Version string `json:"version"`
	// AuthMode 服务 /mcp 端点实际启用的鉴权模式。
	// keystone 注册时据此决定是否为调用动态签发 JWT。
	// 与 manifest.mount.service.auth_mode 应一致；为空时 keystone 视为 none。
	AuthMode AuthMode `json:"auth_mode,omitempty"`
	// ConfigUI 配置界面信息。不存在 UI 的服务此字段为 nil（JSON 中省略）。
	ConfigUI *ConfigUIInfo `json:"config_ui,omitempty"`
	// Tools 服务注册的 MCP 工具清单（可选，用于文档/自检）。
	Tools []ToolInfo `json:"tools,omitempty"`
}

// ConfigUIInfo 描述 service 带界面接入（iframe 嵌入 keystone 后台）的信息。
// 详见 docs/mcp-server-ui-integration.md。
type ConfigUIInfo struct {
	// Enabled 是否提供配置界面。
	Enabled bool `json:"enabled"`
	// URL 界面基础路径。
	// 相对路径（如 "/config-ui/"）基于服务自身地址解析（推荐）；
	// 也可使用绝对 URL（用于 UI 与 API 跨主机部署的场景）。
	URL string `json:"url,omitempty"`
}

// ToolInfo 是 MCP 工具的简要描述。
// 不是 MCP 协议 tools/list 的完整替代，仅用于 /meta 端点的能力展示。
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
