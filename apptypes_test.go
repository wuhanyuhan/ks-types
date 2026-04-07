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
