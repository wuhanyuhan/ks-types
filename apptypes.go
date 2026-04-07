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

// Valid 检查 RuntimeMode 是否合法
func (m RuntimeMode) Valid() bool { return validRuntimeModes[m] }
