package kstypes

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ConfigFieldType 配置项类型枚举
type ConfigFieldType string

const (
	ConfigFieldString        ConfigFieldType = "string"
	ConfigFieldInt           ConfigFieldType = "int"
	ConfigFieldBool          ConfigFieldType = "bool"
	ConfigFieldSelect        ConfigFieldType = "select"
	ConfigFieldAgentSelector ConfigFieldType = "agent_selector"
)

var validConfigFieldTypes = map[ConfigFieldType]bool{
	ConfigFieldString:        true,
	ConfigFieldInt:           true,
	ConfigFieldBool:          true,
	ConfigFieldSelect:        true,
	ConfigFieldAgentSelector: true,
}

// ConfigFieldDef 描述一个用户可见的配置项。
type ConfigFieldDef struct {
	Key      string          `yaml:"key" json:"key"`
	Type     ConfigFieldType `yaml:"type" json:"type"`
	Label    string          `yaml:"label" json:"label"`
	Default  any             `yaml:"default,omitempty" json:"default,omitempty"`
	Required bool            `yaml:"required,omitempty" json:"required,omitempty"`
	Options  []string        `yaml:"options,omitempty" json:"options,omitempty"`
}

// SecretFieldDef 描述一个敏感配置项（值加密存储）。
type SecretFieldDef struct {
	Key      string `yaml:"key" json:"key"`
	Label    string `yaml:"label" json:"label"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

// InstallSpec 描述应用安装时需要的配置 schema（install.yaml）。
type InstallSpec struct {
	ConfigFields []ConfigFieldDef `yaml:"config_fields,omitempty" json:"config_fields,omitempty"`
	SecretFields []SecretFieldDef `yaml:"secret_fields,omitempty" json:"secret_fields,omitempty"`
}

// ParseInstallSpec 从 YAML 字节解析 InstallSpec。
func ParseInstallSpec(data []byte) (*InstallSpec, error) {
	var s InstallSpec
	if len(data) == 0 {
		return &s, nil
	}
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse install.yaml: %w", err)
	}
	return &s, nil
}

// Validate 校验 InstallSpec。
func (s *InstallSpec) Validate() error {
	keys := make(map[string]bool)

	for _, f := range s.ConfigFields {
		if f.Key == "" {
			return fmt.Errorf("install: config_fields 中存在空 key")
		}
		if keys[f.Key] {
			return fmt.Errorf("install: 重复的 key %q", f.Key)
		}
		keys[f.Key] = true

		if !validConfigFieldTypes[f.Type] {
			return fmt.Errorf("install: key %q 的 type %q 不合法", f.Key, f.Type)
		}
		if f.Type == ConfigFieldSelect && len(f.Options) == 0 {
			return fmt.Errorf("install: key %q 的 type=select 必须提供 options", f.Key)
		}
	}

	for _, f := range s.SecretFields {
		if f.Key == "" {
			return fmt.Errorf("install: secret_fields 中存在空 key")
		}
		if keys[f.Key] {
			return fmt.Errorf("install: 重复的 key %q（与 config_fields 冲突）", f.Key)
		}
		keys[f.Key] = true
	}

	return nil
}
