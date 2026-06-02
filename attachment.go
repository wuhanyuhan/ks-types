package kstypes

// KSAttachment 是 MCP 工具结果声明"附件需 keystone 签名出对外 URL"的约定结构。
// MCP 在工具结果 JSON 中以 KSAttachmentFieldName ("ks_attachments") 字段携带 []KSAttachment。
// Keystone 内部 mcp proxy 层会拦截识别 → 替换为 []KSAttachmentResolved。
type KSAttachment struct {
	Path         string `json:"path"`
	MimeType     string `json:"mime_type,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	TTLSeconds   int    `json:"ttl_seconds,omitempty"`
	Inline       bool   `json:"inline,omitempty"`
}

// KSAttachmentResolved 是 keystone 改写后交给上游（chat-ws / orchestrator / squad）的结构。
type KSAttachmentResolved struct {
	URL          string `json:"url"`
	ExpiresAt    int64  `json:"expires_at"`
	MimeType     string `json:"mime_type"`
	OriginalName string `json:"original_name"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
}

// KSAttachmentFieldName 是 MCP 工具结果中 envelope 字段名常量。
// keystone 与 MCP 服务双侧引用此常量，避免硬编码字符串不一致。
const KSAttachmentFieldName = "ks_attachments"
