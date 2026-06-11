package kstypes

import "encoding/json"

// MaxNestingDepth 是 SDUI 节点树允许的最大嵌套深度（fail-fast 上限，防恶意/失误深树拖垮渲染）。
const MaxNestingDepth = 32

// UINode 是 Server-Driven UI 的递归节点。
//
//	Type     查原语注册表（SDUIPrimitiveSchemas / 前端 PRIMITIVE_REGISTRY）
//	Props    该原语的 typed props（按 Type 对应的 SDUI<Name>Props 校验）
//	Children 容器原语的子节点（复合/叶子原语为空）
//	Key      React key / 稳定标识（可选）
//	Data     typed 实时数据源引用（可选）：节点声明它订阅哪个 typed 流，
//	         前端据 Kind 经协作接线解析为实际 URL / 订阅（非自由表达式绑定）
type UINode struct {
	Type     string          `json:"type"`
	Props    json.RawMessage `json:"props,omitempty"`
	Children []UINode        `json:"children,omitempty"`
	Key      string          `json:"key,omitempty"`
	Data     *UIDataSource   `json:"data,omitempty"`
}

// UIDataSource 是 typed 实时数据源引用（非自由表达式绑定）：节点声明「订阅哪个
// 封闭枚举的数据源 + 参数」，前端按 Kind 经协作接线（SDUIRenderContext.collab）
// 解析为可达 URL / 订阅。禁止把任意 query/表达式塞进来——Kind 是封闭枚举。
type UIDataSource struct {
	Kind   string            `json:"kind"`             // 封闭枚举（如 DataSourceTeamProgressStream）
	Params map[string]string `json:"params,omitempty"` // 该 Kind 的具名参数（如 {"run_id": "..."}）
}

// DataSourceTeamProgressStream 是首个数据源 Kind：订阅某个 run 的「团队实时进度流」
// （Params={run_id}）。前端经反代直连 squad 的 /stream SSE，reduceFrame 聚合成 TeamState。
const DataSourceTeamProgressStream = "team_progress_stream"

// Depth 返回以本节点为根的树深度（单节点为 1）。
func (n UINode) Depth() int {
	max := 0
	for _, c := range n.Children {
		if d := c.Depth(); d > max {
			max = d
		}
	}
	return max + 1
}
