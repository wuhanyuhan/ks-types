package kstypes

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// AppSpec 应用 manifest.yaml 的完整结构
type AppSpec struct {
	ID            string                    `yaml:"id" json:"id"`
	Name          string                    `yaml:"name" json:"name"`
	Version       string                    `yaml:"version" json:"version"`
	Type          AppType                   `yaml:"type" json:"type"`
	Summary       string                    `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description   string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Publisher     string                    `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	Category      string                    `yaml:"category,omitempty" json:"category,omitempty"`
	Tags          []string                  `yaml:"tags,omitempty" json:"tags,omitempty"`
	Protection    ProtectionLevel           `yaml:"protection,omitempty" json:"protection,omitempty"`
	Compatibility CompatibilitySpec         `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Pricing       PricingSpec               `yaml:"pricing,omitempty" json:"pricing,omitempty"`
	Runtime       RuntimeSpec               `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Mount         MountSpec                 `yaml:"mount,omitempty" json:"mount,omitempty"`
	Dependencies  DependenciesSpec          `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
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
	Mode           RuntimeMode   `yaml:"mode,omitempty" json:"mode,omitempty"`
	Entry          string        `yaml:"entry,omitempty" json:"entry,omitempty"`
	WorkingDir     string        `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Image          string        `yaml:"image,omitempty" json:"image,omitempty"`
	Port           int           `yaml:"port,omitempty" json:"port,omitempty"`
	Ports          []string      `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes        []string      `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	HealthCheck    string        `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	HealthCheckURL string        `yaml:"health_check_url,omitempty" json:"health_check_url,omitempty"`
	Environment    []string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	Resources      ResourcesSpec `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// DependencyItem 必需依赖
type DependencyItem struct {
	ID      string `yaml:"id" json:"id"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// RecommendItem 推荐依赖
type RecommendItem struct {
	ID     string `yaml:"id" json:"id"`
	Reason string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// ConflictItem 冲突应用
type ConflictItem struct {
	ID     string `yaml:"id" json:"id"`
	Reason string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// DependenciesSpec 依赖声明
type DependenciesSpec struct {
	Requires   []DependencyItem `yaml:"requires,omitempty" json:"requires,omitempty"`
	Recommends []RecommendItem  `yaml:"recommends,omitempty" json:"recommends,omitempty"`
	Conflicts  []ConflictItem   `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
}

// ResourcesSpec 资源限制
type ResourcesSpec struct {
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}

// MountSpec 安装挂载配置
type MountSpec struct {
	Extension *ExtensionMountSpec `yaml:"extension,omitempty" json:"extension,omitempty"`
	Service   *ServiceMountSpec   `yaml:"service,omitempty" json:"service,omitempty"`
	Assistant *AssistantMountSpec `yaml:"assistant,omitempty" json:"assistant,omitempty"`
	Skill     *SkillMountSpec     `yaml:"skill,omitempty" json:"skill,omitempty"`
}

// ExtensionMountSpec extension 类型挂载
type ExtensionMountSpec struct {
	MCPServerName       string   `yaml:"mcp_server_name" json:"mcp_server_name"`
	TransportType       string   `yaml:"transport_type" json:"transport_type"`
	Endpoint            string   `yaml:"endpoint" json:"endpoint"`
	DefaultAllowedTools []string `yaml:"default_allowed_tools,omitempty" json:"default_allowed_tools,omitempty"`
	// AuthMode /mcp 端点鉴权模式。空字符串等价于 AuthModeNone。
	// 生产推荐使用 keystone_jwks；本地 / 不可签发场景用 none 或 static_bearer。
	AuthMode AuthMode `yaml:"auth_mode,omitempty" json:"auth_mode,omitempty"`
}

// ServiceMountSpec service 类型挂载
type ServiceMountSpec struct {
	AutoRegisterMCP     bool             `yaml:"auto_register_mcp,omitempty" json:"auto_register_mcp,omitempty"`
	MCPEndpoint         string           `yaml:"mcp_endpoint,omitempty" json:"mcp_endpoint,omitempty"`
	DefaultAllowedTools []string         `yaml:"default_allowed_tools,omitempty" json:"default_allowed_tools,omitempty"`
	CreateAgent         bool             `yaml:"create_agent,omitempty" json:"create_agent,omitempty"`
	LLMMode             string           `yaml:"llm_mode,omitempty" json:"llm_mode,omitempty"`
	LLMRequirements     *LLMRequirements `yaml:"llm_requirements,omitempty" json:"llm_requirements,omitempty"`
	// AuthMode /mcp 端点鉴权模式。空字符串等价于 AuthModeNone。
	// 生产推荐使用 keystone_jwks；本地 / 不可签发场景用 none 或 static_bearer。
	AuthMode AuthMode `yaml:"auth_mode,omitempty" json:"auth_mode,omitempty"`
}

// SkillMountSpec skill 类型挂载
type SkillMountSpec struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// AssistantMountSpec assistant 类型挂载
type AssistantMountSpec struct {
	CreateAgent  bool   `yaml:"create_agent,omitempty" json:"create_agent,omitempty"`
	Name         string `yaml:"name,omitempty" json:"name,omitempty"`
	SystemPrompt string `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	RoutingPlan  string `yaml:"routing_plan,omitempty" json:"routing_plan,omitempty"`
}

// LLMRequirements LLM 能力要求
type LLMRequirements struct {
	SupportsVision    bool `yaml:"supports_vision,omitempty" json:"supports_vision,omitempty"`
	SupportsToolCalls bool `yaml:"supports_tool_calls,omitempty" json:"supports_tool_calls,omitempty"`
	MinContextTokens  int  `yaml:"min_context_tokens,omitempty" json:"min_context_tokens,omitempty"`
}

// ParseAppSpec 从 YAML 字节解析 AppSpec
func ParseAppSpec(data []byte) (*AppSpec, error) {
	var m AppSpec
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("解析 manifest YAML 失败: %w", err)
	}
	return &m, nil
}

// Validate 校验 AppSpec 的必填字段和枚举值
func (m *AppSpec) Validate() error {
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

	// RuntimeMode 校验
	if !m.Runtime.Mode.Valid() {
		return fmt.Errorf("manifest: invalid runtime mode %q", m.Runtime.Mode)
	}

	// Protection 校验
	if !m.Protection.Valid() {
		return fmt.Errorf("manifest: invalid protection %q", m.Protection)
	}

	// service mount 的 auth_mode 校验
	if m.Mount.Service != nil && !m.Mount.Service.AuthMode.Valid() {
		return fmt.Errorf("manifest: invalid mount.service.auth_mode %q", m.Mount.Service.AuthMode)
	}

	// extension mount 的 auth_mode 校验
	if m.Mount.Extension != nil && !m.Mount.Extension.AuthMode.Valid() {
		return fmt.Errorf("manifest: invalid mount.extension.auth_mode %q", m.Mount.Extension.AuthMode)
	}

	// skill 类型不能有运行时进程
	if m.Type == AppTypeSkill {
		if m.Runtime.Mode != "" && m.Runtime.Mode != RuntimeModeNone {
			return fmt.Errorf("manifest: type=skill 时 runtime.mode 必须为空或 none")
		}
	}

	// extension mount 必须有 mcp_server_name
	if m.Mount.Extension != nil && m.Mount.Extension.MCPServerName == "" {
		return fmt.Errorf("manifest: mount.extension.mcp_server_name 为必填项")
	}

	return nil
}
