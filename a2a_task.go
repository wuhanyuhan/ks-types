package kstypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// A2ATaskState 表示 A2A 任务的生命周期状态，遵循 A2A v1.0 规范。
type A2ATaskState string

const (
	A2ATaskSubmitted     A2ATaskState = "submitted"
	A2ATaskWorking       A2ATaskState = "working"
	A2ATaskCompleted     A2ATaskState = "completed"
	A2ATaskFailed        A2ATaskState = "failed"
	A2ATaskCanceled      A2ATaskState = "canceled"
	A2ATaskInputRequired A2ATaskState = "input-required"
	A2ATaskRejected      A2ATaskState = "rejected"
	A2ATaskAuthRequired  A2ATaskState = "auth-required"
)

// a2aTaskTransitions 定义各状态允许迁移的目标状态集合（canceled 通过 CanTransitionTo 特殊处理）。
var a2aTaskTransitions = map[A2ATaskState]map[A2ATaskState]bool{
	// 注意：→ canceled 的转换由 CanTransitionTo 的特殊分支管控（非终态均可取消），此处不列。
	A2ATaskSubmitted:     {A2ATaskWorking: true, A2ATaskRejected: true, A2ATaskAuthRequired: true},
	A2ATaskWorking:       {A2ATaskCompleted: true, A2ATaskFailed: true, A2ATaskInputRequired: true, A2ATaskRejected: true},
	A2ATaskInputRequired: {A2ATaskWorking: true},
	A2ATaskAuthRequired:  {A2ATaskWorking: true},
}

// CanTransitionTo 返回当前状态是否允许迁移到 next 状态。
// canceled 是通用退出路径：任何非终态均可取消。
func (s A2ATaskState) CanTransitionTo(next A2ATaskState) bool {
	if next == A2ATaskCanceled {
		return s != A2ATaskCompleted && s != A2ATaskFailed && s != A2ATaskCanceled && s != A2ATaskRejected
	}
	m := a2aTaskTransitions[s]
	return m[next]
}

// IsTerminal 返回当前状态是否为终态（终态不可再迁移）。
func (s A2ATaskState) IsTerminal() bool {
	return s == A2ATaskCompleted || s == A2ATaskFailed || s == A2ATaskCanceled || s == A2ATaskRejected
}

// A2ATask 是 A2A 协议中的任务实体，贯穿任务全生命周期。
type A2ATask struct {
	ID        string `json:"id"`
	ContextID string `json:"contextId,omitempty"` // 占位：出站 task chain 关联用
	// SkillID 是首次 tasks/send 时 caller 指定的 skill；input-required 多轮续轮（tasks/messages）
	// 时 server 从 task 上下文恢复，避免 client 续轮时再传一次。
	SkillID                string                   `json:"skillId,omitempty"`
	State                  A2ATaskState             `json:"state"`
	StateTransitionHistory []A2ATaskStateTransition `json:"stateTransitionHistory,omitempty"`
	Messages               []A2AMessage             `json:"messages,omitempty"`
	Artifacts              []A2AArtifact            `json:"artifacts,omitempty"`
	// PushNotification 是 caller 在 tasks/send 时附带的 webhook 通知配置。
	// runner transition 到终态（Completed/Failed/InputRequired/Rejected）后由 framework 异步 POST 通知。
	PushNotification *PushNotificationConfig `json:"pushNotification,omitempty"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
}

// A2ATaskStateTransition 记录一次状态迁移事件，供审计与调试使用。
type A2ATaskStateTransition struct {
	From      A2ATaskState `json:"from"`
	To        A2ATaskState `json:"to"`
	Reason    string       `json:"reason,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// A2AMessage 表示对话中的一条消息，可包含多种类型的 Part。
type A2AMessage struct {
	Role  string           `json:"role"` // user / agent
	Parts []A2AMessagePart `json:"parts"`
}

// A2AMessagePart 是消息的基本组成单元，支持 text / file / data 三种类型。
type A2AMessagePart struct {
	Type string            `json:"type"` // text / file / data
	Text string            `json:"text,omitempty"`
	File *A2AFileReference `json:"file,omitempty"`
	// Data 用于结构化数据（type=data 时使用）。注意：JSON 反序列化会把数字统一变 float64，
	// 直接做 reflect.DeepEqual 时务必只用同源类型的值（如全部 string，或反序列化后再比较）。
	Data     map[string]any `json:"data,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// A2AFileReference 描述消息 Part 中引用的文件信息。
type A2AFileReference struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	URI      string `json:"uri,omitempty"`
	Bytes    string `json:"bytes,omitempty"` // base64
}

// A2AArtifact 是 Agent 执行任务后产出的结构化制品。
type A2AArtifact struct {
	Name     string           `json:"name"`
	Parts    []A2AMessagePart `json:"parts"`
	Metadata map[string]any   `json:"metadata,omitempty"`
}

// === A2A Task Lifecycle Protocol ===

// LifecycleStatus 是 a2a task lifecycle 事件状态枚举（method=tasks/lifecycle 的 status 字段）。
//
// 与 A2ATaskState（任务状态机字段值）分开命名：
//   - LifecycleStatus = 事件名（squad → keystone 反向 RPC 上报的 event type）
//   - A2ATaskState = 任务状态机字段值
//
// 事件 → keystone 端 user-facing 状态机映射：
//   - LifecycleStatusStarted   → state pending → running
//   - LifecycleStatusProgress  → state running 内（不切状态，仅更新 stage_message）
//   - LifecycleStatusCompleted → state running → done
//   - LifecycleStatusFailed    → state running → failed
type LifecycleStatus string

const (
	LifecycleStatusStarted   LifecycleStatus = "started"
	LifecycleStatusProgress  LifecycleStatus = "progress"
	LifecycleStatusCompleted LifecycleStatus = "completed"
	LifecycleStatusFailed    LifecycleStatus = "failed"
)

// LifecycleEvent 是 method=tasks/lifecycle 的 params.event 字段。
//
// 字段语义：
//   - TaskID         必填 — keystone 派发时分配的 UUID v4
//   - Status         必填 — 事件类型（4 enum）
//   - Timestamp      必填 — event 发生时间（squad 端时钟，UTC）
//   - StageMessage   可选 — running 阶段性文案（progress event 必填，其他 event 可选）
//   - FailureReason  条件必填 — failed event 必填，其他 event 应为空
//   - ResultPayload  条件可选 — completed event 可选（squad 自定义结果数据），其他 event 应为空
type LifecycleEvent struct {
	TaskID        string          `json:"task_id"`
	Status        LifecycleStatus `json:"status"`
	Timestamp     time.Time       `json:"timestamp"`
	StageMessage  string          `json:"stage_message,omitempty"`
	FailureReason string          `json:"failure_reason,omitempty"`
	ResultPayload json.RawMessage `json:"result_payload,omitempty"`
}

// Validate 校验 LifecycleEvent 字段集合规则（必填 / 条件必填）。
func (e *LifecycleEvent) Validate() error {
	if e.TaskID == "" {
		return errors.New("task_id 必填")
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp 必填")
	}
	switch e.Status {
	case LifecycleStatusStarted:
		if e.FailureReason != "" {
			return fmt.Errorf("status=%s 不应带 failure_reason", e.Status)
		}
		if len(e.ResultPayload) > 0 {
			return fmt.Errorf("status=%s 不应带 result_payload", e.Status)
		}
	case LifecycleStatusCompleted:
		if e.FailureReason != "" {
			return fmt.Errorf("status=%s 不应带 failure_reason", e.Status)
		}
	case LifecycleStatusProgress:
		if e.StageMessage == "" {
			return errors.New("status=progress 必须带 stage_message")
		}
		if e.FailureReason != "" || len(e.ResultPayload) > 0 {
			return errors.New("status=progress 不应带 failure_reason 或 result_payload")
		}
	case LifecycleStatusFailed:
		if e.FailureReason == "" {
			return errors.New("status=failed 必须带 failure_reason")
		}
	default:
		return fmt.Errorf("unknown lifecycle status: %s", e.Status)
	}
	return nil
}

// LifecycleParams 是 JSON-RPC method=tasks/lifecycle 的 params wrapper。
//
// JSON-RPC request body 形态：
//
//	{
//	  "jsonrpc": "2.0",
//	  "id": "...",
//	  "method": "tasks/lifecycle",
//	  "params": { "event": { "task_id": "...", "status": "started", "timestamp": "...", ... } }
//	}
type LifecycleParams struct {
	Event LifecycleEvent `json:"event"`
}
