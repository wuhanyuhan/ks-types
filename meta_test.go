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

// TestMetaResponse_V050FieldsRoundTrip 覆盖 v0.5.0 新增 5 字段（Nav / Permissions /
// ConfigMode / ProtocolVersion / ConfigStatus）的 JSON 序列化往返一致性。
func TestMetaResponse_V050FieldsRoundTrip(t *testing.T) {
	orig := MetaResponse{
		Name:     "ks-mcp-image-gen",
		Version:  "1.0.0",
		AuthMode: AuthModeKeystoneJWKS,
		Nav: &MetaNavDecl{
			Label:         "图片生成",
			Icon:          "image",
			Category:      "应用",
			Order:         10,
			OpenMode:      "fullpage",
			EntryPath:     "/gallery",
			RequiredPerms: []string{"mcp.image-gen.view", "mcp.image-gen.generate"},
		},
		Permissions: []MetaPermissionDecl{
			{
				Code:         "mcp.image-gen.view",
				Label:        "查看图片",
				DefaultRoles: []string{"admin"},
			},
			{
				Code:         "mcp.image-gen.generate",
				Label:        "生成图片",
				DefaultRoles: []string{"admin"},
			},
		},
		ConfigMode:      "iframe",
		ProtocolVersion: "1.0",
		ConfigStatus:    "via_frontend",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got MetaResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Nav 字段一致性
	if got.Nav == nil {
		t.Fatalf("Nav 解析后为 nil")
	}
	if got.Nav.Label != orig.Nav.Label || got.Nav.Icon != orig.Nav.Icon ||
		got.Nav.Category != orig.Nav.Category || got.Nav.Order != orig.Nav.Order ||
		got.Nav.OpenMode != orig.Nav.OpenMode || got.Nav.EntryPath != orig.Nav.EntryPath {
		t.Errorf("Nav 标量字段不一致: got=%+v want=%+v", got.Nav, orig.Nav)
	}
	if len(got.Nav.RequiredPerms) != len(orig.Nav.RequiredPerms) {
		t.Errorf("Nav.RequiredPerms 长度不一致: got=%d want=%d",
			len(got.Nav.RequiredPerms), len(orig.Nav.RequiredPerms))
	} else {
		for i, p := range orig.Nav.RequiredPerms {
			if got.Nav.RequiredPerms[i] != p {
				t.Errorf("Nav.RequiredPerms[%d]: got=%q want=%q", i, got.Nav.RequiredPerms[i], p)
			}
		}
	}

	// Permissions 字段一致性
	if len(got.Permissions) != len(orig.Permissions) {
		t.Fatalf("Permissions 长度不一致: got=%d want=%d", len(got.Permissions), len(orig.Permissions))
	}
	for i, p := range orig.Permissions {
		if got.Permissions[i].Code != p.Code || got.Permissions[i].Label != p.Label {
			t.Errorf("Permissions[%d] 不一致: got=%+v want=%+v", i, got.Permissions[i], p)
		}
		if len(got.Permissions[i].DefaultRoles) != len(p.DefaultRoles) {
			t.Errorf("Permissions[%d].DefaultRoles 长度不一致", i)
		}
	}

	// 三个字符串字段一致性
	if got.ConfigMode != orig.ConfigMode {
		t.Errorf("ConfigMode: got=%q want=%q", got.ConfigMode, orig.ConfigMode)
	}
	if got.ProtocolVersion != orig.ProtocolVersion {
		t.Errorf("ProtocolVersion: got=%q want=%q", got.ProtocolVersion, orig.ProtocolVersion)
	}
	if got.ConfigStatus != orig.ConfigStatus {
		t.Errorf("ConfigStatus: got=%q want=%q", got.ConfigStatus, orig.ConfigStatus)
	}

	// 验证 omitempty：未设置时序列化产物不应包含 v0.5.0 新字段键名
	empty := MetaResponse{Name: "empty", Version: "1.0.0"}
	emptyData, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal empty: %v", err)
	}
	emptyStr := string(emptyData)
	for _, key := range []string{"\"nav\"", "\"permissions\"", "\"config_mode\"", "\"protocol_version\"", "\"config_status\""} {
		if strings.Contains(emptyStr, key) {
			t.Errorf("v0.5.0 新字段 %s 应被 omitempty 省略，got: %s", key, emptyStr)
		}
	}
}
