package kstypes

import "testing"

func TestReadinessGateKind_IsValid(t *testing.T) {
	cases := []struct {
		kind  ReadinessGateKind
		valid bool
	}{
		{ReadinessGateKindConfig, true},
		{ReadinessGateKindInitTask, true},
		{ReadinessGateKind(""), false},
		{ReadinessGateKind("bogus"), false},
	}
	for _, c := range cases {
		c := c
		name := string(c.kind)
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			if got := c.kind.IsValid(); got != c.valid {
				t.Errorf("ReadinessGateKind(%q).IsValid() = %v, want %v", c.kind, got, c.valid)
			}
		})
	}
}
