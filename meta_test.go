package kstypes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMetaResponse_JSONRoundTrip(t *testing.T) {
	orig := MetaResponse{
		Name:     "ks-mcp-email",
		Version:  "1.0.0",
		AuthMode: AuthModeKeystoneJWKS,
		ConfigUI: &ConfigUIInfo{
			Enabled: true,
			URL:     "/config-ui/",
		},
		Tools: []ToolInfo{
			{Name: "send_email", Description: "发送邮件"},
			{Name: "list_emails", Description: "列出邮件"},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got MetaResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != orig.Name || got.Version != orig.Version {
		t.Errorf("name/version 不一致")
	}
	if got.AuthMode != orig.AuthMode {
		t.Errorf("auth_mode: got %q, want %q", got.AuthMode, orig.AuthMode)
	}
	if got.ConfigUI == nil || !got.ConfigUI.Enabled || got.ConfigUI.URL != "/config-ui/" {
		t.Errorf("config_ui 不一致: %+v", got.ConfigUI)
	}
	if len(got.Tools) != 2 || got.Tools[0].Name != "send_email" {
		t.Errorf("tools 不一致: %+v", got.Tools)
	}
}

func TestMetaResponse_ConfigUIOmittedWhenNil(t *testing.T) {
	m := MetaResponse{
		Name:    "simple-service",
		Version: "1.0.0",
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "config_ui") {
		t.Errorf("ConfigUI 为 nil 时应被 omitempty 省略，got: %s", s)
	}
}
