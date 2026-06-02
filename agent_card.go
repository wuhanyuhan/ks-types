package kstypes

// AgentCard 描述一个 A2A Agent 的对外身份与能力，遵循 A2A v1.0 规范。
type AgentCard struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	ProtocolVersion string            `json:"protocolVersion"`
	URL             string            `json:"url"`
	Provider        AgentProvider     `json:"provider"`
	Capabilities    AgentCapabilities `json:"capabilities"`
	// SecuritySchemes 是 A2A AgentCard 的安全方案集合。
	// 故意不加 omitempty：A2A v1.0 client 可能依赖该字段的存在做协议合规校验，
	// 因此即使为空也输出 "securitySchemes":{}。这是相对 "用 omitempty 向前兼容" 通用建议
	// 的有意偏离。
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
	Skills          []A2ASkill                `json:"skills"`
}

// AgentProvider 描述 Agent 所属的组织或团队。
type AgentProvider struct {
	Organization string `json:"organization"`
	URL          string `json:"url,omitempty"`
}

// AgentCapabilities 声明 Agent 支持的协议能力。
type AgentCapabilities struct {
	Streaming         bool `json:"streaming"`
	PushNotifications bool `json:"pushNotifications"`
}

// HeaderA2ACallChain 是 A2A 调用链 header 名（用于死锁防护）。
// 调用方在出站调用 keystone /a2a 时把当前调用栈作为 JSON-encoded []CallChainEntry 写入此 header；
// keystone 入站 middleware 解析检测元组重复 + 长度阈值，识别死锁/循环调用。
const HeaderA2ACallChain = "X-A2A-Call-Chain"

// CallChainEntry 是调用链中的单跳元素（按元组判定环路）。
// HeaderA2ACallChain 的 value 是 JSON-encoded []CallChainEntry，按时间顺序追加。
//
// 元组判定：单独 AgentID 不足以判环（同一 agent 可能并行处理多个 task），
// 必须用 (AgentID, TaskID) 元组判定循环（同一 agent 处理不同 task 是合法并行）。
type CallChainEntry struct {
	AgentID string `json:"agent_id"`
	TaskID  string `json:"task_id"`
	Ts      int64  `json:"ts"` // 调用方记录的时间戳（毫秒）
}

// MaxA2ACallChainLength 是 A2A 调用链长度阈值（协议级运行时常量）。
// 入站 middleware 收到 chain 长度 ≥ 此值的请求直接 reject（HTTP 508 Loop Detected）；
// 阈值按业务流实测调整（多 agent 串联 5-6 跳已属深度交互场景）。
const MaxA2ACallChainLength = 8

// PushNotificationConfig 描述任务终态后由 server 端 push 到 caller 的 webhook 配置（A2A v1.0 push notification 协议）。
// 用法：caller 在 tasks/send 时把 PushNotificationConfig 传入 task；server runner transition 到终态后
// （Completed / Failed / InputRequired / Rejected）按此配置 POST 通知到 URL。
type PushNotificationConfig struct {
	URL            string                      `json:"url"`
	Token          string                      `json:"token,omitempty"` // Bearer 鉴权 token，server POST 时注入 Authorization header
	Authentication *PushNotificationAuthDetail `json:"authentication,omitempty"`
}

// PushNotificationAuthDetail 是 PushNotificationConfig 的扩展鉴权配置（A2A v1.0 spec 鉴权方案描述）。
// Schemes 列出可用鉴权方案（Bearer / Basic / OAuth2 等），Credentials 是与 schemes[0] 配套的凭据载荷。
type PushNotificationAuthDetail struct {
	Schemes     []string `json:"schemes"`
	Credentials string   `json:"credentials,omitempty"`
}
