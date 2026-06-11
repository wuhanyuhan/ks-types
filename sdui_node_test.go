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

func TestUINode_Depth(t *testing.T) {
	n := UINode{Type: "stack", Children: []UINode{{Type: "card", Children: []UINode{{Type: "text"}}}}}
	if d := n.Depth(); d != 3 {
		t.Fatalf("Depth() = %d, want 3", d)
	}
}
