// Package sdui 提供 Go typed fluent builder：squad 端构造 UINode 树，编译期类型安全、不手搓 JSON。
// 与 UINode/原语 schema 同源（kstypes），是 SDUI 协议的"写"侧单一事实源。
package sdui

import (
	"encoding/json"

	kstypes "github.com/wuhanyuhan/ks-types"
)

const (
	StackVertical   = "vertical"
	StackHorizontal = "horizontal"
)

// mustProps 把任意 props 结构序列化为 json.RawMessage（builder 内部用，输入是本仓 typed 结构，不会失败）。
func mustProps(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		// builder 输入恒为本仓 typed 结构，Marshal 不应失败；失败即编程错误，退回空对象保持节点可渲染。
		return json.RawMessage(`{}`)
	}
	return b
}

// --- 容器 ---

func Stack(direction string, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveStack, Props: mustProps(kstypes.SDUIStackProps{Direction: direction}), Children: children}
}

func Grid(p kstypes.SDUIGridProps, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveGrid, Props: mustProps(p), Children: children}
}

func Card(p kstypes.SDUICardProps, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveCard, Props: mustProps(p), Children: children}
}

func Section(p kstypes.SDUISectionProps, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveSection, Props: mustProps(p), Children: children}
}

func Tabs(p kstypes.SDUITabsProps, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveTabs, Props: mustProps(p), Children: children}
}

func Split(p kstypes.SDUISplitProps, children ...kstypes.UINode) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveSplit, Props: mustProps(p), Children: children}
}

// --- 展示 ---

func Text(p kstypes.SDUITextProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveText, Props: mustProps(p)}
}

func Markdown(p kstypes.SDUIMarkdownProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveMarkdown, Props: mustProps(p)}
}

func FieldGroup(p kstypes.SDUIFieldGroupProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveFieldGroup, Props: mustProps(p)}
}

func Table(p kstypes.SDUITableProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveTable, Props: mustProps(p)}
}

func StatusBadge(p kstypes.SDUIStatusBadgeProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveStatusBadge, Props: mustProps(p)}
}

func Metric(p kstypes.SDUIMetricProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveMetric, Props: mustProps(p)}
}

func EmptyState(p kstypes.SDUIEmptyStateProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveEmptyState, Props: mustProps(p)}
}

// --- 交互 ---

func Button(p kstypes.SDUIButtonProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveButton, Props: mustProps(p)}
}

func Form(p kstypes.SDUIFormProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveForm, Props: mustProps(p)}
}

func Link(p kstypes.SDUILinkProps) kstypes.UINode {
	return kstypes.UINode{Type: kstypes.PrimitiveLink, Props: mustProps(p)}
}
