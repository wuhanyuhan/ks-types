package kstypes

import "testing"

func TestLLMTier_Valid(t *testing.T) {
	cases := []struct {
		name string
		tier LLMTier
		want bool
	}{
		{"economy", LLMTierEconomy, true},
		{"standard", LLMTierStandard, true},
		{"flagship", LLMTierFlagship, true},
		{"empty", LLMTier(""), false}, // 空串不合法：调用方应省略 tier 键取默认，而非传空串
		{"unknown", LLMTier("turbo"), false},
		{"uppercase", LLMTier("ECONOMY"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.tier.Valid(); got != c.want {
				t.Errorf("LLMTier(%q).Valid() = %v, want %v", c.tier, got, c.want)
			}
		})
	}
}

func TestLLMCapability_Valid(t *testing.T) {
	cases := []struct {
		name string
		cap  LLMCapability
		want bool
	}{
		{"vision", LLMCapabilityVision, true},
		{"empty", LLMCapability(""), false},
		{"unimplemented_long_context", LLMCapability("long_context"), false}, // 留枚举扩展位、尚未实现路由
		{"unimplemented_tool_use", LLMCapability("tool_use"), false},
		{"uppercase", LLMCapability("VISION"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cap.Valid(); got != c.want {
				t.Errorf("LLMCapability(%q).Valid() = %v, want %v", c.cap, got, c.want)
			}
		})
	}
}

func TestReasoningMode_Valid(t *testing.T) {
	cases := []struct {
		name string
		mode ReasoningMode
		want bool
	}{
		{"on", ReasoningModeOn, true},
		{"off", ReasoningModeOff, true},
		{"auto", ReasoningModeAuto, true},
		{"empty", ReasoningMode(""), false}, // 空串不合法：不声明时调用方应省略 reasoning 键
		{"unknown", ReasoningMode("medium"), false},
		{"uppercase", ReasoningMode("ON"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.mode.Valid(); got != c.want {
				t.Errorf("ReasoningMode(%q).Valid() = %v, want %v", c.mode, got, c.want)
			}
		})
	}
}
