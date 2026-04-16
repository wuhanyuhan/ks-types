package kstypes

import "testing"

func TestAppTypeValid(t *testing.T) {
	valid := []AppType{AppTypeService, AppTypeSkill, AppTypeAssistant, AppTypeExtension}
	for _, at := range valid {
		if !at.Valid() {
			t.Errorf("expected %q to be valid", at)
		}
	}
}

func TestAppTypeInvalid(t *testing.T) {
	invalid := AppType("unknown")
	if invalid.Valid() {
		t.Error("expected unknown AppType to be invalid")
	}
}

func TestPricingTypeValid(t *testing.T) {
	valid := []PricingType{PricingFree, PricingPaid, PricingFreemium}
	for _, pt := range valid {
		if !pt.Valid() {
			t.Errorf("expected %q to be valid", pt)
		}
	}
}

func TestPricingTypeInvalid(t *testing.T) {
	invalid := PricingType("unknown")
	if invalid.Valid() {
		t.Error("expected unknown PricingType to be invalid")
	}
}

func TestProtectionLevelValid(t *testing.T) {
	valid := []ProtectionLevel{
		"", ProtectionNone, ProtectionPreinstalled, ProtectionProtected, ProtectionSystem,
	}
	for _, p := range valid {
		if !p.Valid() {
			t.Errorf("expected %q to be valid", p)
		}
	}
}

func TestProtectionLevelInvalid(t *testing.T) {
	invalid := ProtectionLevel("unknown")
	if invalid.Valid() {
		t.Error("expected unknown ProtectionLevel to be invalid")
	}
}

func TestRuntimeModeEmptyIsValid(t *testing.T) {
	empty := RuntimeMode("")
	if !empty.Valid() {
		t.Error("expected empty RuntimeMode to be valid")
	}
}

func TestAuthMode_Valid(t *testing.T) {
	cases := []struct {
		name string
		mode AuthMode
		want bool
	}{
		{"none", AuthModeNone, true},
		{"keystone_jwks", AuthModeKeystoneJWKS, true},
		{"static_bearer", AuthModeStaticBearer, true},
		{"empty", AuthMode(""), true}, // 空字符串合法（表示取默认值 none）
		{"invalid", AuthMode("invalid_mode"), false},
		{"uppercase", AuthMode("KEYSTONE_JWKS"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.mode.Valid(); got != c.want {
				t.Errorf("AuthMode(%q).Valid() = %v, want %v", c.mode, got, c.want)
			}
		})
	}
}

func TestAuthMode_Default(t *testing.T) {
	if AuthMode("").Default() != AuthModeNone {
		t.Errorf("empty AuthMode 的 Default 应为 none")
	}
	if AuthModeKeystoneJWKS.Default() != AuthModeKeystoneJWKS {
		t.Errorf("非空 AuthMode 的 Default 应返回自身")
	}
}
