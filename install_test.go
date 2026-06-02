package kstypes

import "testing"

func TestParseInstallSpec(t *testing.T) {
	yaml := `
config_fields:
  - key: storage_dir
    type: string
    label: 存储目录
    default: /tmp/data
    required: false
  - key: quality
    type: select
    label: 质量
    options: [standard, hd]
    default: standard
  - key: bind_agents
    type: agent_selector
    label: 绑定 Agent
    required: false
secret_fields:
  - key: api_key
    label: API Key
    required: true
`
	spec, err := ParseInstallSpec([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(spec.ConfigFields) != 3 {
		t.Fatalf("config_fields count = %d", len(spec.ConfigFields))
	}
	if spec.ConfigFields[0].Key != "storage_dir" {
		t.Errorf("first key = %q", spec.ConfigFields[0].Key)
	}
	if spec.ConfigFields[1].Type != ConfigFieldSelect {
		t.Errorf("second type = %q", spec.ConfigFields[1].Type)
	}
	if len(spec.ConfigFields[1].Options) != 2 {
		t.Errorf("options count = %d", len(spec.ConfigFields[1].Options))
	}
	if len(spec.SecretFields) != 1 {
		t.Fatalf("secret_fields count = %d", len(spec.SecretFields))
	}
	if spec.SecretFields[0].Key != "api_key" {
		t.Errorf("secret key = %q", spec.SecretFields[0].Key)
	}
}

func TestInstallSpecValidate_DuplicateKey(t *testing.T) {
	spec := &InstallSpec{
		ConfigFields: []ConfigFieldDef{
			{Key: "x", Type: ConfigFieldString, Label: "X"},
			{Key: "x", Type: ConfigFieldInt, Label: "X again"},
		},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

func TestInstallSpecValidate_SelectWithoutOptions(t *testing.T) {
	spec := &InstallSpec{
		ConfigFields: []ConfigFieldDef{
			{Key: "q", Type: ConfigFieldSelect, Label: "Q"},
		},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected error for select without options")
	}
}

func TestInstallSpecValidate_InvalidType(t *testing.T) {
	spec := &InstallSpec{
		ConfigFields: []ConfigFieldDef{
			{Key: "x", Type: "unknown_type", Label: "X"},
		},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestInstallSpecValidate_CrossKeyConflict(t *testing.T) {
	spec := &InstallSpec{
		ConfigFields: []ConfigFieldDef{{Key: "token", Type: ConfigFieldString, Label: "T"}},
		SecretFields: []SecretFieldDef{{Key: "token", Label: "T"}},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected error for config/secret key conflict")
	}
}

func TestParseInstallSpec_Empty(t *testing.T) {
	spec, err := ParseInstallSpec([]byte(""))
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(spec.ConfigFields) != 0 || len(spec.SecretFields) != 0 {
		t.Fatal("expected empty spec")
	}
}
