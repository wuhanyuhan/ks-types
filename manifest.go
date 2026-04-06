package kstypes

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ManifestSpec 应用 manifest.yaml 的完整结构
type ManifestSpec struct {
	ID            string                    `yaml:"id" json:"id"`
	Name          string                    `yaml:"name" json:"name"`
	Version       string                    `yaml:"version" json:"version"`
	Type          AppType                   `yaml:"type" json:"type"`
	Summary       string                    `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description   string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Publisher     string                    `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	Category      string                    `yaml:"category,omitempty" json:"category,omitempty"`
	Tags          []string                  `yaml:"tags,omitempty" json:"tags,omitempty"`
	Compatibility CompatibilitySpec         `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Pricing       PricingSpec               `yaml:"pricing,omitempty" json:"pricing,omitempty"`
	Runtime       RuntimeSpec               `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Permissions   map[string]PermissionDecl `yaml:"permissions,omitempty" json:"permissions,omitempty"`
}

// CompatibilitySpec 兼容性约束
type CompatibilitySpec struct {
	Keystone string `yaml:"keystone,omitempty" json:"keystone,omitempty"`
}

// PricingSpec 定价信息
type PricingSpec struct {
	Type        PricingType `yaml:"type,omitempty" json:"type,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
}

// RuntimeSpec 运行时配置
type RuntimeSpec struct {
	Port        int           `yaml:"port,omitempty" json:"port,omitempty"`
	HealthCheck string        `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	Resources   ResourcesSpec `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// ResourcesSpec 资源限制
type ResourcesSpec struct {
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}

// ParseManifest 从 YAML 字节解析 ManifestSpec
func ParseManifest(data []byte) (*ManifestSpec, error) {
	var m ManifestSpec
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest YAML: %w", err)
	}
	return &m, nil
}

// Validate 校验 ManifestSpec 的必填字段和枚举值
func (m *ManifestSpec) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("manifest: id is required")
	}
	if m.Name == "" {
		return fmt.Errorf("manifest: name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest: version is required")
	}
	if !m.Type.Valid() {
		return fmt.Errorf("manifest: invalid type %q", m.Type)
	}
	if m.Pricing.Type != "" && !m.Pricing.Type.Valid() {
		return fmt.Errorf("manifest: invalid pricing type %q", m.Pricing.Type)
	}
	return nil
}
