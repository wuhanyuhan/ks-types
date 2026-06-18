package kstypes

import (
	"fmt"
	"regexp"
)

// readinessGateIDRegex 是就绪门 id 的命名格式（封闭格式，同 app 内引用用）。
var readinessGateIDRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

// ReadinessGateKind 是就绪门的种类。
type ReadinessGateKind string

const (
	// ReadinessGateKindConfig 表示该门靠用户配置（key/secret）满足，由 keystone 侧按事实判定。
	ReadinessGateKindConfig ReadinessGateKind = "config"
	// ReadinessGateKindInitTask 表示该门靠应用执行一次性初始化任务满足（如向量预置），
	// 由 keystone 触发、应用执行、keystone 轮询观测。
	ReadinessGateKindInitTask ReadinessGateKind = "init_task"
)

// IsValid 返回该 kind 是否为受支持的封闭枚举值。
func (k ReadinessGateKind) IsValid() bool {
	switch k {
	case ReadinessGateKindConfig, ReadinessGateKindInitTask:
		return true
	}
	return false
}

// ReadinessSpec 声明应用"可用"的前置条件集合（就绪门）。
// 未声明任何门的应用 = 天然就绪（向后兼容）。
type ReadinessSpec struct {
	Gates []ReadinessGate `yaml:"gates,omitempty" json:"gates,omitempty"`
}

// ReadinessGate 是单个就绪门：要么靠配置满足（kind=config），
// 要么靠一次性初始化任务满足（kind=init_task）。
type ReadinessGate struct {
	ID          string            `yaml:"id" json:"id"`
	Kind        ReadinessGateKind `yaml:"kind" json:"kind"`
	Title       string            `yaml:"title,omitempty" json:"title,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`

	// Blocking 表示该门未满足时是否门控该 app 的能力（默认 true）。
	// 指针区分"未设置(=默认 true)"与"显式 false(=advisory，仅提示不拦截)"。
	Blocking *bool `yaml:"blocking,omitempty" json:"blocking,omitempty"`

	// 以下 kind=config 专用：满足该门所需的配置/secret 字段名，
	// 引用应用已声明的配置字段（不在此重复定义字段 schema）。
	RequiresConfig  []string `yaml:"requires_config,omitempty" json:"requires_config,omitempty"`
	RequiresSecrets []string `yaml:"requires_secrets,omitempty" json:"requires_secrets,omitempty"`

	// 以下 kind=init_task 专用。
	// Idempotent 表示该任务可安全重跑（重试 / 重新初始化）。
	Idempotent bool `yaml:"idempotent,omitempty" json:"idempotent,omitempty"`
	// AutoInit 表示装完是否自动触发（默认 true）。显式 false 则需管理员手动触发。
	AutoInit *bool `yaml:"auto_init,omitempty" json:"auto_init,omitempty"`
	// TimeoutSeconds 是单次初始化的超时（秒，0 表示用平台默认）。
	TimeoutSeconds int `yaml:"timeout_seconds,omitempty" json:"timeout_seconds,omitempty"`
}

// IsBlocking 返回该门是否门控能力（默认 true）。
func (g ReadinessGate) IsBlocking() bool {
	return g.Blocking == nil || *g.Blocking
}

// IsAutoInit 返回 init_task 门是否装完自动触发（默认 true）。
func (g ReadinessGate) IsAutoInit() bool {
	return g.AutoInit == nil || *g.AutoInit
}

// Validate 校验就绪声明：门 id 必填且合法、同 app 内唯一、kind 合法，
// 并按 kind 校验各自专用字段。空门集合合法（天然就绪）。
func (r ReadinessSpec) Validate() error {
	seen := make(map[string]struct{}, len(r.Gates))
	for i, g := range r.Gates {
		if g.ID == "" {
			return fmt.Errorf("readiness.gates[%d].id 为必填项", i)
		}
		if !readinessGateIDRegex.MatchString(g.ID) {
			return fmt.Errorf("readiness.gates[%d].id %q 非法（要求 ^[a-z][a-z0-9_-]{0,63}$）", i, g.ID)
		}
		if _, dup := seen[g.ID]; dup {
			return fmt.Errorf("readiness.gates[%d].id %q 重复", i, g.ID)
		}
		seen[g.ID] = struct{}{}

		if !g.Kind.IsValid() {
			return fmt.Errorf("readiness.gates[%d].kind %q 不合法（允许 config / init_task）", i, g.Kind)
		}

		switch g.Kind {
		case ReadinessGateKindConfig:
			if len(g.RequiresConfig) == 0 && len(g.RequiresSecrets) == 0 {
				return fmt.Errorf("readiness.gates[%d] (kind=config) 必须声明 requires_config 或 requires_secrets 至少一项", i)
			}
		case ReadinessGateKindInitTask:
			if g.TimeoutSeconds < 0 {
				return fmt.Errorf("readiness.gates[%d].timeout_seconds 不能为负", i)
			}
		}
	}
	return nil
}

// ── 就绪端点运行时 wire 契约（GET /ks-readiness、POST /ks-readiness/init）──
// 与 ReadinessSpec（manifest 声明）同仓同文件：声明描述「门是什么」，下列类型描述「门此刻什么状态」。
// 由应用经 ks-devkit SDK 上报、keystone 后端轮询消费——单一事实源的跨仓 wire 契约。

// ReadinessGateStatus 是单个 init_task 就绪门的运行时状态（app 上报 + keystone 持久化/聚合共用）。
type ReadinessGateStatus string

const (
	// ReadinessGateStatusPending 已声明/已 seed，尚未触发初始化。
	ReadinessGateStatusPending ReadinessGateStatus = "pending"
	// ReadinessGateStatusRunning 初始化执行中（progress 反映进度）。
	ReadinessGateStatusRunning ReadinessGateStatus = "running"
	// ReadinessGateStatusReady 已满足。
	ReadinessGateStatusReady ReadinessGateStatus = "ready"
	// ReadinessGateStatusFailed 初始化失败（message 给原因，可重试 / 重跑）。
	ReadinessGateStatusFailed ReadinessGateStatus = "failed"
)

// IsValid 返回该状态是否为受支持的封闭枚举值。
func (s ReadinessGateStatus) IsValid() bool {
	switch s {
	case ReadinessGateStatusPending, ReadinessGateStatusRunning,
		ReadinessGateStatusReady, ReadinessGateStatusFailed:
		return true
	}
	return false
}

// ReadinessGateState 是单个 init_task 门的运行时状态，应用经 GET /ks-readiness 上报。
type ReadinessGateState struct {
	// ID 对应 manifest readiness.gates[].id（kind=init_task）。
	ID string `json:"id"`
	// Status 当前运行时状态。
	Status ReadinessGateStatus `json:"status"`
	// Progress 初始化进度 0-100；未开始 / 无进度时省略。
	Progress *int `json:"progress,omitempty"`
	// Message 人话状态 / 错误原因；可空时省略。
	Message string `json:"message,omitempty"`
}

// ReadinessReport 是 GET /ks-readiness 的响应：本应用全部 init_task 门的运行时状态。
type ReadinessReport struct {
	// Gates 是本应用全部 init_task 门的运行时状态列表。
	Gates []ReadinessGateState `json:"gates"`
}

// ReadinessInitRequest 是 POST /ks-readiness/init 的请求体：触发 / 重触发某 init_task 门。
type ReadinessInitRequest struct {
	// GateID 要触发的 init_task 门 id（对应 manifest readiness.gates[].id）。
	GateID string `json:"gate_id"`
}
