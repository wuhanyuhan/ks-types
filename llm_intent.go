package kstypes

// LLM intent 词表：调用方声明抽象意图（tier + capability + reasoning），由 keystone relay
// 翻译成现场具体 model。三者正交——tier 是可降级偏好、capability 是硬约束、reasoning 是开关。
// 本文件是全生态（keystone / ks-devkit SDK / 各 MCP app）共用的字符串值权威来源。

// LLMTier 表示调用方对 LLM「成本-能力梯度」的抽象偏好。
// 偏好语义：现场缺对应档时 keystone 可见降级到 standard（不报错）。
type LLMTier string

const (
	// LLMTierEconomy 省钱档：低成本模型，适合 query 改写等轻任务。
	LLMTierEconomy LLMTier = "economy"
	// LLMTierStandard 标准档（默认，等同现状通用档）：不声明 tier 时的回落档，永远存在。
	LLMTierStandard LLMTier = "standard"
	// LLMTierFlagship 旗舰档：强指令遵循模型，适合 listwise rerank 等重任务。
	LLMTierFlagship LLMTier = "flagship"
)

var validLLMTiers = map[LLMTier]bool{
	LLMTierEconomy: true, LLMTierStandard: true, LLMTierFlagship: true,
}

// Valid 检查 LLMTier 是否为已定义档位。空串不合法（调用方应省略 tier 键以取默认 standard）。
func (t LLMTier) Valid() bool { return validLLMTiers[t] }

// LLMCapability 表示调用方对 LLM 的「硬能力」要求（如 vision）。
// 与 tier 不同：capability 是硬约束——现场无满足该能力的模型时 keystone 严格报错（422
// capability_unavailable），由调用方自行降级，绝不偷偷路由到不具备该能力的模型。
// 初版只实现 vision；long_context / tool_use 等待 keystone 实现对应路由后再加入本词表。
type LLMCapability string

const (
	// LLMCapabilityVision 图片理解能力（VL 模型），用于 describe_image 等多模态调用。
	LLMCapabilityVision LLMCapability = "vision"
)

var validLLMCapabilities = map[LLMCapability]bool{
	LLMCapabilityVision: true,
}

// Valid 检查 LLMCapability 是否为本词表已实现的能力。
func (c LLMCapability) Valid() bool { return validLLMCapabilities[c] }

// ReasoningMode 表示 LLM 思考开关，与 tier / capability 正交。
// 复用 keystone 现有 reasoning_mode 语义（on/off/auto）。不声明时不干预（沿用 plan/model 默认）。
// 注：keystone 平台侧另有同值的 ReasoningMode（on/off），属各自包定义，
// 跨包只共享字符串语义；本枚举是生态权威，额外提供 auto。
type ReasoningMode string

const (
	// ReasoningModeOn 强制开启思考。
	ReasoningModeOn ReasoningMode = "on"
	// ReasoningModeOff 强制关闭思考。
	ReasoningModeOff ReasoningMode = "off"
	// ReasoningModeAuto 由模型 / plan 自行决定是否思考。
	ReasoningModeAuto ReasoningMode = "auto"
)

var validReasoningModes = map[ReasoningMode]bool{
	ReasoningModeOn: true, ReasoningModeOff: true, ReasoningModeAuto: true,
}

// Valid 检查 ReasoningMode 是否合法。空串不合法（不声明时调用方应省略该键）。
func (m ReasoningMode) Valid() bool { return validReasoningModes[m] }
