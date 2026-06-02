package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUIResource_JSON_RoundTrip_SharedWidget(t *testing.T) {
	t.Parallel()
	in := UIResource{
		Widget: "ks://widgets/diff-review@v1",
		Data:   json.RawMessage(`{"title":"x","diff":[{"type":"context","text":"a"}],"actions":[]}`),
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out UIResource
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Widget != in.Widget {
		t.Errorf("widget: got %q, want %q", out.Widget, in.Widget)
	}
	if string(out.Data) != string(in.Data) {
		t.Errorf("data drift: got %s, want %s", string(out.Data), string(in.Data))
	}
	if out.SrcURL != "" || len(out.SandboxHints) != 0 || out.Error != nil {
		t.Errorf("unexpected non-zero fields: %+v", out)
	}
}

func TestUIResource_JSON_RoundTrip_CustomWidget(t *testing.T) {
	t.Parallel()
	in := UIResource{
		Widget:       "ui://marketing/brand-editor",
		Data:         json.RawMessage(`{"x":1}`),
		SandboxHints: []string{"allow-downloads"},
		SrcURL:       "/v1/admin/mcp-servers/ui/sess-x/brand-editor",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out UIResource
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Widget != in.Widget || out.SrcURL != in.SrcURL {
		t.Errorf("custom widget fields drift: %+v vs %+v", out, in)
	}
	if len(out.SandboxHints) != 1 || out.SandboxHints[0] != "allow-downloads" {
		t.Errorf("sandbox_hints drift: %+v", out.SandboxHints)
	}
}

func TestUIResource_Error_Field(t *testing.T) {
	t.Parallel()
	in := UIResource{
		Widget: "ks://widgets/diff-review@v1",
		Error:  &UIResourceError{Code: "schema_mismatch", Message: "missing field"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"error":{"code":"schema_mismatch"`) {
		t.Errorf("error field not in JSON: %s", string(b))
	}
}

func TestToolUIBinding_OmitEmptySandboxHints(t *testing.T) {
	t.Parallel()
	in := ToolUIBinding{Widget: "ks://widgets/list-actions@v1"}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "sandbox_hints") {
		t.Errorf("expected omitempty, got: %s", string(b))
	}
}

func TestPostMessageMethodNames_Stable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		constant string
		wire     string
	}{
		{PMMethodAppReady, "app.ready"},
		{PMMethodAppData, "app.data"},
		{PMMethodAppResize, "app.resize"},
		{PMMethodAppCallServerTool, "app.callServerTool"},
		{PMMethodAppToolResult, "app.toolResult"},
		{PMMethodAppToolError, "app.toolError"},
		{PMMethodAppUpdateModelContext, "app.updateModelContext"},
		{PMMethodAppClose, "app.close"},
		{PMMethodAppNotify, "app.notify"},
		{PMMethodAppOpenLink, "app.openLink"},
	}
	for _, c := range cases {
		if c.constant != c.wire {
			t.Errorf("post-message wire drift: constant=%q wire=%q", c.constant, c.wire)
		}
	}
}

func TestSandboxFlagWhitelist(t *testing.T) {
	t.Parallel()
	if SandboxFlagAllowDownloads != "allow-downloads" {
		t.Errorf("SandboxFlagAllowDownloads drift: %q", SandboxFlagAllowDownloads)
	}
	if SandboxFlagAllowPopups != "allow-popups" {
		t.Errorf("SandboxFlagAllowPopups drift: %q", SandboxFlagAllowPopups)
	}
	if SandboxFlagAllowModals != "allow-modals" {
		t.Errorf("SandboxFlagAllowModals drift: %q", SandboxFlagAllowModals)
	}
	if SandboxFlagAllowSameOrigin != "allow-same-origin" {
		t.Errorf("SandboxFlagAllowSameOrigin drift: %q", SandboxFlagAllowSameOrigin)
	}
}
