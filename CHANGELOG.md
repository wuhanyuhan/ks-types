# Changelog

`ks-types` 对外类型契约的变更记录。遵循语义化版本；破坏性变更在条目中标注。

## [v0.37.0] - 2026-06-11（前端 tag：widgets-protocol-v2.1.0）

### Added (SDUI 实时数据源)

- **`UIDataSource` typed 实时数据源** + `UINode.Data *UIDataSource` 可选字段（`sdui_node.go`）：节点声明「订阅哪个**封闭枚举**数据源 + 具名参数」，前端按 `Kind` 经协作接线（`SDUIRenderContext.collab`）解析为可达 URL / 订阅——非自由表达式绑定。
- **`DataSourceTeamProgressStream` 常量**（首个 Kind）：订阅某个 run 的「团队实时进度流」（`Params={run_id}`），前端经反代直连 squad `/stream` SSE、`reduceFrame` 聚合成 TeamState。供 SDUI 作战室 `war-room` 原语消费。
- tygo 派生：`dist/widgets.d.ts` 同步出 `UIDataSource` / `data?` / `DataSourceTeamProgressStream`（前端 `@wuhanyuhan/ks-types` 经 `widgets-protocol-v2.1.0` tag 消费）。

加法式：`data` 字段 `omitempty`，既有 UINode wire 格式不变。

## [v0.36.0] - 2026-06-11

### Added (widgets-protocol v2 / SDUI 地基)

- **`UINode` 递归节点协议**（`sdui_node.go`）+ `MaxNestingDepth` 嵌套深度上限常量。`UINode{Type,Props,Children,Key}` 是 Server-Driven UI 的递归节点，前端经 tygo 派生（`dist/widgets.d.ts` 含 `UINode`/`SDUI*Props`）。
- **首批 SDUI 原语 props schema**（`sdui_primitives.go`）：容器 `stack/grid/card/section/tabs/split`、展示 `text/markdown/field-group/table/status-badge/metric/empty-state`、交互 `button/form/link`、复合（遗留 widget 降级）`list-actions/diff-review/timeline/card-grid/image-variants`。每个带值方法 `Validate()`。`SDUIActionIntent` 为 typed 交互意图。
- **原语注册表**（`sdui_registry.go`，不进 tygo）：`SDUIPrimitiveSchemas`（type→props reflect.Type）+ `ContainerPrimitives`（容器标记），供 keystone 后端递归校验器消费。
- **Go typed builder 子包** `github.com/wuhanyuhan/ks-types/sdui`：squad 端编译期类型安全构造 UINode 树，不手搓 JSON。
- 前端独立 tag 同步发布 `widgets-protocol-v2.0.0`（与 Go 模块 tag `v0.36.0` 同一 commit）。

## [v0.33.0] - 2026-06-03

### Removed (breaking)

- **删除全量 a2a 协议类型**：`a2a_task.go`、`a2a_skill.go`、`a2a_security.go`、`agent_card.go` 及其测试。a2a 协议栈已全栈退役，keystone 已删除完整 a2a 实现（后端/5 张表/2 个权限码/前端/capability 枚举 8→7/迁移 0164），全 yuhan 树无任何 a2a 符号残留引用。
- **删除 `manifest.go` 中 a2a 声明层**：`A2AConfig`、`A2ASkillDef` 类型及 `A2ASkillDef.Validate()` 方法。`AppSpec` 已无 a2a 字段（v0.31.0 起已移除 `AppSpec.A2A`），此次清除最后的 Deprecated 残留类型。

## [v0.32.0] - 2026-06-03

### Removed (breaking)

- **内部编排类型迁出公开仓**：`DecisionGate`、`Deliverable`、`ExpertActivity` 及相关类型迁入私有 `ks-internal-contracts` v0.2.0，不再暴露于公开 `ks-types`。

## [v0.31.0] - 2026-06-02

### Changed (breaking)

- **`SignInstanceJWT` 新增 `kid string` 第 4 参数**（`instance_claims.go`）：把 kid 写入 token header，供下游 JWKS 验签按 kid 选公钥。kid 算法由调用方控制（`ks-types` 不重复实现以避免两边漂移）；传 `""` 时不写 kid header（兼容旧测试与不走 JWKS 的链路），生产调用方应始终传非空 kid。

## [v0.29.0] - 2026-05-30

capability 命名去前缀：app 作者在 manifest 写能力**裸名** `name`，全局唯一 `canonical_name` 由注册期派生 `<app_id>.<name>`。

### Added

- **`CapabilitySpec.Name`** 裸名字段（`provides.capabilities[]`，app_provided 必填，单段无点，正则 `^[a-z][a-z0-9_-]*$`）：作者只写裸名（如 `web_search`），全局前缀由注册期派生。
- **`Canonical(appID, name)`** 导出 helper：`<app_id>.<name>` 派生逻辑的单一来源。

### Changed (breaking)

- **`provides.capabilities[]` 去前缀**：app 能力改写裸名 `name`，不再写全名 `canonical_name`；显式写 `canonical_name` 会被 `ProvidesSpec.Validate` 拒绝。
- **`requires.capabilities[].canonical_name` 维持全名**（引用他人已注册能力）；内置 YAML 能力声明维持写全名。

## [v0.28.0] - 2026-05-30

应用 manifest 声明层 clean-break：把 `AppSpec` / `CapabilitySpec` 重定型为四类型、五层精简的最终形态。

### Added

- **四类型体系** `app` / `squad` / `agent` / `skill`（`apptypes.go`）：取代旧 `service` / `assistant` / `extension`。
- **顶层 `auth:` 段**（`AuthSpec.Mode` + `EffectiveMode`）：app 级出入站鉴权契约，secure-by-default——有入站端点时空值派生 `keystone_jwks`，`none` 须显式 opt-out。
- **capability 内联 `output_schema`**：与 `input_schema` 对称，取代被砍的 `output_schema_ref`。
- **capability 内联 `decision_mode`** + `pre_authorize_duration`：三级 `user_only` / `user_authorized` / `agent_autonomous`，默认从 `side_effect_level` 派生（`EffectiveDecisionMode`）；取代被砍的 `requires_approval` + app 级 `compliance`。
- **解析期格式/边界校验**（错误信息均带允许值）：`version` semver（三段 `x.y.z`）、`id` 格式（`^[a-z][a-z0-9-]*$`）、`compatibility.keystone` 区间（如 `">=1.5.0 <2.0.0"`）；`runtime.resources` cpu/memory K8s 风格；capability `timeout_ms` 值域 `100..600000`。
- **字段级 best-practice doc 注释**：保留/新增字段统一标注 是什么 / 何时填 / 默认 / 最佳实践。

### Removed (breaking)

- **`CapabilitySpec` 死字段**：`typical_latency_ms`、`cost_hint`、`default_grant`、`allowed_callers`、`requires_approval`、`input_schema_ref`、`output_schema_ref`、`input_nl`、`output_nl`。
- **`BackendSpec.mcp_endpoint`**。
- **`AppSpec.compliance` 段**（`ComplianceConfig` / `ToolOverride` 类型）：合并为内联 `decision_mode`；保留 `DecisionMode` / `PreAuthorizeDuration` 枚举。
- **`AppSpec.license`**（`LicenseConfig` 类型）。
- **manifest 层 `a2a`**（`AppSpec.A2A` 字段）；`A2AConfig` / `A2ASkillDef` 类型标 `Deprecated`。
- **`runtime.port` / `runtime.ports`**：约定容器内监听 8080，manifest 不声明。
- **`notify_policy.on_started`**。
- **`managed_resources` 强制值样板**：各 spec 的 `required`、`mysql.database` / `mysql.user`、`object_storage.bucket` / `object_storage.prefix`、`cache.provider` / `cache.key_prefix`（只接受 `auto`、平台自动命名）；保留 `inject.*` / `access` / `retain_on_uninstall` / storage scopes。
- **`ProtectionLevel.preinstalled`**。

### Changed (breaking)

- **`mount.assistant` → `mount.agent`**（`AssistantMountSpec` → `AgentMountSpec`），去掉死字段 `routing_plan`。
- **capability `intent_summary` + `natural_description` 合并为单一 `description`**（`CapabilitySpec` + `CapabilityProfileSpec`）。
- **各 `ManagedXxxResourceSpec.Validate` 瘦身**：去掉 `required must be true` 与 `:auto` 检查，仅保留 `inject` env 名映射 / scopes 校验。

## [v0.27.0] - 2026-05-29

### Added

- **Store presentation manifest contract**（`apptypes.go` / `manifest.go`）：新增 `StorePresentation` 枚举与 `AppSpec.Store` 展示字段，支持 `role_agent` / `method_skill` / `toolkit` / `connector` / `service_app` / `expert_team` 六种 Store 呈现形态；`expert_team` 形态校验 `store.team.members[].key/name` 必填。

## [v0.23.0] - 2026-05-20

### Added

- **Capability profile manifest contract**（`manifest.go`）：为 `mount.assistant.profile` / `mount.skill.profile` 增加正式类型 `CapabilityProfileSpec`，承载 `canonical_name` / `display_name` / `intent_summary` / `natural_description` / `aliases` / `user_utterances` / `use_cases` / `domain_terms` / `input_nl` / `output_nl` / `negative_examples`，用于能力的自然语言召回画像。
- **CapabilitySpec 输入输出自然语言说明**：新增 `input_nl` / `output_nl` 字段。

## [v0.22.0] - 2026-05-19

### Removed (breaking)

- **`PlatformEmbeddingServiceSpec.Provider`**、**`.Dim`**、**`PlatformEmbeddingInjectSpec.APIBase`**：embedding 由平台托管并经 SDK 统一访问，相关字段不再允许声明。
- **`ManagedVectorStoreResourceSpec.Provider` / `.CollectionPrefix` / `.Dim` / `.Inject`** 及 **`ManagedVectorStoreInjectSpec`** 整个类型：向量存储统一由平台托管、经 proxy 访问，应用不再直连。

### Changed (breaking)

- **`PlatformEmbeddingServiceSpec.Validate()`**：`Required=true` 时 `Model` 必填。

## [v0.19.0] - 2026-05-19

Capability Mesh manifest schema 的协议层升级。

### Added

- **Capability Mesh manifest schema**（`manifest.go`）：
  - `AppSpec.Provides ProvidesSpec` —— 声明应用对外暴露的 capability 集合（`CapabilitySpec`）。
  - `AppSpec.Requires RequiresSpec` —— 取代旧 `Dependencies.Requires/Recommends`。`RequiresCapabilityItem.Mode` 三档枚举：`required` / `recommended` / `optional`。
  - `AppSpec.Conflicts ConflictsSpec` —— 应用级互斥声明，含 `id` + `reason`。
  - `BackendSpec` —— capability 后端路由：`kind ∈ {mcp_tool, http_endpoint}`。
  - `AppSpec.Validate` 适配：canonical_name 正则 / app_id 前缀强制 / 同 manifest 不重复 / `backend.kind` 与 `runtime.mode` 一致性 / `requires.mode` 与 `execution_mode` 枚举。
- `RuntimeModeExtension RuntimeMode = "extension"` 常量（`apptypes.go`）。

### Removed (breaking)

- **`AppSpec.Dependencies DependenciesSpec` 字段及关联类型**（`DependenciesSpec` / `DependencyItem` / `RecommendItem` / `ConflictItem`）：能力级依赖改用 `AppSpec.Requires.Capabilities[]`；应用级冲突改用 `AppSpec.Conflicts.Apps[]`。
- **`MountSpec.Service` / `MountSpec.Extension` 字段及关联类型**（`ServiceMountSpec` / `ExtensionMountSpec` / `LLMRequirements`，含 `AutoRegisterMCP` / `MCPEndpoint` / `DefaultAllowedTools` / `CreateAgent` / `LLMMode` / `AuthMode` 等字段）：能力暴露改用 `AppSpec.Provides.Capabilities[].backend`。`MountSpec` 仅保留 `assistant` / `skill` 两字段（与 capability 正交）。

## [v0.18.0] - 2026-05-19

### Added

- `KSAttachment` / `KSAttachmentResolved` 类型 + `KSAttachmentFieldName` 常量（`attachment.go`）：MCP 工具结果 envelope 的附件改写约定——MCP 服务在结果中以 `ks_attachments: []KSAttachment` 携带，平台 mcp proxy 拦截改写为带签名 URL 的 `[]KSAttachmentResolved`。

## [v0.17.0] - 2026-05-18

### Added

- `StandaloneFallbackSpec` 类型（`standalone_fallback.go`）：声明应用在 standalone 模式下的本地资源 fallback（MySQL / ObjectStorage / VectorStore / Cache / Storage）。
- `AppSpec.StandaloneFallback` 字段：与 `ManagedResources` 一一对应，managed 模式下忽略，standalone 模式下作为唯一权威源。

## [v0.12.0] - 2026-05-14

### Added

- **`AppSpec.ManagedResources` 字段 + `ManagedResourcesSpec` / `ManagedMySQLResourceSpec` / `ManagedMySQLInjectSpec` 类型**：声明 manifest `managed_resources.mysql`，由平台在安装时分配 MySQL database/user/password 并注入。
- **应用包签名摘要字段改为必填**。
- **配置公钥信任类型补充**。
- **`AppSpec.Compliance` 字段 + `ComplianceConfig` / `ToolOverride` / `DecisionMode` / `PreAuthorizeDuration` 类型**（`compliance.go`）：声明"决策权限三级"（`user_only` / `user_authorized` / `agent_autonomous`）默认值与 tool 级覆盖；`pre_authorize_duration` 五值枚举（`5m` / `30m` / `2h` / `24h` / `forever`），仅 `user_authorized` 模式生效；`omitempty` 保证旧 manifest 向后兼容。

## [v0.9.0] - 2026-05-05

### Added

- **i18n 类型 `LocalizedString` / `LocalizedTags`**（`localized.go`）：YAML 解析支持单形态（scalar / sequence）与 map 形态；`Get(locale)` fallback 链；`MarshalJSON` 永远输出 map 形态。
- **`AppSpec.Changelog` 字段**：当前版本的 changelog markdown。
- **`AttestationClaims` struct + `SignAttestation` / `VerifyAttestation`**：实例身份证明 JWT，`typ=ATT+JWT`、`aud=ks-client`、强制 `kid` header，独立于 Instance JWT。

### Changed

- `AppSpec.Summary` / `Description` 类型 `string` → `LocalizedString`；`AppSpec.Tags` 类型 `[]string` → `LocalizedTags`；`PricingSpec.Description` 类型 `string` → `LocalizedString`。

### Migration

- 老 manifest（单 string / sequence 形态）100% 向后兼容，作者无需改 manifest.yaml。
- 调用方读取上述字段需改为 `.Get(locale)`：

  ```go
  // 旧（v0.8.0）
  s := spec.Summary
  // 新（v0.9.0）
  s := spec.Summary.Get("zh-CN")
  ```

- JSON wire-format 变更：`summary` / `description` / `tags` / `pricing.description` 输出从 scalar/array 变为 map（如 `{"":"..."}` 或 `{"zh-CN":"..."}`），下游消费方需相应适配。

## [widgets-protocol-v1.0.0] - 2026-05-04

### Added

- npm 包分发面 `@wuhanyuhan/ks-types` v1.0.0（types-only 包）：
  - `package.json`（main/types/exports map 指向 `dist/index.{js,d.ts}`）
  - `dist/index.d.ts` 类型 barrel（re-export `widgets.d.ts`）
  - `dist/index.js` 运行时 barrel（手动镜像 `widgets.d.ts` 中 16 个 wire-format 字面量：10 个 PMMethodApp* / 4 个 SandboxFlag* / 2 个 WidgetScheme*）
  - README 补 TypeScript 类型分发说明（git 依赖消费方式）

### Notes

- widgets-protocol-v1 wire schema（`widgets.go` / `widget_uri.go` / `widgets_data.go` / `widget_postmessage.go` / `widgets_registry.go` 等）于 v0.6.0+ 合入；本 tag 不改动 schema，仅建立 JS 分发面。

## [v0.8.0] - 2026-04-30

### Added

- `ginmw.RequireAudience(svc string) Option`：`InstanceJWTMiddleware` 新增 functional option，强制 JWT 的 `aud` 包含指定服务名；不传时默认放行（向后兼容）。
- `ginmw.Option` 类型 + 中间件签名扩展：`InstanceJWTMiddleware(publicPEM, isRevoked, opts ...Option)`，旧调用方无需改动。

### Changed

- `SignInstanceJWT` 的 `Audience` 由 `["ks-hub", "ks-admin"]` 扩展为 `["ks-admin", "ks-hub", "ks-relay", "ks-llm-gateway"]`，覆盖生态全部云服务。

### Breaking Changes

- 无。aud 增加元素与新 option 均向后兼容：旧调用方 `InstanceJWTMiddleware(pub, isRevoked)` 无需改动；旧 token（aud 仅 2 项）只要不启用 `RequireAudience` 仍被接受。

## [v0.6.0] - 2026-04-19

### Added

- 新增 `config_schema.go`（MCP 配置 Schema + E2E 加密协议的共享契约）：
  - `ConfigSchemaResponse` —— MCP `/config-schema` 端点响应 data 字段
  - `ConfigPubkeyResponse` —— MCP `/config-pubkey` 端点响应 data 字段
  - `EncryptedConfigPayload` —— `POST /ks-config/save` 与 `/validate` 的 request body
  - `ConfigApplyResult` —— `POST /ks-config/save` 成功响应 data 字段
  - `AADCanonicalBytes(mcpID string, version uint64, fingerprint string) []byte` —— canonical AAD 字节序列化 helper
  - `Fingerprint(pubkey []byte) string` —— X25519 公钥指纹 helper

### Tests

- `config_schema_test.go`：内联 `aad_canonical` testvectors + `Fingerprint` 断言，字节级对齐验证。

### Breaking Changes

- 无。所有变更为新增类型和函数。

## [v0.5.0] - 2026-04-24

### Added

- `MetaResponse.Nav`（`*MetaNavDecl`）—— MCP 自声明左侧菜单项（label / icon / category / order / open_mode / entry_path / required_perms）
- `MetaResponse.Permissions`（`[]MetaPermissionDecl`）—— 权限码目录数组
- `MetaResponse.ConfigMode` —— `schema` / `iframe` / `none` 枚举
- `MetaResponse.ProtocolVersion` —— SemVer MAJOR.MINOR
- `MetaResponse.ConfigStatus` —— `unconfigured` / `via_frontend` / `via_cli` / `mixed`

### Breaking Changes

- 无。所有变更为 `omitempty` 可选字段；旧消费者解析未含新字段的响应行为不变。

## [v0.4.1] - 2026-04-17

### Added

- `ExtensionMountSpec.AuthMode` 字段（对齐 `ServiceMountSpec.AuthMode`）
- `Validate()` 加 extension mount auth_mode 合法性校验

## [0.4.0] - 2026-04-17

### Added

- `AuthMode` 枚举（`apptypes.go`）：`none` / `keystone_jwks` / `static_bearer`
- `ServiceMountSpec.AuthMode` 字段（`manifest.go`）——声明 MCP service /mcp 端点鉴权模式
- `Validate()` 增强：非法 auth_mode 值在解析校验阶段拒绝
- `MetaResponse` / `ConfigUIInfo` / `ToolInfo` 契约类型（`meta.go`）——`/meta` 端点响应结构
- `AuthMode.Default()` 辅助：空字符串归一为 `AuthModeNone`

### Breaking Changes

- 无。所有变更为向前兼容的新增字段。
