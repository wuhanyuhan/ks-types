package kstypes

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLocalizedString_UnmarshalSingleString(t *testing.T) {
	var s LocalizedString
	if err := yaml.Unmarshal([]byte(`hello world`), &s); err != nil {
		t.Fatal(err)
	}
	if got := s.Get("zh-CN"); got != "hello world" {
		t.Errorf(`Get("zh-CN") = %q, want "hello world"`, got)
	}
	if got := s.Get(""); got != "hello world" {
		t.Errorf(`Get("") = %q, want "hello world"`, got)
	}
}

func TestLocalizedString_UnmarshalMap(t *testing.T) {
	var s LocalizedString
	raw := []byte("zh-CN: 中文摘要\nen-US: English summary\n")
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	if got := s.Get("zh-CN"); got != "中文摘要" {
		t.Errorf(`Get("zh-CN") = %q`, got)
	}
	if got := s.Get("en-US"); got != "English summary" {
		t.Errorf(`Get("en-US") = %q`, got)
	}
}

func TestLocalizedString_GetFallback(t *testing.T) {
	tests := []struct {
		name   string
		input  LocalizedString
		locale string
		want   string
	}{
		{"hit specified", LocalizedString{"zh-CN": "甲", "en-US": "A"}, "zh-CN", "甲"},
		{"fallback to zh-CN default", LocalizedString{"zh-CN": "甲", "fr-FR": "B"}, "ja-JP", "甲"},
		{"fallback to single-form", LocalizedString{"": "single"}, "ja-JP", "single"},
		{"fallback to any non-empty", LocalizedString{"fr-FR": "B"}, "ja-JP", "B"},
		{"all empty returns empty string", LocalizedString{}, "zh-CN", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Get(tt.locale); got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.locale, got, tt.want)
			}
		})
	}
}

func TestLocalizedString_JSONAlwaysMap(t *testing.T) {
	s := LocalizedString{"": "single"}
	buf, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != `{"":"single"}` {
		t.Errorf("Marshal = %s, want {\"\":\"single\"}", buf)
	}
}

func TestLocalizedString_UnmarshalEmpty(t *testing.T) {
	// 空 YAML / null 节点应反序列化为空 map，不应 panic
	var s LocalizedString
	if err := yaml.Unmarshal([]byte(""), &s); err != nil {
		t.Fatalf("unmarshal empty YAML: %v", err)
	}
	if got := s.Get("zh-CN"); got != "" {
		t.Errorf("empty Get = %q, want \"\"", got)
	}
}

func TestLocalizedTags_UnmarshalSingleSlice(t *testing.T) {
	var t1 LocalizedTags
	raw := []byte(`["tdd", "测试"]` + "\n")
	if err := yaml.Unmarshal(raw, &t1); err != nil {
		t.Fatal(err)
	}
	if got := t1.Get("zh-CN"); len(got) != 2 || got[0] != "tdd" {
		t.Errorf("Get = %v", got)
	}
}

func TestLocalizedTags_UnmarshalMap(t *testing.T) {
	var t1 LocalizedTags
	raw := []byte("zh-CN: [测试, 调试]\nen-US: [test, debug]\n")
	if err := yaml.Unmarshal(raw, &t1); err != nil {
		t.Fatal(err)
	}
	if got := t1.Get("en-US"); len(got) != 2 || got[1] != "debug" {
		t.Errorf("Get = %v", got)
	}
}

func TestLocalizedTags_JSONAlwaysMap(t *testing.T) {
	t1 := LocalizedTags{"": {"a", "b"}}
	buf, err := json.Marshal(t1)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != `{"":["a","b"]}` {
		t.Errorf("Marshal = %s", buf)
	}
}

func TestLocalizedTags_GetFallback(t *testing.T) {
	tests := []struct {
		name   string
		input  LocalizedTags
		locale string
		want   []string
	}{
		{"hit specified", LocalizedTags{"zh-CN": {"a"}, "en-US": {"x"}}, "en-US", []string{"x"}},
		{"fallback to zh-CN", LocalizedTags{"zh-CN": {"a"}, "fr-FR": {"x"}}, "ja-JP", []string{"a"}},
		{"fallback to empty key", LocalizedTags{"": {"a"}}, "ja-JP", []string{"a"}},
		{"fallback to any non-empty", LocalizedTags{"fr-FR": {"x"}}, "ja-JP", []string{"x"}},
		{"all empty returns nil", LocalizedTags{}, "zh-CN", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Get(tt.locale)
			if len(got) != len(tt.want) {
				t.Errorf("Get(%q) = %v, want %v", tt.locale, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Get(%q)[%d] = %q, want %q", tt.locale, i, got[i], tt.want[i])
				}
			}
		})
	}
}
