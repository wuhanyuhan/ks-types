package kstypes

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTaskTemplate_Validate_HappyPath(t *testing.T) {
	tpl := TaskTemplate{
		Name: "业务数据周报",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"date_range": map[string]any{"type": "string"},
			},
			"required": []any{"date_range"},
		},
	}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("期望通过，实际 error: %v", err)
	}
}

func TestTaskTemplate_Validate_MissingName(t *testing.T) {
	tpl := TaskTemplate{
		InputSchema: map[string]any{"type": "object"},
	}
	err := tpl.Validate()
	if err == nil {
		t.Fatal("期望 name 缺失校验失败")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("错误信息应含 name，got: %v", err)
	}
}

func TestTaskTemplate_Validate_MissingInputSchema(t *testing.T) {
	tpl := TaskTemplate{Name: "x"}
	err := tpl.Validate()
	if err == nil {
		t.Fatal("期望 input_schema 缺失校验失败")
	}
	if !strings.Contains(err.Error(), "input_schema") {
		t.Errorf("错误信息应含 input_schema，got: %v", err)
	}
}

func TestTaskTemplate_Validate_EmptyInputSchema(t *testing.T) {
	tpl := TaskTemplate{
		Name:        "x",
		InputSchema: map[string]any{},
	}
	if err := tpl.Validate(); err == nil {
		t.Fatal("期望空 input_schema 校验失败")
	}
}

// TestParseAppSpec_TaskTemplates_YAML 验证 manifest.yaml 中 task_templates 段嵌套
// JSON Schema 能正确解析；这是 ks-agents 应用包的核心 wire format。
func TestParseAppSpec_TaskTemplates_YAML(t *testing.T) {
	raw := []byte(`
id: agent-test
name: test
version: 1.0.0
type: agent
runtime:
  mode: none
task_templates:
  - name: 业务数据周报
    description: 对指定时间段做周报分析
    icon: BarChart2
    category: 数据分析
    input_schema:
      type: object
      properties:
        date_range:
          type: string
          title: 分析周期
        metrics:
          type: string
          maxLength: 200
      required:
        - date_range
        - metrics
    default_values:
      metrics: "DAU, 收入"
    sort_order: 10
  - name: 专项数据分析
    icon: TrendingUp
    category: 数据分析
    input_schema:
      type: object
      properties:
        question:
          type: string
      required:
        - question
    sort_order: 20
`)
	spec, err := ParseAppSpec(raw)
	if err != nil {
		t.Fatalf("ParseAppSpec: %v", err)
	}
	if len(spec.TaskTemplates) != 2 {
		t.Fatalf("task_templates len: got %d, want 2", len(spec.TaskTemplates))
	}

	first := spec.TaskTemplates[0]
	if first.Name != "业务数据周报" {
		t.Errorf("[0].Name = %q", first.Name)
	}
	if first.Icon != "BarChart2" {
		t.Errorf("[0].Icon = %q", first.Icon)
	}
	if first.Category != "数据分析" {
		t.Errorf("[0].Category = %q", first.Category)
	}
	if first.SortOrder != 10 {
		t.Errorf("[0].SortOrder = %d", first.SortOrder)
	}
	if first.InputSchema["type"] != "object" {
		t.Errorf("[0].InputSchema.type = %v", first.InputSchema["type"])
	}
	if first.DefaultValues["metrics"] != "DAU, 收入" {
		t.Errorf("[0].DefaultValues.metrics = %v", first.DefaultValues["metrics"])
	}

	// 关键守护：嵌套 map 的 keys 必须是 string（yaml.v3 行为），否则下游 json.Marshal 会失败。
	props, ok := first.InputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("[0].InputSchema.properties 应为 map[string]any，实际类型 %T", first.InputSchema["properties"])
	}
	dateRange, ok := props["date_range"].(map[string]any)
	if !ok {
		t.Fatalf("date_range 应为 map[string]any，实际类型 %T", props["date_range"])
	}
	if dateRange["type"] != "string" {
		t.Errorf("date_range.type = %v", dateRange["type"])
	}

	// 验证嵌套结构能完整 json.Marshal（这是平台侧入库时持久化 input_schema 的前置条件）
	jsonBytes, err := json.Marshal(first.InputSchema)
	if err != nil {
		t.Fatalf("InputSchema json.Marshal: %v", err)
	}
	if !strings.Contains(string(jsonBytes), `"date_range"`) {
		t.Errorf("Marshal 后丢字段，got: %s", jsonBytes)
	}
}

// TestAppSpec_Validate_TaskTemplates 集成校验：AppSpec.Validate() 应级联校验 task_templates。
func TestAppSpec_Validate_TaskTemplates(t *testing.T) {
	spec := &AppSpec{
		ID: "x", Name: "x", Version: "1.0.0", Type: AppTypeAgent,
		Runtime: RuntimeSpec{Mode: RuntimeModeNone},
		TaskTemplates: []TaskTemplate{
			{Name: "ok", InputSchema: map[string]any{"type": "object"}},
			{Name: "", InputSchema: map[string]any{"type": "object"}}, // 第 2 条 name 缺失
		},
	}
	err := spec.Validate()
	if err == nil {
		t.Fatal("期望 task_templates[1] 校验失败")
	}
	if !strings.Contains(err.Error(), "task_templates[1]") {
		t.Errorf("错误信息应定位到下标 [1]，got: %v", err)
	}
}

// TestAppSpec_TaskTemplates_OmitEmpty 不声明 task_templates 时 yaml/json 输出不应出现该字段，
// 保护旧 manifest 解析向后兼容（无该字段的 22 个旧 agent manifest 仍然正常）。
func TestAppSpec_TaskTemplates_OmitEmpty(t *testing.T) {
	spec := AppSpec{
		ID: "x", Name: "x", Version: "1.0.0", Type: AppTypeSkill,
	}
	yamlOut, err := yaml.Marshal(&spec)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	if strings.Contains(string(yamlOut), "task_templates") {
		t.Errorf("空 TaskTemplates 不应出现在 yaml 输出，got: %s", yamlOut)
	}
	jsonOut, err := json.Marshal(&spec)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(jsonOut), "task_templates") {
		t.Errorf("空 TaskTemplates 不应出现在 json 输出，got: %s", jsonOut)
	}
}

// TestAppSpec_TaskTemplates_RoundTrip yaml 序列化往返不丢字段。
func TestAppSpec_TaskTemplates_RoundTrip(t *testing.T) {
	orig := AppSpec{
		ID: "x", Name: "x", Version: "1.0.0", Type: AppTypeAgent,
		Runtime: RuntimeSpec{Mode: RuntimeModeNone},
		TaskTemplates: []TaskTemplate{
			{
				Name:        "周报",
				Description: "对周数据复盘",
				Icon:        "BarChart2",
				Category:    "数据分析",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"date_range": map[string]any{"type": "string"},
					},
				},
				DefaultValues: map[string]any{"date_range": "本周"},
				SortOrder:     10,
			},
		},
	}
	data, err := yaml.Marshal(&orig)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	parsed, err := ParseAppSpec(data)
	if err != nil {
		t.Fatalf("ParseAppSpec: %v", err)
	}
	if len(parsed.TaskTemplates) != 1 {
		t.Fatalf("len = %d", len(parsed.TaskTemplates))
	}
	got := parsed.TaskTemplates[0]
	if got.Name != orig.TaskTemplates[0].Name {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Icon != "BarChart2" {
		t.Errorf("Icon = %q", got.Icon)
	}
	if got.SortOrder != 10 {
		t.Errorf("SortOrder = %d", got.SortOrder)
	}
	if got.DefaultValues["date_range"] != "本周" {
		t.Errorf("DefaultValues round-trip 丢字段: %v", got.DefaultValues)
	}
}
