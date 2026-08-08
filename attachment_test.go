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
		FileID:       "5df7f44caceb46e0a029a67a183899ee",
		ArtifactID:   "0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0",
	}
	raw, err := json.Marshal(in)
	require.NoError(t, err)

	var out KSAttachmentResolved
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, in, out)
}

// keystone AttachmentResolver 在签发 URL 之后把采集结果的 file_id / artifact_id
// **直接塞进同一个 resolved map**（其 internal/mcp/attachment_resolver.go 的 resolveInMap）。
// 本结构体是该 map 的 typed 镜像：少一个字段，消费方 DecodeInto 时就会静默丢掉它。
//
// 这不是假想的——ks-squad-framework 的 imagegen client 正因为解进旧结构体，
// 对每张生成图恒得空 FileID，于是文章封面只能存 7 天就过期的签发 URL。
func TestKSAttachmentResolved_ParsesKeystoneCaptureFields(t *testing.T) {
	t.Parallel()
	// 报文取自 keystone 真实出参形状：签发字段 + 采集回填字段同层。
	raw := `{
		"url": "http://host.docker.internal:5188/v1/managed-files/download?app=ks-mcp-minimax-media&exp=1786761969&path=1786157169_af340c94.png&token=936e2d6a",
		"expires_at": 1786761969,
		"mime_type": "image/jpeg",
		"original_name": "1786157169_af340c94.png",
		"size_bytes": 141414,
		"file_id": "5df7f44caceb46e0a029a67a183899ee",
		"artifact_id": "0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0"
	}`
	var got KSAttachmentResolved
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	assert.Equal(t, "5df7f44caceb46e0a029a67a183899ee", got.FileID,
		"file_id 是产物的持久句柄，丢了就只剩会过期的 URL")
	assert.Equal(t, "0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0", got.ArtifactID)
	assert.Equal(t, int64(1786761969), got.ExpiresAt)
}

// 采集是 best-effort（无 ArtifactInvokeContext / 发起人身份为 0 时整段跳过），
// 故消费方必须容忍缺字段，不能把「没采集到」当成解码失败。
func TestKSAttachmentResolved_CaptureFieldsOptional(t *testing.T) {
	t.Parallel()
	var got KSAttachmentResolved
	require.NoError(t, json.Unmarshal(
		[]byte(`{"url":"https://x/y","expires_at":1,"mime_type":"image/png","original_name":"y.png"}`), &got))
	assert.Empty(t, got.FileID)
	assert.Empty(t, got.ArtifactID)
	assert.Equal(t, "https://x/y", got.URL)
}

func TestKSAttachmentFieldName_Stable(t *testing.T) {
	t.Parallel()
	// 此常量被 keystone 与 MCP 双侧硬依赖，改动属于 BREAKING
	assert.Equal(t, "ks_attachments", KSAttachmentFieldName)
}
