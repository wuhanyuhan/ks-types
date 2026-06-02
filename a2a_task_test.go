package kstypes

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestA2ATaskState_Transitions(t *testing.T) {
	tests := []struct {
		from  A2ATaskState
		to    A2ATaskState
		valid bool
	}{
		{A2ATaskSubmitted, A2ATaskWorking, true},
		{A2ATaskWorking, A2ATaskCompleted, true},
		{A2ATaskWorking, A2ATaskInputRequired, true},
		{A2ATaskWorking, A2ATaskRejected, true}, // 业务侧识别重大违规直接 reject
		{A2ATaskInputRequired, A2ATaskWorking, true},
		{A2ATaskCompleted, A2ATaskWorking, false}, // 终态不可回退
		{A2ATaskFailed, A2ATaskCompleted, false},
		{A2ATaskRejected, A2ATaskWorking, false}, // 终态不可回退
		// 特殊分支：非终态可以取消
		{A2ATaskSubmitted, A2ATaskCanceled, true},
		{A2ATaskInputRequired, A2ATaskCanceled, true},
		// 特殊分支：终态不可取消
		{A2ATaskCompleted, A2ATaskCanceled, false},
	}
	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.valid {
			t.Errorf("CanTransitionTo(%q → %q): got %v, want %v", tt.from, tt.to, got, tt.valid)
		}
	}
}

func TestA2ATaskState_IsTerminal(t *testing.T) {
	cases := []struct {
		state    A2ATaskState
		terminal bool
	}{
		{A2ATaskSubmitted, false},
		{A2ATaskWorking, false},
		{A2ATaskInputRequired, false},
		{A2ATaskAuthRequired, false},
		{A2ATaskCompleted, true},
		{A2ATaskFailed, true},
		{A2ATaskCanceled, true},
		{A2ATaskRejected, true},
	}
	for _, c := range cases {
		if got := c.state.IsTerminal(); got != c.terminal {
			t.Errorf("IsTerminal(%q): got %v, want %v", c.state, got, c.terminal)
		}
	}
}

func TestA2AMessage_RoundTripJSON(t *testing.T) {
	msg := A2AMessage{
		Role: "user",
		Parts: []A2AMessagePart{
			{Type: "text", Text: "hello"},
			{Type: "file", File: &A2AFileReference{
				Name:     "a.txt",
				MimeType: "text/plain",
				URI:      "https://example.com/a.txt",
			}},
			{Type: "data", Data: map[string]any{"key": "value"}},
			{Type: "file", File: &A2AFileReference{
				Name:     "b.bin",
				MimeType: "application/octet-stream",
				Bytes:    "aGVsbG8=", // "hello" in base64
			}},
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got A2AMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  = %+v\n want = %+v", got, msg)
	}
}

// TestA2ATask_PushNotification_JSON 验证 A2ATask.PushNotification 字段 omitempty + 嵌套 round-trip 行为。
// nil 时 JSON 不出现 pushNotification key（兼容既有 caller / store 序列化）；
// 非 nil 时嵌套字段（URL / Token）正确保留。
func TestA2ATask_PushNotification_JSON(t *testing.T) {
	// 不带 PushNotification — 输出不包含 pushNotification key
	task := A2ATask{ID: "t1", State: A2ATaskWorking}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	if _, has := raw["pushNotification"]; has {
		t.Errorf("expected pushNotification omitted when nil, got %s", string(data))
	}

	// 带 PushNotification — round-trip 字段保留
	task.PushNotification = &PushNotificationConfig{
		URL:   "https://marketing.example.com/webhook/a2a",
		Token: "tok-1",
	}
	data, err = json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal with notification: %v", err)
	}
	var got A2ATask
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PushNotification == nil {
		t.Fatalf("PushNotification missing after round-trip; data=%s", data)
	}
	if got.PushNotification.URL != task.PushNotification.URL {
		t.Errorf("URL mismatch: got %q, want %q", got.PushNotification.URL, task.PushNotification.URL)
	}
	if got.PushNotification.Token != task.PushNotification.Token {
		t.Errorf("Token mismatch: got %q, want %q", got.PushNotification.Token, task.PushNotification.Token)
	}
}

func TestA2AArtifact_RoundTripJSON(t *testing.T) {
	art := A2AArtifact{
		Name:  "result",
		Parts: []A2AMessagePart{{Type: "text", Text: "done"}},
		Metadata: map[string]any{
			"timestamp": "2026-05-04",
		},
	}
	data, err := json.Marshal(art)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got A2AArtifact
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, art) {
		t.Errorf("round-trip mismatch:\n got  = %+v\n want = %+v", got, art)
	}
}

// === A2A Task Lifecycle Protocol tests ===

// TestLifecycleEvent_RoundTripJSON 验证 LifecycleEvent 全字段 marshal / unmarshal 可逆（含 omitempty 字段非空时回填）。
func TestLifecycleEvent_RoundTripJSON(t *testing.T) {
	// 含全部字段（completed event with result_payload）
	original := LifecycleEvent{
		TaskID:        "550e8400-e29b-41d4-a716-446655440000",
		Status:        LifecycleStatusCompleted,
		Timestamp:     time.Date(2026, 5, 8, 10, 23, 50, 123_000_000, time.UTC),
		StageMessage:  "已完成",
		ResultPayload: json.RawMessage(`{"pr_url":"https://github.com/x/y/pull/1"}`),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got LifecycleEvent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TaskID != original.TaskID {
		t.Errorf("TaskID mismatch: got %q, want %q", got.TaskID, original.TaskID)
	}
	if got.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", got.Status, original.Status)
	}
	if !got.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", got.Timestamp, original.Timestamp)
	}
	if got.StageMessage != original.StageMessage {
		t.Errorf("StageMessage mismatch: got %q, want %q", got.StageMessage, original.StageMessage)
	}
	if string(got.ResultPayload) != string(original.ResultPayload) {
		t.Errorf("ResultPayload mismatch: got %s, want %s", got.ResultPayload, original.ResultPayload)
	}

	// omitempty 验证：started event 不带 stage_message / failure_reason / result_payload → JSON 不出现这些 key
	started := LifecycleEvent{
		TaskID:    "task-1",
		Status:    LifecycleStatusStarted,
		Timestamp: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
	}
	data2, _ := json.Marshal(started)
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data2, &raw)
	for _, key := range []string{"stage_message", "failure_reason", "result_payload"} {
		if _, has := raw[key]; has {
			t.Errorf("expected %q omitted when empty, got %s", key, string(data2))
		}
	}
}

// TestLifecycleStatus_Constants 验证 4 个 LifecycleStatus 常量值与协议规范一致。
func TestLifecycleStatus_Constants(t *testing.T) {
	cases := []struct {
		c    LifecycleStatus
		want string
	}{
		{LifecycleStatusStarted, "started"},
		{LifecycleStatusProgress, "progress"},
		{LifecycleStatusCompleted, "completed"},
		{LifecycleStatusFailed, "failed"},
	}
	for _, c := range cases {
		if string(c.c) != c.want {
			t.Errorf("LifecycleStatus const = %q, want %q", c.c, c.want)
		}
	}
}

// TestLifecycleEvent_Validate_Started_Happy 验证 started event happy path（无 stage_message / 无 failure_reason）。
func TestLifecycleEvent_Validate_Started_Happy(t *testing.T) {
	e := &LifecycleEvent{
		TaskID:    "task-1",
		Status:    LifecycleStatusStarted,
		Timestamp: time.Now().UTC(),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("Validate started happy: got err %v", err)
	}

	// 含 stage_message 也允许（started event stage_message 是可选）
	e.StageMessage = "开始处理"
	if err := e.Validate(); err != nil {
		t.Errorf("Validate started with stage_message: got err %v", err)
	}
}

// TestLifecycleEvent_Validate_Progress_Happy 验证 progress event 必填 stage_message happy path。
func TestLifecycleEvent_Validate_Progress_Happy(t *testing.T) {
	e := &LifecycleEvent{
		TaskID:       "task-1",
		Status:       LifecycleStatusProgress,
		Timestamp:    time.Now().UTC(),
		StageMessage: "正在 clone repo",
	}
	if err := e.Validate(); err != nil {
		t.Errorf("Validate progress happy: got err %v", err)
	}
}

// TestLifecycleEvent_Validate_Progress_MissingStageMessage_Err 验证 progress 缺 stage_message → err 含 "progress"。
func TestLifecycleEvent_Validate_Progress_MissingStageMessage_Err(t *testing.T) {
	e := &LifecycleEvent{
		TaskID:    "task-1",
		Status:    LifecycleStatusProgress,
		Timestamp: time.Now().UTC(),
	}
	err := e.Validate()
	if err == nil {
		t.Fatalf("Validate progress missing stage_message: expected err, got nil")
	}
	if !strings.Contains(err.Error(), "progress") {
		t.Errorf("Validate progress missing: err %q does not mention 'progress'", err)
	}

	// progress event 不应带 failure_reason / result_payload
	e.StageMessage = "ok"
	e.FailureReason = "should not be here"
	if err := e.Validate(); err == nil || !strings.Contains(err.Error(), "progress") {
		t.Errorf("Validate progress with failure_reason: expected err mentioning progress, got %v", err)
	}
}

// TestLifecycleEvent_Validate_Failed_MissingFailureReason_Err 验证 failed 缺 failure_reason → err 含 "failed"。
func TestLifecycleEvent_Validate_Failed_MissingFailureReason_Err(t *testing.T) {
	e := &LifecycleEvent{
		TaskID:    "task-1",
		Status:    LifecycleStatusFailed,
		Timestamp: time.Now().UTC(),
	}
	err := e.Validate()
	if err == nil {
		t.Fatalf("Validate failed missing failure_reason: expected err, got nil")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("Validate failed missing: err %q does not mention 'failed'", err)
	}

	// 含 failure_reason 后通过
	e.FailureReason = "build error: missing dep"
	if err := e.Validate(); err != nil {
		t.Errorf("Validate failed with failure_reason: got err %v", err)
	}
}

// TestLifecycleEvent_Validate_EdgeCases 验证表驱动边角 case：task_id 空 / Timestamp 零 / unknown status / completed 带 failure_reason / started 带 result_payload。
func TestLifecycleEvent_Validate_EdgeCases(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name       string
		event      LifecycleEvent
		errSubstr  string
	}{
		{
			name:      "task_id 空",
			event:     LifecycleEvent{Status: LifecycleStatusStarted, Timestamp: now},
			errSubstr: "task_id",
		},
		{
			name:      "Timestamp 零值",
			event:     LifecycleEvent{TaskID: "t1", Status: LifecycleStatusStarted},
			errSubstr: "timestamp",
		},
		{
			name:      "unknown status",
			event:     LifecycleEvent{TaskID: "t1", Status: LifecycleStatus("bogus"), Timestamp: now},
			errSubstr: "unknown",
		},
		{
			name: "completed 带 failure_reason",
			event: LifecycleEvent{
				TaskID: "t1", Status: LifecycleStatusCompleted, Timestamp: now,
				FailureReason: "should not be here",
			},
			errSubstr: "completed",
		},
		{
			name: "started 带 result_payload",
			event: LifecycleEvent{
				TaskID: "t1", Status: LifecycleStatusStarted, Timestamp: now,
				ResultPayload: json.RawMessage(`{"x":1}`),
			},
			errSubstr: "started",
		},
	}
	for _, c := range cases {
		err := c.event.Validate()
		if err == nil {
			t.Errorf("%s: expected err, got nil", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.errSubstr) {
			t.Errorf("%s: err %q does not contain %q", c.name, err, c.errSubstr)
		}
	}
}

// TestLifecycleEvent_Validate_Completed_Happy 验证 completed event happy path（带 / 不带 result_payload 都允许）。
func TestLifecycleEvent_Validate_Completed_Happy(t *testing.T) {
	e := &LifecycleEvent{
		TaskID:    "task-1",
		Status:    LifecycleStatusCompleted,
		Timestamp: time.Now().UTC(),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("Validate completed happy (no payload): got err %v", err)
	}

	e.ResultPayload = json.RawMessage(`{"pr_url":"x"}`)
	if err := e.Validate(); err != nil {
		t.Errorf("Validate completed happy (with payload): got err %v", err)
	}

	e.StageMessage = "已完成"
	if err := e.Validate(); err != nil {
		t.Errorf("Validate completed happy (with stage_message): got err %v", err)
	}
}

// TestLifecycleParams_RoundTripJSON 验证 LifecycleParams 嵌套 marshal/unmarshal 可逆（JSON-RPC params 包装层）。
func TestLifecycleParams_RoundTripJSON(t *testing.T) {
	original := LifecycleParams{
		Event: LifecycleEvent{
			TaskID:       "task-1",
			Status:       LifecycleStatusProgress,
			Timestamp:    time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
			StageMessage: "正在 clone repo",
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// 验证 JSON 结构含 nested event key
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	if _, has := raw["event"]; !has {
		t.Errorf("expected nested 'event' key, got %s", string(data))
	}

	var got LifecycleParams
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Event.TaskID != original.Event.TaskID {
		t.Errorf("Event.TaskID mismatch")
	}
	if got.Event.Status != original.Event.Status {
		t.Errorf("Event.Status mismatch")
	}
	if got.Event.StageMessage != original.Event.StageMessage {
		t.Errorf("Event.StageMessage mismatch")
	}
}
