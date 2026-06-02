package kstypes

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestAgentCard_RoundTripJSON(t *testing.T) {
	card := AgentCard{
		Name:            "Marketing Squad",
		Description:     "B2B 市场营销团队",
		Version:         "1.0.0",
		ProtocolVersion: "1.0.0",
		URL:             "https://keystone.example.com/a2a/squad-marketing",
		Provider:        AgentProvider{Organization: "self"},
		Capabilities:    AgentCapabilities{Streaming: true, PushNotifications: true},
		SecuritySchemes: map[string]SecurityScheme{},
		Skills:          []A2ASkill{},
	}

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// 验证 skills 字段在 JSON 中为空数组（而非 null）。
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if string(raw["skills"]) != "[]" {
		t.Errorf("expected skills=[] in JSON, got %s", string(raw["skills"]))
	}

	var got AgentCard
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 用 JSON 规范形式比较，避免 nil vs empty slice 的 DeepEqual 问题。
	gotData, _ := json.Marshal(got)
	if !bytes.Equal(data, gotData) {
		t.Errorf("round-trip mismatch:\n got  = %s\n want = %s", gotData, data)
	}
}

// TestPushNotificationConfig_JSONRoundtrip 验证 PushNotificationConfig JSON 编解码 round-trip 字段保留。
// PushNotification 经过 task 落库 / runner 终态读取 / webhook payload 序列化等多个跨边界场景，必须保持稳定字段语义。
func TestPushNotificationConfig_JSONRoundtrip(t *testing.T) {
	cases := []struct {
		name string
		cfg  PushNotificationConfig
	}{
		{
			name: "minimal — 仅 URL",
			cfg:  PushNotificationConfig{URL: "https://marketing.example.com/webhook/a2a"},
		},
		{
			name: "with token — Bearer 鉴权",
			cfg: PushNotificationConfig{
				URL:   "https://marketing.example.com/webhook/a2a",
				Token: "secret-token-abc",
			},
		},
		{
			name: "with authentication — 扩展鉴权",
			cfg: PushNotificationConfig{
				URL:   "https://marketing.example.com/webhook/a2a",
				Token: "fallback-token",
				Authentication: &PushNotificationAuthDetail{
					Schemes:     []string{"Bearer", "Basic"},
					Credentials: "encoded-creds",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := json.Marshal(c.cfg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got PushNotificationConfig
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			gotData, _ := json.Marshal(got)
			if !bytes.Equal(data, gotData) {
				t.Errorf("round-trip mismatch:\n got  = %s\n want = %s", gotData, data)
			}
		})
	}
}

// TestCallChainEntry_JSONRoundtrip 验证 CallChainEntry 切片 JSON 编解码 round-trip 字段保留。
// HeaderA2ACallChain 的 value 是 JSON-encoded []CallChainEntry，调用链跨仓传递时必须保持稳定字段语义。
func TestCallChainEntry_JSONRoundtrip(t *testing.T) {
	entries := []CallChainEntry{
		{AgentID: "marketing", TaskID: "tm1", Ts: 1714800000000},
		{AgentID: "legal", TaskID: "tl1", Ts: 1714800001000},
	}

	buf, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got []CallChainEntry
	if err := json.Unmarshal(buf, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got) != len(entries) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(entries))
	}
	for i := range entries {
		if got[i] != entries[i] {
			t.Errorf("entry %d mismatch: got %+v, want %+v", i, got[i], entries[i])
		}
	}
}
