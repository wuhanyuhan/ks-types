package kstypes

// AppType 应用类型枚举
type AppType string

const (
	// AppTypeApp 跑 MCP server 的工具服务（外部进程/容器），keystone 视作黑盒能力
	// 提供者，能力经 provides.capabilities 暴露。
	// 何时选：纯工具集、无 LLM 负责人编排——只对外提供一批可调用能力。
	// （合并了旧 service + extension：两者安装路径相同，区分是伪命题。）
	AppTypeApp AppType = "app"
	// AppTypeSquad 一个 agent 团队（lead + 专家，外部进程/容器），keystone 视作
	// 单一"团队前门"黑盒；团队内部由 ks-squad-framework 自治，不泄漏给平台。
	// 何时选：有 LLM 负责人编排多个专家技能——以一个前门能力对外供给。
	AppTypeSquad AppType = "squad"
	// AppTypeAgent 单体 LLM 智能体（运行在 keystone 内，runtime.mode=none、无独立进程），
	// 消费能力网格完成用户意图。
	// 何时选：平台内的角色化助手，自身不跑 server、只调别人的能力。（旧 assistant 改名。）
	AppTypeAgent AppType = "agent"
	// AppTypeSkill 技能资源（运行在 keystone 内、挂载形态），被语义召回后调用，轻量供给。
	// 何时选：单一可被召回的技能/方法资源，不构成独立 agent 或服务。
	AppTypeSkill AppType = "skill"
)

var validAppTypes = map[AppType]bool{
	AppTypeApp: true, AppTypeSquad: true, AppTypeAgent: true, AppTypeSkill: true,
}

// Valid 检查 AppType 是否合法
func (t AppType) Valid() bool { return validAppTypes[t] }

// PricingType 定价类型枚举
type PricingType string

const (
	PricingFree     PricingType = "free"
	PricingPaid     PricingType = "paid"
	PricingFreemium PricingType = "freemium"
)

var validPricingTypes = map[PricingType]bool{
	PricingFree: true, PricingPaid: true, PricingFreemium: true,
}

// Valid 检查 PricingType 是否合法
func (t PricingType) Valid() bool { return validPricingTypes[t] }

// RuntimeMode 运行时模式
type RuntimeMode string

const (
	RuntimeModeNone      RuntimeMode = "none"
	RuntimeModeProcess   RuntimeMode = "process"
	RuntimeModeContainer RuntimeMode = "container"
	// RuntimeModeExtension 表示应用以 stdio MCP server 等扩展进程形态运行，
	// dispatcher 通过 mcp_tool kind 的 capability backend 路由调用。
	// v0.19.0 capability mesh 引入。
	RuntimeModeExtension RuntimeMode = "extension"
)

var validRuntimeModes = map[RuntimeMode]bool{
	RuntimeModeNone: true, RuntimeModeProcess: true, RuntimeModeContainer: true, RuntimeModeExtension: true,
}

// Valid 检查 RuntimeMode 是否合法；空值视为合法（等同 none）
func (m RuntimeMode) Valid() bool { return m == "" || validRuntimeModes[m] }

// StorePresentation 描述 Store 商品展示形态，不改变安装和运行时类型。
type StorePresentation string

const (
	StorePresentationRoleAgent   StorePresentation = "role_agent"
	StorePresentationMethodSkill StorePresentation = "method_skill"
	StorePresentationToolkit     StorePresentation = "toolkit"
	StorePresentationConnector   StorePresentation = "connector"
	StorePresentationServiceApp  StorePresentation = "service_app"
	StorePresentationExpertTeam  StorePresentation = "expert_team"
)

var validStorePresentations = map[StorePresentation]bool{
	StorePresentationRoleAgent:   true,
	StorePresentationMethodSkill: true,
	StorePresentationToolkit:     true,
	StorePresentationConnector:   true,
	StorePresentationServiceApp:  true,
	StorePresentationExpertTeam:  true,
}

// Valid 检查 StorePresentation 是否合法；空值视为合法（由消费方派生默认展示）。
func (p StorePresentation) Valid() bool {
	return p == "" || validStorePresentations[p]
}

// ProtectionLevel 保护级别（none / protected / system）。
//
// 归属平台：protection 决定卸载策略，由平台对内置/预装应用打标，**第三方 manifest 即使
// 写 protected/system 也被忽略、按 none 处理**（否则任何 app 自称 system 即不可卸载=安全漏洞）。
// enforcement：归属从 manifest 读迁移到平台打标。preinstalled 死值已清除。
type ProtectionLevel string

const (
	// ProtectionNone 无特殊保护，可正常卸载——第三方 app 的有效默认（也是留空时的语义）。
	ProtectionNone ProtectionLevel = "none"
	// ProtectionProtected 受保护，卸载需额外确认；由平台对关键预装应用打标，作者声明无效。
	ProtectionProtected ProtectionLevel = "protected"
	// ProtectionSystem 系统级不可卸载；由平台对内置应用打标，第三方 manifest 写了无效。
	ProtectionSystem ProtectionLevel = "system"
)

var validProtectionLevels = map[ProtectionLevel]bool{
	ProtectionNone: true, ProtectionProtected: true, ProtectionSystem: true,
}

// Valid 检查 ProtectionLevel 是否合法；空值视为合法（等同 none）
func (p ProtectionLevel) Valid() bool {
	return p == "" || validProtectionLevels[p]
}

// AuthMode 描述 app 入站 /mcp 端点的鉴权模式。
//
// 声明位置：顶层 auth.mode 段（见 manifest.go AuthSpec）；旧 mount.service.auth_mode 已废。
// 默认值：空字符串经 AuthSpec.EffectiveMode 按 secure-by-default 派生——有入站端点
// （container/process）→ keystone_jwks，无入站端点（none：agent/skill）→ none。
// Default() 仅做"空→none"的朴素归一，不含 secure-by-default 派生；新代码用 EffectiveMode。
type AuthMode string

const (
	// AuthModeNone /mcp 端点不做鉴权，依赖网络边界。
	// 仅在受控内网 + keystone 是唯一调用方时可用。
	AuthModeNone AuthMode = "none"

	// AuthModeKeystoneJWKS service 通过 keystone /.well-known/jwks.json 公钥
	// 验证调用者 Authorization 头的 RS256 JWT（推荐，strict-by-default）。
	AuthModeKeystoneJWKS AuthMode = "keystone_jwks"

	// AuthModeStaticBearer service 比对静态 Bearer token（由平台侧在调用方连接配置的
	// auth_headers 注入）。用于本地工具或不可签发 JWT 的场景。
	AuthModeStaticBearer AuthMode = "static_bearer"
)

var validAuthModes = map[AuthMode]bool{
	AuthModeNone:         true,
	AuthModeKeystoneJWKS: true,
	AuthModeStaticBearer: true,
}

// Valid 返回 AuthMode 是否合法。空字符串视为合法（解析为默认值 none）。
func (m AuthMode) Valid() bool { return m == "" || validAuthModes[m] }

// Default 返回归一化后的 AuthMode：空字符串返回 AuthModeNone，否则返回自身。
// manifest 解析后 auth.mode 可能为 ""，调用端用此函数取实际值。
func (m AuthMode) Default() AuthMode {
	if m == "" {
		return AuthModeNone
	}
	return m
}
