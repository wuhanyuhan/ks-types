package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWidgetListActionsV1_Validate_Happy(t *testing.T) {
	t.Parallel()
	w := WidgetListActionsV1{
		Title: "草稿列表",
		Items: []WidgetListItem{
			{ID: "1", Title: "5月营销月报"},
		},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetListActionsV1_Validate_EmptyItems_OK(t *testing.T) {
	t.Parallel()
	w := WidgetListActionsV1{Items: []WidgetListItem{}, Empty: &WidgetEmptyState{Title: "暂无草稿"}}
	if err := w.Validate(); err != nil {
		t.Fatalf("empty items + empty state should pass: %v", err)
	}
}

func TestWidgetListActionsV1_Validate_ItemMissingID(t *testing.T) {
	t.Parallel()
	w := WidgetListActionsV1{Items: []WidgetListItem{{Title: "x"}}}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "items[0].id") {
		t.Fatalf("expected items[0].id error, got %v", err)
	}
}

func TestWidgetListActionsV1_Validate_ItemMissingTitle(t *testing.T) {
	t.Parallel()
	w := WidgetListActionsV1{Items: []WidgetListItem{{ID: "1"}}}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "items[0].title") {
		t.Fatalf("expected items[0].title error, got %v", err)
	}
}

func TestWidgetListActionsV1_JSON_OmitEmpty(t *testing.T) {
	t.Parallel()
	w := WidgetListActionsV1{Items: []WidgetListItem{{ID: "1", Title: "x"}}}
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if strings.Contains(s, `"actions"`) || strings.Contains(s, `"empty"`) {
		t.Errorf("expected omitempty for actions/empty, got: %s", s)
	}
}

func TestWidgetDiffReviewV1_Validate_Happy(t *testing.T) {
	t.Parallel()
	w := WidgetDiffReviewV1{
		Title: "审稿",
		Diff: []WidgetDiffSegment{
			{Type: "context", Text: "前文"},
			{Type: "insert", Text: "新段落"},
		},
		Actions: []WidgetActionDescriptor{{ID: "approve", Label: "通过"}},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetDiffReviewV1_Validate_EmptyDiff_Reject(t *testing.T) {
	t.Parallel()
	w := WidgetDiffReviewV1{Title: "x", Diff: []WidgetDiffSegment{}, Actions: []WidgetActionDescriptor{{ID: "a", Label: "b"}}}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "diff requires at least 1 segment") {
		t.Fatalf("expected diff-empty error, got %v", err)
	}
}

func TestWidgetDiffReviewV1_Validate_BadSegmentType(t *testing.T) {
	t.Parallel()
	w := WidgetDiffReviewV1{
		Title:   "x",
		Diff:    []WidgetDiffSegment{{Type: "modify", Text: "x"}},
		Actions: []WidgetActionDescriptor{{ID: "a", Label: "b"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid segment type") {
		t.Fatalf("expected invalid segment type error, got %v", err)
	}
}

func TestWidgetDiffReviewV1_Validate_AnnotationOutOfRange(t *testing.T) {
	t.Parallel()
	w := WidgetDiffReviewV1{
		Title:       "x",
		Diff:        []WidgetDiffSegment{{Type: "context", Text: "x"}},
		Actions:     []WidgetActionDescriptor{{ID: "a", Label: "b"}},
		Annotations: []WidgetDiffAnnotation{{AnchorIndex: 99, Severity: "info", Message: "y"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("expected out-of-range error, got %v", err)
	}
}

func TestWidgetDiffReviewV1_Validate_AnnotationBadSeverity(t *testing.T) {
	t.Parallel()
	w := WidgetDiffReviewV1{
		Title:       "x",
		Diff:        []WidgetDiffSegment{{Type: "context", Text: "x"}},
		Actions:     []WidgetActionDescriptor{{ID: "a", Label: "b"}},
		Annotations: []WidgetDiffAnnotation{{AnchorIndex: 0, Severity: "critical", Message: "y"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid severity") {
		t.Fatalf("expected invalid severity error, got %v", err)
	}
}

func TestWidgetTimelineV1_Validate_Happy(t *testing.T) {
	t.Parallel()
	w := WidgetTimelineV1{
		Title: "Pipeline",
		Events: []WidgetTimelineNode{
			{ID: "n1", Time: "2026-05-04T10:00:00Z", Title: "build", Status: "success"},
		},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetTimelineV1_Validate_BadTimeFormat(t *testing.T) {
	t.Parallel()
	w := WidgetTimelineV1{
		Events: []WidgetTimelineNode{
			{ID: "n1", Time: "2026-05-04 10:00:00", Title: "x", Status: "success"},
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid time") {
		t.Fatalf("expected invalid time error, got %v", err)
	}
}

func TestWidgetTimelineV1_Validate_BadStatus(t *testing.T) {
	t.Parallel()
	w := WidgetTimelineV1{
		Events: []WidgetTimelineNode{
			{ID: "n1", Time: "2026-05-04T10:00:00Z", Title: "x", Status: "succeeded"},
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestWidgetTimelineV1_Validate_MissingID(t *testing.T) {
	t.Parallel()
	w := WidgetTimelineV1{
		Events: []WidgetTimelineNode{
			{Time: "2026-05-04T10:00:00Z", Title: "x", Status: "success"},
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "events[0].id") {
		t.Fatalf("expected events[0].id error, got %v", err)
	}
}

func ptrFloat64(v float64) *float64 { return &v }

func TestWidgetCardGridV1_Validate_Happy_DefaultColumns(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Cards: []WidgetCard{{ID: "c1", Title: "x"}},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetCardGridV1_Validate_ColumnsValid(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Columns: 2,
		Cards:   []WidgetCard{{ID: "c1", Title: "x"}},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetCardGridV1_Validate_ColumnsTooMany(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Columns: 5,
		Cards:   []WidgetCard{{ID: "c1", Title: "x"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "columns must be in [1,4]") {
		t.Fatalf("expected columns range error, got %v", err)
	}
}

func TestWidgetCardGridV1_Validate_BadSourceURLScheme(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Cards: []WidgetCard{{ID: "c1", Title: "x", SourceURL: "javascript:alert(1)"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "unsupported source_url scheme") {
		t.Fatalf("expected unsupported scheme error, got %v", err)
	}
}

func TestWidgetCardGridV1_Validate_ScoreOutOfRange(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Cards: []WidgetCard{{ID: "c1", Title: "x", Score: ptrFloat64(1.5)}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "score must be in [0,1]") {
		t.Fatalf("expected score range error, got %v", err)
	}
}

func TestWidgetCardGridV1_Validate_MissingCardID(t *testing.T) {
	t.Parallel()
	w := WidgetCardGridV1{
		Cards: []WidgetCard{{Title: "x"}},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "cards[0].id") {
		t.Fatalf("expected cards[0].id error, got %v", err)
	}
}

// TestWidgetCardGridV1_Validate_ImageURLBadScheme 校验 WidgetCard.ImageURL 严格 https。
// ImageURL 直接进入前端 <img src> 渲染，必须拒绝 javascript:/data:/http: 等非 https 协议
// 防止 XSS。空字符串视为合法（omitempty 字段）。
//
// 注意：与 SourceURL 语义不同——SourceURL 是用户点击跳转，允许 http+https；
// ImageURL 是 inline 渲染，仅允许 https。
func TestWidgetCardGridV1_Validate_ImageURLBadScheme(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		imageURL  string
		wantErr   bool
		wantSubst string
	}{
		{
			name:      "javascript scheme rejected",
			imageURL:  "javascript:alert(1)",
			wantErr:   true,
			wantSubst: "must be https URL",
		},
		{
			name:      "http scheme rejected",
			imageURL:  "http://example.com/x.png",
			wantErr:   true,
			wantSubst: "must be https URL",
		},
		{
			name:      "data scheme rejected",
			imageURL:  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
			wantErr:   true,
			wantSubst: "must be https URL",
		},
		{
			name:     "https scheme accepted",
			imageURL: "https://example.com/x.png",
			wantErr:  false,
		},
		{
			name:     "empty string accepted (omitempty)",
			imageURL: "",
			wantErr:  false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := WidgetCardGridV1{
				Cards: []WidgetCard{{ID: "c1", Title: "x", ImageURL: tc.imageURL}},
			}
			err := w.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.imageURL)
				}
				if !strings.Contains(err.Error(), tc.wantSubst) {
					t.Fatalf("expected substring %q in error, got %v", tc.wantSubst, err)
				}
				if !strings.Contains(err.Error(), "cards[0].image_url") {
					t.Fatalf("expected error to reference cards[0].image_url, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil for %q, got %v", tc.imageURL, err)
				}
			}
		})
	}
}

func TestWidgetImageVariantsV1_Validate_Happy(t *testing.T) {
	t.Parallel()
	w := WidgetImageVariantsV1{
		Primary: WidgetImageItem{
			ID: "p", URL: "https://cdn.example.com/x.png", AltText: "封面",
		},
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWidgetImageVariantsV1_Validate_HttpURL_Reject(t *testing.T) {
	t.Parallel()
	w := WidgetImageVariantsV1{
		Primary: WidgetImageItem{
			ID: "p", URL: "http://cdn.example.com/x.png", AltText: "x",
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "must be https URL") {
		t.Fatalf("expected https-only error, got %v", err)
	}
}

func TestWidgetImageVariantsV1_Validate_MissingAltText(t *testing.T) {
	t.Parallel()
	w := WidgetImageVariantsV1{
		Primary: WidgetImageItem{ID: "p", URL: "https://cdn.example.com/x.png"},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "primary.alt_text required") {
		t.Fatalf("expected primary.alt_text error, got %v", err)
	}
}

func TestWidgetImageVariantsV1_Validate_NegativeWidth(t *testing.T) {
	t.Parallel()
	w := WidgetImageVariantsV1{
		Primary: WidgetImageItem{
			ID: "p", URL: "https://cdn.example.com/x.png", AltText: "x", Width: -1,
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "width must be >= 0") {
		t.Fatalf("expected width error, got %v", err)
	}
}

func TestWidgetImageVariantsV1_Validate_VariantBadURL(t *testing.T) {
	t.Parallel()
	w := WidgetImageVariantsV1{
		Primary: WidgetImageItem{
			ID: "p", URL: "https://cdn.example.com/x.png", AltText: "x",
		},
		Variants: []WidgetImageItem{
			{ID: "v0", URL: "ftp://x", AltText: "y"},
		},
	}
	err := w.Validate()
	if err == nil || !strings.Contains(err.Error(), "variants[0].url") {
		t.Fatalf("expected variants[0].url error, got %v", err)
	}
}
