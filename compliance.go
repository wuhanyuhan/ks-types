package kstypes

// 决策模式枚举（被 CapabilitySpec.decision_mode 复用；app 级 compliance 段已砍除）。
//
// 旧 AppSpec.Compliance（default_decision_mode + tool_overrides）已合并为 CapabilitySpec
// 上的可选内联 decision_mode（默认从 side_effect_level 派生，见 manifest.go
// EffectiveDecisionMode）。这里只保留三级决策语义枚举本身，供契约与 keystone 复用。
//
// 政策依据：网信办 / 国家发改委 / 工信部《智能体规范应用与创新发展实施意见》（2026-05-08）—
// "确保用户对智能体自主决策享有知情权和最终决策权"。中国 AI 合规三级语义即由此承载。

// DecisionMode 智能体决策权限三级（中国 AI 合规三级语义的真值源）。
//
// 与 keystone 平台侧的 DecisionMode 字符串值一致，跨仓只共享枚举语义。
type DecisionMode string

const (
	// DecisionModeUserOnly 每次调用都需用户当场点确认（最高管控；side_effect=hard_irreversible 的派生默认）。
	DecisionModeUserOnly DecisionMode = "user_only"
	// DecisionModeUserAuthorized 用户预授权后，pre_authorize_duration 时段内自动放行
	//（side_effect=soft_reversible 的派生默认）。
	DecisionModeUserAuthorized DecisionMode = "user_authorized"
	// DecisionModeAgentAutonomous 智能体可自主调用、无需人介入（最低管控；side_effect=none 的派生默认）。
	DecisionModeAgentAutonomous DecisionMode = "agent_autonomous"
)

// IsValid 判断 DecisionMode 是否在枚举集合中。
func (m DecisionMode) IsValid() bool {
	switch m {
	case DecisionModeUserOnly, DecisionModeUserAuthorized, DecisionModeAgentAutonomous:
		return true
	}
	return false
}

// PreAuthorizeDuration 预授权时长枚举。
//
// 仅在 DecisionMode = user_authorized 模式下有意义：用户在弹窗里选择"自动放行 N 时长"，
// 平台侧入库后，N 时段内同 (user, agent, tool) 调用直接通过。
type PreAuthorizeDuration string

const (
	PreAuth5m      PreAuthorizeDuration = "5m"
	PreAuth30m     PreAuthorizeDuration = "30m"
	PreAuth2h      PreAuthorizeDuration = "2h"
	PreAuth24h     PreAuthorizeDuration = "24h"
	PreAuthForever PreAuthorizeDuration = "forever"
)

// IsValid 判断 PreAuthorizeDuration 是否在枚举集合中。
func (d PreAuthorizeDuration) IsValid() bool {
	switch d {
	case PreAuth5m, PreAuth30m, PreAuth2h, PreAuth24h, PreAuthForever:
		return true
	}
	return false
}
