package kstypes

import "testing"

func TestCheckNavConfigConsistency(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		navState    NavState
		openMode    string
		configMode  string
		hasConfigUI bool
		wantOK      bool
	}{
		{"absent_none", NavAbsent, "", "none", false, true},
		{"absent_empty", NavAbsent, "", "", false, true},
		{"absent_schema", NavAbsent, "", "schema", false, false},
		{"absent_iframe", NavAbsent, "", "iframe", false, false},
		{"invalid", NavInvalid, "", "schema", false, false},
		{"dialog_schema", NavValid, "dialog", "schema", false, true},
		{"dialog_iframe_with_ui", NavValid, "dialog", "iframe", true, true},
		{"dialog_iframe_no_ui", NavValid, "dialog", "iframe", false, false},
		{"dialog_none_dead", NavValid, "dialog", "none", false, false},
		{"fullpage_none", NavValid, "fullpage", "none", false, true},
		{"fullpage_empty_defaults_none", NavValid, "fullpage", "", false, true},
		{"fullpage_iframe_illegal", NavValid, "fullpage", "iframe", false, false},
		{"fullpage_schema_dead", NavValid, "fullpage", "schema", false, false},
		{"tab_none", NavValid, "tab", "none", false, true},
		{"tab_iframe_illegal", NavValid, "tab", "iframe", false, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reason, ok := CheckNavConfigConsistency(tc.navState, tc.openMode, tc.configMode, tc.hasConfigUI)
			if ok != tc.wantOK {
				t.Fatalf("ok=%v, want %v (reason=%q)", ok, tc.wantOK, reason)
			}
			if !ok && reason == "" {
				t.Errorf("不一致时 reason 不应为空")
			}
			if ok && reason != "" {
				t.Errorf("一致时 reason 应为空，得 %q", reason)
			}
		})
	}
}
