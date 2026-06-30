package sdui_test

import (
	"encoding/json"
	"testing"

	kstypes "github.com/wuhanyuhan/ks-types"
	"github.com/wuhanyuhan/ks-types/sdui"
)

func TestBuilder_StackWithChildren(t *testing.T) {
	node := sdui.Stack(
		sdui.StackVertical,
		sdui.Card(kstypes.SDUICardProps{Title: "T"},
			sdui.Text(kstypes.SDUITextProps{Text: "hello"}),
		),
	)
	if node.Type != kstypes.PrimitiveStack {
		t.Fatalf("root type = %q", node.Type)
	}
	if len(node.Children) != 1 || node.Children[0].Type != kstypes.PrimitiveCard {
		t.Fatalf("children mismatch: %+v", node.Children)
	}
	// props 应被序列化进 Props
	var cp kstypes.SDUICardProps
	if err := json.Unmarshal(node.Children[0].Props, &cp); err != nil || cp.Title != "T" {
		t.Fatalf("card props mismatch: %v %+v", err, cp)
	}
	// 叶子 text 节点 props 也应可解回
	leaf := node.Children[0].Children[0]
	if leaf.Type != kstypes.PrimitiveText {
		t.Fatalf("leaf type = %q", leaf.Type)
	}
	var tp kstypes.SDUITextProps
	if err := json.Unmarshal(leaf.Props, &tp); err != nil || tp.Text != "hello" {
		t.Fatalf("text props mismatch: %v %+v", err, tp)
	}
}

func TestBuilder_ButtonAndForm(t *testing.T) {
	btn := sdui.Button(kstypes.SDUIButtonProps{Label: "点我", Action: kstypes.SDUIActionIntent{ActionID: "ping"}})
	if btn.Type != kstypes.PrimitiveButton || len(btn.Children) != 0 {
		t.Fatalf("button node malformed: %+v", btn)
	}
	form := sdui.Form(kstypes.SDUIFormProps{Submit: kstypes.SDUIActionIntent{ActionID: "go"}})
	if form.Type != kstypes.PrimitiveForm {
		t.Fatalf("form type = %q", form.Type)
	}
}

func TestBuilder_TableWithPrimaryAction(t *testing.T) {
	table := sdui.Table(kstypes.SDUITableProps{
		Columns: []kstypes.SDUITableColumn{{Key: "display_id", Label: "ID"}},
		Rows:    []map[string]string{{"display_id": "5"}},
		PrimaryAction: &kstypes.SDUITableActionTemplate{
			Label: "查看条文",
			Action: kstypes.SDUIActionIntent{
				ActionID: "open-regulation",
				Intent:   kstypes.SDUIActionIntentConsoleNavigate,
				Route: &kstypes.SDUIConsoleRouteTarget{
					ViewKey: "regulation-detail",
					Params:  map[string]string{"regulation_id": "{{display_id}}"},
				},
			},
		},
	})
	if table.Type != kstypes.PrimitiveTable {
		t.Fatalf("table type = %q", table.Type)
	}
	var props kstypes.SDUITableProps
	if err := json.Unmarshal(table.Props, &props); err != nil {
		t.Fatalf("props: %v", err)
	}
	if props.PrimaryAction == nil || props.PrimaryAction.Action.Route.Params["regulation_id"] != "{{display_id}}" {
		t.Fatalf("primary action mismatch: %+v", props.PrimaryAction)
	}
}

func TestP3Builders(t *testing.T) {
	chart := sdui.Chart(kstypes.SDUIChartProps{ChartType: "line", Categories: []string{"a"}, Series: []kstypes.SDUIChartSeries{{Name: "x", Values: []float64{1}}}})
	if chart.Type != kstypes.PrimitiveChart {
		t.Fatalf("chart type: %s", chart.Type)
	}
	rv := sdui.ReportViewer(kstypes.SDUIReportViewerProps{Title: "周报"}, sdui.Text(kstypes.SDUITextProps{Text: "正文"}))
	if rv.Type != kstypes.PrimitiveReportViewer || len(rv.Children) != 1 {
		t.Fatalf("report-viewer children: %d", len(rv.Children))
	}
	live := sdui.WarRoomLive("run-1")
	if live.Type != kstypes.PrimitiveWarRoom || live.Data == nil || live.Data.Kind != kstypes.DataSourceTeamProgressStream || live.Data.Params["run_id"] != "run-1" {
		t.Fatalf("war-room live data: %+v", live.Data)
	}
	snap := sdui.WarRoomSnapshot("run-2")
	if snap.Data.Kind != kstypes.DataSourceTeamProgressSnapshot {
		t.Fatalf("war-room snapshot kind: %s", snap.Data.Kind)
	}
	taskCollaboration := sdui.TaskCollaborationSnapshot("run-3")
	if taskCollaboration.Type != kstypes.PrimitiveWarRoom || taskCollaboration.Data == nil {
		t.Fatalf("task collaboration node malformed: %+v", taskCollaboration)
	}
	if taskCollaboration.Data.Kind != kstypes.DataSourceTaskCollaborationSnapshot {
		t.Fatalf("task collaboration data kind: %s", taskCollaboration.Data.Kind)
	}
	if taskCollaboration.Data.Params["run_id"] != "run-3" {
		t.Fatalf("task collaboration run_id param: %+v", taskCollaboration.Data.Params)
	}
	shell := sdui.ConsoleShell(kstypes.SDUIConsoleShellProps{Nav: kstypes.NavTree{Items: []kstypes.NavItem{{Key: "d", Label: "总览", Kind: kstypes.NavKindSDUI}}}})
	if shell.Type != kstypes.PrimitiveConsoleShell {
		t.Fatalf("console-shell type: %s", shell.Type)
	}
	if sdui.Slot(kstypes.SDUISlotProps{Path: "/brand"}).Type != kstypes.PrimitiveSlot {
		t.Fatal("slot type")
	}
}
