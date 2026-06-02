package kstypes

// ExpertActivity 是过程面里一个专家或编排官单 worker 的工作面快照。
type ExpertActivity struct {
	ExpertID    string         `json:"expert_id"`
	DisplayName string         `json:"display_name"`
	Role        string         `json:"role,omitempty"`
	Avatar      string         `json:"avatar,omitempty"`
	Phase       ActivityPhase  `json:"phase"`
	Status      ActivityStatus `json:"status"`
	Current     *ActivityStep  `json:"current,omitempty"`
	Steps       []ActivityStep `json:"steps,omitempty"`
	Drafts      []ArtifactRef  `json:"drafts,omitempty"`
}

// ActivityStep 是工作面时间线的一条记录。
type ActivityStep struct {
	Kind ActivityStepKind `json:"kind"`
	Text string           `json:"text"`
	Done bool             `json:"done"`
	TS   int64            `json:"ts"`
}

// ActivityStepKind 区分动作、理由、消息和等待四类过程信息。
type ActivityStepKind string

const (
	StepKindAction    ActivityStepKind = "action"
	StepKindReasoning ActivityStepKind = "reasoning"
	StepKindMessage   ActivityStepKind = "message"
	StepKindWait      ActivityStepKind = "wait"
)

// ActivityPhase 是专家工作的粗粒度业务阶段。
type ActivityPhase string

const (
	PhaseResearch   ActivityPhase = "research"
	PhaseDrafting   ActivityPhase = "drafting"
	PhaseReviewing  ActivityPhase = "reviewing"
	PhaseFinalizing ActivityPhase = "finalizing"
	PhaseDone       ActivityPhase = "done"
	PhaseBlocked    ActivityPhase = "blocked"
)

// ActivityStatus 是专家当前运行态。
type ActivityStatus string

const (
	ActivityStatusRunning ActivityStatus = "running"
	ActivityStatusDone    ActivityStatus = "done"
	ActivityStatusBlocked ActivityStatus = "blocked"
	ActivityStatusFailed  ActivityStatus = "failed"
)
