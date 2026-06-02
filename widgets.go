// Package kstypes widgets-protocol-v1：MCP tool widget binding 与 runtime UIResource 类型。
//
// 设计参考 docs/widgets-protocol-v1.md。
package kstypes

import "encoding/json"

// ToolUIBinding 是 squad 在 /meta.tools[]._meta.ui 里声明的 widget 绑定。
// mount 时由平台侧入库。
type ToolUIBinding struct {
	Widget       string   `json:"widget"`
	SandboxHints []string `json:"sandbox_hints,omitempty"`
}

// UIResource 是 keystone proxy normalize 后输出的 runtime 字段，
// 挂在 ToolCallResult.UIResource。前端 ToolUIRenderer 直接消费此结构。
type UIResource struct {
	Widget       string           `json:"widget"`
	Data         json.RawMessage  `json:"data"`
	SandboxHints []string         `json:"sandbox_hints,omitempty"`
	SrcURL       string           `json:"src_url,omitempty"`
	Error        *UIResourceError `json:"error,omitempty"`
}

// UIResourceError 描述 keystone 在 normalize 阶段对 widget 数据校验失败的诊断信息。
type UIResourceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MetaUIDecl 是 CallToolResult._meta.ui 的字段（squad runtime 响应里出现）。
type MetaUIDecl struct {
	Widget       string   `json:"widget,omitempty"`
	SandboxHints []string `json:"sandbox_hints,omitempty"`
}

// MetaKeystoneUIDecl 是 CallToolResult._meta.keystone.ui 的字段。
type MetaKeystoneUIDecl struct {
	Data json.RawMessage `json:"data"`
}

// CapabilitiesUI 是 squad /meta.capabilities.ui 字段。
type CapabilitiesUI struct {
	Enabled          bool     `json:"enabled"`
	RequestedSandbox []string `json:"requested_sandbox,omitempty"`
}
