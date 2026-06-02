package kstypes

import (
	"fmt"
	"net/url"
	"time"
)

// 本文件定义 widgets-protocol-v1 的 5 个 MVP widget 数据 schema 与通用辅助类型。
//
// 每个 widget schema 类型实现 Validate() error，由 keystone NormalizeToolResult
// 在落库前统一调用。Schema 注册表见 widgets_registry.go。

// WidgetActionDescriptor 描述 widget 上一个可点击 action（按钮）。
// 渲染前端用 ActionButton 组件统一处理。
type WidgetActionDescriptor struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Variant       string `json:"variant,omitempty"` // "primary" | "default" | "destructive" | "ghost" | "link"
	Icon          string `json:"icon,omitempty"`
	Disabled      bool   `json:"disabled,omitempty"`
	Tooltip       string `json:"tooltip,omitempty"`
	ConfirmPrompt string `json:"confirm_prompt,omitempty"` // 非空时点击需二次确认
}

// WidgetBadge 是 widget 上的角标标签（如列表行的状态、卡片的 tag）。
type WidgetBadge struct {
	Label   string `json:"label"`
	Variant string `json:"variant,omitempty"`
}

// WidgetEmptyState 是 widget 列表为空时的占位提示。
type WidgetEmptyState struct {
	Icon    string `json:"icon,omitempty"`
	Title   string `json:"title"`
	Message string `json:"message,omitempty"`
}

// validateHTTPSURL 校验 urlStr 是 https URL；空字符串视为合法（caller 决定是否必填）。
// scheme 严格限定为 https，拒绝 http / ftp / javascript / data 等。
// 适用于直接进入前端 inline 渲染的 URL（如 <img src>），防止 XSS 风险。
func validateHTTPSURL(prefix, urlStr string) error {
	if urlStr == "" {
		return nil
	}
	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme != "https" {
		return fmt.Errorf("%s: must be https URL, got %q", prefix, urlStr)
	}
	return nil
}

// =========================================================================
// ks://widgets/list-actions@v1
// 列表 + 行级操作；适用 calendar/email/git issues/legal compliance/web-fetch
// =========================================================================

// WidgetListActionsV1 是 list-actions@v1 widget 的数据 schema。
type WidgetListActionsV1 struct {
	Title   string                   `json:"title,omitempty"`
	Items   []WidgetListItem         `json:"items"`
	Actions []WidgetActionDescriptor `json:"actions,omitempty"`
	Empty   *WidgetEmptyState        `json:"empty,omitempty"`
}

// WidgetListItem 是列表里的单行条目。
type WidgetListItem struct {
	ID         string                   `json:"id"`
	Title      string                   `json:"title"`
	Subtitle   string                   `json:"subtitle,omitempty"`
	Icon       string                   `json:"icon,omitempty"`
	Badges     []WidgetBadge            `json:"badges,omitempty"`
	Metadata   map[string]any           `json:"metadata,omitempty"`
	RowActions []WidgetActionDescriptor `json:"row_actions,omitempty"`
}

// Validate 校验 list-actions widget 数据：每行必须有 id 与 title。
func (w WidgetListActionsV1) Validate() error {
	for i, it := range w.Items {
		if it.ID == "" {
			return fmt.Errorf("items[%d].id is required", i)
		}
		if it.Title == "" {
			return fmt.Errorf("items[%d].title is required", i)
		}
	}
	return nil
}

// =========================================================================
// ks://widgets/diff-review@v1
// Diff + 审批；适用 marketing 审稿/dev PR review/legal 合同对比
// =========================================================================

// WidgetDiffReviewV1 是 diff-review@v1 widget 的数据 schema。
type WidgetDiffReviewV1 struct {
	Title       string                   `json:"title"`
	Subtitle    string                   `json:"subtitle,omitempty"`
	Diff        []WidgetDiffSegment      `json:"diff"`
	Actions     []WidgetActionDescriptor `json:"actions"`
	Annotations []WidgetDiffAnnotation   `json:"annotations,omitempty"`
}

// WidgetDiffSegment 是一段 diff 文本。
type WidgetDiffSegment struct {
	Type string `json:"type"` // "context" | "insert" | "delete"
	Text string `json:"text"`
}

// WidgetDiffAnnotation 是挂在 diff 某段上的批注（可选）。
type WidgetDiffAnnotation struct {
	AnchorIndex int    `json:"anchor_index"`
	Severity    string `json:"severity"` // "info" | "warning" | "error"
	Message     string `json:"message"`
}

var diffSegmentTypes = map[string]bool{"context": true, "insert": true, "delete": true}
var diffAnnotationSeverities = map[string]bool{"info": true, "warning": true, "error": true}

// Validate 校验 diff-review widget 数据：diff 段必须 ≥1 且类型合法；annotation
// 的 anchor_index 必须在 diff 范围内、severity 必须合法。
func (w WidgetDiffReviewV1) Validate() error {
	if len(w.Diff) < 1 {
		return fmt.Errorf("diff requires at least 1 segment")
	}
	for i, seg := range w.Diff {
		if !diffSegmentTypes[seg.Type] {
			return fmt.Errorf("diff[%d]: invalid segment type %q", i, seg.Type)
		}
	}
	for i, ann := range w.Annotations {
		if ann.AnchorIndex < 0 || ann.AnchorIndex >= len(w.Diff) {
			return fmt.Errorf("annotations[%d].anchor_index %d out of range [0,%d)", i, ann.AnchorIndex, len(w.Diff))
		}
		if !diffAnnotationSeverities[ann.Severity] {
			return fmt.Errorf("annotations[%d]: invalid severity %q", i, ann.Severity)
		}
	}
	return nil
}

// =========================================================================
// ks://widgets/timeline@v1
// 时间轴 + 节点详情；适用 dev pipeline/legal deadline/marketing campaign
// =========================================================================

// WidgetTimelineV1 是 timeline@v1 widget 的数据 schema。
type WidgetTimelineV1 struct {
	Title  string               `json:"title,omitempty"`
	Events []WidgetTimelineNode `json:"events"`
}

// WidgetTimelineNode 是时间轴上的单个节点。
// Time 字段必须为 RFC3339（ISO8601 UTC）格式。
type WidgetTimelineNode struct {
	ID       string                   `json:"id"`
	Time     string                   `json:"time"` // ISO8601 UTC（RFC3339）
	Title    string                   `json:"title"`
	Status   string                   `json:"status"` // "pending" | "running" | "success" | "failed" | "skipped"
	Subtitle string                   `json:"subtitle,omitempty"`
	Detail   string                   `json:"detail,omitempty"`
	Actions  []WidgetActionDescriptor `json:"actions,omitempty"`
}

var timelineNodeStatuses = map[string]bool{
	"pending": true,
	"running": true,
	"success": true,
	"failed":  true,
	"skipped": true,
}

// Validate 校验 timeline widget 数据：每个 event 必须有 id；time 必须为 RFC3339；
// status 必须在 {pending, running, success, failed, skipped} 范围内。
func (w WidgetTimelineV1) Validate() error {
	for i, n := range w.Events {
		if n.ID == "" {
			return fmt.Errorf("events[%d].id is required", i)
		}
		if n.Title == "" {
			return fmt.Errorf("events[%d].title is required", i)
		}
		if _, err := time.Parse(time.RFC3339, n.Time); err != nil {
			return fmt.Errorf("events[%d]: invalid time %q (expect RFC3339)", i, n.Time)
		}
		if !timelineNodeStatuses[n.Status] {
			return fmt.Errorf("events[%d]: invalid status %q", i, n.Status)
		}
	}
	return nil
}

// =========================================================================
// ks://widgets/card-grid@v1
// 卡片网格；适用 knowledge 检索结果/web-fetch 搜索结果/git 仓库列表
// =========================================================================

// WidgetCardGridV1 是 card-grid@v1 widget 的数据 schema。
// Columns 取值范围 [1,4]；0 视为未设置（前端默认 2 列）。
type WidgetCardGridV1 struct {
	Title   string       `json:"title,omitempty"`
	Columns int          `json:"columns,omitempty"` // 默认 2，1-4
	Cards   []WidgetCard `json:"cards"`
}

// WidgetCard 是网格中的单张卡片。
// SourceURL 必须为 http/https；Score 若设置必须在 [0,1] 区间。
type WidgetCard struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Excerpt     string                   `json:"excerpt,omitempty"`
	ImageURL    string                   `json:"image_url,omitempty"`
	SourceLabel string                   `json:"source_label,omitempty"`
	SourceURL   string                   `json:"source_url,omitempty"`
	Score       *float64                 `json:"score,omitempty"`
	Badges      []WidgetBadge            `json:"badges,omitempty"`
	Actions     []WidgetActionDescriptor `json:"actions,omitempty"`
}

// Validate 校验 card-grid widget 数据：columns 范围合法；每张卡有 id/title；
// image_url 严格 https（inline 渲染防 XSS）；source_url 允许 http/https
// （用户点击跳转）；score 必须在 [0,1] 区间。
func (w WidgetCardGridV1) Validate() error {
	if w.Columns != 0 && (w.Columns < 1 || w.Columns > 4) {
		return fmt.Errorf("columns must be in [1,4], got %d", w.Columns)
	}
	for i, c := range w.Cards {
		if c.ID == "" {
			return fmt.Errorf("cards[%d].id is required", i)
		}
		if c.Title == "" {
			return fmt.Errorf("cards[%d].title is required", i)
		}
		if err := validateHTTPSURL(fmt.Sprintf("cards[%d].image_url", i), c.ImageURL); err != nil {
			return err
		}
		if c.SourceURL != "" {
			u, err := url.Parse(c.SourceURL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				return fmt.Errorf("cards[%d]: unsupported source_url scheme in %q", i, c.SourceURL)
			}
		}
		if c.Score != nil && (*c.Score < 0 || *c.Score > 1) {
			return fmt.Errorf("cards[%d]: score must be in [0,1], got %v", i, *c.Score)
		}
	}
	return nil
}

// =========================================================================
// ks://widgets/image-variants@v1
// 图片 + 变体操作；适用 image-gen/browser screenshot/document 预览
// =========================================================================

// WidgetImageVariantsV1 是 image-variants@v1 widget 的数据 schema。
type WidgetImageVariantsV1 struct {
	Title    string                   `json:"title,omitempty"`
	Primary  WidgetImageItem          `json:"primary"`
	Variants []WidgetImageItem        `json:"variants,omitempty"`
	Actions  []WidgetActionDescriptor `json:"actions,omitempty"`
}

// WidgetImageItem 是单张图片条目。
// URL 必须为 https；AltText 必填；Width/Height 若非 0 必须 > 0。
type WidgetImageItem struct {
	ID      string                   `json:"id"`
	URL     string                   `json:"url"` // https only
	AltText string                   `json:"alt_text"`
	Width   int                      `json:"width,omitempty"`
	Height  int                      `json:"height,omitempty"`
	Caption string                   `json:"caption,omitempty"`
	Actions []WidgetActionDescriptor `json:"actions,omitempty"`
}

func validateImageItem(prefix string, it WidgetImageItem) error {
	if it.URL == "" {
		return fmt.Errorf("%s.url is required", prefix)
	}
	if err := validateHTTPSURL(prefix+".url", it.URL); err != nil {
		return err
	}
	if it.AltText == "" {
		return fmt.Errorf("%s.alt_text required", prefix)
	}
	if it.Width < 0 {
		return fmt.Errorf("%s.width must be >= 0, got %d", prefix, it.Width)
	}
	if it.Height < 0 {
		return fmt.Errorf("%s.height must be >= 0, got %d", prefix, it.Height)
	}
	return nil
}

// Validate 校验 image-variants widget 数据：primary URL 必须 https；
// alt_text 必填；width/height 必须 > 0；variants 同样校验。
func (w WidgetImageVariantsV1) Validate() error {
	if err := validateImageItem("primary", w.Primary); err != nil {
		return err
	}
	for i, v := range w.Variants {
		if err := validateImageItem(fmt.Sprintf("variants[%d]", i), v); err != nil {
			return err
		}
	}
	return nil
}
