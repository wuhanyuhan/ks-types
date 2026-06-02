package kstypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCanonical_Derive 锁定去前缀派生 helper：<app_id>.<bare_name>。
func TestCanonical_Derive(t *testing.T) {
	assert.Equal(t, "ks-mcp-browser.web_search", Canonical("ks-mcp-browser", "web_search"))
}

// TestProvidesSpec_BareName 锁定 app_provided 校验分流：作者写裸名 name 通过；
// 写全名 canonical_name 被拒；裸名带点（多段）被拒。
func TestProvidesSpec_BareName(t *testing.T) {
	base := func(c CapabilitySpec) ProvidesSpec { return ProvidesSpec{Capabilities: []CapabilitySpec{c}} }
	good := CapabilitySpec{
		Name:          "web_search",
		ExecutionMode: "sync",
		Backend:       BackendSpec{Kind: "mcp_tool", ToolName: "web_search"},
	}
	require.NoError(t, base(good).Validate("ks-mcp-browser", RuntimeModeContainer))

	// 作者写全名 canonical_name → 被拒（去前缀后不允许）
	full := good
	full.Name = ""
	full.CanonicalName = "ks-mcp-browser.web_search"
	assert.ErrorContains(t, base(full).Validate("ks-mcp-browser", RuntimeModeContainer), "name")

	// 裸名带点（多段）→ 被拒（要求单段无点）
	dotted := good
	dotted.Name = "browser.search"
	assert.ErrorContains(t, base(dotted).Validate("ks-mcp-browser", RuntimeModeContainer), "name")
}
