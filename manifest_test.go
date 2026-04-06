package kstypes

import (
	"os"
	"testing"
)

func TestParseManifest_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/valid_manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}

	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if m.ID != "my-translator" {
		t.Errorf("id: got %q", m.ID)
	}
	if m.Type != AppTypeService {
		t.Errorf("type: got %q", m.Type)
	}
	if m.Version != "1.2.0" {
		t.Errorf("version: got %q", m.Version)
	}
	if m.Compatibility.Keystone != ">=1.5.0" {
		t.Errorf("compat: got %q", m.Compatibility.Keystone)
	}
	if m.Pricing.Type != PricingFree {
		t.Errorf("pricing: got %q", m.Pricing.Type)
	}
	if m.Runtime.Port != 8080 {
		t.Errorf("port: got %d", m.Runtime.Port)
	}
	if len(m.Permissions) != 4 {
		t.Errorf("permissions count: got %d", len(m.Permissions))
	}
	if m.Permissions["network"].Level != "restricted" {
		t.Errorf("network level: got %q", m.Permissions["network"].Level)
	}
	if len(m.Permissions["network"].AllowedDomains) != 1 {
		t.Errorf("network domains: got %v", m.Permissions["network"].AllowedDomains)
	}
}

func TestParseManifest_IncompleteFieldsParseOK(t *testing.T) {
	data, _ := os.ReadFile("testdata/invalid_manifest.yaml")
	_, err := ParseManifest(data)
	if err != nil {
		t.Fatal("YAML parsing should succeed even for semantically incomplete manifests")
	}
}

func TestValidateManifest_Valid(t *testing.T) {
	data, _ := os.ReadFile("testdata/valid_manifest.yaml")
	m, _ := ParseManifest(data)
	if err := m.Validate(); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func TestValidateManifest_MissingID(t *testing.T) {
	data, _ := os.ReadFile("testdata/invalid_manifest.yaml")
	m, _ := ParseManifest(data)
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for missing id")
	}
}

func TestValidateManifest_InvalidType(t *testing.T) {
	m := &ManifestSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type: AppType("invalid"),
	}
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for invalid type")
	}
}

func TestValidateManifest_InvalidPricing(t *testing.T) {
	m := &ManifestSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:    AppTypeService,
		Pricing: PricingSpec{Type: PricingType("invalid")},
	}
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for invalid pricing")
	}
}
