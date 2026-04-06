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
