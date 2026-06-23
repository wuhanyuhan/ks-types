package kstypes

// postMessage 协议方法名（host ↔ iframe）。
// 仅自定义 widget（ui:// scheme）路径用；共享 widget 走同源 React，不走 postmessage。
//
// 协议层只定义方法名（wire format string），具体 payload 由前端 SDK 与 squad SDK 自行规范。
const (
	PMMethodAppReady              = "app.ready"
	PMMethodAppData               = "app.data"
	PMMethodAppResize             = "app.resize"
	PMMethodAppCallServerTool     = "app.callServerTool"
	PMMethodAppToolResult         = "app.toolResult"
	PMMethodAppToolError          = "app.toolError"
	PMMethodAppUpdateModelContext = "app.updateModelContext"
	PMMethodAppClose              = "app.close"
	PMMethodAppNotify             = "app.notify"
	PMMethodAppOpenLink           = "app.openLink"
)

// mounted fullpage 子路由同步 postMessage 方法名（host ↔ iframe）。
const (
	// PMMethodMountedRouteChanged 表示 iframe 内应用子路由已变化，host 应更新父级 URL。
	PMMethodMountedRouteChanged = "keystone.mounted.route.changed"
	// PMMethodMountedRouteRestore 表示 host 要求 iframe 恢复到指定子路由。
	PMMethodMountedRouteRestore = "keystone.mounted.route.restore"
)

// MountedRouteChangedMessage 是 mounted 应用发给 Keystone host 的子路由变化消息。
type MountedRouteChangedMessage struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	AppID   string `json:"appId"`
	Path    string `json:"path"`
	Hash    string `json:"hash,omitempty"`
	Title   string `json:"title,omitempty"`
	Replace bool   `json:"replace,omitempty"`
}

// MountedRouteRestoreMessage 是 Keystone host 发给 mounted 应用的子路由恢复消息。
type MountedRouteRestoreMessage struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Path    string `json:"path"`
	Hash    string `json:"hash,omitempty"`
	Replace bool   `json:"replace,omitempty"`
}

// iframe sandbox flag 常量（自定义 widget 容器用）。
//
// 默认基础集（不可变）：allow-scripts allow-forms
// 故意去掉：allow-same-origin / allow-top-navigation / allow-modals / allow-popups
//
// squad 在 ToolUIBinding.SandboxHints 申请追加 flag 时，必须落在本白名单内；
// 平台侧 mount 阶段会校验。
const (
	SandboxFlagAllowDownloads  = "allow-downloads"   // keystone 默认在白名单
	SandboxFlagAllowPopups     = "allow-popups"      // 默认不在
	SandboxFlagAllowModals     = "allow-modals"      // 默认不在
	SandboxFlagAllowSameOrigin = "allow-same-origin" // 默认严禁，特殊审批
)
