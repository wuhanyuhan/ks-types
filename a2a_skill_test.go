package kstypes

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestA2ASkill_RoundTripJSON(t *testing.T) {
	skill := A2ASkill{
		ID:              "content-writer",
		Name:            "内容撰写",
		Description:     "根据主题生成结构化营销文案",
		InputModes:      []string{"text"},
		OutputModes:     []string{"text"},
		Examples:        []string{"为新品上市写一段 200 字的朋友圈文案"},
		Tags:            []string{"writing", "marketing"},
		AcceptsBindings: []string{"landing_page_repo"},
		ExecutionMode:   "sync",
	}

	data, err := json.Marshal(skill)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got A2ASkill
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 用 JSON 规范形式比较，避免 nil vs empty slice 的 DeepEqual 问题。
	gotData, _ := json.Marshal(got)
	if !bytes.Equal(data, gotData) {
		t.Errorf("round-trip mismatch:\n got  = %s\n want = %s", gotData, data)
	}

	// AcceptsBindings 字段 round-trip 不破现有结构（minor 兼容扩展）
	if strings.Join(got.AcceptsBindings, "|") != "landing_page_repo" {
		t.Errorf("round-trip AcceptsBindings: got %v want [landing_page_repo]", got.AcceptsBindings)
	}
}

// TestA2ASkill_AcceptsBindings_BackwardCompat 验证旧 AgentCard wire format
// （不声明 acceptsBindings 字段）仍能正常解析为 nil。
func TestA2ASkill_AcceptsBindings_BackwardCompat(t *testing.T) {
	legacy := `{
		"id": "search",
		"name": "搜索",
		"description": "全文检索",
		"inputModes": ["text"],
		"outputModes": ["text"],
		"examples": []
	}`

	var skill A2ASkill
	if err := json.Unmarshal([]byte(legacy), &skill); err != nil {
		t.Fatalf("unmarshal legacy AgentCard skill: %v", err)
	}
	if skill.AcceptsBindings != nil {
		t.Errorf("expected AcceptsBindings nil for legacy AgentCard, got %v", skill.AcceptsBindings)
	}
}

// TestA2ASkill_AcceptsBindings_RoundTripJSON 验证 acceptsBindings 字段
// camelCase JSON tag 与 ks-types A2ASkillDef.AcceptsBindings 一致透传。
func TestA2ASkill_AcceptsBindings_RoundTripJSON(t *testing.T) {
	wire := `{"id":"develop","name":"开发","description":"","inputModes":["text"],"outputModes":["text"],"examples":[],"acceptsBindings":["landing_page_repo","scripts_repo"]}`

	var skill A2ASkill
	if err := json.Unmarshal([]byte(wire), &skill); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if strings.Join(skill.AcceptsBindings, "|") != "landing_page_repo|scripts_repo" {
		t.Errorf("AcceptsBindings parse: got %v want [landing_page_repo, scripts_repo]", skill.AcceptsBindings)
	}

	got, err := json.Marshal(skill)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(got), `"acceptsBindings":["landing_page_repo","scripts_repo"]`) {
		t.Errorf("expected acceptsBindings camelCase tag in output, got: %s", got)
	}
}

// TestA2ASkill_ExecutionMode_RoundTripJSON 验证 executionMode 字段
// camelCase JSON tag + 无 omitempty 严格必填语义。minor 升级。
func TestA2ASkill_ExecutionMode_RoundTripJSON(t *testing.T) {
	// async 路径
	asyncSkill := A2ASkill{
		ID:            "develop",
		Name:          "开发",
		InputModes:    []string{"text"},
		OutputModes:   []string{"text"},
		ExecutionMode: "async",
	}
	data, err := json.Marshal(asyncSkill)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"executionMode":"async"`) {
		t.Errorf("expected executionMode camelCase tag in output, got: %s", data)
	}

	var got A2ASkill
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ExecutionMode != "async" {
		t.Errorf("round-trip ExecutionMode: got %q want async", got.ExecutionMode)
	}

	// sync 路径
	syncSkill := A2ASkill{
		ID:            "search",
		Name:          "搜索",
		InputModes:    []string{"text"},
		OutputModes:   []string{"text"},
		ExecutionMode: "sync",
	}
	syncData, _ := json.Marshal(syncSkill)
	if !strings.Contains(string(syncData), `"executionMode":"sync"`) {
		t.Errorf("expected executionMode=sync in output, got: %s", syncData)
	}

	// 无 omitempty：空值（虽然语义不允许出现）也会被 marshal 出来
	emptySkill := A2ASkill{
		ID:          "x",
		Name:        "x",
		InputModes:  []string{"text"},
		OutputModes: []string{"text"},
	}
	emptyData, _ := json.Marshal(emptySkill)
	if !strings.Contains(string(emptyData), `"executionMode":""`) {
		t.Errorf("expected executionMode field present (no omitempty), got: %s", emptyData)
	}
}

func TestA2ASkill_Validate_Success(t *testing.T) {
	skill := A2ASkill{
		ID:          "search",
		Name:        "搜索",
		InputModes:  []string{"text"},
		OutputModes: []string{"text"},
	}
	if err := skill.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestA2ASkill_Validate_Errors(t *testing.T) {
	cases := []struct {
		name      string
		skill     A2ASkill
		wantError string
	}{
		{
			"missing ID",
			A2ASkill{Name: "n", InputModes: []string{"text"}, OutputModes: []string{"text"}},
			"id is required",
		},
		{
			"missing Name",
			A2ASkill{ID: "i", InputModes: []string{"text"}, OutputModes: []string{"text"}},
			"name is required",
		},
		{
			"missing InputModes",
			A2ASkill{ID: "i", Name: "n", OutputModes: []string{"text"}},
			"input_modes is required",
		},
		{
			"missing OutputModes",
			A2ASkill{ID: "i", Name: "n", InputModes: []string{"text"}},
			"output_modes is required",
		},
	}
	for _, c := range cases {
		err := c.skill.Validate()
		if err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.wantError) {
			t.Errorf("%s: error %q doesn't contain %q", c.name, err.Error(), c.wantError)
		}
	}
}
