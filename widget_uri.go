package kstypes

import (
	"fmt"
	"regexp"
	"strings"
)

// WidgetURIScheme 是 widget URI 支持的两种 scheme：ks（共享 widget）/ ui（自定义 widget）。
type WidgetURIScheme string

const (
	// WidgetSchemeKS 表示共享 widget URI（ks://widgets/{name}@{vN}）。
	WidgetSchemeKS WidgetURIScheme = "ks"
	// WidgetSchemeCustom 表示自定义 widget URI（ui://{squad-id}/{path}）。
	WidgetSchemeCustom WidgetURIScheme = "ui"
)

// WidgetURI 是 widget URI 解析后的结构。
//
//	ks 形态: scheme=ks, name=<widget-name>, version=<vN>
//	ui 形态: scheme=ui, name=<squad-id>/<path>, version=""
type WidgetURI struct {
	Scheme  WidgetURIScheme
	Name    string
	Version string
	Raw     string
}

var ksWidgetURIRe = regexp.MustCompile(`^ks://widgets/([a-z][a-z0-9-]*)@(v\d+)$`)

// ParseWidgetURI 解析 widget URI。
//
//	合法:
//	  ks://widgets/{name}@{version}    name 小写连字符；version v\d+
//	  ui://{squad-id}/{path}           squad-id 与 path 非空；path 可含 /
//	非法:
//	  query / fragment / 空 / 未知 scheme / ks 不带 version / ks 不在 widgets 命名空间
func ParseWidgetURI(s string) (*WidgetURI, error) {
	if s == "" {
		return nil, fmt.Errorf("widget URI is empty")
	}
	if strings.ContainsAny(s, "?#") {
		return nil, fmt.Errorf("widget URI must not contain query/fragment: %q", s)
	}
	switch {
	case strings.HasPrefix(s, "ks://"):
		m := ksWidgetURIRe.FindStringSubmatch(s)
		if m == nil {
			return nil, fmt.Errorf("invalid ks widget URI: %q (expect ks://widgets/{name}@{vN})", s)
		}
		return &WidgetURI{Scheme: WidgetSchemeKS, Name: m[1], Version: m[2], Raw: s}, nil
	case strings.HasPrefix(s, "ui://"):
		rest := strings.TrimPrefix(s, "ui://")
		if rest == "" || !strings.Contains(rest, "/") {
			return nil, fmt.Errorf("invalid ui widget URI: %q (expect ui://{squad}/{path})", s)
		}
		squad, path, _ := strings.Cut(rest, "/")
		if squad == "" || path == "" {
			return nil, fmt.Errorf("invalid ui widget URI: %q (squad/path empty)", s)
		}
		return &WidgetURI{Scheme: WidgetSchemeCustom, Name: rest, Raw: s}, nil
	default:
		return nil, fmt.Errorf("unsupported widget URI scheme: %q", s)
	}
}

// SquadID 返回自定义 widget 的 squad-id 段（仅 ui:// scheme 有效）。
func (u *WidgetURI) SquadID() string {
	if u.Scheme != WidgetSchemeCustom {
		return ""
	}
	squad, _, _ := strings.Cut(u.Name, "/")
	return squad
}

// Path 返回自定义 widget 的 path 段（仅 ui:// scheme 有效）。
func (u *WidgetURI) Path() string {
	if u.Scheme != WidgetSchemeCustom {
		return ""
	}
	_, path, _ := strings.Cut(u.Name, "/")
	return path
}
