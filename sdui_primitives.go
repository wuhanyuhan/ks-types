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

	// 视图原语（P3 统一作战台）
	PrimitiveChart        = "chart"
	PrimitiveReportViewer = "report-viewer"
	PrimitiveConsoleShell = "console-shell"
	PrimitiveSlot         = "slot"
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
	Columns int    `json:"columns"` // 1..6
	Gap     string `json:"gap,omitempty"`
}

func (p SDUIGridProps) Validate() error {
	if p.Columns < 1 || p.Columns > 6 {
		return fmt.Errorf("grid.columns out of range 1..6: %d", p.Columns)
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
	Key        string `json:"key"`
	Label      string `json:"label"`
	Type       string `json:"type,omitempty"`  // text | datetime | status | number | id
	Width      int    `json:"width,omitempty"` // px；0 表示由前端按类型分配
	Align      string `json:"align,omitempty"` // left | center | right
	Ellipsis   *bool  `json:"ellipsis,omitempty"`
	Sortable   *bool  `json:"sortable,omitempty"`
	Filterable bool   `json:"filterable,omitempty"`
}

type SDUITableProps struct {
	Columns           []SDUITableColumn         `json:"columns"`
	Rows              []map[string]string       `json:"rows"`
	PrimaryAction     *SDUITableActionTemplate  `json:"primary_action,omitempty"`
	RowActions        []SDUITableActionTemplate `json:"row_actions,omitempty"`
	PageSize          int                       `json:"page_size,omitempty"`
	StateKey          string                    `json:"state_key,omitempty"`
	Searchable        *bool                     `json:"searchable,omitempty"`
	SearchPlaceholder string                    `json:"search_placeholder,omitempty"`
	SearchKeys        []string                  `json:"search_keys,omitempty"`
}

func (p SDUITableProps) Validate() error {
	if len(p.Columns) == 0 {
		return fmt.Errorf("table.columns empty")
	}
	for i, col := range p.Columns {
		switch col.Type {
		case "", "text", "datetime", "status", "number", "id":
		default:
			return fmt.Errorf("table.columns[%d].type invalid: %q", i, col.Type)
		}
		switch col.Align {
		case "", "left", "center", "right":
		default:
			return fmt.Errorf("table.columns[%d].align invalid: %q", i, col.Align)
		}
		if col.Width < 0 {
			return fmt.Errorf("table.columns[%d].width negative: %d", i, col.Width)
		}
	}
	if p.PageSize < 0 {
		return fmt.Errorf("table.page_size negative: %d", p.PageSize)
	}
	if p.StateKey != "" && !isConsoleStateKey(p.StateKey) {
		return fmt.Errorf("table.state_key invalid: %q", p.StateKey)
	}
	if p.PrimaryAction != nil && p.PrimaryAction.Action.ActionID == "" {
		return fmt.Errorf("table.primary_action.action.action_id empty")
	}
	for i, action := range p.RowActions {
		if action.Action.ActionID == "" {
			return fmt.Errorf("table.row_actions[%d].action.action_id empty", i)
		}
	}
	return nil
}

func isConsoleStateKey(value string) bool {
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
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

const (
	SDUIActionIntentConsoleNavigate = "console.navigate"
	SDUIActionIntentConsoleBack     = "console.back"
)

type SDUIConsoleRouteTarget struct {
	ViewKey   string            `json:"view_key"`
	Params    map[string]string `json:"params,omitempty"`
	ActiveNav string            `json:"active_nav,omitempty"`
	Replace   bool              `json:"replace,omitempty"`
}

// ConsoleBreadcrumb 描述宿主 console 外层面包屑。Route 为空表示当前页叶子节点。
type ConsoleBreadcrumb struct {
	Label string                  `json:"label"`
	Route *SDUIConsoleRouteTarget `json:"route,omitempty"`
}

// ConsoleViewMeta 是 squad console view 返回给宿主 shell 的页面元信息。
// 宿主负责把它渲染为外层 PageHeader / breadcrumb，不要求各 squad 在 SDUI 树内放返回按钮。
type ConsoleViewMeta struct {
	Title       string                  `json:"title,omitempty"`
	Breadcrumbs []ConsoleBreadcrumb     `json:"breadcrumbs,omitempty"`
	ReturnTo    *SDUIConsoleRouteTarget `json:"return_to,omitempty"`
}

func (m ConsoleViewMeta) Validate() error {
	for i, item := range m.Breadcrumbs {
		if item.Label == "" {
			return fmt.Errorf("console_meta.breadcrumbs[%d].label empty", i)
		}
	}
	return nil
}

// SDUIActionIntent 是 typed 交互意图（禁散文意图 + 前端字符串解析）。
// ToolName 空则回退到当前 widget 的 toolName（见前端 action-dispatcher）。
// 既有 button/form 调用方可继续只传 action_id/tool_name；console 导航调用方传 intent/route。
type SDUIActionIntent struct {
	ActionID string                  `json:"action_id"`
	ToolName string                  `json:"tool_name,omitempty"`
	Intent   string                  `json:"intent,omitempty"`
	Route    *SDUIConsoleRouteTarget `json:"route,omitempty"`
	Payload  map[string]string       `json:"payload,omitempty"`
}

type SDUITableActionTemplate struct {
	Label    string           `json:"label,omitempty"`
	Icon     string           `json:"icon,omitempty"`
	Variant  string           `json:"variant,omitempty"`
	Disabled bool             `json:"disabled,omitempty"`
	Tooltip  string           `json:"tooltip,omitempty"`
	Action   SDUIActionIntent `json:"action"`
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

// --- 视图原语 props（P3 统一作战台）---

type SDUIChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
	Color  string    `json:"color,omitempty"` // 可选；缺省走主题调色板
}

// SDUIChartProps：中性图表（bar/line/pie），domain-neutral，不认识任何领域语义。
type SDUIChartProps struct {
	ChartType  string            `json:"chart_type"` // bar | line | pie
	Title      string            `json:"title,omitempty"`
	Categories []string          `json:"categories"` // x 轴 / 扇区类目
	Series     []SDUIChartSeries `json:"series"`
}

func (p SDUIChartProps) Validate() error {
	switch p.ChartType {
	case "bar", "line", "pie":
	default:
		return fmt.Errorf("chart.chart_type invalid: %q", p.ChartType)
	}
	if len(p.Series) == 0 {
		return fmt.Errorf("chart.series empty")
	}
	for i, s := range p.Series {
		if len(s.Values) != len(p.Categories) {
			return fmt.Errorf("chart.series[%d].values length %d != categories length %d", i, len(s.Values), len(p.Categories))
		}
	}
	return nil
}

// SDUIReportViewerProps：结构化区块文档容器（children 是中性区块），加报告 chrome。
type SDUIReportViewerProps struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
}

func (p SDUIReportViewerProps) Validate() error { return nil }

// SDUIConsoleShellProps：console 外壳。Nav 是二级导航树；ActiveKey 是当前选中项（宿主据路由设）。
// 内容 = 本节点唯一 child（当前视图，由宿主取并注入）。
type SDUIConsoleShellProps struct {
	Title     string  `json:"title,omitempty"`
	ActiveKey string  `json:"active_key,omitempty"`
	Nav       NavTree `json:"nav"`
}

func (p SDUIConsoleShellProps) Validate() error { return p.Nav.Validate() }

// SDUISlotProps：专有岛嵌入。Path 指向 squad 自服务面板，经反代以同源 iframe 承载。
type SDUISlotProps struct {
	Title string `json:"title,omitempty"`
	Path  string `json:"path"`
}

func (p SDUISlotProps) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("slot.path empty")
	}
	return nil
}
