package kstypes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKSAttachment_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	in := KSAttachment{
		Path:         "output/abc.png",
		FileID:       "3f2a77c1d9e64b0f8c5a12de34ab56cd",
		MimeType:     "image/png",
		OriginalName: "abc.png",
		TTLSeconds:   900,
		Inline:       true,
	}
	raw, err := json.Marshal(in)
	require.NoError(t, err)

	var out KSAttachment
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, in, out)
}

func TestKSAttachment_FileID_BackwardCompatible(t *testing.T) {
	t.Parallel()
	var got KSAttachment
	require.NoError(t, json.Unmarshal([]byte(`{"path":"out.mp3","mime_type":"audio/mpeg"}`), &got))
	assert.Empty(t, got.FileID)
	assert.Equal(t, "out.mp3", got.Path)
}

func TestKSAttachment_OmitEmpty(t *testing.T) {
	t.Parallel()
	in := KSAttachment{Path: "output/abc.png"}
	raw, err := json.Marshal(in)
	require.NoError(t, err)
	// 仅 path 字段必输出；其余空值省略
	assert.JSONEq(t, `{"path":"output/abc.png"}`, string(raw))
}

func TestKSAttachmentResolved_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	in := KSAttachmentResolved{
		URL:          "https://example.com/v1/managed-files/download?app=x&path=y&token=z&exp=1",
		ExpiresAt:    1747654321,
		MimeType:     "image/png",
		OriginalName: "abc.png",
		SizeBytes:    1024,
	}
	raw, err := json.Marshal(in)
	require.NoError(t, err)

	var out KSAttachmentResolved
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, in, out)
}

func TestKSAttachmentFieldName_Stable(t *testing.T) {
	t.Parallel()
	// 此常量被 keystone 与 MCP 双侧硬依赖，改动属于 BREAKING
	assert.Equal(t, "ks_attachments", KSAttachmentFieldName)
}
