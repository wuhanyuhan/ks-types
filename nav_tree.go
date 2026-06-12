package kstypes

import "fmt"

// console 二级导航项内容种类（封闭枚举，禁自由文本路由）。
const (
	NavKindSDUI   = "sdui"   // 内容 = keystone 原生渲染的 UINode 视图（经 /api/console/views/:key 取）
	NavKindIsland = "island" // 内容 = squad 自服务专有面板，经反代以 slot 嵌入
)

// NavItem 是 console 二级导航树的一个节点（typed 封闭结构，非自由文本）。
type NavItem struct {
	Key      string    `json:"key"`            // 稳定标识 + 路由段
	Label    string    `json:"label"`          // 侧栏显示名
	Icon     string    `json:"icon,omitempty"` // lucide-react 图标名
	Kind     string    `json:"kind"`           // sdui | island
	Path     string    `json:"path,omitempty"` // 仅 island：自服务面板反代子路径；sdui 视图按 key 取
	Children []NavItem `json:"children,omitempty"`
}

func (n NavItem) Validate() error {
	if n.Key == "" {
		return fmt.Errorf("nav item key empty")
	}
	if n.Label == "" {
		return fmt.Errorf("nav item %q label empty", n.Key)
	}
	switch n.Kind {
	case NavKindSDUI:
	case NavKindIsland:
		if n.Path == "" {
			return fmt.Errorf("nav item %q kind=island requires path", n.Key)
		}
	default:
		return fmt.Errorf("nav item %q kind invalid: %q", n.Key, n.Kind)
	}
	for _, c := range n.Children {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// NavTree 是 console 二级导航树：squad 声明，keystone 据此渲染 console-shell 侧栏 + 路由分发。
type NavTree struct {
	Items []NavItem `json:"items"`
}

func (t NavTree) Validate() error {
	if len(t.Items) == 0 {
		return fmt.Errorf("nav tree items empty")
	}
	seen := map[string]bool{}
	for _, it := range t.Items {
		if seen[it.Key] {
			return fmt.Errorf("nav tree duplicate key: %q", it.Key)
		}
		seen[it.Key] = true
		if err := it.Validate(); err != nil {
			return err
		}
	}
	return nil
}
