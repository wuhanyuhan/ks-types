package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
)

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

func TestReadinessSpec_Validate(t *testing.T) {
	blockingFalse := false
	cases := []struct {
		name  string
		spec  ReadinessSpec
		valid bool
	}{
		{
			name:  "empty gates ok",
			spec:  ReadinessSpec{},
			valid: true,
		},
		{
			name: "valid config gate",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "api_key", Kind: ReadinessGateKindConfig, RequiresSecrets: []string{"api_key"}},
			}},
			valid: true,
		},
		{
			name: "valid init_task gate",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "corpus_embed", Kind: ReadinessGateKindInitTask, TimeoutSeconds: 600},
			}},
			valid: true,
		},
		{
			name: "missing id",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{Kind: ReadinessGateKindConfig, RequiresSecrets: []string{"k"}},
			}},
			valid: false,
		},
		{
			name: "bad id format",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "Bad ID", Kind: ReadinessGateKindConfig, RequiresSecrets: []string{"k"}},
			}},
			valid: false,
		},
		{
			name: "duplicate id",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "g1", Kind: ReadinessGateKindConfig, RequiresSecrets: []string{"k"}},
				{ID: "g1", Kind: ReadinessGateKindInitTask},
			}},
			valid: false,
		},
		{
			name: "invalid kind",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "g1", Kind: ReadinessGateKind("bogus")},
			}},
			valid: false,
		},
		{
			name: "config gate without requires",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "g1", Kind: ReadinessGateKindConfig},
			}},
			valid: false,
		},
		{
			name: "init_task negative timeout",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "g1", Kind: ReadinessGateKindInitTask, TimeoutSeconds: -1},
			}},
			valid: false,
		},
		{
			name: "blocking false is fine",
			spec: ReadinessSpec{Gates: []ReadinessGate{
				{ID: "g1", Kind: ReadinessGateKindInitTask, Blocking: &blockingFalse},
			}},
			valid: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := c.spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("Validate() valid=%v want %v (err=%v)", gotValid, c.valid, err)
			}
		})
	}
}

func TestAppSpec_Validate_SurfacesReadinessError(t *testing.T) {
	spec := &AppSpec{
		ID: "test-app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
		Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
		Provides: ProvidesSpec{Capabilities: []CapabilitySpec{
			{Name: "x", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "x"}},
		}},
		Readiness: ReadinessSpec{Gates: []ReadinessGate{
			{ID: "g1", Kind: ReadinessGateKindConfig}, // 缺 requires_* → 非法
		}},
	}
	err := spec.Validate()
	if err == nil {
		t.Fatal("expected AppSpec.Validate() to surface readiness error, got nil")
	}
	if !strings.Contains(err.Error(), "readiness") {
		t.Errorf("error should mention readiness, got: %v", err)
	}
}

func TestReadinessGateStatus_IsValid(t *testing.T) {
	cases := []struct {
		status ReadinessGateStatus
		valid  bool
	}{
		{ReadinessGateStatusPending, true},
		{ReadinessGateStatusRunning, true},
		{ReadinessGateStatusReady, true},
		{ReadinessGateStatusFailed, true},
		{ReadinessGateStatus(""), false},
		{ReadinessGateStatus("bogus"), false},
	}
	for _, c := range cases {
		if got := c.status.IsValid(); got != c.valid {
			t.Errorf("ReadinessGateStatus(%q).IsValid() = %v, want %v", c.status, got, c.valid)
		}
	}
}

func TestReadinessReport_JSONShape(t *testing.T) {
	// ready 门无 progress/message -> omitempty 应省略，锁定跨语言 wire 形状。
	readyOnly, err := json.Marshal(ReadinessGateState{ID: "warm_cache", Status: ReadinessGateStatusReady})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(readyOnly) != `{"id":"warm_cache","status":"ready"}` {
		t.Errorf("ready gate JSON = %s, want omitempty progress/message", readyOnly)
	}

	p := 42
	orig := ReadinessReport{Gates: []ReadinessGateState{
		{ID: "corpus_embed", Status: ReadinessGateStatusRunning, Progress: &p, Message: "已嵌入 1200/2900 条"},
		{ID: "warm_cache", Status: ReadinessGateStatusReady},
	}}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `{"gates":[{"id":"corpus_embed","status":"running","progress":42,"message":"已嵌入 1200/2900 条"},{"id":"warm_cache","status":"ready"}]}` {
		t.Errorf("report JSON = %s, want exact gates wire shape", data)
	}
	var got ReadinessReport
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Gates) != 2 || got.Gates[0].ID != "corpus_embed" || got.Gates[0].Progress == nil || *got.Gates[0].Progress != 42 {
		t.Fatalf("roundtrip mismatch: %#v", got)
	}

	initData, err := json.Marshal(ReadinessInitRequest{GateID: "corpus_embed"})
	if err != nil {
		t.Fatalf("marshal init req: %v", err)
	}
	if string(initData) != `{"gate_id":"corpus_embed"}` {
		t.Errorf("ReadinessInitRequest JSON = %s, want gate_id wire key", initData)
	}

	var ir ReadinessInitRequest
	if err := json.Unmarshal([]byte(`{"gate_id":"corpus_embed"}`), &ir); err != nil {
		t.Fatalf("unmarshal init req: %v", err)
	}
	if ir.GateID != "corpus_embed" {
		t.Errorf("ReadinessInitRequest.GateID = %q, want corpus_embed", ir.GateID)
	}
}
