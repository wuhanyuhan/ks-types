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
type UINode struct {
	Type     string          `json:"type"`
	Props    json.RawMessage `json:"props,omitempty"`
	Children []UINode        `json:"children,omitempty"`
	Key      string          `json:"key,omitempty"`
}

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
