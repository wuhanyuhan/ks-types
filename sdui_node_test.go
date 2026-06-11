package kstypes

import (
	"encoding/json"
	"testing"
)

func TestUINode_JSONRoundTrip(t *testing.T) {
	n := UINode{
		Type: "stack",
		Children: []UINode{
			{Type: "text", Props: json.RawMessage(`{"text":"hi"}`), Key: "t1"},
		},
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got UINode
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != "stack" || len(got.Children) != 1 || got.Children[0].Type != "text" {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestUINode_DataSource(t *testing.T) {
	n := UINode{Type: "war-room", Data: &UIDataSource{Kind: DataSourceTeamProgressStream, Params: map[string]string{"run_id": "r1"}}}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got UINode
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Data == nil || got.Data.Kind != DataSourceTeamProgressStream || got.Data.Params["run_id"] != "r1" {
		t.Fatalf("data source round trip mismatch: %+v", got.Data)
	}
}

func TestUINode_NoDataSourceOmitted(t *testing.T) {
	// 无数据源时 data 字段应被 omitempty 省略（不污染既有节点 wire 格式）。
	b, err := json.Marshal(UINode{Type: "text"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `{"type":"text"}` {
		t.Fatalf("expected data omitted, got %s", b)
	}
}

func TestUINode_Depth(t *testing.T) {
	n := UINode{Type: "stack", Children: []UINode{{Type: "card", Children: []UINode{{Type: "text"}}}}}
	if d := n.Depth(); d != 3 {
		t.Fatalf("Depth() = %d, want 3", d)
	}
}
