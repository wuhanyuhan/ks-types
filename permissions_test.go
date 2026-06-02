package kstypes

import (
	"fmt"
	"testing"
)

func TestPermissionDecl_Basic(t *testing.T) {
	p := PermissionDecl{
		Level:          "restricted",
		AllowedDomains: []string{"api.deepl.com"},
	}
	if p.Level != "restricted" {
		t.Errorf("level: got %q", p.Level)
	}
}

func TestValidatePermissions_Valid(t *testing.T) {
	registry := NewPermissionRegistry()
	registry.Register("network", PermissionDimension{
		DisplayName: "网络访问",
		Levels:      []string{"none", "restricted", "unrestricted"},
		RiskWeight:  5,
	})
	registry.Register("llm", PermissionDimension{
		DisplayName: "LLM 调用",
		Levels:      []string{"none", "host_proxy", "self_managed"},
		RiskWeight:  3,
	})

	perms := map[string]PermissionDecl{
		"network": {Level: "restricted", AllowedDomains: []string{"api.deepl.com"}},
		"llm":     {Level: "host_proxy"},
	}

	warnings, err := registry.Validate(perms)
	if err != nil {
		t.Errorf("validate: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestValidatePermissions_UnknownDimension(t *testing.T) {
	registry := NewPermissionRegistry()
	registry.Register("network", PermissionDimension{
		Levels: []string{"none"},
	})

	perms := map[string]PermissionDecl{
		"network":   {Level: "none"},
		"bluetooth": {Level: "on"}, // 未注册的维度
	}

	warnings, err := registry.Validate(perms)
	if err != nil {
		t.Errorf("should not error, got: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Dimension != "bluetooth" {
		t.Errorf("warning dimension: got %q", warnings[0].Dimension)
	}
}

func TestValidatePermissions_InvalidLevel(t *testing.T) {
	registry := NewPermissionRegistry()
	registry.Register("network", PermissionDimension{
		Levels: []string{"none", "restricted", "unrestricted"},
	})

	perms := map[string]PermissionDecl{
		"network": {Level: "full_access"}, // 不在合法 levels 中
	}

	_, err := registry.Validate(perms)
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestHighRiskPermissions(t *testing.T) {
	registry := NewPermissionRegistry()
	registry.Register("filesystem", PermissionDimension{
		Levels:     []string{"none", "read_scoped", "scoped", "full"},
		RiskWeight: 8,
	})

	perms := map[string]PermissionDecl{
		"filesystem": {Level: "full"},
	}

	risks := registry.HighRiskPermissions(perms, 5)
	if len(risks) != 1 {
		t.Fatalf("expected 1 high risk, got %d", len(risks))
	}
	if risks[0] != "filesystem" {
		t.Errorf("got %q", risks[0])
	}
}

func TestDefaultPermissionRegistry(t *testing.T) {
	r := DefaultPermissionRegistry()
	if r == nil {
		t.Fatal("DefaultPermissionRegistry() returned nil")
	}

	// 验证预置的 4 个维度都已注册
	expectedDims := []string{"network", "llm", "filesystem", "user_context"}
	for _, key := range expectedDims {
		dim, ok := r.dimensions[key]
		if !ok {
			t.Errorf("missing dimension %q", key)
			continue
		}
		if len(dim.Levels) == 0 {
			t.Errorf("dimension %q has no levels", key)
		}
		if dim.RiskWeight <= 0 {
			t.Errorf("dimension %q: RiskWeight should be > 0, got %d", key, dim.RiskWeight)
		}
		if dim.DisplayName == "" {
			t.Errorf("dimension %q has empty DisplayName", key)
		}
	}

	// 用一个合法的权限声明验证 Validate 无 error
	perms := map[string]PermissionDecl{
		"network": {Level: "restricted"},
		"llm":     {Level: "host_proxy"},
	}
	warnings, err := r.Validate(perms)
	if err != nil {
		t.Errorf("Validate: unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("Validate: unexpected warnings: %v", warnings)
	}
}

func TestPermissionRegistryConcurrency(t *testing.T) {
	r := DefaultPermissionRegistry()

	done := make(chan struct{})
	// 并发写：注册新维度
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			r.Register(fmt.Sprintf("dim_%d", i), PermissionDimension{
				DisplayName: fmt.Sprintf("维度%d", i),
				Levels:      []string{"none", "full"},
				RiskWeight:  i,
			})
		}
	}()

	// 并发读：校验权限
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			perms := map[string]PermissionDecl{
				"network": {Level: "restricted"},
			}
			r.Validate(perms)
		}
	}()

	// 并发读：高风险检测
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			perms := map[string]PermissionDecl{
				"filesystem": {Level: "full"},
			}
			r.HighRiskPermissions(perms, 5)
		}
	}()

	<-done
	<-done
	<-done
}
