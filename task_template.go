package kstypes

import "fmt"

// TaskTemplate manifest.yaml 中 task_templates 段的单条任务模板配置。
//
// Assistant / service 类型应用安装后，平台侧按此清单为新建 agent 回填任务模板，
// 用户在 agent 详情页可看到"开箱即用"的任务卡片，无需手动新建模板。
//
// 历史：早期版本里任务模板硬编码在平台侧、跟内置 preset agent 强耦合。后续 preset agent
// 外移到应用市场，任务模板归属权随 agent 一起转移到应用包 manifest，避免平台仓继续维护跟
// agent 高度耦合的内容。
type TaskTemplate struct {
	// Name 模板标题（如"业务数据周报"），UI 卡片标题。必填。
	Name string `yaml:"name" json:"name"`
	// Description 任务用途简述，UI 卡片副标题。
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Icon Lucide React 图标名（如 BarChart2 / TrendingUp / Code2 / ClipboardList），
	// 不是文件路径——跟 AppSpec.Icon 字段语义独立。
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// Category 任务分类标签（如"数据分析"），UI 用于分组展示。
	// 跟 AppSpec.Category（应用分类）语义独立，两者可同名也可不同。
	Category string `yaml:"category,omitempty" json:"category,omitempty"`
	// InputSchema 输入表单的 JSON Schema（type / properties / required 等），平台侧入库后
	// 前端按此渲染表单。必填，至少含 type+properties。
	InputSchema map[string]any `yaml:"input_schema" json:"input_schema"`
	// DefaultValues 输入字段默认值（map[字段名]默认值）。
	DefaultValues map[string]any `yaml:"default_values,omitempty" json:"default_values,omitempty"`
	// InputMapping 用户输入 → agent prompt 变量的映射规则，keystone task engine 消费。
	InputMapping map[string]any `yaml:"input_mapping,omitempty" json:"input_mapping,omitempty"`
	// SortOrder 排序权重，UI 按升序排列。同值按 created_at 兜底。
	SortOrder int `yaml:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// Validate 校验 TaskTemplate 必填字段。
func (t TaskTemplate) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(t.InputSchema) == 0 {
		return fmt.Errorf("input_schema is required")
	}
	return nil
}
