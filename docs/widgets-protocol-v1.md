# widgets-protocol-v1 — 协议参考

| 字段 | 值 |
|------|-----|
| **协议版本** | widgets-protocol-v1 |
| **首发 ks-types 版本** | v0.6.0（minor） |
| **状态** | Draft |
| **跨语言载体** | Go 端 `github.com/wuhanyuhan/ks-types`；TS 端 `dist/widgets.d.ts`（tygo 派生） |

本文档是 **协议层契约**：squad 端如何在 manifest / runtime 声明 widget；keystone 后端 / 前端 / SDK 如何消费。仅描述协议字段，不描述内部实现细节。

---

## 目录

- [§1. 概述](#1-概述)
- [§2. 协议字段](#2-协议字段)
- [§3. 五个 MVP widget 数据 schema](#3-五个-mvp-widget-数据-schema)
- [§4. WidgetURI 解析规则](#4-widgeturi-解析规则)
- [§5. 版本兼容矩阵](#5-版本兼容矩阵)
- [§6. postMessage 协议（仅自定义 widget）](#6-postmessage-协议仅自定义-widget)
- [§7. 错误码列表](#7-错误码列表)
- [§8. 演进规则](#8-演进规则)

---

## §1. 概述

### 1.1 widgets-protocol-v1 是什么

widgets-protocol-v1 定义 Keystone 生态里 **MCP tool 调用结果如何在对话流气泡里以 widget 形式渲染** 的协议层契约。典型链路：

```
LLM 调 squad.review_draft(draft_id=42)
  └─► squad runtime 返回 CallToolResult{content, _meta:{ui, keystone.ui.data}}
       └─► keystone proxy NormalizeToolResult → 校验 → UIResource
            └─► 前端 ToolUIRenderer → 同源 React 组件 / iframe escape hatch
```

5 个 MVP widget 覆盖 80% 用例（见 §3）；自定义 widget 走 `ui://` iframe escape hatch（见 §6 postMessage 协议）。

### 1.2 与 MCP Apps spec 的关系

Anthropic 2026-01-26 发布的 MCP Apps spec 让 tool 返回 iframe 资源。本协议**没有照搬**：主路径走同源 React（不开 iframe，性能/a11y/调试体验优于 iframe）；iframe 仅作 ≤ 20% escape hatch；协议字段定义在 ks-types 强类型 + `Validate()`（spec 自身不约束 schema）；通过 tygo 派生 TS d.ts 实现跨语言 single source。

### 1.3 与 nav 反代壳 UI 的关系

**完全正交**。nav 是"用户主动打开 squad 应用"的产品线（`open_mode = dialog/fullpage/tab`），承载 squad 自有 SPA；widgets-protocol-v1 是"LLM 自动调 tool 结果渲染"的产品线。两者协议字段、proxy 路径、前端组件树、数据库表互不感知。

---

## §2. 协议字段

### 2.1 字段总览

| 阶段 | 出现位置 | 字段 | 承载类型 |
|------|---------|------|---------|
| **Manifest**（mount 时） | `/meta` 响应 | `capabilities.ui` | `MetaCapabilities` |
| Manifest | `/meta` 响应 | `tools[]._meta.ui` | `ToolUIBinding` |
| **Runtime**（每次 tool 调用） | `CallToolResult._meta` | `ui` | `MetaUIDecl`（per-call override） |
| Runtime | `CallToolResult._meta` | `keystone.ui.data` | `MetaKeystoneUIDecl` |
| **后端 normalize 输出** | `ToolCallResult.UIResource` | 整体 | `UIResource` |

### 2.2 Manifest 字段（squad `/meta` 端点）

squad 在 `/meta` 端点声明 capability + tool widget binding，keystone 在 mount 时拉取并登记到平台注册表。

```json
{
  "name": "ks-mcp-squad-marketing",
  "version": "0.6.0",
  "capabilities": {"ui": {"enabled": true, "requested_sandbox": ["allow-downloads"]}},
  "tools": [
    {"name": "review_draft", "_meta": {"ui": {"widget": "ks://widgets/diff-review@v1"}}},
    {"name": "list_drafts",  "_meta": {"ui": {"widget": "ks://widgets/list-actions@v1"}}},
    {"name": "get_brand_manual", "_meta": {"ui": {
        "widget": "ui://marketing/brand-editor", "sandbox_hints": ["allow-downloads"]}}},
    {"name": "get_weekly_retro"}
  ]
}
```

**Go 类型**（引用 ks-types）：

| 类型 | 文件位置 | 字段 | 必填 | 说明 |
|------|----------|------|------|------|
| `MetaResponse` | `meta.go:31` | `Capabilities *MetaCapabilities` | 否 | omitempty |
| `MetaCapabilities` | `meta.go:95` | `UI *CapabilitiesUI` | 否 | omitempty |
| `CapabilitiesUI` | `widgets.go:43` | `Enabled bool`、`RequestedSandbox []string` | enabled 必填 | tool_ui 总开关 + 申请 sandbox |
| `ToolInfo` | `meta.go:86` | `Meta *ToolInfoMeta` | 否 | omitempty |
| `ToolInfoMeta` | `meta.go:100` | `UI *ToolUIBinding` | 否 | omitempty |
| `ToolUIBinding` | `widgets.go:10` | `Widget string`、`SandboxHints []string` | widget 必填 | URI 合法（§4）；hints 仅 ui:// 用 |

**字段约束**：

- `capabilities.ui.enabled = false` / 字段缺失 → keystone 忽略 `tools[]._meta.ui`
- 某 tool 不带 `_meta.ui` → 不走 widget 渲染（保持纯文本/JSON）
- `requested_sandbox` 仅 `ui://` 生效；mount 按信任级别审批
- `widget` URI 校验失败 → mount 拒绝该 binding + WARN（不杀 squad）

### 2.3 Runtime 字段（每次 tool 调用响应）

squad tool 实现返回的 `CallToolResult` 通过 `_meta` 携带 widget 数据：

```json
{
  "content": [{"type": "text", "text": "已审阅 draft 42，有 3 处建议改动。"}],
  "_meta": {
    "ui": {"widget": "ks://widgets/diff-review@v1"},
    "keystone": {"ui": {"data": {
      "title": "5月营销月报",
      "diff": [
        {"type": "context", "text": "这是开篇..."},
        {"type": "delete", "text": "我们的产品很好用"},
        {"type": "insert", "text": "我们的产品在 X 行业有 30% 性能优势"}
      ],
      "actions": [
        {"id": "approve", "label": "批准", "variant": "primary"},
        {"id": "reject", "label": "拒绝", "variant": "destructive"}
      ]
    }}}
  }
}
```

| 类型 | 文件位置 | 字段 | 说明 |
|------|----------|------|------|
| `MetaUIDecl` | `widgets.go:32` | `Widget string`、`SandboxHints []string` | per-call override（多数场景留空） |
| `MetaKeystoneUIDecl` | `widgets.go:38` | `Data json.RawMessage` | widget 数据；schema 严格匹配 widget URI |

**字段语义**：

- `content[]`：MCP spec 约定字段——**必填**，即使有 widget 也要填一句人类可读摘要给 LLM/transcript
- `_meta.ui.widget`：可省略；省略时用 manifest binding；填了等于 per-call override
- `_meta.ui.sandbox_hints`：可省略；per-call 申请；必须是 `requested_sandbox` 子集
- `_meta.keystone.ui.data`：走 widget 渲染时**必填**；schema 严格匹配 `WidgetXxxV1`

### 2.4 后端 normalize 输出（前端消费）

keystone proxy `NormalizeToolResult` 把上面的 raw `_meta` 转成统一的 `UIResource` 给前端：

| 类型 | 文件路径 | 字段 | 说明 |
|------|----------|------|------|
| `UIResource` | `widgets.go:17` | `Widget string` | normalize 后的 widget URI |
| `UIResource` | `widgets.go:17` | `Data json.RawMessage` | data payload（已校验） |
| `UIResource` | `widgets.go:17` | `SandboxHints []string` | omitempty；仅 ui:// 有效 |
| `UIResource` | `widgets.go:17` | `SrcURL string` | omitempty；仅 ui:// 时填反代 URL |
| `UIResource` | `widgets.go:17` | `Error *UIResourceError` | omitempty；宽容模式渲染失败占位符 |
| `UIResourceError` | `widgets.go:26` | `Code string` | 错误码（§7） |
| `UIResourceError` | `widgets.go:26` | `Message string` | 人类可读消息 |

---

## §3. 五个 MVP widget 数据 schema

每个 widget 对应一个 Go struct（`widgets_data.go`）+ 注册到 `SharedWidgetSchemas`（`widgets_registry.go:10`），同时派生为 TS d.ts（`dist/widgets.d.ts`，前端 SDK 消费）。

校验入口：keystone `NormalizeToolResult` 调 `json.Decoder.DisallowUnknownFields()` 严格 decode，再调可选的 `Validate() error`。

### 3.1 list-actions@v1

**Go 类型**：`WidgetListActionsV1`（`widgets_data.go:59`）；用途：列表 + 行级操作（calendar/email/git issues/legal compliance/web-fetch）。

**字段表**（顶层 + 行）：

| 路径 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 否 | omitempty |
| `items` | `[]WidgetListItem` | 是 | 列表行 |
| `actions` | `[]WidgetActionDescriptor` | 否 | 列表级 action |
| `empty` | `*WidgetEmptyState` | 否 | 空态占位 |
| `items[].id` | string | 是 | 行标识 |
| `items[].title` | string | 是 | 行标题 |
| `items[].subtitle` / `icon` | string | 否 | 副标题；lucide-react 图标名 |
| `items[].badges` | `[]WidgetBadge` | 否 | 状态徽章 |
| `items[].metadata` | `map[string]any` | 否 | 自由扩展字段 |
| `items[].row_actions` | `[]WidgetActionDescriptor` | 否 | 行级 action |

**Validate 规则**（`widgets_data.go:78`）：每个 item 的 `id` 与 `title` 必须非空。

**Happy path 示例**（拉取草稿列表）：

```json
{
  "title": "待审阅草稿",
  "items": [
    {
      "id": "42", "title": "5月营销月报", "subtitle": "2 小时前", "icon": "file-text",
      "badges": [{"label": "待审", "variant": "warning"}],
      "row_actions": [
        {"id": "review", "label": "审阅", "variant": "primary"},
        {"id": "discard", "label": "丢弃", "variant": "destructive"}
      ]
    },
    {"id": "43", "title": "新品发布稿", "badges": [{"label": "草稿"}]}
  ],
  "empty": {"icon": "inbox", "title": "暂无待审稿件"}
}
```

**Error path**：缺 `items[0].title` → `Validate()` 返回 `items[0].title is required`。

### 3.2 diff-review@v1

**Go 类型**：`WidgetDiffReviewV1`（`widgets_data.go:96`）；用途：Diff + 审批（marketing 审稿/dev PR review/legal 合同对比）。

**字段表**：

| 路径 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | diff 标题 |
| `subtitle` | string | 否 | omitempty |
| `diff` | `[]WidgetDiffSegment` | 是 | 至少 1 段 |
| `actions` | `[]WidgetActionDescriptor` | 是 | 审批按钮 |
| `annotations` | `[]WidgetDiffAnnotation` | 否 | 段级批注 |
| `diff[].type` | string | 是 | "context" \| "insert" \| "delete" |
| `diff[].text` | string | 是 | 段文本 |
| `annotations[].anchor_index` | int | 是 | 指向 diff 段索引 |
| `annotations[].severity` | string | 是 | "info" \| "warning" \| "error" |
| `annotations[].message` | string | 是 | 批注内容 |

**Validate 规则**（`widgets_data.go:122`）：
- `diff` 至少 1 段
- 每段 `type` 必须在 `{context, insert, delete}` 内
- 每个 annotation 的 `anchor_index` 必须 ∈ `[0, len(diff))`
- 每个 annotation 的 `severity` 必须在 `{info, warning, error}` 内

**Happy path 示例**：

```json
{
  "title": "5月营销月报 — 修改建议",
  "diff": [
    {"type": "context", "text": "这是开篇..."},
    {"type": "delete", "text": "我们的产品很好用"},
    {"type": "insert", "text": "我们的产品在 X 行业有 30% 性能优势"}
  ],
  "annotations": [
    {"anchor_index": 1, "severity": "warning", "message": "原句过于空泛"}
  ],
  "actions": [
    {"id": "approve", "label": "批准", "variant": "primary"},
    {"id": "reject", "label": "拒绝", "variant": "destructive"}
  ]
}
```

**Error path**：`anchor_index = 5` 但 `diff` 只有 3 段 → `annotations[0].anchor_index 5 out of range [0,3)`。

### 3.3 timeline@v1

**Go 类型**：`WidgetTimelineV1`（`widgets_data.go:148`）；用途：时间轴 + 节点详情（dev pipeline/legal deadline/marketing campaign）。

**字段表**：

| 路径 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 否 | omitempty |
| `events` | `[]WidgetTimelineNode` | 是 | 时间轴节点（按 time 升序） |
| `events[].id` | string | 是 | 节点 ID |
| `events[].time` | string | 是 | RFC3339 UTC（严格格式） |
| `events[].title` | string | 是 | 节点标题 |
| `events[].status` | string | 是 | "pending" \| "running" \| "success" \| "failed" \| "skipped" |
| `events[].subtitle` / `detail` | string | 否 | omitempty |
| `events[].actions` | `[]WidgetActionDescriptor` | 否 | 节点级 action |

**Validate 规则**（`widgets_data.go:175`）：
- 每个 event 的 `id` / `title` 非空
- `time` 必须 RFC3339 合法（Go 端用 `time.Parse(time.RFC3339, ...)`；Python 端拒绝 naive datetime）
- `status` 必须在 5 个状态之一

**Happy path 示例**：

```json
{
  "title": "CI 流水线 — feature/widget-poc",
  "events": [
    {"id": "build", "time": "2026-05-04T08:00:00Z", "title": "build", "status": "success", "subtitle": "12s"},
    {"id": "test", "time": "2026-05-04T08:00:12Z", "title": "test", "status": "running",
     "actions": [{"id": "cancel", "label": "取消", "variant": "destructive"}]},
    {"id": "deploy", "time": "2026-05-04T08:02:30Z", "title": "deploy", "status": "pending"}
  ]
}
```

**Error path**：`time = "2026-05-04 08:00:00"`（用空格分隔）→ Go RFC3339 拒绝 → `events[0]: invalid time "2026-05-04 08:00:00" (expect RFC3339)`。

### 3.4 card-grid@v1

**Go 类型**：`WidgetCardGridV1`（`widgets_data.go:200`）；用途：卡片网格（knowledge 检索结果/web-fetch 搜索结果/git 仓库列表）。

**字段表**：

| 路径 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 否 | omitempty |
| `columns` | int | 否 | 0 = 默认 2 列；非 0 必须 ∈ `[1,4]` |
| `cards` | `[]WidgetCard` | 是 | 卡片列表 |
| `cards[].id` | string | 是 | 卡片 ID |
| `cards[].title` | string | 是 | 卡片标题 |
| `cards[].excerpt` | string | 否 | omitempty |
| `cards[].image_url` | string | 否 | 必须 https（inline 渲染防 XSS） |
| `cards[].source_label` | string | 否 | omitempty |
| `cards[].source_url` | string | 否 | 允许 http/https（用户跳转） |
| `cards[].score` | `*float64` | 否 | omitempty；非 nil 必须 ∈ `[0,1]` |
| `cards[].badges` | `[]WidgetBadge` | 否 | omitempty |
| `cards[].actions` | `[]WidgetActionDescriptor` | 否 | omitempty |

**Validate 规则**（`widgets_data.go:223`）：
- `columns ≠ 0` 时必须 ∈ `[1,4]`
- 每张卡的 `id` / `title` 非空
- `image_url` 必须 `https://`（拒绝 http/data/javascript）
- `source_url` 必须 `http://` 或 `https://`
- `score` 非 nil 必须 ∈ `[0,1]`

**Happy path 示例**（知识检索）：

```json
{
  "title": "检索结果（top 4）",
  "columns": 2,
  "cards": [
    {"id": "kb-1", "title": "新员工入职流程", "excerpt": "onboarding checklist...",
     "image_url": "https://cdn.example.com/onboarding.png",
     "source_label": "HR 知识库", "source_url": "https://kb.example.com/onboarding",
     "score": 0.92, "badges": [{"label": "官方"}]},
    {"id": "kb-2", "title": "工卡申请流程",
     "source_url": "https://kb.example.com/badge", "score": 0.78}
  ]
}
```

**Error path**：`image_url = "http://x"` → `cards[0].image_url: must be https URL, got "http://x"`。

### 3.5 image-variants@v1

**Go 类型**：`WidgetImageVariantsV1`（`widgets_data.go:256`）；用途：图片 + 变体操作（image-gen/browser screenshot/document 预览）。

**字段表**：

| 路径 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 否 | omitempty |
| `primary` | `WidgetImageItem` | 是 | 主图 |
| `variants` | `[]WidgetImageItem` | 否 | 变体（不同尺寸/风格） |
| `actions` | `[]WidgetActionDescriptor` | 否 | 图片级 action |
| `(WidgetImageItem).id` | string | 是 | 图片 ID |
| `(WidgetImageItem).url` | string | 是 | 必须 https |
| `(WidgetImageItem).alt_text` | string | 是 | 必填（a11y 强制） |
| `(WidgetImageItem).width` / `height` | int | 否 | omitempty；非负（0 视为未设置） |
| `(WidgetImageItem).caption` | string | 否 | omitempty |
| `(WidgetImageItem).actions` | `[]WidgetActionDescriptor` | 否 | 单图级 action |

**Validate 规则**（`widgets_data.go:296`）：
- `primary.url` / 每个 variant 的 url 必须 `https://`
- `alt_text` 必填（无障碍强制）
- `width` / `height` 可省略，但负数拒绝

**Happy path 示例**（图片生成）：

```json
{
  "title": "Logo 候选",
  "primary": {"id": "v1", "url": "https://cdn.example.com/logo-v1.png",
              "alt_text": "蓝色简约 logo，方形", "width": 1024, "height": 1024},
  "variants": [
    {"id": "v2", "url": "https://cdn.example.com/logo-v2.png", "alt_text": "蓝色简约 logo，圆形"},
    {"id": "v3", "url": "https://cdn.example.com/logo-v3.png", "alt_text": "深紫渐变 logo"}
  ],
  "actions": [
    {"id": "regenerate", "label": "再生成"},
    {"id": "save", "label": "保存", "variant": "primary"}
  ]
}
```

**Error path**：`primary.alt_text = ""` → `primary.alt_text required`。

### 3.6 通用辅助类型

| 类型 | 文件路径 | 字段 |
|------|----------|------|
| `WidgetActionDescriptor` | `widgets_data.go:16` | `id`（必填）、`label`（必填）、`variant`、`icon`、`disabled`、`tooltip`、`confirm_prompt` |
| `WidgetBadge` | `widgets_data.go:27` | `label`（必填）、`variant` |
| `WidgetEmptyState` | `widgets_data.go:33` | `icon`、`title`（必填）、`message` |

`variant` 取值：`primary` / `default` / `destructive` / `ghost` / `link`。

---

## §4. WidgetURI 解析规则

实现：`widget_uri.go:39 ParseWidgetURI`。

### 4.1 两种 scheme

| Scheme | 形态 | 用途 | 渲染 |
|--------|------|------|------|
| `ks://` | `ks://widgets/{name}@{version}` | 共享 widget（5 个 MVP） | 同源 React 组件 |
| `ui://` | `ui://{squad-id}/{path}` | 自定义 widget（escape hatch） | iframe sandbox |

### 4.2 ks:// 形态

**正则**（`widget_uri.go:30`）：`^ks://widgets/([a-z][a-z0-9-]*)@(v\d+)$`

合法示例：
- `ks://widgets/diff-review@v1`
- `ks://widgets/list-actions@v1`
- `ks://widgets/timeline@v1`
- `ks://widgets/card-grid@v1`
- `ks://widgets/image-variants@v1`

非法示例：
- `ks://widgets/diff-review` — 缺 version
- `ks://widgets/diff-review@1` — version 必须 `v\d+`
- `ks://widgets/Diff-Review@v1` — name 必须小写
- `ks://others/foo@v1` — 必须在 `widgets/` 命名空间下
- `ks://widgets/diff-review@v1?x=1` — 禁止 query/fragment

### 4.3 ui:// 形态

**约束**：`ui://{squad-id}/{path}`，`squad-id` 与 `path` 都不能为空；`path` 可含 `/` 多段。

合法示例：
- `ui://marketing/brand-editor`
- `ui://dev-tools/screenshot/region-picker`

非法示例：
- `ui://marketing` — 缺 path
- `ui:///brand-editor` — squad-id 空
- `ui://marketing/` — path 空
- `ui://marketing/brand-editor#x` — 禁止 fragment

### 4.4 跨 squad 引用规则（ui://）

**核心约束**：`ui://{squad-id}/...` 的 `squad-id` 必须等于发起 widget 的 squad ID（即 mount 时 squad manifest 的 `name`），否则拒绝。

| 场景 | 行为 |
|------|------|
| Manifest 阶段 squad A 引用 `ui://squadB/widget` | mount 时拒绝 + WARN |
| Runtime 阶段 squad A 响应 `_meta.ui.widget = "ui://squadB/..."` | normalize 拒绝 → `UIResource.Error.Code = cross_squad_widget` |

辅助方法：`(*WidgetURI).SquadID()`（`widget_uri.go:69`）、`(*WidgetURI).Path()`（`widget_uri.go:78`）。

---

## §5. 版本兼容矩阵

### 5.1 widget URI 版本规则（ks://）

| 变更类型 | 是否需要升 widget major | ks-types semver 影响 | 前端影响 |
|---------|----------------------|--------------------|-----------------|
| 给 `WidgetXxxV1` 加可选字段（`omitempty`） | 否（仍 `@v1`） | minor +1 | minor +1（兼容老组件） |
| 改字段类型 / 删字段 / 改字段语义 | 是（`@v2` 共存） | minor +1（保留 v1） | minor +1（同时 ship v1 + v2） |
| 全面下线 widget `@v1` | 是 | major +1 | major +1 |

### 5.2 widgets-protocol-v1 与老版本对比

widgets-protocol-v1 是**首发协议层**，不存在更老的兼容对端：

- 不声明 `capabilities.ui` 的 squad（如所有 ks-mcp-* 现状）→ keystone 完全跳过 normalize（fast path），与升级前行为完全一致
- `capabilities.ui.enabled = false` → 同上
- 部分启用：`capabilities.ui.enabled = true`，部分 tool 不带 `_meta.ui` → 这些 tool fast path

### 5.3 兼容窗口承诺

- 同 widget 多 major 共存：≥ 6 个月
- ks-types breaking：每年最多 1 次 major
- 前端 breaking：跟 ks-types 节奏

下线 widget `@vN`：在 `widgets_registry.go` `DeprecatedWidgets`（`widgets_registry.go:27`）登记 `DeprecationInfo{DeprecatedSince, EOLDate, Replacement, MigrationGuide}`，提前 6 个月预告。

### 5.4 加新 capability 字段

`capabilities.ui.*` 子字段可加；老 squad 不声明 = 默认值（兼容）。例如未来加 `capabilities.ui.theme_hints`：老 squad 不声明 → 等于不申请，不影响。

---

## §6. postMessage 协议（仅自定义 widget）

共享 widget（`ks://`）走同源 React，**不用** postMessage。本节仅适用 `ui://` iframe 路径。

### 6.1 方法名常量

定义在 `widget_postmessage.go`：

| 常量 | 字符串 | 方向 | 说明 |
|------|--------|------|------|
| `PMMethodAppReady` | `app.ready` | iframe → host | iframe mount 完成，请求注入数据 |
| `PMMethodAppData` | `app.data` | host → iframe | 注入 widget 数据（响应 ready） |
| `PMMethodAppResize` | `app.resize` | iframe → host | 上报需要的高度（≤ 800px） |
| `PMMethodAppCallServerTool` | `app.callServerTool` | iframe → host | 反向调 tool（同 squad 限定） |
| `PMMethodAppToolResult` | `app.toolResult` | host → iframe | callServerTool 成功响应 |
| `PMMethodAppToolError` | `app.toolError` | host → iframe | callServerTool 失败响应 |
| `PMMethodAppUpdateModelContext` | `app.updateModelContext` | iframe → host | 通知 LLM 上下文更新 |
| `PMMethodAppClose` | `app.close` | iframe → host | iframe 主动收起 |
| `PMMethodAppNotify` | `app.notify` | iframe → host | 在 keystone 顶部统一通知 |
| `PMMethodAppOpenLink` | `app.openLink` | iframe → host | 申请打开外部链接 |

未列方法静默忽略。完整 payload 协议、JSON-RPC 信封、origin/source 校验算法、targetOrigin 规则由前端消费方实现，不在协议层强制。

### 6.2 sandbox flag 常量

定义在 `widget_postmessage.go:27`：

| 常量 | 字符串 | 默认白名单 |
|------|--------|----------|
| `SandboxFlagAllowDownloads` | `allow-downloads` | 是 |
| `SandboxFlagAllowPopups` | `allow-popups` | 否（按 squad 审批） |
| `SandboxFlagAllowModals` | `allow-modals` | 否（按 squad 审批） |
| `SandboxFlagAllowSameOrigin` | `allow-same-origin` | 否（特殊审批，能不用就别用） |

基础 sandbox 集（不可变）：`allow-scripts allow-forms`。

---

## §7. 错误码列表

`UIResource.Error.Code` 在宽容模式下出现，前端渲染"widget 加载失败"占位符。错误码字符串由平台后端落地；本节列出协议层契约的稳定字符串。

| Code | 触发场景 | 协议层影响 |
|------|---------|-----------|
| `widget_uri_invalid` | URI 解析失败（§4 任一非法形态） | normalize 失败 |
| `schema_mismatch` | data payload 与 widget URI 对应 schema 不符（含 `DisallowUnknownFields` 触发的多余字段、`Validate()` 业务错） | normalize 失败 |
| `ui_data_missing` | tool 声明了 widget 但 `_meta.keystone.ui.data` 缺失 | normalize 失败 |
| `cross_squad_widget` | `ui://` 的 squad-id 不等于发起 squad（§4.4） | normalize 拒绝 |
| `capability_disabled` | `_meta.ui.widget` 出现但 squad `capabilities.ui.enabled = false` | normalize 跳过 |
| `binding_not_registered` | tool 无 mount-time binding 且 runtime 也未声明 widget URI | fast path（不算错误） |
| `sandbox_hint_unapproved` | runtime `sandbox_hints` 含未审批 flag | normalize 过滤 + WARN |

严格模式（`strict_mode = true`）下任何错误都会让 tool call 整体失败；宽容模式（默认）转占位符。

---

## §8. 演进规则

### 8.1 加 widget

按 RFC 流程评审（≥ 2 squad 用例 + 数据 schema 提案 + a11y 设计），通过后：

1. ks-types：往 `widgets_data.go` 加 `WidgetXxxV1` struct（含 `Validate() error`）
2. ks-types：往 `widgets_registry.go` `SharedWidgetSchemas`（`widgets_registry.go:10`）注册条目
3. ks-types：升 minor 版本 + tygo 派生 TS d.ts + CHANGELOG
4. 前端消费方：同步实现 React 组件 + Storybook + axe-core
5. 前端消费方：往 `WidgetRegistry` 注册映射

加新 widget 不影响老 squad（兼容增量）。

### 8.2 改 widget

只允许 **加可选字段**（带 `omitempty`/`default_factory`）；改字段类型、删字段、改字段语义必须新建 `@v2` 共存。加可选字段流程：给 `WidgetXxxV1` struct 加 `omitempty` 字段、`Validate()` 不强制、ks-types 升 minor、前端优雅降级（老 squad 不发新字段也能跑）。

### 8.3 加 widget URI scheme

`ks://` / `ui://` 之外要加（如未来的 `npm://` 远程组件包加载）必须升 ks-types major。

### 8.4 加 capabilities 子字段 / postMessage 方法名

`capabilities.ui.*` 子字段可加；老 squad 不声明 = 默认值（兼容）。新增 postMessage 方法常量：minor +1；改既有语义：major +1 + 保留老方法 ≥ 6 个月。

### 8.5 永远不做（设计原则禁区）

- LLM 自己生成 widget HTML/data
- 单 squad 专用共享 widget（必须 ≥ 2 squad 用例）
- 跨 squad widget 调用（LLM 跨 squad 可，widget 不行）
- widget 嵌套 widget
- widget 反向控制 LLM 决策

---

## 参考资料

- 协议层代码（同仓）：
  - `widgets.go` — 绑定类型 + UIResource
  - `widget_uri.go` — URI 解析
  - `widgets_data.go` — 5 个 widget data schema
  - `widgets_registry.go` — schema 注册表
  - `widget_postmessage.go` — postMessage 方法名 + sandbox flag 常量
  - `meta.go` — MetaResponse 扩展（`Capabilities`、`ToolInfo.Meta`）
- TS 派生产物：`dist/widgets.d.ts`（前端 SDK 消费契约）
- squad 端 quickstart：参见 squad SDK 的 tool UI quickstart 文档
