package kstypes

import "testing"

func TestNavTreeValidate(t *testing.T) {
	ok := NavTree{Items: []NavItem{
		{Key: "dashboard", Label: "总览", Kind: NavKindSDUI},
		{Key: "brand", Label: "品牌手册", Kind: NavKindIsland, Path: "/brand"},
	}}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid tree rejected: %v", err)
	}
	cases := map[string]NavTree{
		"empty":         {Items: nil},
		"island-nopath": {Items: []NavItem{{Key: "x", Label: "X", Kind: NavKindIsland}}},
		"bad-kind":      {Items: []NavItem{{Key: "x", Label: "X", Kind: "weird"}}},
		"dup-key":       {Items: []NavItem{{Key: "x", Label: "A", Kind: NavKindSDUI}, {Key: "x", Label: "B", Kind: NavKindSDUI}}},
		"no-label":      {Items: []NavItem{{Key: "x", Kind: NavKindSDUI}}},
	}
	for name, tree := range cases {
		if err := tree.Validate(); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}
