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
