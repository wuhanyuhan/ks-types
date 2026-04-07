package kstypes

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
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
	if m.Runtime.Mode != "container" {
		t.Errorf("runtime.mode: got %q", m.Runtime.Mode)
	}
	if m.Runtime.Image != "my-team/translator:1.2.0" {
		t.Errorf("runtime.image: got %q", m.Runtime.Image)
	}
	if m.Runtime.Port != 8080 {
		t.Errorf("port: got %d", m.Runtime.Port)
	}
	if len(m.Runtime.Ports) != 1 || m.Runtime.Ports[0] != "9090:9090" {
		t.Errorf("runtime.ports: got %v", m.Runtime.Ports)
	}
	if len(m.Runtime.Volumes) != 1 || m.Runtime.Volumes[0] != "/data/models:/models" {
		t.Errorf("runtime.volumes: got %v", m.Runtime.Volumes)
	}
	if len(m.Dependencies.Requires) != 1 || m.Dependencies.Requires[0].ID != "ks-mcp-lili" {
		t.Errorf("dependencies.requires: got %v", m.Dependencies.Requires)
	}
	if len(m.Dependencies.Recommends) != 1 || m.Dependencies.Recommends[0].ID != "ks-mcp-writer" {
		t.Errorf("dependencies.recommends: got %v", m.Dependencies.Recommends)
	}
	if len(m.Dependencies.Conflicts) != 1 || m.Dependencies.Conflicts[0].ID != "old-translator" {
		t.Errorf("dependencies.conflicts: got %v", m.Dependencies.Conflicts)
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

func TestParseRuntimeSpec_ProcessMode(t *testing.T) {
	input := `
runtime:
  mode: process
  entry: ./bin/myapp
  working_dir: /opt/app
  port: 3000
  health_check: /health
`
	type wrapper struct {
		Runtime RuntimeSpec `yaml:"runtime"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(input), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if w.Runtime.Mode != "process" {
		t.Errorf("mode: got %q", w.Runtime.Mode)
	}
	if w.Runtime.Entry != "./bin/myapp" {
		t.Errorf("entry: got %q", w.Runtime.Entry)
	}
	if w.Runtime.WorkingDir != "/opt/app" {
		t.Errorf("working_dir: got %q", w.Runtime.WorkingDir)
	}
	if w.Runtime.Port != 3000 {
		t.Errorf("port: got %d", w.Runtime.Port)
	}
}

func TestParseRuntimeSpec_ContainerMode(t *testing.T) {
	input := `
runtime:
  mode: container
  image: registry.local/myapp:latest
  port: 8080
  ports:
    - "9090:9090"
    - "9091:9091"
  volumes:
    - /host/data:/container/data
    - /host/config:/container/config
  health_check: /healthz
  resources:
    cpu: "1.0"
    memory: 1Gi
`
	type wrapper struct {
		Runtime RuntimeSpec `yaml:"runtime"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(input), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if w.Runtime.Mode != "container" {
		t.Errorf("mode: got %q", w.Runtime.Mode)
	}
	if w.Runtime.Image != "registry.local/myapp:latest" {
		t.Errorf("image: got %q", w.Runtime.Image)
	}
	if len(w.Runtime.Ports) != 2 {
		t.Errorf("ports len: got %d", len(w.Runtime.Ports))
	}
	if len(w.Runtime.Volumes) != 2 {
		t.Errorf("volumes len: got %d", len(w.Runtime.Volumes))
	}
	if w.Runtime.Resources.CPU != "1.0" {
		t.Errorf("cpu: got %q", w.Runtime.Resources.CPU)
	}
}

func TestParseDependenciesSpec(t *testing.T) {
	input := `
dependencies:
  requires:
    - id: ks-mcp-lili
      version: ">=1.0.0"
    - id: ks-mcp-document
  recommends:
    - id: ks-mcp-writer
      reason: 提升翻译润色质量
  conflicts:
    - id: old-translator
      reason: 端口冲突
`
	type wrapper struct {
		Dependencies DependenciesSpec `yaml:"dependencies"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(input), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	deps := w.Dependencies
	if len(deps.Requires) != 2 {
		t.Fatalf("requires len: got %d", len(deps.Requires))
	}
	if deps.Requires[0].ID != "ks-mcp-lili" || deps.Requires[0].Version != ">=1.0.0" {
		t.Errorf("requires[0]: got %+v", deps.Requires[0])
	}
	if deps.Requires[1].ID != "ks-mcp-document" || deps.Requires[1].Version != "" {
		t.Errorf("requires[1]: got %+v", deps.Requires[1])
	}
	if len(deps.Recommends) != 1 || deps.Recommends[0].Reason != "提升翻译润色质量" {
		t.Errorf("recommends: got %+v", deps.Recommends)
	}
	if len(deps.Conflicts) != 1 || deps.Conflicts[0].ID != "old-translator" {
		t.Errorf("conflicts: got %+v", deps.Conflicts)
	}
}

func TestManifestSpec_RoundTrip(t *testing.T) {
	orig := ManifestSpec{
		ID:      "round-trip-test",
		Name:    "Round Trip",
		Version: "2.0.0",
		Type:    AppTypeSkill,
		Runtime: RuntimeSpec{
			Mode:       "container",
			Image:      "myimage:latest",
			Port:       8080,
			Ports:      []string{"9090:9090"},
			Volumes:    []string{"/data:/data"},
			HealthCheck: "/healthz",
			Resources:  ResourcesSpec{CPU: "0.5", Memory: "256Mi"},
		},
		Dependencies: DependenciesSpec{
			Requires:   []DependencyItem{{ID: "dep-a", Version: ">=1.0.0"}},
			Recommends: []RecommendItem{{ID: "rec-b", Reason: "nice to have"}},
			Conflicts:  []ConflictItem{{ID: "conf-c", Reason: "port clash"}},
		},
	}

	data, err := yaml.Marshal(&orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.ID != orig.ID {
		t.Errorf("id: got %q want %q", parsed.ID, orig.ID)
	}
	if parsed.Runtime.Mode != "container" {
		t.Errorf("runtime.mode: got %q", parsed.Runtime.Mode)
	}
	if parsed.Runtime.Image != "myimage:latest" {
		t.Errorf("runtime.image: got %q", parsed.Runtime.Image)
	}
	if len(parsed.Runtime.Ports) != 1 || parsed.Runtime.Ports[0] != "9090:9090" {
		t.Errorf("runtime.ports: got %v", parsed.Runtime.Ports)
	}
	if len(parsed.Runtime.Volumes) != 1 || parsed.Runtime.Volumes[0] != "/data:/data" {
		t.Errorf("runtime.volumes: got %v", parsed.Runtime.Volumes)
	}
	if len(parsed.Dependencies.Requires) != 1 || parsed.Dependencies.Requires[0].ID != "dep-a" {
		t.Errorf("deps.requires: got %v", parsed.Dependencies.Requires)
	}
	if len(parsed.Dependencies.Recommends) != 1 || parsed.Dependencies.Recommends[0].ID != "rec-b" {
		t.Errorf("deps.recommends: got %v", parsed.Dependencies.Recommends)
	}
	if len(parsed.Dependencies.Conflicts) != 1 || parsed.Dependencies.Conflicts[0].ID != "conf-c" {
		t.Errorf("deps.conflicts: got %v", parsed.Dependencies.Conflicts)
	}
}

func TestParseManifest_MountExtension(t *testing.T) {
	input := `
id: ext-app
name: Ext App
version: 1.0.0
type: extension
mount:
  extension:
    mcp_server_name: ext-app
    transport_type: streamable_http
    endpoint: http://localhost:8080/mcp
    default_allowed_tools: [greet]
`
	m, err := ParseManifest([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Mount.Extension == nil {
		t.Fatal("mount.extension 不应为 nil")
	}
	if m.Mount.Extension.MCPServerName != "ext-app" {
		t.Errorf("mcp_server_name: got %q", m.Mount.Extension.MCPServerName)
	}
	if m.Mount.Extension.TransportType != "streamable_http" {
		t.Errorf("transport_type: got %q", m.Mount.Extension.TransportType)
	}
	if len(m.Mount.Extension.DefaultAllowedTools) != 1 || m.Mount.Extension.DefaultAllowedTools[0] != "greet" {
		t.Errorf("default_allowed_tools: got %v", m.Mount.Extension.DefaultAllowedTools)
	}
}

func TestParseManifest_MountService(t *testing.T) {
	input := `
id: svc-app
name: Svc App
version: 1.0.0
type: service
mount:
  service:
    auto_register_mcp: true
    mcp_endpoint: http://localhost:8080/mcp
    llm_mode: keystone_relay
    llm_requirements:
      supports_tool_calls: true
      min_context_tokens: 4000
`
	m, err := ParseManifest([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Mount.Service == nil {
		t.Fatal("mount.service 不应为 nil")
	}
	if !m.Mount.Service.AutoRegisterMCP {
		t.Error("auto_register_mcp 应为 true")
	}
	if m.Mount.Service.MCPEndpoint != "http://localhost:8080/mcp" {
		t.Errorf("mcp_endpoint: got %q", m.Mount.Service.MCPEndpoint)
	}
	if m.Mount.Service.LLMMode != "keystone_relay" {
		t.Errorf("llm_mode: got %q", m.Mount.Service.LLMMode)
	}
	if m.Mount.Service.LLMRequirements == nil {
		t.Fatal("llm_requirements 不应为 nil")
	}
	if !m.Mount.Service.LLMRequirements.SupportsToolCalls {
		t.Error("supports_tool_calls 应为 true")
	}
	if m.Mount.Service.LLMRequirements.MinContextTokens != 4000 {
		t.Errorf("min_context_tokens: got %d", m.Mount.Service.LLMRequirements.MinContextTokens)
	}
}

func TestRuntimeMode_Valid(t *testing.T) {
	cases := []struct {
		mode  RuntimeMode
		valid bool
	}{
		{RuntimeModeNone, true},
		{RuntimeModeProcess, true},
		{RuntimeModeContainer, true},
		{RuntimeMode("invalid"), false},
		{RuntimeMode(""), false},
	}
	for _, c := range cases {
		if got := c.mode.Valid(); got != c.valid {
			t.Errorf("RuntimeMode(%q).Valid() = %v, want %v", c.mode, got, c.valid)
		}
	}
}

func TestValidateManifest_InvalidRuntimeMode(t *testing.T) {
	m := &ManifestSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:    AppTypeService,
		Runtime: RuntimeSpec{Mode: "invalid"},
	}
	if err := m.Validate(); err == nil {
		t.Error("期望 runtime mode 校验失败")
	}
}

func TestValidateManifest_InvalidProtection(t *testing.T) {
	m := &ManifestSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:       AppTypeService,
		Protection: "invalid",
	}
	if err := m.Validate(); err == nil {
		t.Error("期望 protection 校验失败")
	}
}

func TestValidateManifest_ValidProtection(t *testing.T) {
	for _, p := range []string{"", "none", "preinstalled", "protected", "system"} {
		m := &ManifestSpec{
			ID: "test", Name: "test", Version: "1.0.0",
			Type:       AppTypeService,
			Protection: p,
		}
		if err := m.Validate(); err != nil {
			t.Errorf("protection %q 应通过校验: %v", p, err)
		}
	}
}

func TestValidateManifest_ExtensionMountMissingName(t *testing.T) {
	m := &ManifestSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type: AppTypeExtension,
		Mount: MountSpec{Extension: &ExtensionMountSpec{
			TransportType: "streamable_http",
			Endpoint:      "http://localhost:8080/mcp",
		}},
	}
	if err := m.Validate(); err == nil {
		t.Error("期望 extension mount 缺少 mcp_server_name 校验失败")
	}
}
