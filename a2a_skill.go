package kstypes

import "fmt"

// A2ASkill 描述 A2A Agent 对外暴露的单项能力，遵循 A2A v1.0 规范。
type A2ASkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	InputModes  []string `json:"inputModes"`
	OutputModes []string `json:"outputModes"`
	Examples    []string `json:"examples"`
	Tags        []string `json:"tags,omitempty"`

	// AcceptsBindings 列出本 skill 可消费的集成绑定语义键（binding_kind，
	// 如 "landing_page_repo"），由 manifest A2ASkillDef.AcceptsBindings 透传到
	// AgentCard wire format。keystone Agent 拉到本字段后启用 binding 注入路径。
	// minor 兼容扩展（缺省字段不影响旧 AgentCard 解析）。
	AcceptsBindings []string `json:"acceptsBindings,omitempty"`

	// ExecutionMode 与 A2ASkillDef.ExecutionMode 透传（minor 升级，破坏性）。
	// sync / async 二选一；wire format 中**总是**有值 — 旧 squad 升 ks-types v0.9.5+ 后
	// 不带 execution_mode 启动会 fail（manifest layer 严格必填，启动期 fail-fast），
	// 因此 wire format 永远不会有空值。无 omitempty。
	ExecutionMode string `json:"executionMode"`
}

// Validate 校验 A2ASkill 的必填字段。
func (s A2ASkill) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("a2a skill id is required")
	}
	if s.Name == "" {
		return fmt.Errorf("a2a skill name is required")
	}
	if len(s.InputModes) == 0 {
		return fmt.Errorf("a2a skill input_modes is required")
	}
	if len(s.OutputModes) == 0 {
		return fmt.Errorf("a2a skill output_modes is required")
	}
	return nil
}
