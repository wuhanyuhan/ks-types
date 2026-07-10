package kstypes

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseAppSpec_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/valid_manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}

	m, err := ParseAppSpec(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if m.ID != "my-translator" {
		t.Errorf("id: got %q", m.ID)
	}
	if m.Type != AppTypeApp {
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
	if m.Runtime.Mode != RuntimeModeContainer {
		t.Errorf("runtime.mode: got %q", m.Runtime.Mode)
	}
	if m.Runtime.Image != "my-team/translator:1.2.0" {
		t.Errorf("runtime.image: got %q", m.Runtime.Image)
	}
	if len(m.Runtime.Volumes) != 1 || m.Runtime.Volumes[0] != "/data/models:/models" {
		t.Errorf("runtime.volumes: got %v", m.Runtime.Volumes)
	}
	if m.Store.Presentation != StorePresentationServiceApp {
		t.Errorf("store.presentation: got %q", m.Store.Presentation)
	}
	if len(m.Store.Highlights) != 1 || m.Store.Highlights[0] != "提供文章生成和封面配图能力" {
		t.Errorf("store.highlights: got %v", m.Store.Highlights)
	}
	if len(m.Store.TryPrompts) != 1 || m.Store.TryPrompts[0] != "帮我写一篇关于 AI Agent 的文章" {
		t.Errorf("store.try_prompts: got %v", m.Store.TryPrompts)
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

	// provides.capabilities[]（v0.19.0 capability mesh）
	if len(m.Provides.Capabilities) != 1 {
		t.Fatalf("provides.capabilities len: got %d, want 1", len(m.Provides.Capabilities))
	}
	cap0 := m.Provides.Capabilities[0]
	if cap0.Name != "translate" {
		t.Errorf("provides.capabilities[0].name: got %q", cap0.Name)
	}
	if cap0.ExecutionMode != "sync" {
		t.Errorf("provides.capabilities[0].execution_mode: got %q", cap0.ExecutionMode)
	}
	if cap0.Backend.Kind != "mcp_tool" || cap0.Backend.ToolName != "translate" {
		t.Errorf("provides.capabilities[0].backend: got %+v", cap0.Backend)
	}
	if cap0.NotifyPolicy.OnFailed == nil || !*cap0.NotifyPolicy.OnFailed {
		t.Errorf("provides.capabilities[0].notify_policy.on_failed 应为 true: got %v", cap0.NotifyPolicy.OnFailed)
	}

	// requires.capabilities[]（required + recommended 各一）
	if len(m.Requires.Capabilities) != 2 {
		t.Fatalf("requires.capabilities len: got %d, want 2", len(m.Requires.Capabilities))
	}
	if m.Requires.Capabilities[0].CanonicalName != "ks-mcp-lili.search" || m.Requires.Capabilities[0].Mode != "required" {
		t.Errorf("requires.capabilities[0]: got %+v", m.Requires.Capabilities[0])
	}
	if m.Requires.Capabilities[1].CanonicalName != "ks-mcp-writer.polish" || m.Requires.Capabilities[1].Mode != "recommended" {
		t.Errorf("requires.capabilities[1]: got %+v", m.Requires.Capabilities[1])
	}

	// conflicts.apps[]
	if len(m.Conflicts.Apps) != 1 || m.Conflicts.Apps[0].ID != "old-translator" {
		t.Errorf("conflicts.apps: got %+v", m.Conflicts.Apps)
	}

	// public_http.paths[]
	if len(m.PublicHTTP.Paths) != 2 || m.PublicHTTP.Paths[0] != "/healthz" || m.PublicHTTP.Paths[1] != "/api/publisher/plugin/*" {
		t.Errorf("public_http.paths: got %+v", m.PublicHTTP.Paths)
	}
}

func TestParseAppSpec_IncompleteFieldsParseOK(t *testing.T) {
	data, _ := os.ReadFile("testdata/invalid_manifest.yaml")
	_, err := ParseAppSpec(data)
	if err != nil {
		t.Fatal("YAML parsing should succeed even for semantically incomplete manifests")
	}
}

func TestValidateManifest_Valid(t *testing.T) {
	data, _ := os.ReadFile("testdata/valid_manifest.yaml")
	m, _ := ParseAppSpec(data)
	if err := m.Validate(); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func TestValidateManifest_MissingID(t *testing.T) {
	data, _ := os.ReadFile("testdata/invalid_manifest.yaml")
	m, _ := ParseAppSpec(data)
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for missing id")
	}
}

func TestValidateManifest_InvalidType(t *testing.T) {
	m := &AppSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type: AppType("invalid"),
	}
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for invalid type")
	}
}

func TestValidateManifest_InvalidPricing(t *testing.T) {
	m := &AppSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:    AppTypeApp,
		Pricing: PricingSpec{Type: PricingType("invalid")},
	}
	err := m.Validate()
	if err == nil {
		t.Error("expected validation error for invalid pricing")
	}
}

func TestAppSpec_VersionAndIDValidation(t *testing.T) {
	base := func() *AppSpec {
		return &AppSpec{ID: "my-app", Name: "n", Version: "1.2.0", Type: AppTypeApp}
	}
	require.NoError(t, base().Validate())

	// 合法区间（双边）应通过格式校验
	ok := base()
	ok.Compatibility.Keystone = ">=1.5.0 <2.0.0"
	require.NoError(t, ok.Validate())

	bad := base()
	bad.Version = "v1.2" // 非 semver
	assert.ErrorContains(t, bad.Validate(), "version")

	bad = base()
	bad.ID = "My_App" // 非法 id（大写字母 + 下划线）
	assert.ErrorContains(t, bad.Validate(), "id")

	bad = base()
	bad.Compatibility.Keystone = "1.5" // 非区间格式（缺第三段）
	assert.ErrorContains(t, bad.Validate(), "compatibility")
}

func TestAppSpecValidate_StorePresentationExpertTeamRequiresMembers(t *testing.T) {
	spec := AppSpec{
		ID:      "ks-mcp-squad-marketing",
		Name:    "Marketing Squad · 市场团队",
		Version: "0.1.0",
		Type:    AppTypeApp,
		Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
		Store: StoreSpec{
			Presentation: StorePresentationExpertTeam,
			Team: &StoreTeamSpec{
				LeadRole: "marketing-lead",
				Members: []StoreTeamMemberSpec{
					{Key: "market-researcher", Name: "市场研究员", Title: "负责趋势、竞品和用户洞察", Avatar: "avatars/market-researcher.png"},
				},
			},
		},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("expert_team with member should pass: %v", err)
	}

	spec.Store.Team = nil
	if err := spec.Validate(); err == nil {
		t.Fatal("expert_team without team should fail")
	}
}

func TestAppSpecValidate_RejectsUnknownStorePresentation(t *testing.T) {
	spec := AppSpec{
		ID:      "agent-backend-engineer",
		Name:    "后端工程师",
		Version: "1.0.0",
		Type:    AppTypeAgent,
		Runtime: RuntimeSpec{Mode: RuntimeModeNone},
		Store:   StoreSpec{Presentation: StorePresentation("squad")},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("unknown store.presentation should fail")
	}
}

func TestParseAppSpec_ManagedMySQLResource(t *testing.T) {
	input := `
id: db-backed-app
name: DB Backed App
version: 1.0.0
type: app
managed_resources:
  mysql:
    retain_on_uninstall: true
    inject:
      host: DB_HOST
      port: DB_PORT
      database: DB_NAME
      user: DB_USER
      password: DB_PASSWORD
`
	m, err := ParseAppSpec([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.ManagedResources.MySQL == nil {
		t.Fatal("managed_resources.mysql 不应为 nil")
	}
	mysql := m.ManagedResources.MySQL
	if !mysql.RetainOnUninstall {
		t.Error("retain_on_uninstall 应为 true")
	}
	if mysql.Inject.Database != "DB_NAME" || mysql.Inject.Password != "DB_PASSWORD" {
		t.Fatalf("unexpected inject mapping: %+v", mysql.Inject)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestParseAppSpec_ManagedResourceSuite(t *testing.T) {
	input := `
id: resource-heavy-app
name: Resource Heavy App
version: 1.0.0
type: app
managed_resources:
  object_storage:
    access: private
    retain_on_uninstall: true
    inject:
      endpoint: STORAGE_ENDPOINT
      bucket: STORAGE_BUCKET
      prefix: STORAGE_PREFIX
      access_key: STORAGE_ACCESS_KEY
      secret_key: STORAGE_SECRET_KEY
  vector_store:
    retain_on_uninstall: true
  storage:
    files:
      size_mb: 1024
      retain_on_uninstall: true
    logs:
      size_mb: 256
      retain_on_uninstall: true
  cache:
    inject:
      url: CACHE_URL
      key_prefix: CACHE_KEY_PREFIX
managed_secrets:
  items:
    - name: preview_hmac
      generate: random_hex_32
      inject: HMAC_SECRET
platform_services:
  embedding:
    required: true
    model: bge-m3
    inject:
      model: KS_EMBEDDING_MODEL
      dim: KS_EMBEDDING_DIM
`
	m, err := ParseAppSpec([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.ManagedResources.ObjectStorage == nil || m.ManagedResources.ObjectStorage.Inject.Prefix != "STORAGE_PREFIX" {
		t.Fatalf("object_storage 未正确解析: %+v", m.ManagedResources.ObjectStorage)
	}
	if m.ManagedResources.VectorStore == nil || !m.ManagedResources.VectorStore.RetainOnUninstall {
		t.Fatalf("vector_store 未正确解析: %+v", m.ManagedResources.VectorStore)
	}
	if m.ManagedResources.Storage == nil || m.ManagedResources.Storage.Files == nil || m.ManagedResources.Storage.Logs == nil {
		t.Fatalf("storage 未正确解析: %+v", m.ManagedResources.Storage)
	}
	if m.ManagedResources.Storage.Files.SizeMB != 1024 || !m.ManagedResources.Storage.Files.RetainOnUninstall {
		t.Fatalf("storage.files 未正确解析: %+v", m.ManagedResources.Storage.Files)
	}
	if m.ManagedResources.Cache == nil || m.ManagedResources.Cache.Inject.KeyPrefix != "CACHE_KEY_PREFIX" {
		t.Fatalf("cache 未正确解析: %+v", m.ManagedResources.Cache)
	}
	if len(m.ManagedSecrets.Items) != 1 || m.ManagedSecrets.Items[0].Inject != "HMAC_SECRET" {
		t.Fatalf("managed_secrets 未正确解析: %+v", m.ManagedSecrets)
	}
	if m.PlatformServices.Embedding == nil || m.PlatformServices.Embedding.Model != "bge-m3" {
		t.Fatalf("platform_services.embedding 未正确解析: %+v", m.PlatformServices.Embedding)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateManifest_ManagedStorageRequiresAtLeastOneScope(t *testing.T) {
	m := &AppSpec{
		ID: "storage-app", Name: "Storage App", Version: "1.0.0",
		Type: AppTypeApp,
		ManagedResources: ManagedResourcesSpec{
			Storage: &ManagedStorageResourceSpec{},
		},
	}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty managed storage")
	}
}

func TestValidateManifest_ManagedSecretRequiresGenerateAndInject(t *testing.T) {
	m := &AppSpec{
		ID:      "test",
		Name:    "test",
		Version: "1.0.0",
		Type:    AppTypeApp,
		ManagedSecrets: ManagedSecretsSpec{Items: []ManagedSecretSpec{{
			Name:     "preview_hmac",
			Generate: "random_hex_32",
		}}},
	}
	if err := m.Validate(); err == nil {
		t.Error("expected validation error for missing managed secret inject")
	}
}

func TestValidateManifest_PlatformEmbeddingRequiresInjectMapping(t *testing.T) {
	m := &AppSpec{
		ID:      "test",
		Name:    "test",
		Version: "1.0.0",
		Type:    AppTypeApp,
		PlatformServices: PlatformServicesSpec{
			Embedding: &PlatformEmbeddingServiceSpec{
				Required: true,
				Model:    "bge-m3",
				Inject:   PlatformEmbeddingInjectSpec{Model: "KS_EMBEDDING_MODEL"},
			},
		},
	}
	if err := m.Validate(); err == nil {
		t.Error("expected validation error for missing embedding dim injection")
	}
}

func TestPlatformEmbeddingServiceSpec_Validate_ModelRequired(t *testing.T) {
	cases := []struct {
		name    string
		spec    PlatformEmbeddingServiceSpec
		wantErr string
	}{
		{
			name:    "required without model",
			spec:    PlatformEmbeddingServiceSpec{Required: true},
			wantErr: "model 不能为空",
		},
		{
			name: "required without inject.model",
			spec: PlatformEmbeddingServiceSpec{
				Required: true,
				Model:    "bge-m3",
				Inject:   PlatformEmbeddingInjectSpec{Dim: "KS_EMBEDDING_DIM"},
			},
			wantErr: "inject.model 不能为空",
		},
		{
			name: "valid",
			spec: PlatformEmbeddingServiceSpec{
				Required: true,
				Model:    "bge-m3",
				Inject: PlatformEmbeddingInjectSpec{
					Model: "KS_EMBEDDING_MODEL",
					Dim:   "KS_EMBEDDING_DIM",
				},
			},
		},
		{
			name: "not required no model is fine",
			spec: PlatformEmbeddingServiceSpec{Required: false},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want nil, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want err contains %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateManifest_ManagedMySQLRequiresInjectMapping(t *testing.T) {
	m := &AppSpec{
		ID:      "test",
		Name:    "test",
		Version: "1.0.0",
		Type:    AppTypeApp,
		ManagedResources: ManagedResourcesSpec{
			MySQL: &ManagedMySQLResourceSpec{
				Inject: ManagedMySQLInjectSpec{
					Host: "DB_HOST", Port: "DB_PORT", Database: "DB_NAME", User: "DB_USER",
				},
			},
		},
	}
	if err := m.Validate(); err == nil {
		t.Error("expected validation error for missing password injection")
	}
}

func TestParseRuntimeSpec_ProcessMode(t *testing.T) {
	input := `
runtime:
  mode: process
  entry: ./bin/myapp
  working_dir: /opt/app
  health_check: /health
`
	type wrapper struct {
		Runtime RuntimeSpec `yaml:"runtime"`
	}
	var w wrapper
	if err := yaml.Unmarshal([]byte(input), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if w.Runtime.Mode != RuntimeModeProcess {
		t.Errorf("mode: got %q", w.Runtime.Mode)
	}
	if w.Runtime.Entry != "./bin/myapp" {
		t.Errorf("entry: got %q", w.Runtime.Entry)
	}
	if w.Runtime.WorkingDir != "/opt/app" {
		t.Errorf("working_dir: got %q", w.Runtime.WorkingDir)
	}
}

func TestParseRuntimeSpec_ContainerMode(t *testing.T) {
	input := `
runtime:
  mode: container
  image: registry.local/myapp:latest
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
	if w.Runtime.Mode != RuntimeModeContainer {
		t.Errorf("mode: got %q", w.Runtime.Mode)
	}
	if w.Runtime.Image != "registry.local/myapp:latest" {
		t.Errorf("image: got %q", w.Runtime.Image)
	}
	if len(w.Runtime.Volumes) != 2 {
		t.Errorf("volumes len: got %d", len(w.Runtime.Volumes))
	}
	if w.Runtime.Resources.CPU != "1.0" {
		t.Errorf("cpu: got %q", w.Runtime.Resources.CPU)
	}
}

func TestParseRuntimeSpec_WritableRootFS(t *testing.T) {
	type wrapper struct {
		Runtime RuntimeSpec `yaml:"runtime"`
	}

	// 声明 writable_root_fs: true -> 字段为 true
	optIn := `
runtime:
  mode: container
  image: registry.local/sandbox:4.0.0
  writable_root_fs: true
`
	var w wrapper
	if err := yaml.Unmarshal([]byte(optIn), &w); err != nil {
		t.Fatalf("unmarshal opt-in: %v", err)
	}
	if !w.Runtime.WritableRootFS {
		t.Errorf("WritableRootFS opt-in: got false, want true")
	}

	// 未声明 -> 安全默认 false（只读）
	def := `
runtime:
  mode: container
  image: registry.local/other:1.0.0
`
	var d wrapper
	if err := yaml.Unmarshal([]byte(def), &d); err != nil {
		t.Fatalf("unmarshal default: %v", err)
	}
	if d.Runtime.WritableRootFS {
		t.Errorf("WritableRootFS default: got true, want false（安全默认只读）")
	}
}

// TestAppSpec_AuthorIcon_Preserved 守护 Author + Icon 字段 yaml round-trip 后保留。
//
// 历史 bug：AppSpec 没这两字段时，ks-devkit WriteManifestYAML 调用
// yaml.Marshal(spec) 会把作者填的 author / icon 行丢掉。作者 commit
// publish 后写回的 manifest.yaml 即永久失去这两个字段，store 详情页
// 失去图标显示。
func TestAppSpec_AuthorIcon_Preserved(t *testing.T) {
	orig := AppSpec{
		ID:      "x",
		Name:    "x",
		Version: "1.0.0",
		Type:    AppTypeSkill,
		Author:  "keystone-ecosystem",
		Icon:    "icon.svg",
	}

	data, err := yaml.Marshal(&orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := ParseAppSpec(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Author != "keystone-ecosystem" {
		t.Errorf("author: got %q want keystone-ecosystem", parsed.Author)
	}
	if parsed.Icon != "icon.svg" {
		t.Errorf("icon: got %q want icon.svg", parsed.Icon)
	}

	// 同时验证 yaml 输出包含字段（防 omitempty 误把非空值视为空）
	yamlStr := string(data)
	if !strings.Contains(yamlStr, "author: keystone-ecosystem") {
		t.Errorf("yaml output 缺 author 行: %s", yamlStr)
	}
	if !strings.Contains(yamlStr, "icon: icon.svg") {
		t.Errorf("yaml output 缺 icon 行: %s", yamlStr)
	}
}

func TestAppSpec_RoundTrip(t *testing.T) {
	orig := AppSpec{
		ID:      "round-trip-test",
		Name:    "Round Trip",
		Version: "2.0.0",
		Type:    AppTypeSkill,
		Runtime: RuntimeSpec{
			Mode:        RuntimeModeContainer,
			Image:       "myimage:latest",
			Volumes:     []string{"/data:/data"},
			HealthCheck: "/healthz",
			Resources:   ResourcesSpec{CPU: "0.5", Memory: "256Mi"},
		},
		Provides: ProvidesSpec{
			Capabilities: []CapabilitySpec{
				{
					CanonicalName: "round-trip-test.echo",
					ExecutionMode: "sync",
					Backend:       BackendSpec{Kind: "mcp_tool", ToolName: "echo"},
				},
			},
		},
		Requires: RequiresSpec{
			Capabilities: []RequiresCapabilityItem{
				{CanonicalName: "ks-mcp-lili.search", Mode: "required", Reason: "查词典"},
			},
		},
		Conflicts: ConflictsSpec{
			Apps: []ConflictsAppItem{{ID: "conf-c", Reason: "port clash"}},
		},
	}

	data, err := yaml.Marshal(&orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := ParseAppSpec(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.ID != orig.ID {
		t.Errorf("id: got %q want %q", parsed.ID, orig.ID)
	}
	if parsed.Runtime.Mode != RuntimeModeContainer {
		t.Errorf("runtime.mode: got %q", parsed.Runtime.Mode)
	}
	if parsed.Runtime.Image != "myimage:latest" {
		t.Errorf("runtime.image: got %q", parsed.Runtime.Image)
	}
	if len(parsed.Runtime.Volumes) != 1 || parsed.Runtime.Volumes[0] != "/data:/data" {
		t.Errorf("runtime.volumes: got %v", parsed.Runtime.Volumes)
	}
	if len(parsed.Provides.Capabilities) != 1 || parsed.Provides.Capabilities[0].CanonicalName != "round-trip-test.echo" {
		t.Errorf("provides.capabilities: got %v", parsed.Provides.Capabilities)
	}
	if parsed.Provides.Capabilities[0].Backend.Kind != "mcp_tool" || parsed.Provides.Capabilities[0].Backend.ToolName != "echo" {
		t.Errorf("provides.capabilities[0].backend: got %+v", parsed.Provides.Capabilities[0].Backend)
	}
	if len(parsed.Requires.Capabilities) != 1 || parsed.Requires.Capabilities[0].CanonicalName != "ks-mcp-lili.search" {
		t.Errorf("requires.capabilities: got %v", parsed.Requires.Capabilities)
	}
	if len(parsed.Conflicts.Apps) != 1 || parsed.Conflicts.Apps[0].ID != "conf-c" {
		t.Errorf("conflicts.apps: got %v", parsed.Conflicts.Apps)
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
		{RuntimeModeExtension, true},
		{RuntimeMode("invalid"), false},
		{RuntimeMode(""), true},
	}
	for _, c := range cases {
		if got := c.mode.Valid(); got != c.valid {
			t.Errorf("RuntimeMode(%q).Valid() = %v, want %v", c.mode, got, c.valid)
		}
	}
}

func TestValidateManifest_InvalidRuntimeMode(t *testing.T) {
	m := &AppSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:    AppTypeApp,
		Runtime: RuntimeSpec{Mode: RuntimeMode("invalid")},
	}
	if err := m.Validate(); err == nil {
		t.Error("期望 runtime mode 校验失败")
	}
}

func TestValidateManifest_ValidRuntimeMode(t *testing.T) {
	for _, mode := range []RuntimeMode{"", RuntimeModeNone, RuntimeModeProcess, RuntimeModeContainer} {
		m := &AppSpec{
			ID: "test", Name: "test", Version: "1.0.0",
			Type:    AppTypeApp,
			Runtime: RuntimeSpec{Mode: mode},
		}
		if err := m.Validate(); err != nil {
			t.Errorf("runtime mode %q 应通过校验: %v", mode, err)
		}
	}
}

func TestValidateManifest_InvalidProtection(t *testing.T) {
	m := &AppSpec{
		ID: "test", Name: "test", Version: "1.0.0",
		Type:       AppTypeApp,
		Protection: ProtectionLevel("invalid"),
	}
	if err := m.Validate(); err == nil {
		t.Error("期望 protection 校验失败")
	}
}

func TestValidateManifest_ValidProtection(t *testing.T) {
	for _, p := range []ProtectionLevel{"", ProtectionNone, ProtectionProtected, ProtectionSystem} {
		m := &AppSpec{
			ID: "test", Name: "test", Version: "1.0.0",
			Type:       AppTypeApp,
			Protection: p,
		}
		if err := m.Validate(); err != nil {
			t.Errorf("protection %q 应通过校验: %v", p, err)
		}
	}
}

func TestParseAppSpec_Skill(t *testing.T) {
	input := `
id: test-skill
name: 测试技能
version: 1.0.0
type: skill
runtime:
  mode: none
mount:
  skill:
    name: 测试技能
    description: 一个测试用 skill
`
	spec, err := ParseAppSpec([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if spec.Mount.Skill == nil {
		t.Fatal("mount.skill is nil")
	}
	if spec.Mount.Skill.Name != "测试技能" {
		t.Errorf("skill.name = %q", spec.Mount.Skill.Name)
	}
}

func TestAppSpecValidate_SkillMustHaveNoneMode(t *testing.T) {
	spec := &AppSpec{
		ID: "s", Name: "S", Version: "1.0.0", Type: AppTypeSkill,
		Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected error: skill with container mode")
	}
}

func TestAppSpecValidate_SkillMountNotRequired(t *testing.T) {
	spec := &AppSpec{
		ID: "s", Name: "S", Version: "1.0.0", Type: AppTypeSkill,
		Runtime: RuntimeSpec{Mode: RuntimeModeNone},
	}
	// skill 不要求 mount.skill 必须存在（纯文件型 skill 没有挂载配置也行）
	if err := spec.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestParseAppSpec_LocalizedFieldsBackwardsCompatible 老 manifest（单 string 形态）
// 必须仍能解析，向后兼容是 v0.9.0 的硬约束。
func TestParseAppSpec_LocalizedFieldsBackwardsCompatible(t *testing.T) {
	raw := []byte(`
id: skill-tdd
name: skill-tdd
version: 0.1.0
type: skill
runtime:
  mode: none
summary: 在每个新代码改动前强制 RED-GREEN-REFACTOR 循环
description: |
  ## 功能
  详细描述
tags: ["tdd", "测试"]
pricing:
  type: free
  description: 永久免费
`)
	spec, err := ParseAppSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.Summary.Get("zh-CN"); got != "在每个新代码改动前强制 RED-GREEN-REFACTOR 循环" {
		t.Errorf("Summary.Get = %q", got)
	}
	if got := spec.Summary.Get(""); got != "在每个新代码改动前强制 RED-GREEN-REFACTOR 循环" {
		t.Errorf("Summary.Get(\"\") = %q", got)
	}
	if got := spec.Description.Get("zh-CN"); !strings.Contains(got, "详细描述") {
		t.Errorf("Description.Get = %q", got)
	}
	if got := spec.Tags.Get("zh-CN"); len(got) != 2 || got[0] != "tdd" {
		t.Errorf("Tags.Get = %v", got)
	}
	if got := spec.Pricing.Description.Get("zh-CN"); got != "永久免费" {
		t.Errorf("Pricing.Description.Get = %q", got)
	}
	if err := spec.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

// TestParseAppSpec_LocalizedFieldsI18nForm 多 locale map 形态。
func TestParseAppSpec_LocalizedFieldsI18nForm(t *testing.T) {
	raw, err := os.ReadFile("testdata/i18n_manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseAppSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.Summary.Get("zh-CN"); got != "中文摘要" {
		t.Errorf("Summary zh-CN = %q", got)
	}
	if got := spec.Summary.Get("en-US"); got != "English summary" {
		t.Errorf("Summary en-US = %q", got)
	}
	if got := spec.Tags.Get("en-US"); len(got) != 2 || got[1] != "debug" {
		t.Errorf("Tags en-US = %v", got)
	}
	if got := spec.Pricing.Description.Get("zh-CN"); got != "按月订阅 ¥99" {
		t.Errorf("Pricing.Description zh-CN = %q", got)
	}
}

// TestParseAppSpec_ChangelogField AppSpec.Changelog 字段读写。
func TestParseAppSpec_ChangelogField(t *testing.T) {
	raw := []byte(`
id: skill-tdd
name: skill-tdd
version: 0.3.0
type: skill
runtime:
  mode: none
summary: TDD
changelog: |
  ### Added
  - 支持 examples 抽取
`)
	spec, err := ParseAppSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(spec.Changelog, "examples 抽取") {
		t.Errorf("Changelog missing content: %q", spec.Changelog)
	}
}

func TestAppSpec_StandaloneFallback_Parse(t *testing.T) {
	input := `id: test-squad
name: Test Squad
version: 1.0.0
type: mcp-service
managed_resources:
  mysql:
    required: true
    database: auto
    user: auto
    inject:
      host: DB_HOST
      port: DB_PORT
      database: DB_NAME
      user: DB_USER
      password: DB_PASSWORD
standalone_fallback:
  mysql:
    host: localhost
    port: 3306
    database: test_squad_dev
    user: root
    password: root
`
	var spec AppSpec
	require.NoError(t, yaml.Unmarshal([]byte(input), &spec))

	require.NotNil(t, spec.StandaloneFallback)
	require.NotNil(t, spec.StandaloneFallback.MySQL)
	assert.Equal(t, "localhost", spec.StandaloneFallback.MySQL.Host)
	assert.Equal(t, 3306, spec.StandaloneFallback.MySQL.Port)
}

// TestAppSpec_LocalizedFields_JSONMapShape 验证 JSON 序列化永远输出 map 形态。
// 这是 v0.9.0 故意引入的 wire-format 变更：前端拿到的总是 map，不用 union 类型。
// 下游消费方（ks-hub backend handler）需要 `.Get(locale)` 取值再回填到老 wire schema。
func TestAppSpec_LocalizedFields_JSONMapShape(t *testing.T) {
	spec := &AppSpec{
		ID:      "x",
		Name:    "x",
		Version: "1.0.0",
		Type:    AppTypeSkill,
		Summary: LocalizedString{"": "single"},
		Tags:    LocalizedTags{"": {"a", "b"}},
	}
	buf, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	out := string(buf)
	if !strings.Contains(out, `"summary":{"":"single"}`) {
		t.Errorf("summary 应输出 map 形态，got: %s", out)
	}
	if !strings.Contains(out, `"tags":{"":["a","b"]}`) {
		t.Errorf("tags 应输出 map 形态，got: %s", out)
	}
}

// ────────── v0.19.0 capability mesh schema 单测 ──────────

// TestAppSpec_ParseProvidesRequires 验证 provides / requires / conflicts 三段
// 完整 YAML 解析 round-trip。
func TestAppSpec_ParseProvidesRequires(t *testing.T) {
	input := `
id: ks-mcp-writer
name: AI 写手服务
version: 1.0.0
type: app
runtime:
  mode: container
  image: ks-mcp-writer:1.0.0
provides:
  capabilities:
    - name: create_article
      display_name: AI 文章创作
      execution_mode: long_running
      backend:
        kind: mcp_tool
        tool_name: create_article
      timeout_ms: 300000
      concurrency_limit: 3
      side_effect_level: hard_irreversible
      resumable: true
      guardrail_profile: content_creation
      tags: [writing]
      notify_policy:
        on_done: true
        on_failed: true
      description: 根据主题与风格生成完整文章草稿，接收 topic + style + platform 调用 LLM 生成
      aliases: [写文章, AI 创作]
      user_utterances:
        - 帮我写一篇关于 AI 的科技新闻
      use_cases: [新闻稿, 公众号]
      domain_terms: [topic, style, platform]
      negative_examples:
        - 帮我润色这段话
    - name: list_articles
      execution_mode: sync
      backend:
        kind: http_endpoint
        path: /api/articles
        method: GET
requires:
  capabilities:
    - canonical_name: ks-mcp-image-gen.generate
      mode: required
      reason: 文章封面与配图
    - canonical_name: ks-mcp-lili.search
      mode: optional
conflicts:
  apps:
    - id: ks-mcp-legacy-writer
      reason: 兼容性冲突
`
	spec, err := ParseAppSpec([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(spec.Provides.Capabilities) != 2 {
		t.Fatalf("provides len: got %d", len(spec.Provides.Capabilities))
	}
	c0 := spec.Provides.Capabilities[0]
	if c0.Name != "create_article" || c0.ExecutionMode != "long_running" {
		t.Errorf("provides[0] core: %+v", c0)
	}
	if c0.Backend.Kind != "mcp_tool" || c0.Backend.ToolName != "create_article" {
		t.Errorf("provides[0].backend: %+v", c0.Backend)
	}
	if c0.TimeoutMs != 300000 || c0.ConcurrencyLimit != 3 {
		t.Errorf("provides[0] perf: timeout=%d concurrency=%d",
			c0.TimeoutMs, c0.ConcurrencyLimit)
	}
	if !c0.Resumable {
		t.Errorf("provides[0].resumable 应为 true")
	}
	if c0.GuardrailProfile != "content_creation" {
		t.Errorf("provides[0].guardrail_profile: %q", c0.GuardrailProfile)
	}
	if c0.NotifyPolicy.OnDone == nil || !*c0.NotifyPolicy.OnDone {
		t.Errorf("provides[0].notify_policy.on_done 应 *true")
	}
	if len(c0.UserUtterances) != 1 || c0.UserUtterances[0] != "帮我写一篇关于 AI 的科技新闻" {
		t.Errorf("provides[0].user_utterances: %v", c0.UserUtterances)
	}
	c1 := spec.Provides.Capabilities[1]
	if c1.Backend.Kind != "http_endpoint" || c1.Backend.Path != "/api/articles" || c1.Backend.Method != "GET" {
		t.Errorf("provides[1].backend: %+v", c1.Backend)
	}

	if len(spec.Requires.Capabilities) != 2 {
		t.Fatalf("requires len: %d", len(spec.Requires.Capabilities))
	}
	if spec.Requires.Capabilities[1].Mode != "optional" {
		t.Errorf("requires[1].mode: %q", spec.Requires.Capabilities[1].Mode)
	}
	if len(spec.Conflicts.Apps) != 1 || spec.Conflicts.Apps[0].ID != "ks-mcp-legacy-writer" {
		t.Errorf("conflicts.apps: %v", spec.Conflicts.Apps)
	}

	if err := spec.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

// TestRequiresSpec_ValidateCanonicalNameRegex 验证 requires.canonical_name 正则校验。
// 去前缀后 provides 写裸名 name（见 TestProvidesSpec_BareName）；全名 canonical_name
// 仅保留于 requires（引用他人已注册能力，故不对称）。
func TestRequiresSpec_ValidateCanonicalNameRegex(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"app.verb", true},
		{"my-app.do-something", true},
		{"app.verb.sub", true},
		{"app.verb_with_underscore", true},
		{"app.verb-with-dash", true},
		{"App.Verb", false},  // 大写首字母
		{".verb", false},     // 起始 .
		{"app.", false},      // 末尾 .
		{"app", false},       // 缺少 .
		{"1app.verb", false}, // 数字开头
		{"app..verb", false}, // 双 .
		{"app verb", false},  // 空格
		{"app.中文", false},    // 非 ASCII
		{"my-translator.translate", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			spec := &AppSpec{
				ID: "consumer-app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
				Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
				Requires: RequiresSpec{Capabilities: []RequiresCapabilityItem{
					{CanonicalName: c.name, Mode: "required"},
				}},
			}
			err := spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("requires.canonical_name=%q valid=%v want %v (err=%v)", c.name, gotValid, c.valid, err)
			}
		})
	}
}

// TestPublicHTTPSpec_Validate 覆盖 public_http 白名单路径校验：精确/前缀通配合法；
// 缺前导斜杠、.. 穿越、根通配、中段通配、query/fragment、空白、空路径、重复、超上限均应拒绝。
func TestPublicHTTPSpec_Validate(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		valid bool
	}{
		{"exact", []string{"/healthz"}, true},
		{"prefix-wildcard", []string{"/api/publisher/plugin/*"}, true},
		{"exact-plus-wildcard", []string{"/healthz", "/api/publisher/oauth/*"}, true},
		{"empty-list", nil, true},
		{"missing-leading-slash", []string{"healthz"}, false},
		{"dotdot-traversal", []string{"/api/../secret"}, false},
		{"root-wildcard", []string{"/*"}, false},
		{"mid-wildcard", []string{"/api/*/plugin"}, false},
		{"query", []string{"/api?x=1"}, false},
		{"fragment", []string{"/api#frag"}, false},
		{"whitespace", []string{"/api /plugin"}, false},
		{"empty-path", []string{""}, false},
		{"duplicate", []string{"/healthz", "/healthz"}, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := PublicHTTPSpec{Paths: c.paths}.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("public_http.paths=%v valid=%v want %v (err=%v)", c.paths, gotValid, c.valid, err)
			}
		})
	}

	// 超上限（长度检查先于逐条校验触发，条目内容不影响结论）。
	over := make([]string, maxPublicHTTPPaths+1)
	for i := range over {
		over[i] = "/p"
	}
	if err := (PublicHTTPSpec{Paths: over}).Validate(); err == nil {
		t.Errorf("期望超过 %d 条上限校验失败", maxPublicHTTPPaths)
	}
}

// TestProvidesSpec_ValidateDuplicateName 验证同 manifest 内 capability 裸名 name 不重复。
func TestProvidesSpec_ValidateDuplicateName(t *testing.T) {
	spec := &AppSpec{
		ID: "ks-mcp-writer", Name: "x", Version: "1.0.0", Type: AppTypeApp,
		Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
		Provides: ProvidesSpec{Capabilities: []CapabilitySpec{
			{Name: "create", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "create"}},
			{Name: "create", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "create2"}},
		}},
	}
	err := spec.Validate()
	if err == nil {
		t.Fatal("期望 name 重复校验失败")
	}
	if !strings.Contains(err.Error(), "重复") {
		t.Errorf("错误信息应提示重复: %v", err)
	}
}

// TestAppSpec_ValidateBackendKindRuntimeConsistency 验证 backend.kind 与 runtime.mode 一致性。
func TestAppSpec_ValidateBackendKindRuntimeConsistency(t *testing.T) {
	cases := []struct {
		name        string
		runtimeMode RuntimeMode
		backendKind string
		valid       bool
	}{
		{"http_endpoint+container", RuntimeModeContainer, "http_endpoint", true},
		{"http_endpoint+extension", RuntimeModeExtension, "http_endpoint", false},
		{"http_endpoint+none", RuntimeModeNone, "http_endpoint", false},
		{"mcp_tool+container", RuntimeModeContainer, "mcp_tool", true},
		{"mcp_tool+extension", RuntimeModeExtension, "mcp_tool", true},
		{"mcp_tool+none", RuntimeModeNone, "mcp_tool", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var be BackendSpec
			switch c.backendKind {
			case "http_endpoint":
				be = BackendSpec{Kind: "http_endpoint", Path: "/x", Method: "GET"}
			case "mcp_tool":
				be = BackendSpec{Kind: "mcp_tool", ToolName: "x"}
			}
			spec := &AppSpec{
				ID: "app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
				Runtime: RuntimeSpec{Mode: c.runtimeMode},
				Provides: ProvidesSpec{Capabilities: []CapabilitySpec{
					{Name: "x", ExecutionMode: "sync", Backend: be},
				}},
			}
			err := spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("%s: valid=%v want %v (err=%v)", c.name, gotValid, c.valid, err)
			}
		})
	}
}

// TestAppSpec_ValidateRequiresModeEnum 验证 requires.mode 枚举。
func TestAppSpec_ValidateRequiresModeEnum(t *testing.T) {
	cases := []struct {
		mode  string
		valid bool
	}{
		{"required", true},
		{"recommended", true},
		{"optional", true},
		{"", false},
		{"REQUIRED", false},
		{"must", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.mode, func(t *testing.T) {
			spec := &AppSpec{
				ID: "app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
				Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
				Requires: RequiresSpec{Capabilities: []RequiresCapabilityItem{
					{CanonicalName: "other.cap", Mode: c.mode},
				}},
			}
			err := spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("mode=%q valid=%v want %v (err=%v)", c.mode, gotValid, c.valid, err)
			}
		})
	}
}

// TestAppSpec_ValidateExecutionModeEnum 验证 execution_mode 枚举。
func TestAppSpec_ValidateExecutionModeEnum(t *testing.T) {
	cases := []struct {
		mode  string
		valid bool
	}{
		{"sync", true},
		{"long_running", true},
		{"hybrid", true},
		{"", false},
		{"async", false}, // a2a 用 async；capability 用 long_running
		{"streaming", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.mode, func(t *testing.T) {
			spec := &AppSpec{
				ID: "app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
				Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
				Provides: ProvidesSpec{Capabilities: []CapabilitySpec{
					{
						Name:          "x",
						ExecutionMode: c.mode,
						Backend:       BackendSpec{Kind: "mcp_tool", ToolName: "x"},
					},
				}},
			}
			err := spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("execution_mode=%q valid=%v want %v (err=%v)", c.mode, gotValid, c.valid, err)
			}
		})
	}
}

// TestAppSpec_ValidateBackendKindEnum 验证 backend.kind 枚举。
func TestAppSpec_ValidateBackendKindEnum(t *testing.T) {
	cases := []struct {
		kind  string
		valid bool
	}{
		{"mcp_tool", true},
		{"http_endpoint", true},
		{"", false},
		{"grpc", false},
		{"MCP_TOOL", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.kind, func(t *testing.T) {
			be := BackendSpec{Kind: c.kind}
			if c.kind == "mcp_tool" {
				be.ToolName = "x"
			}
			if c.kind == "http_endpoint" {
				be.Path = "/x"
				be.Method = "GET"
			}
			spec := &AppSpec{
				ID: "app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
				Runtime: RuntimeSpec{Mode: RuntimeModeContainer},
				Provides: ProvidesSpec{Capabilities: []CapabilitySpec{
					{Name: "x", ExecutionMode: "sync", Backend: be},
				}},
			}
			err := spec.Validate()
			gotValid := err == nil
			if gotValid != c.valid {
				t.Errorf("backend.kind=%q valid=%v want %v (err=%v)", c.kind, gotValid, c.valid, err)
			}
		})
	}
}

// TestAppSpec_ValidateConflicts 验证 conflicts.apps[].id 必填。
func TestAppSpec_ValidateConflicts(t *testing.T) {
	spec := &AppSpec{
		ID: "app", Name: "x", Version: "1.0.0", Type: AppTypeApp,
		Runtime:   RuntimeSpec{Mode: RuntimeModeContainer},
		Conflicts: ConflictsSpec{Apps: []ConflictsAppItem{{ID: "", Reason: "x"}}},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("期望 conflicts.apps[].id 必填校验失败")
	}
}

// TestCapabilitySpec_InputSchema 验证 manifest yaml 的 input_schema 内联字段
// 能正确反序列化到 CapabilitySpec.InputSchema map。
func TestCapabilitySpec_InputSchema(t *testing.T) {
	yamlStr := `
canonical_name: test.echo
execution_mode: sync
backend:
  kind: mcp_tool
  tool_name: echo
input_schema:
  type: object
  properties:
    message:
      type: string
      description: "要回显的内容"
  required: [message]
`
	var spec CapabilitySpec
	if err := yaml.Unmarshal([]byte(yamlStr), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if spec.InputSchema == nil {
		t.Fatal("InputSchema is nil")
	}
	if spec.InputSchema["type"] != "object" {
		t.Errorf("InputSchema.type = %v, want object", spec.InputSchema["type"])
	}
	props, ok := spec.InputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties type: %T", spec.InputSchema["properties"])
	}
	if _, ok := props["message"]; !ok {
		t.Error("properties.message missing")
	}
}

func TestAppSpec_ParseMountCapabilityProfiles(t *testing.T) {
	input := `
id: agent-backend-engineer
name: 后端工程师
version: 1.2.0
type: agent
runtime:
  mode: none
mount:
  agent:
    create_agent: true
    name: 后端工程师
    system_prompt: 先理数据流。
    profile:
      canonical_name: agent.backend-engineer
      display_name: 后端工程师
      description: 设计后端接口、数据流与数据库方案；用户需要后端架构、接口契约、SQL 优化或故障根因排查时使用
      aliases: [后端助手, API 设计]
      user_utterances:
        - 帮我设计这个接口
      use_cases: [接口设计, SQL 优化]
      domain_terms: [API, SQL, 幂等]
      negative_examples:
        - 不负责前端视觉细节
`
	spec, err := ParseAppSpec([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, spec.Mount.Agent)
	require.NotNil(t, spec.Mount.Agent.Profile)
	assert.Equal(t, "agent.backend-engineer", spec.Mount.Agent.Profile.CanonicalName)
	assert.Equal(t, "设计后端接口、数据流与数据库方案；用户需要后端架构、接口契约、SQL 优化或故障根因排查时使用", spec.Mount.Agent.Profile.Description)
	assert.Equal(t, []string{"后端助手", "API 设计"}, spec.Mount.Agent.Profile.Aliases)
	assert.NoError(t, spec.Validate())

	skillInput := `
id: skill-investigate
name: 深度调查
version: 1.2.0
type: skill
runtime:
  mode: none
mount:
  skill:
    name: skill-investigate
    profile:
      canonical_name: skill.skill-investigate
      display_name: 深度调查
      description: 系统化排查故障根因
`
	skillSpec, err := ParseAppSpec([]byte(skillInput))
	require.NoError(t, err)
	require.NotNil(t, skillSpec.Mount.Skill)
	require.NotNil(t, skillSpec.Mount.Skill.Profile)
	assert.Equal(t, "skill.skill-investigate", skillSpec.Mount.Skill.Profile.CanonicalName)
	assert.NoError(t, skillSpec.Validate())
}

func TestProvidesSpec_Validate_SideEffectEnum(t *testing.T) {
	base := func(se string) ProvidesSpec {
		return ProvidesSpec{Capabilities: []CapabilitySpec{{
			Name:            "do_thing",
			ExecutionMode:   "sync",
			Backend:         BackendSpec{Kind: "mcp_tool", ToolName: "do_thing"},
			SideEffectLevel: se,
		}}}
	}
	// 合法值通过
	if err := base("none").Validate("", RuntimeModeContainer); err != nil {
		t.Fatalf("合法枚举应通过, got %v", err)
	}
	if err := base("hard_irreversible").Validate("", RuntimeModeContainer); err != nil {
		t.Fatalf("合法枚举应通过, got %v", err)
	}
	// 空值允许（keystone 侧再强制必填）
	if err := base("").Validate("", RuntimeModeContainer); err != nil {
		t.Fatalf("空值在 ks-types 层应放行, got %v", err)
	}
	// 旧枚举值（readonly）应被拒
	if err := base("readonly").Validate("", RuntimeModeContainer); err == nil {
		t.Fatal("side_effect_level=readonly 应被拒")
	}
}

func TestAuthSpec_EffectiveMode(t *testing.T) {
	// 未声明 + 有入站端点（container/process）→ secure-by-default keystone_jwks
	assert.Equal(t, AuthModeKeystoneJWKS, AuthSpec{}.EffectiveMode(RuntimeModeContainer))
	assert.Equal(t, AuthModeKeystoneJWKS, AuthSpec{}.EffectiveMode(RuntimeModeProcess))
	// 未声明 + 无入站端点（none：agent/skill）→ none
	assert.Equal(t, AuthModeNone, AuthSpec{}.EffectiveMode(RuntimeModeNone))
	assert.Equal(t, AuthModeNone, AuthSpec{}.EffectiveMode(""))
	// 显式声明 → 尊重作者（none opt-out / 非默认模式）
	assert.Equal(t, AuthModeNone, AuthSpec{Mode: AuthModeNone}.EffectiveMode(RuntimeModeContainer))
	assert.Equal(t, AuthModeStaticBearer, AuthSpec{Mode: AuthModeStaticBearer}.EffectiveMode(RuntimeModeContainer))
}

// TestCapability_OutputSchemaInline 验证内联 output_schema 能反序列化到
// CapabilitySpec.OutputSchema map（与 input_schema 对称；取代被砍的 output_schema_ref）。
func TestCapability_OutputSchemaInline(t *testing.T) {
	yamlStr := `
canonical_name: test.echo
execution_mode: sync
backend:
  kind: mcp_tool
  tool_name: echo
output_schema:
  type: object
  properties:
    status:
      type: string
`
	var spec CapabilitySpec
	require.NoError(t, yaml.Unmarshal([]byte(yamlStr), &spec))
	assert.Equal(t, "object", spec.OutputSchema["type"])
	props, ok := spec.OutputSchema["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "status")
}

// TestCapability_DescriptionMerged 验证合并后的单一 description 字段解析
// （取代旧 intent_summary + natural_description 双字段）。
func TestCapability_DescriptionMerged(t *testing.T) {
	yamlStr := `
provides:
  capabilities:
    - canonical_name: x.y
      execution_mode: sync
      backend:
        kind: mcp_tool
        tool_name: y
      description: 一句语义真话
`
	m, err := ParseAppSpec([]byte(yamlStr))
	require.NoError(t, err)
	assert.Equal(t, "一句语义真话", m.Provides.Capabilities[0].Description)
}

func TestCapability_EffectiveDecisionMode(t *testing.T) {
	// 默认从 side_effect_level 派生
	assert.Equal(t, DecisionModeAgentAutonomous, CapabilitySpec{SideEffectLevel: "none"}.EffectiveDecisionMode())
	assert.Equal(t, DecisionModeUserAuthorized, CapabilitySpec{SideEffectLevel: "soft_reversible"}.EffectiveDecisionMode())
	assert.Equal(t, DecisionModeUserOnly, CapabilitySpec{SideEffectLevel: "hard_irreversible"}.EffectiveDecisionMode())
	// 未声明 side_effect_level → agent_autonomous
	assert.Equal(t, DecisionModeAgentAutonomous, CapabilitySpec{}.EffectiveDecisionMode())
	// 作者覆盖（敏感读：side_effect=none 但要 user_only）
	assert.Equal(t, DecisionModeUserOnly, CapabilitySpec{SideEffectLevel: "none", DecisionMode: DecisionModeUserOnly}.EffectiveDecisionMode())
}

func TestCapability_DecisionModeValidate(t *testing.T) {
	bad := CapabilitySpec{Name: "y", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "y"}, DecisionMode: "bogus"}
	err := ProvidesSpec{Capabilities: []CapabilitySpec{bad}}.Validate("x", RuntimeModeContainer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision_mode")

	// pre_authorize_duration 仅在 decision_mode=user_authorized 时有效
	preAuthNoMode := CapabilitySpec{Name: "y", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "y"}, PreAuthorizeDuration: PreAuth24h}
	err = ProvidesSpec{Capabilities: []CapabilitySpec{preAuthNoMode}}.Validate("x", RuntimeModeContainer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre_authorize_duration")

	// 合法组合通过
	ok := CapabilitySpec{Name: "y", ExecutionMode: "sync", Backend: BackendSpec{Kind: "mcp_tool", ToolName: "y"}, DecisionMode: DecisionModeUserAuthorized, PreAuthorizeDuration: PreAuth24h}
	require.NoError(t, ProvidesSpec{Capabilities: []CapabilitySpec{ok}}.Validate("x", RuntimeModeContainer))
}

func TestResourcesSpec_Validate(t *testing.T) {
	require.NoError(t, ResourcesSpec{CPU: "0.5", Memory: "512Mi"}.Validate())
	require.NoError(t, ResourcesSpec{}.Validate()) // 空=用平台默认，合法
	assert.Error(t, ResourcesSpec{CPU: "abc"}.Validate())
	assert.Error(t, ResourcesSpec{Memory: "512"}.Validate()) // 缺单位
}

func TestCapability_TimeoutBounds(t *testing.T) {
	mk := func(ms int) ProvidesSpec {
		return ProvidesSpec{Capabilities: []CapabilitySpec{{
			Name: "y", ExecutionMode: "sync",
			Backend: BackendSpec{Kind: "mcp_tool", ToolName: "y"}, TimeoutMs: ms,
		}}}
	}
	assert.Error(t, mk(-1).Validate("x", RuntimeModeContainer))
	assert.Error(t, mk(99).Validate("x", RuntimeModeContainer)) // < 下界
	assert.NoError(t, mk(30000).Validate("x", RuntimeModeContainer))
}

func TestManagedMySQL_SlimValidate(t *testing.T) {
	// 瘦身后：无 required/database/user 强制值样板，只校验 inject.* 必填
	ok := ManagedMySQLResourceSpec{Inject: ManagedMySQLInjectSpec{
		Host: "DB_HOST", Port: "DB_PORT", Database: "DB_NAME", User: "DB_USER", Password: "DB_PASSWORD",
	}}
	require.NoError(t, ok.Validate())
	bad := ManagedMySQLResourceSpec{} // 缺 inject
	assert.Error(t, bad.Validate())
}

func TestCapabilitySpec_EntryField(t *testing.T) {
	y := []byte("name: create_campaign\nexecution_mode: long_running\nentry: true\nbackend:\n  kind: mcp_tool\n  tool_name: create_campaign\n")
	var c CapabilitySpec
	if err := yaml.Unmarshal(y, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c.Entry {
		t.Fatalf("expected Entry=true, got false")
	}
}
