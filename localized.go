package kstypes

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// LocalizedString 是 i18n 字段的标准类型。YAML 既支持单 string 形态也支持 map 形态。
//   - 单 string：summary: "一句话摘要"   → LocalizedString{"": "一句话摘要"}
//   - i18n map：summary: { zh-CN: "...", en-US: "..." }
//
// JSON 序列化永远输出 map 形态，简化前端处理。
type LocalizedString map[string]string

// UnmarshalYAML 支持 scalar / map 双形态；空节点（null）反序列化为空 map。
func (l *LocalizedString) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*l = LocalizedString{}
		return nil
	}
	switch value.Kind {
	case 0:
		// 空文档 / null 节点
		*l = LocalizedString{}
		return nil
	case yaml.ScalarNode:
		// !!null tag 同样视为空 map
		if value.Tag == "!!null" {
			*l = LocalizedString{}
			return nil
		}
		*l = LocalizedString{"": value.Value}
		return nil
	case yaml.MappingNode:
		m := make(map[string]string)
		if err := value.Decode(&m); err != nil {
			return fmt.Errorf("LocalizedString: decode map: %w", err)
		}
		*l = LocalizedString(m)
		return nil
	default:
		return fmt.Errorf("LocalizedString: unsupported YAML kind %v", value.Kind)
	}
}

// MarshalJSON 永远输出 map 形态，避免前端 union 类型处理。
func (l LocalizedString) MarshalJSON() ([]byte, error) {
	if l == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(l))
}

// Get 按 locale fallback 链取值：
//
//	指定 locale → "zh-CN"（默认） → ""（单语言形态）→ 任意非空 → ""
func (l LocalizedString) Get(locale string) string {
	if v, ok := l[locale]; ok && v != "" {
		return v
	}
	if v, ok := l["zh-CN"]; ok && v != "" {
		return v
	}
	if v, ok := l[""]; ok && v != "" {
		return v
	}
	for _, v := range l {
		if v != "" {
			return v
		}
	}
	return ""
}

// LocalizedTags 是 tag 列表的 i18n 版本。
//   - 单形态：tags: [a, b]              → LocalizedTags{"": ["a", "b"]}
//   - i18n map：tags: { zh-CN: [...], en-US: [...] }
type LocalizedTags map[string][]string

func (t *LocalizedTags) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*t = LocalizedTags{}
		return nil
	}
	switch value.Kind {
	case 0:
		*t = LocalizedTags{}
		return nil
	case yaml.ScalarNode:
		if value.Tag == "!!null" {
			*t = LocalizedTags{}
			return nil
		}
		return fmt.Errorf("LocalizedTags: unexpected scalar value %q", value.Value)
	case yaml.SequenceNode:
		var slice []string
		if err := value.Decode(&slice); err != nil {
			return fmt.Errorf("LocalizedTags: decode sequence: %w", err)
		}
		*t = LocalizedTags{"": slice}
		return nil
	case yaml.MappingNode:
		m := make(map[string][]string)
		if err := value.Decode(&m); err != nil {
			return fmt.Errorf("LocalizedTags: decode map: %w", err)
		}
		*t = LocalizedTags(m)
		return nil
	default:
		return fmt.Errorf("LocalizedTags: unsupported YAML kind %v", value.Kind)
	}
}

func (t LocalizedTags) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string][]string(t))
}

// Get 按 locale fallback 链取值，与 LocalizedString.Get 同语义。
func (t LocalizedTags) Get(locale string) []string {
	if v, ok := t[locale]; ok && len(v) > 0 {
		return v
	}
	if v, ok := t["zh-CN"]; ok && len(v) > 0 {
		return v
	}
	if v, ok := t[""]; ok && len(v) > 0 {
		return v
	}
	for _, v := range t {
		if len(v) > 0 {
			return v
		}
	}
	return nil
}
