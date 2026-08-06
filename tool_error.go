package kstypes

import "fmt"

// ErrorCategory 是 MCP 应用工具错误的稳定分类，「错误类别」的单一事实源。
//
// 自 ks-internal-contracts 物理下沉至此（R4）：错误契约要被全生态 MCP 应用消费——squad 系
// 经 ks-squad-framework 写 `_meta.keystone.error`，普通 ks-mcp-* 经 ks-devkit SDK 白名单
// 结构化通道——性质已是公开协议，而 ks-types 是 ks-internal-contracts 与 ks-devkit 唯一
// 公共下层。ks-internal-contracts 保留类型别名 re-export，既有下游 import 路径不破坏。
//
// 流转链：app handler 返回 *ToolError → SDK/framework 写 wire 结构化字段 → keystone
// executor → dispatcher → WorkerResult → finalize 排障话术。禁止用字符串匹配 error 文案
// 推断类别（《LLM/自由文本边界纪律》）：类别只能由 errors.As(*ToolError) 或结构化字段携带。
//
// 与本仓 BizError（HTTP 业务码维度）正交：BizError 面向 keystone HTTP API 的业务响应码，
// ToolError 面向工具执行错误的跨服务 wire 分类，互不替代。
type ErrorCategory string

const (
	ErrorCategoryPermission  ErrorCategory = "permission"   // 无权限 / 需审批 / 被护栏拦
	ErrorCategoryNotFound    ErrorCategory = "not_found"    // 找不到目标资源
	ErrorCategoryValidation  ErrorCategory = "validation"   // 参数缺失 / 非法枚举 / schema 不过
	ErrorCategoryDependency  ErrorCategory = "dependency"   // 我方登记的下游 / 后端暂不可用
	ErrorCategoryTimeout     ErrorCategory = "timeout"      // 超时
	ErrorCategoryRateLimited ErrorCategory = "rate_limited" // 限流
	ErrorCategoryInternal    ErrorCategory = "internal"     // 未归类 / 内部错误（未知一律归此）
	// ErrorCategoryUpstream 是对端站点 / 第三方 API 故障（如抓取目标返回 521/503）。
	// 与 dependency 划清：dependency 指「我方登记的下游」（自家 gateway / 存储 / 兄弟服务），
	// upstream 指「任务对象本身在外部世界的对端」——它不该计入我方 5xx 告警，用户话术应说
	// 「目标站点不可达，建议换源」而非替我方道歉（R4 扩维，治缺陷 F 的类别缺维）。
	ErrorCategoryUpstream ErrorCategory = "upstream"
)

var knownErrorCategories = map[ErrorCategory]struct{}{
	ErrorCategoryPermission:  {},
	ErrorCategoryNotFound:    {},
	ErrorCategoryValidation:  {},
	ErrorCategoryDependency:  {},
	ErrorCategoryTimeout:     {},
	ErrorCategoryRateLimited: {},
	ErrorCategoryInternal:    {},
	ErrorCategoryUpstream:    {},
}

// IsKnownErrorCategory 报告 c 是否为已知类别。
func IsKnownErrorCategory(c ErrorCategory) bool {
	_, ok := knownErrorCategories[c]
	return ok
}

// NormalizeErrorCategory 把空 / 未知类别归一为 internal。
//
// 缺省不归 dependency——dependency 暗示已知下游宕机，对未知错误用它会复制旧误诊
// （历史 bug：squad 错误一律误判 dependency → 给「等几分钟」错误排障建议）。
// 同理缺省不归 upstream：对端故障必须由生产端显式声明（携带 UpstreamStatus 证据），
// 不许消费端对未知错误乐观推断「是对方的锅」。
func NormalizeErrorCategory(c ErrorCategory) ErrorCategory {
	if IsKnownErrorCategory(c) {
		return c
	}
	return ErrorCategoryInternal
}

// ToolError 是 MCP 应用工具返回的结构化错误：既是 wire payload（framework / SDK 写入
// 结构化字段，keystone normalizer 据此还原类别），又实现 error 供 errors.As 在生产 /
// 消费两侧取出 Category。只有本类型的结构化字段允许出 wire；非 typed error 的自由文本
// 正文仍只进服务端日志（脱敏边界，白名单通道而非放开透传）。
type ToolError struct {
	Category ErrorCategory `json:"category"`
	Message  string        `json:"message"`
	// Retryable 标记该错误是否值得重试；nil = 生产端未表态（可选，与类别正交——
	// upstream 可能可重试（源站抖动）也可能不可（源站封禁））。
	Retryable *bool `json:"retryable,omitempty"`
	// UpstreamStatus 承载对端原始 HTTP 状态码（如 521），供排障与话术引用；
	// 0 = 未设置（omitempty，不出 wire）。仅 upstream / dependency 语境有意义。
	UpstreamStatus int `json:"upstream_status,omitempty"`
}

// NewToolError 构造 *ToolError。
func NewToolError(category ErrorCategory, message string) *ToolError {
	return &ToolError{Category: category, Message: message}
}

// NewUpstreamError 构造对端站点 / 第三方 API 故障错误，携带对端原始状态码。
func NewUpstreamError(status int, message string) *ToolError {
	return &ToolError{Category: ErrorCategoryUpstream, Message: message, UpstreamStatus: status}
}

// Error 实现 error 接口。
func (e *ToolError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}
