package kstypes

import "testing"

func TestParseWidgetURI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		input      string
		wantErr    bool
		wantScheme WidgetURIScheme
		wantName   string
		wantVer    string
	}{
		{"shared-valid", "ks://widgets/diff-review@v1", false, WidgetSchemeKS, "diff-review", "v1"},
		{"shared-with-major", "ks://widgets/list-actions@v3", false, WidgetSchemeKS, "list-actions", "v3"},
		{"custom-valid", "ui://marketing/brand-editor", false, WidgetSchemeCustom, "marketing/brand-editor", ""},
		{"custom-nested-path", "ui://legal/contract/diff-board", false, WidgetSchemeCustom, "legal/contract/diff-board", ""},
		{"shared-no-version", "ks://widgets/diff-review", true, "", "", ""},
		{"shared-bad-version", "ks://widgets/diff-review@v1.0", true, "", "", ""},
		{"shared-no-widgets-prefix", "ks://other/diff-review@v1", true, "", "", ""},
		{"unknown-scheme", "npm://x/y", true, "", "", ""},
		{"empty", "", true, "", "", ""},
		{"with-fragment", "ks://widgets/diff-review@v1#x", true, "", "", ""},
		{"with-query", "ks://widgets/diff-review@v1?x=1", true, "", "", ""},
		{"custom-empty-path", "ui://marketing/", true, "", "", ""},
		{"custom-no-squad", "ui:///path", true, "", "", ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			uri, err := ParseWidgetURI(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil with uri=%+v", uri)
				}
				if uri != nil {
					t.Errorf("expected nil uri on error, got %+v", uri)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if uri == nil {
				t.Fatal("expected non-nil uri")
			}
			if uri.Scheme != c.wantScheme {
				t.Errorf("scheme: got %q, want %q", uri.Scheme, c.wantScheme)
			}
			if uri.Name != c.wantName {
				t.Errorf("name: got %q, want %q", uri.Name, c.wantName)
			}
			if uri.Version != c.wantVer {
				t.Errorf("version: got %q, want %q", uri.Version, c.wantVer)
			}
			if uri.Raw != c.input {
				t.Errorf("raw: got %q, want %q", uri.Raw, c.input)
			}
		})
	}
}

func TestWidgetURI_SquadID_Path(t *testing.T) {
	t.Parallel()
	uri, err := ParseWidgetURI("ui://marketing/brand-editor")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := uri.SquadID(); got != "marketing" {
		t.Errorf("SquadID: got %q, want marketing", got)
	}
	if got := uri.Path(); got != "brand-editor" {
		t.Errorf("Path: got %q, want brand-editor", got)
	}

	nested, err := ParseWidgetURI("ui://legal/contract/diff-board")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := nested.SquadID(); got != "legal" {
		t.Errorf("nested SquadID: got %q", got)
	}
	if got := nested.Path(); got != "contract/diff-board" {
		t.Errorf("nested Path: got %q", got)
	}

	ks, err := ParseWidgetURI("ks://widgets/diff-review@v1")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ks.SquadID() != "" || ks.Path() != "" {
		t.Errorf("ks-scheme should have empty SquadID/Path, got %q/%q", ks.SquadID(), ks.Path())
	}
}
