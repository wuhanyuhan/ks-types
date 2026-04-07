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
