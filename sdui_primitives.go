package kstypes

import "fmt"

// 首批原语 type 名常量（小写连字符，跨前后端一致）。
const (
	// 容器原语（含 children，递归）
	PrimitiveStack   = "stack"
	PrimitiveGrid    = "grid"
	PrimitiveCard    = "card"
	PrimitiveSection = "section"
	PrimitiveTabs    = "tabs"
	PrimitiveSplit   = "split"
	// 展示原语（叶子）
	PrimitiveText        = "text"
	PrimitiveMarkdown    = "markdown"
	PrimitiveFieldGroup  = "field-group"
	PrimitiveTable       = "table"
	PrimitiveStatusBadge = "status-badge"
	PrimitiveMetric      = "metric"
	PrimitiveEmptyState  = "empty-state"
	// 交互原语
	PrimitiveButton = "button"
	PrimitiveForm   = "form"
	PrimitiveLink   = "link"
	// 复合原语（v1 遗留 widget 降级；前端复用 widgets_data.go 现有 WidgetXxxV1 类型）
	PrimitiveListActions   = "list-actions"
	PrimitiveDiffReview    = "diff-review"
	PrimitiveTimeline      = "timeline"
	PrimitiveCardGrid      = "card-grid"
	PrimitiveImageVariants = "image-variants"
)

// --- 容器原语 props ---

type SDUIStackProps struct {
	Direction string `json:"direction,omitempty"` // vertical(默认) | horizontal
	Gap       string `json:"gap,omitempty"`       // sm | md | lg
}

func (p SDUIStackProps) Validate() error {
	switch p.Direction {
	case "", "vertical", "horizontal":
	default:
		return fmt.Errorf("stack.direction invalid: %q", p.Direction)
	}
	return nil
}

type SDUIGridProps struct {
	Columns int    `json:"columns"` // 1..4
	Gap     string `json:"gap,omitempty"`
}

func (p SDUIGridProps) Validate() error {
	if p.Columns < 1 || p.Columns > 4 {
		return fmt.Errorf("grid.columns out of range 1..4: %d", p.Columns)
	}
	return nil
}

type SDUICardProps struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
}

func (p SDUICardProps) Validate() error { return nil }

type SDUISectionProps struct {
	Title string `json:"title,omitempty"`
}

func (p SDUISectionProps) Validate() error { return nil }

type SDUITabsItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// SDUITabsProps：tabs 的每个 item 对应 children 同序节点。
type SDUITabsProps struct {
	Items []SDUITabsItem `json:"items"`
}

func (p SDUITabsProps) Validate() error {
	if len(p.Items) == 0 {
		return fmt.Errorf("tabs.items empty")
	}
	return nil
}

type SDUISplitProps struct {
	Ratio string `json:"ratio,omitempty"` // 如 "1:1" | "2:1"
}

func (p SDUISplitProps) Validate() error { return nil }

// --- 展示原语 props ---

type SDUITextProps struct {
	Text    string `json:"text"`
	Variant string `json:"variant,omitempty"` // body(默认) | title | subtitle | caption
}

func (p SDUITextProps) Validate() error {
	switch p.Variant {
	case "", "body", "title", "subtitle", "caption":
		return nil
	default:
		return fmt.Errorf("text.variant invalid: %q", p.Variant)
	}
}

type SDUIMarkdownProps struct {
	Markdown string `json:"markdown"`
}

func (p SDUIMarkdownProps) Validate() error { return nil }

type SDUIField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SDUIFieldGroupProps struct {
	Fields []SDUIField `json:"fields"`
}

func (p SDUIFieldGroupProps) Validate() error { return nil }

type SDUITableColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type SDUITableProps struct {
	Columns []SDUITableColumn   `json:"columns"`
	Rows    []map[string]string `json:"rows"`
}

func (p SDUITableProps) Validate() error {
	if len(p.Columns) == 0 {
		return fmt.Errorf("table.columns empty")
	}
	return nil
}

type SDUIStatusBadgeProps struct {
	Label string `json:"label"`
	Tone  string `json:"tone,omitempty"` // neutral | success | warning | danger | info
}

func (p SDUIStatusBadgeProps) Validate() error {
	switch p.Tone {
	case "", "neutral", "success", "warning", "danger", "info":
		return nil
	default:
		return fmt.Errorf("status-badge.tone invalid: %q", p.Tone)
	}
}

type SDUIMetricProps struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint,omitempty"`
}

func (p SDUIMetricProps) Validate() error { return nil }

type SDUIEmptyStateProps struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

func (p SDUIEmptyStateProps) Validate() error { return nil }

// --- 交互原语 props ---

// SDUIActionIntent 是 typed 交互意图（禁散文意图 + 前端字符串解析）。
// ToolName 空则回退到当前 widget 的 toolName（见前端 action-dispatcher）。
type SDUIActionIntent struct {
	ActionID string `json:"action_id"`
	ToolName string `json:"tool_name,omitempty"`
}

type SDUIButtonProps struct {
	Label   string           `json:"label"`
	Variant string           `json:"variant,omitempty"` // default | primary | destructive | ghost
	Action  SDUIActionIntent `json:"action"`
}

func (p SDUIButtonProps) Validate() error {
	if p.Label == "" {
		return fmt.Errorf("button.label empty")
	}
	if p.Action.ActionID == "" {
		return fmt.Errorf("button.action.action_id empty")
	}
	return nil
}

type SDUIFormField struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Kind     string   `json:"kind"` // text | textarea | select | number
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"` // select 用
}

type SDUIFormProps struct {
	Fields []SDUIFormField  `json:"fields"`
	Submit SDUIActionIntent `json:"submit"`
}

func (p SDUIFormProps) Validate() error {
	if p.Submit.ActionID == "" {
		return fmt.Errorf("form.submit.action_id empty")
	}
	return nil
}

type SDUILinkProps struct {
	Label string `json:"label"`
	Href  string `json:"href"` // 仅 https（前端再校验）
}

func (p SDUILinkProps) Validate() error {
	if p.Href == "" {
		return fmt.Errorf("link.href empty")
	}
	return nil
}
