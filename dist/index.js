// @wuhanyuhan/ks-types runtime barrel：手动镜像 dist/widgets.d.ts 顶层 export const 字面量。
//
// 设计权衡：tygo 仅生成 .d.ts（types-only），但 widgets.d.ts 中含 wire-format 字面量
// (postmessage method / sandbox flag / WidgetURIScheme)，前端 SDK 与 keystone host 在运行时
// 都需要这些值。本文件手动镜像；当 widgets.d.ts 新增 export const 时必须同步更新。
//
// 漂移守护：CI 可加 grep 比对 widgets.d.ts 与本文件 const 名一致性。

// postMessage 方法名（host ↔ iframe）。
export const PMMethodAppReady = 'app.ready'
export const PMMethodAppData = 'app.data'
export const PMMethodAppResize = 'app.resize'
export const PMMethodAppCallServerTool = 'app.callServerTool'
export const PMMethodAppToolResult = 'app.toolResult'
export const PMMethodAppToolError = 'app.toolError'
export const PMMethodAppUpdateModelContext = 'app.updateModelContext'
export const PMMethodAppClose = 'app.close'
export const PMMethodAppNotify = 'app.notify'
export const PMMethodAppOpenLink = 'app.openLink'
export const PMMethodMountedRouteChanged = 'keystone.mounted.route.changed'
export const PMMethodMountedRouteRestore = 'keystone.mounted.route.restore'

// iframe sandbox flag（自定义 widget 容器）。
export const SandboxFlagAllowDownloads = 'allow-downloads'
export const SandboxFlagAllowPopups = 'allow-popups'
export const SandboxFlagAllowModals = 'allow-modals'
export const SandboxFlagAllowSameOrigin = 'allow-same-origin'

// widget URI scheme。
export const WidgetSchemeKS = 'ks'
export const WidgetSchemeCustom = 'ui'

// 统一交付体验：决策门。
export const GateModeConfirm = 'confirm'
export const GateModeInput = 'input'
export const GateModeChoice = 'choice'
export const GateStatePending = 'pending'
export const GateStateAnswered = 'answered'
export const GateStateExpired = 'expired'
export const GateAnswerSourceWarroom = 'warroom'
export const GateAnswerSourceChatText = 'chat_text'
export const GateAnswerSourceChatCard = 'chat_card'

// 统一交付体验：交付物。
export const DeliverableArticle = 'article'
export const DeliverableReport = 'report'
export const DeliverableSpreadsheet = 'spreadsheet'
export const DeliverableCode = 'code'
export const DeliverableImageSet = 'image_set'
export const DeliverableDocument = 'document'
export const DeliverableDataset = 'dataset'
export const DeliverableComposite = 'composite'
export const DeliverableStateDraft = 'draft'
export const DeliverableStateFinalized = 'finalized'
export const DeliverableStateSuperseded = 'superseded'
export const PreviewKindMarkdown = 'markdown'
export const PreviewKindUIResource = 'ui_resource'
export const PreviewKindThumbnail = 'thumbnail'
export const ArtifactKindFile = 'file'
export const ArtifactKindURL = 'url'
export const ArtifactKindWidget = 'widget'
export const ArtifactKindCode = 'code'
export const ArtifactKindImage = 'image'
export const ArtifactKindData = 'data'
export const ActionKindNavigate = 'navigate'
export const ActionKindDownload = 'download'
export const ActionKindTriggerCapability = 'trigger_capability'
export const ActionKindExport = 'export'

// 统一交付体验：专家活动。
export const StepKindAction = 'action'
export const StepKindReasoning = 'reasoning'
export const StepKindMessage = 'message'
export const StepKindWait = 'wait'
export const PhaseResearch = 'research'
export const PhaseDrafting = 'drafting'
export const PhaseReviewing = 'reviewing'
export const PhaseFinalizing = 'finalizing'
export const PhaseDone = 'done'
export const PhaseBlocked = 'blocked'
export const ActivityStatusRunning = 'running'
export const ActivityStatusDone = 'done'
export const ActivityStatusBlocked = 'blocked'
export const ActivityStatusFailed = 'failed'
