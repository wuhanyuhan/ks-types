package kstypes

// /meta 端点响应契约。
// 所有 MCP service 实现 /meta 端点时应返回此结构的 JSON 序列化。
// keystone 在动态注册时或 ConfigUI 代理时调用此端点发现服务能力。

// MetaNavDecl 是 MCP 自声明的左侧菜单项（v0.5.0 新增，由 Spec B 引入）。
// keystone 拉取 /meta 后据此在管理后台导航中插入菜单项。
type MetaNavDecl struct {
	Label         string   `json:"label"`                    // <= 12 字符，中文
	Icon          string   `json:"icon,omitempty"`           // lucide-react 图标名
	Category      string   `json:"category"`                 // 应用 / 工具 / 配置 / 集成
	Order         int      `json:"order,omitempty"`          // 默认 99
	OpenMode      string   `json:"open_mode"`                // dialog / fullpage
	EntryPath     string   `json:"entry_path,omitempty"`     // 默认 '/'
	RequiredPerms []string `json:"required_perms,omitempty"` // AND 语义；空数组 = admin 直通
}

// MetaPermissionDecl 是 MCP 自声明的权限码目录条目（v0.5.0 新增，由 Spec B 引入）。
// keystone 据此在权限管理后台展示并分配给角色。
type MetaPermissionDecl struct {
	Code         string   `json:"code"`                    // mcp.{mcp_id}.{action}
	Label        string   `json:"label"`
	DefaultRoles []string `json:"default_roles,omitempty"` // MVP 只 ['admin']
}

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
	// Nav MCP 自声明的左侧菜单项（v0.5.0 新增）。
	Nav *MetaNavDecl `json:"nav,omitempty"`
	// Permissions MCP 自声明的权限码目录数组（v0.5.0 新增）。
	Permissions []MetaPermissionDecl `json:"permissions,omitempty"`
	// ConfigMode 与 ConfigUI 的语义关系（v0.5.0 共存约定）：
	//   ConfigMode 是"配置模式分类"（schema / iframe / none / ""）
	//   ConfigUI   是"iframe 模式的接入信息"（URL 等）
	//
	// 兼容规则（keystone 后端按此优先级判断）：
	//   - ConfigMode == ""（老 SDK）→ 看 ConfigUI != nil && ConfigUI.Enabled 决定（保持 v0.4 语义）
	//   - ConfigMode == "iframe"   → 必须同时填 ConfigUI.URL（Spec B I-7 启动校验）
	//   - ConfigMode == "schema"   → ConfigUI 应为 nil（Spec A 配置由 SchemaForm 渲染）
	//   - ConfigMode == "none"     → ConfigUI 为 nil 或 Enabled=false
	//
	// 长期演进（v1.0.0 独立 spec 处理）：
	//   - ConfigUI.Enabled 标 @deprecated
	//   - 可能合并入 Config { Mode, IframeURL } 或保持原样
	//   - 本次 v0.5.0 不做重构，避免 breaking
	ConfigMode string `json:"config_mode,omitempty"`
	// ProtocolVersion SemVer "MAJOR.MINOR"，MVP "1.0"（v0.5.0 新增）。
	ProtocolVersion string `json:"protocol_version,omitempty"`
	// ConfigStatus unconfigured / via_frontend / via_cli / mixed（Spec A §6.4，v0.5.0 新增）。
	ConfigStatus string `json:"config_status,omitempty"`
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
