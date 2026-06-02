package kstypes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGateState_CanTransitionTo(t *testing.T) {
	if !GateStatePending.CanTransitionTo(GateStateAnswered) {
		t.Fatal("pending should transition to answered")
	}
	if !GateStatePending.CanTransitionTo(GateStateExpired) {
		t.Fatal("pending should transition to expired")
	}
	if GateStateAnswered.CanTransitionTo(GateStatePending) {
		t.Fatal("answered must be terminal")
	}
	if GateStateExpired.CanTransitionTo(GateStateAnswered) {
		t.Fatal("expired must be terminal")
	}
}

func TestDecisionGate_IsExpiredAt(t *testing.T) {
	g := &DecisionGate{ExpiresAt: 0}
	if g.IsExpiredAt(1000) {
		t.Fatal("expires_at=0 should never expire")
	}
	g.ExpiresAt = 1000
	if g.IsExpiredAt(1000) {
		t.Fatal("same millisecond is still valid")
	}
	if !g.IsExpiredAt(1001) {
		t.Fatal("after expires_at should expire")
	}
}

func TestDeliveryContracts_JSONRoundTrip(t *testing.T) {
	deliverable := Deliverable{
		ID:            "del_1",
		RunID:         "run_1",
		CanonicalName: "squad-marketing.run_content_piece",
		Type:          DeliverableArticle,
		Title:         "新品发布推文",
		Summary:       "已生成一版可发布的中文推文。",
		Preview: PreviewBlock{
			Kind:    PreviewKindMarkdown,
			Content: "## 推文\n今天发布新品。",
		},
		Artifacts: []Artifact{{
			Kind:  ArtifactKindFile,
			Ref:   "file_1",
			Title: "推文正文.md",
		}},
		State:   DeliverableStateFinalized,
		Version: 1,
		Metadata: DeliverableMetadata{
			CreatedAt:   1710000000000,
			CompletedAt: 1710000005000,
			SourceKind:  "mcp_tool",
		},
	}
	activity := ExpertActivity{
		ExpertID:    "writer",
		DisplayName: "内容策划",
		Role:        "市场文案专家",
		Phase:       PhaseDrafting,
		Status:      ActivityStatusRunning,
		Current: &ActivityStep{
			Kind: StepKindAction,
			Text: "正在起草推文",
			Done: false,
			TS:   1710000001000,
		},
		Steps: []ActivityStep{{
			Kind: StepKindReasoning,
			Text: "面向老客户强调升级价值。",
			Done: true,
			TS:   1710000000000,
		}},
		Drafts: []ArtifactRef{{
			Kind:  ArtifactKindFile,
			Ref:   "file_draft_1",
			Title: "推文草稿.md",
		}},
	}
	gate := DecisionGate{
		ID:            "gate_choice_run_1_v1",
		RunID:         "run_1",
		CanonicalName: "squad-marketing.run_content_piece",
		Mode:          GateModeChoice,
		Prompt:        "选择推文方向",
		Options: []Option{{
			ID:    "tone_a",
			Label: "稳健商务",
		}},
		State:     GateStatePending,
		CreatedAt: 1710000002000,
	}

	var gotDeliverable Deliverable
	roundTripJSON(t, deliverable, &gotDeliverable)
	if gotDeliverable.RunID != "run_1" {
		t.Fatalf("run_id mismatch: %s", gotDeliverable.RunID)
	}
	if gotDeliverable.Preview.Kind != PreviewKindMarkdown {
		t.Fatalf("preview.kind mismatch: %s", gotDeliverable.Preview.Kind)
	}

	var gotActivity ExpertActivity
	roundTripJSON(t, activity, &gotActivity)
	if len(gotActivity.Drafts) != 1 || gotActivity.Drafts[0].Ref != "file_draft_1" {
		t.Fatalf("drafts[0].ref mismatch: %#v", gotActivity.Drafts)
	}

	var gotGate DecisionGate
	roundTripJSON(t, gate, &gotGate)
	if gotGate.Mode != GateModeChoice {
		t.Fatalf("gate.mode mismatch: %s", gotGate.Mode)
	}
}

func TestDeliveryContracts_CanonicalFixtures(t *testing.T) {
	fixtures := []struct {
		name string
		ptr  any
	}{
		{"deliverable.json", &Deliverable{}},
		{"expert_activity.json", &ExpertActivity{}},
		{"decision_gate.json", &DecisionGate{}},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata", "delivery", fixture.name))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(b, fixture.ptr); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDeliveryFixtures_IncludedInPackageFiles(t *testing.T) {
	b, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatal(err)
	}

	var pkg struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		t.Fatal(err)
	}

	for _, file := range pkg.Files {
		if file == "testdata/delivery" {
			return
		}
	}
	t.Fatalf("package.json files must include testdata/delivery, got %#v", pkg.Files)
}

func roundTripJSON(t *testing.T, in any, out any) {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}
