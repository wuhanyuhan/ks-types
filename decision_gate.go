package kstypes

// GateMode 是决策门的三态语义。
type GateMode string

const (
	GateModeConfirm GateMode = "confirm"
	GateModeInput   GateMode = "input"
	GateModeChoice  GateMode = "choice"
)

// GateState 是决策门生命周期状态。
type GateState string

const (
	GateStatePending  GateState = "pending"
	GateStateAnswered GateState = "answered"
	GateStateExpired  GateState = "expired"
)

var gateTransitions = map[GateState]map[GateState]bool{
	GateStatePending: {GateStateAnswered: true, GateStateExpired: true},
}

// CanTransitionTo 返回当前状态是否允许迁移到 next。
func (s GateState) CanTransitionTo(next GateState) bool {
	return gateTransitions[s][next]
}

// IsTerminal 返回当前状态是否为终态。
func (s GateState) IsTerminal() bool {
	return s == GateStateAnswered || s == GateStateExpired
}

// DecisionGate 是交互面的统一表达：一次能力运行中需要用户拍板的阻塞点。
type DecisionGate struct {
	ID              string        `json:"id"`
	RunID           string        `json:"runId"`
	CanonicalName   string        `json:"canonicalName"`
	Mode            GateMode      `json:"mode"`
	Prompt          string        `json:"prompt"`
	Options         []Option      `json:"options,omitempty"`
	InputSchema     *InputSchema  `json:"inputSchema,omitempty"`
	Preview         *PreviewBlock `json:"preview,omitempty"`
	SideEffectLevel string        `json:"sideEffectLevel,omitempty"`
	State           GateState     `json:"state"`
	Answer          *GateAnswer   `json:"answer,omitempty"`
	ExpiresAt       int64         `json:"expiresAt,omitempty"`
	CreatedAt       int64         `json:"createdAt"`
	AnsweredAt      int64         `json:"answeredAt,omitempty"`
}

// IsExpiredAt 按 expires_at 判断 gate 在某时刻是否已过新鲜度窗口。
func (g *DecisionGate) IsExpiredAt(nowUnixMs int64) bool {
	return g.ExpiresAt > 0 && nowUnixMs > g.ExpiresAt
}

// Option 是 mode=choice 的一个候选。
type Option struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// InputSchema 是 mode=input 的补充输入契约。
type InputSchema struct {
	Fields    []InputField `json:"fields"`
	AllowSkip bool         `json:"allowSkip,omitempty"`
}

// InputField 是一个补充输入字段。
type InputField struct {
	Name          string   `json:"name"`
	Kind          string   `json:"kind"`
	Label         string   `json:"label"`
	Description   string   `json:"description,omitempty"`
	Required      bool     `json:"required"`
	Options       []string `json:"options,omitempty"`
	AllowFreeText bool     `json:"allowFreeText,omitempty"`
	Accept        []string `json:"accept,omitempty"`
	MinFiles      int      `json:"minFiles,omitempty"`
	MaxFiles      int      `json:"maxFiles,omitempty"`
}

// GateAnswer 是回流后的用户拍板内容。
type GateAnswer struct {
	ChosenOptionID string           `json:"chosenOptionId,omitempty"`
	FreeText       string           `json:"freeText,omitempty"`
	InputPayload   map[string]any   `json:"inputPayload,omitempty"`
	Confirmed      *bool            `json:"confirmed,omitempty"`
	Source         GateAnswerSource `json:"source,omitempty"`
}

// GateAnswerSource 标记回流入口。
type GateAnswerSource string

const (
	GateAnswerSourceWarroom  GateAnswerSource = "warroom"
	GateAnswerSourceChatText GateAnswerSource = "chat_text"
	GateAnswerSourceChatCard GateAnswerSource = "chat_card"
)
