package kstypes

// AppType 应用类型枚举
type AppType string

const (
	AppTypeService   AppType = "service"
	AppTypeSkill     AppType = "skill"
	AppTypeAssistant AppType = "assistant"
	AppTypeExtension AppType = "extension"
)

var validAppTypes = map[AppType]bool{
	AppTypeService: true, AppTypeSkill: true,
	AppTypeAssistant: true, AppTypeExtension: true,
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
)

var validRuntimeModes = map[RuntimeMode]bool{
	RuntimeModeNone: true, RuntimeModeProcess: true, RuntimeModeContainer: true,
}

// Valid 检查 RuntimeMode 是否合法；空值视为合法（等同 none）
func (m RuntimeMode) Valid() bool { return m == "" || validRuntimeModes[m] }

// ProtectionLevel 保护级别
type ProtectionLevel string

const (
	ProtectionNone         ProtectionLevel = "none"
	ProtectionPreinstalled ProtectionLevel = "preinstalled"
	ProtectionProtected    ProtectionLevel = "protected"
	ProtectionSystem       ProtectionLevel = "system"
)

var validProtectionLevels = map[ProtectionLevel]bool{
	ProtectionNone: true, ProtectionPreinstalled: true,
	ProtectionProtected: true, ProtectionSystem: true,
}

// Valid 检查 ProtectionLevel 是否合法；空值视为合法（等同 none）
func (p ProtectionLevel) Valid() bool {
	return p == "" || validProtectionLevels[p]
}

// AuthMode 描述 service 类型应用的 /mcp 端点鉴权模式。
//
// 默认值：空字符串等价于 AuthModeNone（由 Default() 统一）。
// 声明位置：manifest.yaml 的 mount.service.auth_mode 字段。
type AuthMode string

const (
	// AuthModeNone /mcp 端点不做鉴权，依赖网络边界。
	// 仅在受控内网 + keystone 是唯一调用方时可用。
	AuthModeNone AuthMode = "none"

	// AuthModeKeystoneJWKS service 通过 keystone /.well-known/jwks.json 公钥
	// 验证调用者 Authorization 头的 RS256 JWT（推荐，strict-by-default）。
	AuthModeKeystoneJWKS AuthMode = "keystone_jwks"

	// AuthModeStaticBearer service 比对静态 Bearer token（调用方在 t_mcp_servers.
	// connection_config.auth_headers 注入）。用于本地工具或不可签发 JWT 的场景。
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
// manifest 解析后 ServiceMountSpec.AuthMode 可能为 ""，调用端用此函数取实际值。
func (m AuthMode) Default() AuthMode {
	if m == "" {
		return AuthModeNone
	}
	return m
}
