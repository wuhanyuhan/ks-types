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

func TestParseAppSpec_Readiness(t *testing.T) {
	yamlData := []byte(`
id: test-readiness-app
name: Test Readiness App
version: 1.0.0
type: app
runtime:
  mode: none
readiness:
  gates:
    - id: api_key
      kind: config
      title: 配置 API Key
      requires_secrets: [api_key]
    - id: corpus_embed
      kind: init_task
      blocking: false
      auto_init: false
      idempotent: true
      timeout_seconds: 600
`)
	m, err := ParseAppSpec(yamlData)
	if err != nil {
		t.Fatalf("ParseAppSpec: %v", err)
	}
	if len(m.Readiness.Gates) != 2 {
		t.Fatalf("gates len = %d, want 2", len(m.Readiness.Gates))
	}

	g0 := m.Readiness.Gates[0]
	if g0.ID != "api_key" || g0.Kind != ReadinessGateKindConfig {
		t.Errorf("gate0 = {%q,%q}, want {api_key,config}", g0.ID, g0.Kind)
	}
	if len(g0.RequiresSecrets) != 1 || g0.RequiresSecrets[0] != "api_key" {
		t.Errorf("gate0.RequiresSecrets = %v, want [api_key]", g0.RequiresSecrets)
	}
	// 未设置 blocking/auto_init → 默认 true
	if !g0.IsBlocking() {
		t.Errorf("gate0.IsBlocking() = false, want true (default)")
	}
	if !g0.IsAutoInit() {
		t.Errorf("gate0.IsAutoInit() = false, want true (default)")
	}

	g1 := m.Readiness.Gates[1]
	if g1.ID != "corpus_embed" || g1.Kind != ReadinessGateKindInitTask {
		t.Errorf("gate1 = {%q,%q}, want {corpus_embed,init_task}", g1.ID, g1.Kind)
	}
	// 显式 blocking:false / auto_init:false 必须被尊重
	if g1.IsBlocking() {
		t.Errorf("gate1.IsBlocking() = true, want false (explicit)")
	}
	if g1.IsAutoInit() {
		t.Errorf("gate1.IsAutoInit() = true, want false (explicit)")
	}
	if !g1.Idempotent || g1.TimeoutSeconds != 600 {
		t.Errorf("gate1 idempotent=%v timeout=%d, want true/600", g1.Idempotent, g1.TimeoutSeconds)
	}
}
