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
