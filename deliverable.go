package kstypes

import "encoding/json"

// DeliverableType 是交付物语义类型，驱动前端卡片的图标、预览样式与默认动作。
type DeliverableType string

const (
	DeliverableArticle     DeliverableType = "article"     // 文章、推文、文案等长文本创作
	DeliverableReport      DeliverableType = "report"      // 结构化报告
	DeliverableSpreadsheet DeliverableType = "spreadsheet" // 表格数据
	DeliverableCode        DeliverableType = "code"        // 代码、配置、脚本
	DeliverableImageSet    DeliverableType = "image_set"   // 一组图片
	DeliverableDocument    DeliverableType = "document"    // 文件型产物
	DeliverableDataset     DeliverableType = "dataset"     // 数据集或批量结构化数据
	DeliverableComposite   DeliverableType = "composite"   // 多个异质子产物聚合
)

// DeliverableState 是交付物生命周期。
type DeliverableState string

const (
	DeliverableStateDraft      DeliverableState = "draft"      // 内部过渡态，不回写主对话
	DeliverableStateFinalized  DeliverableState = "finalized"  // 终态，回写主对话使用
	DeliverableStateSuperseded DeliverableState = "superseded" // 预留给后续版本链
)

// PreviewKind 决定 PreviewBlock 哪个字段生效。
type PreviewKind string

const (
	PreviewKindMarkdown   PreviewKind = "markdown"
	PreviewKindUIResource PreviewKind = "ui_resource"
	PreviewKindThumbnail  PreviewKind = "thumbnail"
)

// ArtifactKind 是附件、子产物、中间产物的资源种类。
type ArtifactKind string

const (
	ArtifactKindFile   ArtifactKind = "file"
	ArtifactKindURL    ArtifactKind = "url"
	ArtifactKindWidget ArtifactKind = "widget"
	ArtifactKindCode   ArtifactKind = "code"
	ArtifactKindImage  ArtifactKind = "image"
	ArtifactKindData   ArtifactKind = "data"
)

// ActionKind 是交付卡操作按钮种类。
type ActionKind string

const (
	ActionKindNavigate          ActionKind = "navigate"
	ActionKindDownload          ActionKind = "download"
	ActionKindTriggerCapability ActionKind = "trigger_capability"
	ActionKindExport            ActionKind = "export"
)

// Thumbnail 是 PreviewBlock(kind=thumbnail) 的单张缩略图。
type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Alt    string `json:"alt,omitempty"`
}

// PreviewBlock 是主对话卡片主体和 DecisionGate dry-run 预览的共享展示块。
type PreviewBlock struct {
	Kind       PreviewKind `json:"kind"`
	Content    string      `json:"content,omitempty"`
	UIResource *UIResource `json:"ui_resource,omitempty"`
	Thumbnails []Thumbnail `json:"thumbnails,omitempty"`
	Truncated  bool        `json:"truncated,omitempty"`
}

// ArtifactRef 是资源轻量引用，仅含定位信息。
type ArtifactRef struct {
	Kind  ArtifactKind `json:"kind"`
	Ref   string       `json:"ref"`
	Title string       `json:"title"`
}

// Artifact 是交付物附件或子产物，可下载、可打开，也可带内联预览。
type Artifact struct {
	Kind      ArtifactKind  `json:"kind"`
	Ref       string        `json:"ref"`
	Title     string        `json:"title"`
	Preview   *PreviewBlock `json:"preview,omitempty"`
	Mime      string        `json:"mime,omitempty"`
	SizeBytes int64         `json:"size_bytes,omitempty"`
}

// Action 是交付卡操作按钮，Payload 结构随 Kind 变化。
type Action struct {
	Label   string          `json:"label"`
	Kind    ActionKind      `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// DeliverableMetadata 是交付物时间戳与来源标识。
type DeliverableMetadata struct {
	CreatedAt   int64  `json:"created_at"`
	CompletedAt int64  `json:"completed_at"`
	SourceKind  string `json:"source_kind,omitempty"`
}

// Deliverable 是一次能力运行的最终产物结构化表达。
type Deliverable struct {
	ID            string              `json:"id"`
	RunID         string              `json:"run_id"`
	CanonicalName string              `json:"canonical_name"`
	Type          DeliverableType     `json:"type"`
	Title         string              `json:"title"`
	Summary       string              `json:"summary"`
	Preview       PreviewBlock        `json:"preview"`
	Artifacts     []Artifact          `json:"artifacts,omitempty"`
	Actions       []Action            `json:"actions,omitempty"`
	State         DeliverableState    `json:"state"`
	Version       int                 `json:"version"`
	Supersedes    string              `json:"supersedes,omitempty"`
	Metadata      DeliverableMetadata `json:"metadata"`
}
