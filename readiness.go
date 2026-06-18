package kstypes

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
