package kstypes

import "fmt"

// NavState 是调用方从各自数据形态（keystone JSON map / Go NavDecl / Python dict）
// 归一出的 nav 声明状态，使矩阵函数与具体数据形态解耦。
type NavState int

const (
	NavAbsent  NavState = iota // 无 nav 块
	NavInvalid                 // 有 nav 块但字段不合法（缺 label/category/open_mode 或枚举错）
	NavValid                   // nav 合法，调用方保证 openMode ∈ {dialog,fullpage,tab}
)

// nav/config_mode/open_mode 枚举（与 meta.go 既有语义一致）。
const (
	navConfigModeNone   = "none"
	navConfigModeSchema = "schema"
	navConfigModeIframe = "iframe"
	navOpenModeDialog   = "dialog"
	navOpenModeFullpage = "fullpage"
	navOpenModeTab      = "tab"
)

// CheckNavConfigConsistency 是 nav / config_mode / config_ui 组合能否产出"可用入口"的单一事实源。
// 把 app-entry.ts（前端分类）+ keystone buildConfigUIAccess（后端反代）的合法矩阵收口到一处，
// 供 keystone 摄入诊断、Go SDK / Python SDK 启动期 fail-fast 三方复用。
//
// configMode=="" 内部归一为 "none"（对齐 keystone buildNavRegistryRow 落库语义）。
// navState=NavValid 时调用方保证 openMode ∈ {dialog,fullpage,tab}。
// 返回 (reason, ok)：ok=false 时 reason 是人话诊断。
func CheckNavConfigConsistency(navState NavState, openMode, configMode string, hasConfigUI bool) (reason string, ok bool) {
	if configMode == "" {
		configMode = navConfigModeNone
	}

	switch navState {
	case NavAbsent:
		if configMode == navConfigModeSchema || configMode == navConfigModeIframe {
			return fmt.Sprintf("声明了 config_mode=%s 却未声明 nav 导航入口：配置界面将无法打开（列表显示「无入口」，强制访问报 40041）。请补 nav（schema 配置类用 open_mode=dialog）", configMode), false
		}
		return "", true
	case NavInvalid:
		return "nav 声明不合法（缺 label/category/open_mode 或 open_mode 非 dialog/fullpage/tab），nav 行会被丢弃 → 应用「无入口」", false
	}

	// NavValid
	switch openMode {
	case navOpenModeDialog:
		switch configMode {
		case navConfigModeSchema:
			return "", true
		case navConfigModeIframe:
			if hasConfigUI {
				return "", true
			}
			return "open_mode=dialog + config_mode=iframe 需要 config_ui.enabled=true 且 url 非空，当前缺失 → 点击配置会报 40041", false
		default:
			return fmt.Sprintf("open_mode=dialog + config_mode=%s 无效：dialog 入口只支持 schema/iframe 配置弹窗 → 应用「无入口」", configMode), false
		}
	case navOpenModeFullpage, navOpenModeTab:
		switch configMode {
		case navConfigModeNone:
			return "", true
		case navConfigModeSchema:
			return fmt.Sprintf("open_mode=%s + config_mode=schema 非法：schema 配置只能在 dialog 弹窗内渲染，此 nav 会被菜单丢弃 → 应用「无入口」。配置类请改 open_mode=dialog", openMode), false
		case navConfigModeIframe:
			return fmt.Sprintf("open_mode=%s + config_mode=iframe 非法：点击会报 40041。fullpage/tab 业务前端应 config_mode=none；配置界面应 open_mode=dialog", openMode), false
		default:
			return fmt.Sprintf("open_mode=%s + config_mode=%s 组合未知", openMode, configMode), false
		}
	}
	return "", true // 不可达：NavValid 保证 openMode ∈ 三枚举
}
