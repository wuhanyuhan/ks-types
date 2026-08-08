package kstypes

// KSAttachment 是 MCP 工具结果声明"附件需 keystone 签名出对外 URL"的约定结构。
// MCP 在工具结果 JSON 中以 KSAttachmentFieldName ("ks_attachments") 字段携带 []KSAttachment。
// Keystone 内部 mcp proxy 层会拦截识别 → 替换为 []KSAttachmentResolved。
type KSAttachment struct {
	Path string `json:"path"`
	// FileID 可选：产物已由 app 主动回存入库时，填写对应的 t_files.file_id。
	// Keystone 能力产物采集可在校验归属后直引已有文件，避免从托管卷二次搬运；
	// 校验失败时仍可回退 Path。该字段不改变附件 URL 签发的 Path 契约。
	FileID       string `json:"file_id,omitempty"`
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
	// FileID 是 keystone 能力产物采集后的持久文件句柄（t_files.file_id）。
	//
	// **它与 URL 的时效性完全不同，别混用**：URL 是按 KSAttachment.TTLSeconds 签发的
	// 短期地址（keystone 缺省仅 15m），过期即失效；FileID 指向 keystone 已把字节
	// 从 app 托管卷搬进 t_files 的那份副本，不随签发 URL 或 app 侧文件清理而消失。
	// 要把产物**落库长存**（文章封面、报告插图）的消费方应当存 FileID，需要字节时
	// 再经 query.file.download_url 现换一个新签名 URL；只存 URL 等于存了一个会烂掉的引用。
	//
	// 仅在产物采集成功时非空：采集要求调用链上有 ArtifactInvokeContext 且
	// 发起人身份非零（keystone AttachmentResolver best-effort，失败只记日志不阻断签发），
	// 故消费方必须按「可能为空」处理，不得假定它一定有值。
	FileID string `json:"file_id,omitempty"`
	// ArtifactID 是「我的空间」里对应产物行的 UUID；workspace 采集失败时为空。
	// 只用于跳转与展示，不作为读取字节的句柄——那是 FileID 的职责。
	ArtifactID string `json:"artifact_id,omitempty"`
}

// KSAttachmentFieldName 是 MCP 工具结果中 envelope 字段名常量。
// keystone 与 MCP 服务双侧引用此常量，避免硬编码字符串不一致。
const KSAttachmentFieldName = "ks_attachments"
