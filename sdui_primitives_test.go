package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSDUIPrimitiveSchemas_Coverage(t *testing.T) {
	// 首批原语必须全部注册
	want := []string{
		PrimitiveStack, PrimitiveGrid, PrimitiveCard, PrimitiveSection, PrimitiveTabs, PrimitiveSplit,
		PrimitiveText, PrimitiveMarkdown, PrimitiveFieldGroup, PrimitiveTable, PrimitiveStatusBadge, PrimitiveMetric, PrimitiveEmptyState,
		PrimitiveButton, PrimitiveForm, PrimitiveLink,
		PrimitiveListActions, PrimitiveDiffReview, PrimitiveTimeline, PrimitiveCardGrid, PrimitiveImageVariants,
	}
	for _, k := range want {
		if _, ok := SDUIPrimitiveSchemas[k]; !ok {
			t.Errorf("primitive %q not registered in SDUIPrimitiveSchemas", k)
		}
	}
}

func TestContainerPrimitives_OnlyContainers(t *testing.T) {
	for _, c := range []string{PrimitiveStack, PrimitiveGrid, PrimitiveCard, PrimitiveSection, PrimitiveTabs, PrimitiveSplit} {
		if !ContainerPrimitives[c] {
			t.Errorf("container primitive %q missing from ContainerPrimitives", c)
		}
	}
	for _, leaf := range []string{PrimitiveText, PrimitiveButton, PrimitiveTable, PrimitiveListActions} {
		if ContainerPrimitives[leaf] {
			t.Errorf("leaf primitive %q wrongly marked as container", leaf)
		}
	}
}

func TestSDUIStackProps_Validate(t *testing.T) {
	if err := (SDUIStackProps{Direction: "vertical"}).Validate(); err != nil {
		t.Errorf("valid stack rejected: %v", err)
	}
	if err := (SDUIStackProps{Direction: "diagonal"}).Validate(); err == nil {
		t.Error("invalid direction accepted")
	}
}

func TestSDUIGridProps_Validate(t *testing.T) {
	// 1..6 均有效（概览卡放开 5/6 列：消除 5 张概览卡 4+1 右侧留白）
	for _, n := range []int{1, 3, 4, 5, 6} {
		if err := (SDUIGridProps{Columns: n}).Validate(); err != nil {
			t.Errorf("valid grid columns=%d rejected: %v", n, err)
		}
	}
	// 越界（<1 或 >6）拒绝
	for _, n := range []int{0, 7, 9} {
		if err := (SDUIGridProps{Columns: n}).Validate(); err == nil {
			t.Errorf("grid.columns=%d accepted（应越界拒绝）", n)
		}
	}
}

func TestSDUIButtonProps_Validate(t *testing.T) {
	if err := (SDUIButtonProps{Label: "确认", Action: SDUIActionIntent{ActionID: "approve"}}).Validate(); err != nil {
		t.Errorf("valid button rejected: %v", err)
	}
	if err := (SDUIButtonProps{Label: "", Action: SDUIActionIntent{ActionID: "x"}}).Validate(); err == nil {
		t.Error("empty label accepted")
	}
	if err := (SDUIButtonProps{Label: "确认", Action: SDUIActionIntent{}}).Validate(); err == nil {
		t.Error("empty action_id accepted")
	}
}

func TestSDUIActionIntent_ConsoleNavigateJSON(t *testing.T) {
	action := SDUIActionIntent{
		ActionID: "open-regulation",
		Intent:   SDUIActionIntentConsoleNavigate,
		Route: &SDUIConsoleRouteTarget{
			ViewKey:   "regulation-detail",
			ActiveNav: "regulations",
			Params:    map[string]string{"regulation_id": "{{display_id}}"},
		},
	}

	b, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"intent":"console.navigate"`) {
		t.Fatalf("missing intent in %s", b)
	}
	if !strings.Contains(string(b), `"view_key":"regulation-detail"`) {
		t.Fatalf("missing view key in %s", b)
	}

	var got SDUIActionIntent
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Route == nil || got.Route.Params["regulation_id"] != "{{display_id}}" {
		t.Fatalf("route round trip mismatch: %+v", got.Route)
	}
}

func TestSDUITableProps_WithExplicitRowActions(t *testing.T) {
	props := SDUITableProps{
		Columns: []SDUITableColumn{{Key: "display_id", Label: "ID"}, {Key: "law", Label: "法规"}},
		Rows:    []map[string]string{{"display_id": "5", "law": "中华人民共和国民法典"}},
		PrimaryAction: &SDUITableActionTemplate{
			Label: "查看条文",
			Icon:  "book-open",
			Action: SDUIActionIntent{
				ActionID: "open-regulation",
				Intent:   SDUIActionIntentConsoleNavigate,
				Route: &SDUIConsoleRouteTarget{
					ViewKey:   "regulation-detail",
					ActiveNav: "regulations",
					Params:    map[string]string{"regulation_id": "{{display_id}}"},
				},
			},
		},
	}

	if err := props.Validate(); err != nil {
		t.Fatalf("valid table rejected: %v", err)
	}

	b, err := json.Marshal(props)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"primary_action"`) {
		t.Fatalf("primary action omitted: %s", b)
	}
}

func TestSDUITableProps_BackwardCompatibleRowsOnly(t *testing.T) {
	props := SDUITableProps{
		Columns: []SDUITableColumn{{Key: "name", Label: "名称"}},
		Rows:    []map[string]string{{"name": "旧表格"}},
	}
	if err := props.Validate(); err != nil {
		t.Fatalf("legacy table rejected: %v", err)
	}
}

func TestP3ViewPrimitiveProps(t *testing.T) {
	// chart：series 值长度须 == categories
	good := SDUIChartProps{ChartType: "bar", Categories: []string{"一月", "二月"}, Series: []SDUIChartSeries{{Name: "GMV", Values: []float64{1, 2}}}}
	if err := good.Validate(); err != nil {
		t.Fatalf("good chart rejected: %v", err)
	}
	bad := []interface{ Validate() error }{
		SDUIChartProps{ChartType: "pyramid", Categories: []string{"a"}, Series: []SDUIChartSeries{{Values: []float64{1}}}},
		SDUIChartProps{ChartType: "bar", Categories: []string{"a", "b"}, Series: []SDUIChartSeries{{Values: []float64{1}}}}, // 长度不齐
		SDUIChartProps{ChartType: "bar", Categories: []string{"a"}, Series: nil},                                            // 空 series
		SDUISlotProps{Path: ""},               // slot 缺 path
		SDUIConsoleShellProps{Nav: NavTree{}}, // nav 空
	}
	for i, p := range bad {
		if err := p.Validate(); err == nil {
			t.Errorf("bad[%d]: expected error", i)
		}
	}
	// report-viewer / 合法 console-shell 无必填
	if err := (SDUIReportViewerProps{Title: "周报"}).Validate(); err != nil {
		t.Errorf("report-viewer: %v", err)
	}
	if err := (SDUIConsoleShellProps{Title: "营销台", Nav: NavTree{Items: []NavItem{{Key: "d", Label: "总览", Kind: NavKindSDUI}}}}).Validate(); err != nil {
		t.Errorf("console-shell: %v", err)
	}
}
