package kstypes

import "testing"

func TestDecisionMode_IsValid(t *testing.T) {
	cases := []struct {
		mode DecisionMode
		want bool
	}{
		{DecisionModeUserOnly, true},
		{DecisionModeUserAuthorized, true},
		{DecisionModeAgentAutonomous, true},
		{"", false},
		{"unknown", false},
	}
	for _, c := range cases {
		if got := c.mode.IsValid(); got != c.want {
			t.Errorf("DecisionMode(%q).IsValid() = %v, want %v", c.mode, got, c.want)
		}
	}
}

func TestPreAuthorizeDuration_IsValid(t *testing.T) {
	cases := []struct {
		d    PreAuthorizeDuration
		want bool
	}{
		{PreAuth5m, true},
		{PreAuth30m, true},
		{PreAuth2h, true},
		{PreAuth24h, true},
		{PreAuthForever, true},
		{"", false},
		{"1h", false},
	}
	for _, c := range cases {
		if got := c.d.IsValid(); got != c.want {
			t.Errorf("PreAuthorizeDuration(%q).IsValid() = %v, want %v", c.d, got, c.want)
		}
	}
}
