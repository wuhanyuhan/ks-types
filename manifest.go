package kstypes

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// canonicalNameRegex 是 capability canonical_name 的命名格式正则。
//
// 形态：<provider>.<verb>[.<sub>]...
//   - 段间用 . 分隔，至少两段
//   - 每段首字母小写，后续可含 a-z / 0-9 / _ / -
var canonicalNameRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]*(\.[a-z][a-z0-9_-]*)+$`)

// bareNameRegex 是 app_provided capability 裸名格式（单段无点）。
// app manifest 的 provides.capabilities[].name 作者写裸名（如 web_search），
// 全局 canonical_name 由注册期派生 <app_id>.<name>，作者不写前缀（DRY：app_id 已在顶层声明）。
var bareNameRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Canonical 由 app_id + 裸名派生全局唯一 canonical_name。
// keystone 注册期与三语言 SDK 共用同一派生逻辑，避免各处各拼前缀。
func Canonical(appID, name string) string { return appID + "." + name }

// AppSpec 身份字段的格式校验正则。三者均为解析期的“格式校验”——
// 只保证写法合法；version 升级单调性、compatibility install gate 由平台在后续阶段执行（见各字段 doc 注释）。
var (
	// semverRegex 严格三段 x.y.z，与 keystone 的 semver 校验同款。
	semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	// appIDRegex app id 合法形态：小写字母开头 + 小写字母/数字/连字符，无下划线。
	// 比 token 签发处更严——id 是 capability canonical_name 的前缀根，必须严。
	appIDRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	// compatRangeRegex 校验 compatibility.keystone 区间写法：单约束 ">=1.5.0"、
	// 双约束 ">=1.5.0 <2.0.0"、或裸版本 "1.5.0"（约束符 >=|<=|>|<|= 可选）。
	compatRangeRegex = regexp.MustCompile(`^(>=|<=|>|<|=)?\d+\.\d+\.\d+(\s+(>=|<=|>|<|=)?\d+\.\d+\.\d+)?$`)
)

// validCapabilityKinds 是 BackendSpec.Kind 当前支持的枚举集合。
var validCapabilityKinds = map[string]bool{
	"mcp_tool":      true,
	"http_endpoint": true,
}

// validExecutionModes 是 CapabilitySpec.ExecutionMode 当前支持的枚举集合。
var validExecutionModes = map[string]bool{
	"sync":         true,
	"long_running": true,
	"hybrid":       true,
}

// validRequiresModes 是 RequiresCapabilityItem.Mode 当前支持的枚举集合。
var validRequiresModes = map[string]bool{
	"required":    true,
	"recommended": true,
	"optional":    true,
}

// validSideEffectLevels 是 CapabilitySpec.SideEffectLevel 支持的枚举集合。
// 与 keystone 平台侧 side_effect_level / 编排官读写准入一致（可逆性档）。
var validSideEffectLevels = map[string]bool{
	"none":              true,
	"soft_reversible":   true,
	"hard_irreversible": true,
}

// AppSpec 应用 manifest.yaml 的完整结构
//
// v0.9.0 起 Summary / Description 升级为 LocalizedString、Tags 升级为 LocalizedTags，
// 同时新增 Changelog 字段。YAML 单 string / sequence 形态向后兼容，
// JSON 序列化统一为 map 形态，下游读取时调用 .Get(locale) 取值。
type AppSpec struct {
	// ID 应用唯一标识，也是本应用所有 capability canonical_name 的前缀根
	// （如 id=translator → 能力 translator.translate）。形态 ^[a-z][a-z0-9-]*$：
	// 小写字母开头、仅含小写字母/数字/连字符、无下划线——因是前缀根故必须严，解析期即校验。
	ID string `yaml:"id" json:"id"`
	// Name 应用稳定标识名（不 i18n，与 i18n 的 Summary/Description/Tags 形态不同）。
	Name string `yaml:"name" json:"name"`
	// Version 应用“包版本”，三段 semver（x.y.z）。与能力契约“名字即身份”（破坏性
	// 变更靠铸新 canonical_name、requires 永不带 version）是两条正交轴——此处管 app 包
	// 发布版本，非能力契约版本。enforcement：解析期仅校验 semver 格式；升级单调性由平台后续阶段执行。
	Version string `yaml:"version" json:"version"`
	// Type 应用类型 app/squad/agent/skill（怎么选见 apptypes.go AppType 常量注释）。
	Type        AppType         `yaml:"type" json:"type"`
	Summary     LocalizedString `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description LocalizedString `yaml:"description,omitempty" json:"description,omitempty"`
	Publisher   string          `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	Category    string          `yaml:"category,omitempty" json:"category,omitempty"`
	Tags        LocalizedTags   `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Protection 卸载保护级别（归属平台：第三方 manifest 声明无效、按 none 处理，见 ProtectionLevel）。
	Protection ProtectionLevel `yaml:"protection,omitempty" json:"protection,omitempty"`
	// Compatibility 声明对 keystone 平台的兼容性约束（见 CompatibilitySpec）。
	Compatibility CompatibilitySpec `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Pricing       PricingSpec       `yaml:"pricing,omitempty" json:"pricing,omitempty"`
	Runtime       RuntimeSpec       `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Store         StoreSpec         `yaml:"store,omitempty" json:"store,omitempty"`
	Mount         MountSpec         `yaml:"mount,omitempty" json:"mount,omitempty"`
	// Auth app 级鉴权契约（顶层 auth 段）。详见 AuthSpec。
	Auth AuthSpec `yaml:"auth,omitempty" json:"auth,omitempty"`
	// Provides 声明本应用对外暴露的 capability 集合（v0.19.0 capability mesh）。
	// 取代旧 mount.service.auto_register_mcp / mount.extension 的能力暴露语义。
	Provides ProvidesSpec `yaml:"provides,omitempty" json:"provides,omitempty"`
	// Requires 声明本应用调用其他应用 capability 的依赖（v0.19.0 capability mesh）。
	// 取代旧 dependencies.requires / dependencies.recommends 的应用级依赖语义。
	Requires RequiresSpec `yaml:"requires,omitempty" json:"requires,omitempty"`
	// Conflicts 声明应用级互斥（同时安装会冲突的应用），保留应用级语义。
	Conflicts ConflictsSpec `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// ManagedResources 声明由 Keystone 在安装时统一分配和注入的托管资源。
	// 应用继续拥有自己的 schema/migration/业务数据语义，平台只负责基础资源
	// 生命周期、隔离、台账与运行时注入。
	ManagedResources ManagedResourcesSpec `yaml:"managed_resources,omitempty" json:"managed_resources,omitempty"`
	ManagedSecrets   ManagedSecretsSpec   `yaml:"managed_secrets,omitempty" json:"managed_secrets,omitempty"`
	PlatformServices PlatformServicesSpec `yaml:"platform_services,omitempty" json:"platform_services,omitempty"`
	Readiness        ReadinessSpec        `yaml:"readiness,omitempty" json:"readiness,omitempty"`
	// StandaloneFallback 声明非 keystone 托管模式下的本地资源 fallback 配置。
	// 与 ManagedResources 一一对应：managed 模式下忽略此段，standalone 模式下作为
	// 唯一权威源。
	StandaloneFallback *StandaloneFallbackSpec   `yaml:"standalone_fallback,omitempty" json:"standalone_fallback,omitempty"`
	Permissions        map[string]PermissionDecl `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	// Changelog 当前版本的 changelog markdown，由 publish 流程从 CHANGELOG.md 抽取
	// 或开发者直接在 manifest 内填写。Store 详情页直接渲染。
	Changelog string `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	// Author 应用作者标识（人或组织名）。Publisher 是发布主体（由平台分配），
	// Author 可独立标注开发者署名。
	Author string `yaml:"author,omitempty" json:"author,omitempty"`
	// Icon 图标资源相对路径（相对 manifest.yaml 所在目录），如 "icon.svg"。
	// publish 时随 tarball 上传，平台登记图标资源，store 详情页据此
	// 渲染图标 URL。yaml/json tag 与历史 manifest.yaml 字段名兼容。
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// TaskTemplates 开箱即用任务模板清单。agent 类应用安装后，keystone 按此清单在平台侧
	// 回填任务模板，用户进入 agent 详情页即可看到现成任务卡片。
	// 详见 task_template.go。
	TaskTemplates []TaskTemplate `yaml:"task_templates,omitempty" json:"task_templates,omitempty"`
}

// AuthSpec app 级鉴权契约（取代 v0.19.0 删掉的 mount.service.auth_mode）。
//
// 是什么：app 入站 /mcp 端点与 keystone 出站调用之间的鉴权约定——app 级安全契约
// （非 per-capability、非 run-config）。keystone 出站据此决定是否签注短期 JWT，
// app 入站用 SDK 校验，双边。必须在 manifest 声明：这是 a-priori 握手，
// keystone 调 app 前就要知道，不能只靠 /meta。
//
// 何时填：app / squad（有入站端点）按需显式声明；agent / skill（runtime.mode=none、
// 无入站端点）省略即可。
//
// 默认：留空时由 EffectiveMode 按 secure-by-default 派生——有入站端点 → keystone_jwks。
// 最佳实践：别关。keystone_jwks 是推荐项；none 仅本地逃生口（须配 KS_APP_AUTH_MODE=insecure），
// 绝不在生产对外端点显式 none。
type AuthSpec struct {
	// Mode 鉴权模式；为空时按 EffectiveMode 派生（secure-by-default）。
	// 枚举 none | keystone_jwks | static_bearer，见 apptypes.go AuthMode。
	Mode AuthMode `yaml:"mode,omitempty" json:"mode,omitempty"`
}

// EffectiveMode 解析有效鉴权模式（secure-by-default）：
//   - 显式声明 → 尊重作者（含 none 的显式 opt-out）；
//   - 未声明 + 有入站端点（container / process）→ keystone_jwks；
//   - 未声明 + 无入站端点（none：agent / skill）→ none。
func (a AuthSpec) EffectiveMode(runtimeMode RuntimeMode) AuthMode {
	if a.Mode != "" {
		return a.Mode
	}
	if runtimeMode == RuntimeModeNone || runtimeMode == "" {
		return AuthModeNone
	}
	return AuthModeKeystoneJWKS
}

// CompatibilitySpec 兼容性约束
type CompatibilitySpec struct {
	// Keystone 要求的 keystone 版本区间，写法：单约束 ">=1.5.0"、双约束
	// ">=1.5.0 <2.0.0"、或裸版本 "1.5.0"（约束符 >=|<=|>|<|= 可选）。何时填：依赖
	// 特定平台能力时；留空=不约束。enforcement：解析期仅校验区间格式，安装期 install
	// gate（拒装不兼容版本）由平台在后续阶段执行。
	Keystone string `yaml:"keystone,omitempty" json:"keystone,omitempty"`
}

// PricingSpec 定价信息
//
// v0.9.0 起 Description 升级为 LocalizedString，与 AppSpec 同构。
type PricingSpec struct {
	Type        PricingType     `yaml:"type,omitempty" json:"type,omitempty"`
	Description LocalizedString `yaml:"description,omitempty" json:"description,omitempty"`
}

// RuntimeSpec 运行时配置
type RuntimeSpec struct {
	Mode       RuntimeMode `yaml:"mode,omitempty" json:"mode,omitempty"`
	Entry      string      `yaml:"entry,omitempty" json:"entry,omitempty"`
	WorkingDir string      `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Image      string      `yaml:"image,omitempty" json:"image,omitempty"`
	// WritableRootFS 显式 opt-in 容器可写 rootfs（能在容器内 apt/pip 装软件、写系统目录）。
	// 零值 false = 只读 rootfs（安全默认，所有未声明的 container app 保持只读）。
	// keystone buildAgentAppSpec 据此填 agentclient.AppSpec.ReadOnlyRootFS = !WritableRootFS。
	WritableRootFS bool `yaml:"writable_root_fs,omitempty" json:"writable_root_fs,omitempty"`
	// 端口：keystone 拉起约定容器内监听 8080，manifest 不声明 port/ports。
	Volumes        []string      `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	HealthCheck    string        `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	HealthCheckURL string        `yaml:"health_check_url,omitempty" json:"health_check_url,omitempty"`
	Environment    []string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	Resources      ResourcesSpec `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// StoreSpec 声明应用在 Store 中的展示材料。
//
// 字段只表达商品呈现，不承载执行、权限、健康状态等运行事实。
type StoreSpec struct {
	Presentation StorePresentation `yaml:"presentation,omitempty" json:"presentation,omitempty"`
	Audience     []string          `yaml:"audience,omitempty" json:"audience,omitempty"`
	Highlights   []string          `yaml:"highlights,omitempty" json:"highlights,omitempty"`
	TryPrompts   []string          `yaml:"try_prompts,omitempty" json:"try_prompts,omitempty"`
	Badges       []string          `yaml:"badges,omitempty" json:"badges,omitempty"`
	Team         *StoreTeamSpec    `yaml:"team,omitempty" json:"team,omitempty"`
	Media        []StoreMediaSpec  `yaml:"media,omitempty" json:"media,omitempty"`
}

// StoreTeamSpec 描述专家团等多角色展示形态。
type StoreTeamSpec struct {
	LeadRole string                `yaml:"lead_role,omitempty" json:"lead_role,omitempty"`
	Members  []StoreTeamMemberSpec `yaml:"members,omitempty" json:"members,omitempty"`
}

// StoreTeamMemberSpec 描述 Store 中展示的单个虚拟团队成员。
type StoreTeamMemberSpec struct {
	Key    string `yaml:"key" json:"key"`
	Name   string `yaml:"name" json:"name"`
	Title  string `yaml:"title,omitempty" json:"title,omitempty"`
	Avatar string `yaml:"avatar,omitempty" json:"avatar,omitempty"`
}

// StoreMediaSpec 描述 Store 展示媒体资源。
type StoreMediaSpec struct {
	Path    string `yaml:"path" json:"path"`
	Role    string `yaml:"role,omitempty" json:"role,omitempty"`
	Caption string `yaml:"caption,omitempty" json:"caption,omitempty"`
}

// ResourcesSpec 容器资源限制。
//
// 格式：K8s 风格字符串——cpu 用核数 "0.5" / "2" 或毫核 "500m"；memory 用带单位的量
// "512Mi" / "1Gi"（单位 Ki|Mi|Gi|K|M|G）。何时填：多租户下需给容器设资源上限时；
// 留空=用平台默认。enforcement：解析期仅校验格式；解析 K8s 字符串并在拉起容器时应用
// 由平台在后续阶段执行。
type ResourcesSpec struct {
	// CPU 核数 "0.5" / "2" 或毫核 "500m"。
	CPU string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	// Memory 内存量 "512Mi" / "1Gi"（单位 Ki|Mi|Gi|K|M|G 必填）。
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}

// resources 格式校验正则。
var (
	// cpuRegex 接受核数 "0.5" / "2" 或毫核 "500m"。
	cpuRegex = regexp.MustCompile(`^(\d+(\.\d+)?|\d+m)$`)
	// memRegex 接受带单位的 K8s 内存量 "512Mi" / "1Gi"。
	memRegex = regexp.MustCompile(`^\d+(Ki|Mi|Gi|K|M|G)$`)
)

// Validate 校验容器资源限制格式（解析期；解析并在拉起时应用由平台后续阶段执行）。
func (r ResourcesSpec) Validate() error {
	if r.CPU != "" && !cpuRegex.MatchString(r.CPU) {
		return fmt.Errorf("runtime.resources.cpu %q 非法（要求核数 \"0.5\"/\"2\" 或毫核 \"500m\"）", r.CPU)
	}
	if r.Memory != "" && !memRegex.MatchString(r.Memory) {
		return fmt.Errorf("runtime.resources.memory %q 非法（要求带单位，如 \"512Mi\"/\"1Gi\"）", r.Memory)
	}
	return nil
}

// ManagedResourcesSpec 声明应用安装时需要 Keystone 托管的基础资源。
type ManagedResourcesSpec struct {
	MySQL         *ManagedMySQLResourceSpec         `yaml:"mysql,omitempty" json:"mysql,omitempty"`
	ObjectStorage *ManagedObjectStorageResourceSpec `yaml:"object_storage,omitempty" json:"object_storage,omitempty"`
	VectorStore   *ManagedVectorStoreResourceSpec   `yaml:"vector_store,omitempty" json:"vector_store,omitempty"`
	Storage       *ManagedStorageResourceSpec       `yaml:"storage,omitempty" json:"storage,omitempty"`
	Cache         *ManagedCacheResourceSpec         `yaml:"cache,omitempty" json:"cache,omitempty"`
}

// ManagedMySQLResourceSpec 声明一个由 Keystone 分配的 MySQL database + user。
//
// razor 瘦身：已砍 required（块存在即=需要，must-be-true 零信息）、database/user
// （只接受 auto、平台无条件自动命名）等强制值样板。作者只声明 inject env 名映射 + 保留策略。
type ManagedMySQLResourceSpec struct {
	// RetainOnUninstall 卸载时是否保留底层 database（默认 false=随卸载回收）。
	RetainOnUninstall bool `yaml:"retain_on_uninstall,omitempty" json:"retain_on_uninstall,omitempty"`
	// Inject env 名映射：平台据此把分配好的连接信息注入容器（真载荷，必填）。
	Inject ManagedMySQLInjectSpec `yaml:"inject,omitempty" json:"inject,omitempty"`
}

// ManagedMySQLInjectSpec 声明 MySQL 连接信息注入到应用运行时的 env key。
type ManagedMySQLInjectSpec struct {
	Host     string `yaml:"host,omitempty" json:"host,omitempty"`
	Port     string `yaml:"port,omitempty" json:"port,omitempty"`
	Database string `yaml:"database,omitempty" json:"database,omitempty"`
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// ManagedObjectStorageResourceSpec 声明一个由 Keystone 分配的对象存储命名空间。
//
// razor 瘦身：已砍 required、bucket/prefix（只接受 auto、平台自动命名）。保留 access
// （private/public 可见性，真语义）+ inject env 名映射 + 保留策略。
type ManagedObjectStorageResourceSpec struct {
	// Access 对象可见性：空（默认 private）| private | public。
	Access string `yaml:"access,omitempty" json:"access,omitempty"`
	// RetainOnUninstall 卸载时是否保留底层 bucket 内容。
	RetainOnUninstall bool `yaml:"retain_on_uninstall,omitempty" json:"retain_on_uninstall,omitempty"`
	// Inject env 名映射：平台据此把对象存储连接信息注入容器（真载荷，必填）。
	Inject ManagedObjectStorageInjectSpec `yaml:"inject,omitempty" json:"inject,omitempty"`
}

// ManagedObjectStorageInjectSpec 声明对象存储连接信息注入到应用运行时的 env key。
type ManagedObjectStorageInjectSpec struct {
	Endpoint      string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Bucket        string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	Prefix        string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Region        string `yaml:"region,omitempty" json:"region,omitempty"`
	AccessKey     string `yaml:"access_key,omitempty" json:"access_key,omitempty"`
	SecretKey     string `yaml:"secret_key,omitempty" json:"secret_key,omitempty"`
	PublicBaseURL string `yaml:"public_base_url,omitempty" json:"public_base_url,omitempty"`
}

// ManagedVectorStoreResourceSpec 声明一个由 Keystone 分配的向量存储命名空间。
// 平台保证 provider / dim / collection 命名 / 鉴权全部自动注入。
//
// 先例：v0.22.0 起本就是 razor 瘦身样本（连 inject 都由平台决定）；本批顺手砍 required。
type ManagedVectorStoreResourceSpec struct {
	// RetainOnUninstall 卸载时是否保留底层向量集合。
	RetainOnUninstall bool `yaml:"retain_on_uninstall,omitempty" json:"retain_on_uninstall,omitempty"`
}

// ManagedStorageResourceSpec 声明 Keystone 管理的应用私有文件系统 scope。
// Keystone 负责决定宿主机路径、容器内路径和标准 env 注入，应用只声明需要哪些语义目录。
//
// razor 瘦身：已砍 required（声明任一 scope 即=需要）；作者按需声明 files/cache/tmp/logs/config
// 五类语义目录中的至少一个。
type ManagedStorageResourceSpec struct {
	Files  *ManagedStorageScopeSpec `yaml:"files,omitempty" json:"files,omitempty"`
	Cache  *ManagedStorageScopeSpec `yaml:"cache,omitempty" json:"cache,omitempty"`
	Tmp    *ManagedStorageScopeSpec `yaml:"tmp,omitempty" json:"tmp,omitempty"`
	Logs   *ManagedStorageScopeSpec `yaml:"logs,omitempty" json:"logs,omitempty"`
	Config *ManagedStorageScopeSpec `yaml:"config,omitempty" json:"config,omitempty"`
}

// ManagedStorageScopeSpec 描述单个应用私有目录的配额和保留策略。
type ManagedStorageScopeSpec struct {
	SizeMB            int  `yaml:"size_mb,omitempty" json:"size_mb,omitempty"`
	RetainOnUninstall bool `yaml:"retain_on_uninstall,omitempty" json:"retain_on_uninstall,omitempty"`
	ReadOnly          bool `yaml:"read_only,omitempty" json:"read_only,omitempty"`
}

// ManagedCacheResourceSpec 声明一个由 Keystone 分配的缓存 key namespace。
//
// razor 瘦身：已砍 required、provider（平台决定，当前 redis）、key_prefix（只接受 auto、
// 平台自动命名）。保留 inject env 名映射 + 保留策略。
type ManagedCacheResourceSpec struct {
	// RetainOnUninstall 卸载时是否保留缓存内容（一般 false）。
	RetainOnUninstall bool `yaml:"retain_on_uninstall,omitempty" json:"retain_on_uninstall,omitempty"`
	// Inject env 名映射：平台据此把缓存连接信息注入容器（真载荷，必填）。
	Inject ManagedCacheInjectSpec `yaml:"inject,omitempty" json:"inject,omitempty"`
}

// ManagedCacheInjectSpec 声明缓存连接信息注入到应用运行时的 env key。
type ManagedCacheInjectSpec struct {
	URL       string `yaml:"url,omitempty" json:"url,omitempty"`
	KeyPrefix string `yaml:"key_prefix,omitempty" json:"key_prefix,omitempty"`
}

// ManagedSecretsSpec 声明由 Keystone 生成、保存、轮换和注入的应用密钥。
type ManagedSecretsSpec struct {
	Items []ManagedSecretSpec `yaml:"items,omitempty" json:"items,omitempty"`
}

// ManagedSecretSpec 声明单个生成型密钥。
type ManagedSecretSpec struct {
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Generate string `yaml:"generate,omitempty" json:"generate,omitempty"`
	Inject   string `yaml:"inject,omitempty" json:"inject,omitempty"`
	Rotate   string `yaml:"rotate,omitempty" json:"rotate,omitempty"`
}

// PlatformServicesSpec 声明应用需要 Keystone 提供的共享平台能力入口。
type PlatformServicesSpec struct {
	Embedding *PlatformEmbeddingServiceSpec `yaml:"embedding,omitempty" json:"embedding,omitempty"`
}

// PlatformEmbeddingServiceSpec 声明统一 embedding 服务接入需求。
type PlatformEmbeddingServiceSpec struct {
	Required bool                        `yaml:"required,omitempty" json:"required,omitempty"`
	Model    string                      `yaml:"model" json:"model"`
	Inject   PlatformEmbeddingInjectSpec `yaml:"inject,omitempty" json:"inject,omitempty"`
}

// PlatformEmbeddingInjectSpec 声明 embedding 能力入口注入到应用运行时的 env key。
type PlatformEmbeddingInjectSpec struct {
	Model string `yaml:"model,omitempty" json:"model,omitempty"`
	Dim   string `yaml:"dim,omitempty" json:"dim,omitempty"`
}

// MountSpec 安装挂载配置（仅 agent / skill 两路；app / squad 走 provides.capabilities）。
//
// app / squad 类型对外暴露的能力由 AppSpec.Provides.Capabilities[] 显式声明，不走 mount。
// agent / skill 与 capability 是正交职责轴（keystone 内 agent 实例 / skill 资源），
// 经 mount.agent / mount.skill 在安装期挂载。
type MountSpec struct {
	Agent *AgentMountSpec `yaml:"agent,omitempty" json:"agent,omitempty"`
	Skill *SkillMountSpec `yaml:"skill,omitempty" json:"skill,omitempty"`
}

// SkillMountSpec skill 类型挂载
type SkillMountSpec struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Profile     *CapabilityProfileSpec `yaml:"profile,omitempty" json:"profile,omitempty"`
}

// AgentMountSpec agent 类型挂载（原 AssistantMountSpec 改名；去 routing_plan 死字段，
// 它从未驱动团队/路由，被数字型 RelayRoutingPlanID 取代）。
type AgentMountSpec struct {
	CreateAgent  bool                   `yaml:"create_agent,omitempty" json:"create_agent,omitempty"`
	Name         string                 `yaml:"name,omitempty" json:"name,omitempty"`
	SystemPrompt string                 `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Profile      *CapabilityProfileSpec `yaml:"profile,omitempty" json:"profile,omitempty"`
}

// ProvidesSpec 声明应用对外暴露的 capability 集合（v0.19.0 capability mesh）。
type ProvidesSpec struct {
	Capabilities []CapabilitySpec `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
}

// CapabilityProfileSpec 声明一个已挂载资源的自然语言召回 profile。
// 它只描述“用户会如何表达这个能力”，不承载权限、健康、执行参数或运行状态事实。
type CapabilityProfileSpec struct {
	CanonicalName string `yaml:"canonical_name,omitempty" json:"canonical_name,omitempty"`
	DisplayName   string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	// Description 一段语义真话——这个能力做什么、何时用（合并旧 intent_summary +
	// natural_description，一处声明别拆两段）。最佳实践：写一句真话别堆砌；aliases /
	// user_utterances / use_cases / domain_terms 由 LLM 从 description + schema 生成草稿、
	// 作者审核，不手凑。
	Description      string   `yaml:"description,omitempty" json:"description,omitempty"`
	Aliases          []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	UserUtterances   []string `yaml:"user_utterances,omitempty" json:"user_utterances,omitempty"`
	UseCases         []string `yaml:"use_cases,omitempty" json:"use_cases,omitempty"`
	DomainTerms      []string `yaml:"domain_terms,omitempty" json:"domain_terms,omitempty"`
	NegativeExamples []string `yaml:"negative_examples,omitempty" json:"negative_examples,omitempty"`
}

// CapabilitySpec 单个 capability 的完整声明（manifest 写入侧形态，
// install 时由 keystone 注册期解析并在平台侧入库）。
//
// 字段集对应平台侧的 capability mesh 数据模型（clean-break 后已精简）；
// 自然语言段（description / aliases / user_utterances 等）由 dispatcher
// 语义命中阶段消费。
type CapabilitySpec struct {
	// Name 能力裸名（app_provided 必填，单段无点，正则 ^[a-z][a-z0-9_-]*$）。
	// 作者只写裸名（如 web_search）；全局 canonical_name 由 keystone 注册期派生
	// <app_id>.<name>，作者不写、不能写前缀（app_id 已在顶层声明）。
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// CanonicalName 全局唯一能力名。app_provided 由注册期从 Name 派生（作者不写）；
	// 内置 YAML 维持现状（写全名，如 query.list_my_workflows）。
	CanonicalName string `yaml:"canonical_name,omitempty" json:"canonical_name,omitempty"`
	// Entry 标记该能力是否为 squad 专家团对编排官的「入口动作」。
	// 仅 type=squad 应用有意义：keystone 安装时据此派生团队锚点并标记成员；
	// 普通 app/agent/skill 留默认 false，行为不变。
	Entry bool `yaml:"entry,omitempty" json:"entry,omitempty"`
	// DisplayName UI 展示名（多语言通过 LocalizedString 升级留后续）。
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	// ExecutionMode 执行模式枚举：sync | long_running | hybrid。
	ExecutionMode string `yaml:"execution_mode" json:"execution_mode"`
	// Backend 后端路由声明：dispatcher 据此决定调用通道。
	Backend BackendSpec `yaml:"backend" json:"backend"`
	// TimeoutMs 单次调用超时（毫秒），值域 100..600000；0/留空=用平台默认。
	// 何时填：能力执行时间显著偏离默认时显式声明。
	TimeoutMs int `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty"`
	// ConcurrencyLimit 并发上限（0 表示不限）。
	ConcurrencyLimit int `yaml:"concurrency_limit,omitempty" json:"concurrency_limit,omitempty"`
	// SideEffectLevel 副作用/可逆性档：none | soft_reversible | hard_irreversible。
	SideEffectLevel string `yaml:"side_effect_level,omitempty" json:"side_effect_level,omitempty"`
	// DecisionMode 决策/审批策略（取代被砍的 requires_approval + app 级 compliance）。
	// 派生字段：为空时由 EffectiveDecisionMode 从 side_effect_level 派生
	// （none→agent_autonomous / soft_reversible→user_authorized / hard_irreversible→user_only），
	// 仅在派生失效处（如敏感读：side_effect=none 却要 user_only）显式覆盖。枚举见 compliance.go。
	DecisionMode DecisionMode `yaml:"decision_mode,omitempty" json:"decision_mode,omitempty"`
	// PreAuthorizeDuration 预授权时长；仅 decision_mode=user_authorized 时有意义。枚举见 compliance.go。
	PreAuthorizeDuration PreAuthorizeDuration `yaml:"pre_authorize_duration,omitempty" json:"pre_authorize_duration,omitempty"`
	// Resumable 是否支持续跑（long_running task 重启后恢复）。
	Resumable bool `yaml:"resumable,omitempty" json:"resumable,omitempty"`
	// GuardrailProfile 内容/安全 guardrail profile 名（如 content_creation）。
	GuardrailProfile string `yaml:"guardrail_profile,omitempty" json:"guardrail_profile,omitempty"`
	// Tags 自由标签集。
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// NotifyPolicy 调用生命周期通知策略（指针区分未设置 vs false）。
	NotifyPolicy NotifyPolicySpec `yaml:"notify_policy,omitempty" json:"notify_policy,omitempty"`
	// CompletionMessageTemplate 完成时回写消息模板（占位变量待 dispatcher 渲染）。
	CompletionMessageTemplate string `yaml:"completion_message_template,omitempty" json:"completion_message_template,omitempty"`
	// InputSchema 内联输入 JSON Schema（manifest 直接声明，唯一来源；ref 形态已砍）。
	// 主要用途：SDK 透传到 MCP ToolDef.InputSchema，让编排官 tools/list 拿到完整参数 schema。
	InputSchema map[string]any `yaml:"input_schema,omitempty" json:"input_schema,omitempty"`
	// OutputSchema 内联输出 JSON Schema（与 InputSchema 对称；取代被砍的 output_schema_ref）。
	// 何时填：能力有结构化输出即填——SDK 透传给编排官，便于其规划下游消费。
	OutputSchema map[string]any `yaml:"output_schema,omitempty" json:"output_schema,omitempty"`
	// Description 一段语义真话——这个能力做什么、何时用（合并旧 intent_summary +
	// natural_description，一处声明别拆两段）。最佳实践：写一句真话别堆砌；aliases /
	// user_utterances / use_cases / domain_terms 由 LLM 从 description + schema 生成草稿、
	// 作者审核，不手凑。
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Aliases 别名集（同一能力的口语化叫法）。
	Aliases []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	// UserUtterances 真实用户语料示例。
	UserUtterances []string `yaml:"user_utterances,omitempty" json:"user_utterances,omitempty"`
	// UseCases 典型使用场景。
	UseCases []string `yaml:"use_cases,omitempty" json:"use_cases,omitempty"`
	// DomainTerms 领域术语集（让 LLM 识别上下文专有名词）。
	DomainTerms []string `yaml:"domain_terms,omitempty" json:"domain_terms,omitempty"`
	// NegativeExamples 反例集（让 LLM 学会拒绝相似但应路由到他处的请求）。
	NegativeExamples []string `yaml:"negative_examples,omitempty" json:"negative_examples,omitempty"`
}

// EffectiveDecisionMode 未显式声明 decision_mode 时从 side_effect_level 派生：
//   - hard_irreversible → user_only；
//   - soft_reversible   → user_authorized；
//   - none / 未声明     → agent_autonomous。
//
// 作者仅在派生失效处（如敏感读）显式覆盖。
func (c CapabilitySpec) EffectiveDecisionMode() DecisionMode {
	if c.DecisionMode != "" {
		return c.DecisionMode
	}
	switch c.SideEffectLevel {
	case "hard_irreversible":
		return DecisionModeUserOnly
	case "soft_reversible":
		return DecisionModeUserAuthorized
	default: // none / 未声明
		return DecisionModeAgentAutonomous
	}
}

// BackendSpec 描述 capability 的后端路由方式。
//
// kind 枚举：
//   - mcp_tool：dispatcher 通过应用内部 MCP server 暴露的 tool 调用，需声明 tool_name。
//   - http_endpoint：dispatcher 直接 HTTP 调应用容器，需声明 path + method。
type BackendSpec struct {
	Kind string `yaml:"kind" json:"kind"`
	// ToolName mcp_tool kind 必填，对应 MCP server 暴露的 tool 名。
	ToolName string `yaml:"tool_name,omitempty" json:"tool_name,omitempty"`
	// Path http_endpoint kind 必填，应用容器内的 HTTP 路径。
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Method http_endpoint kind 必填，HTTP 方法（GET/POST/...）。
	Method string `yaml:"method,omitempty" json:"method,omitempty"`
}

// NotifyPolicySpec 调用生命周期通知策略（保留 on_done / on_failed 两态）。
//
// 指针类型区分「未设置」与「显式 false」：
//   - nil：consumer 默认行为（execution_mode=sync 不发通知；long_running 发 done/failed）
//   - *false：显式禁用该阶段通知
//   - *true：显式启用
//
// on_started 已砍除：keystone 侧无消费、且 sync 模式禁用，属伪声明。
type NotifyPolicySpec struct {
	// OnDone 任务完成通知（long_running 闭环已接线）。
	OnDone *bool `yaml:"on_done,omitempty" json:"on_done,omitempty"`
	// OnFailed 任务失败通知。
	OnFailed *bool `yaml:"on_failed,omitempty" json:"on_failed,omitempty"`
}

// RequiresSpec 声明应用调用其他应用 capability 的依赖（v0.19.0 capability mesh）。
type RequiresSpec struct {
	Capabilities []RequiresCapabilityItem `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
}

// RequiresCapabilityItem 单条 capability 依赖项。
type RequiresCapabilityItem struct {
	// CanonicalName 被依赖 capability 的全局名（必须存在于 Registry）。
	CanonicalName string `yaml:"canonical_name" json:"canonical_name"`
	// Mode 依赖强度：required | recommended | optional。
	//
	// 三档语义：
	//   - required：install 时缺则阻断；运行时缺返 CapabilityNotFound
	//   - recommended：install 时缺则 warning + UI 推荐；运行时缺返 CapabilityNotFound
	//   - optional：install 时不检查；运行时缺由调用方自行处理
	Mode string `yaml:"mode" json:"mode"`
	// Reason 依赖原因（便于 UI 展示与 install warning 文案）。
	Reason string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// ConflictsSpec 应用级互斥声明（同时安装会冲突的应用集合）。
type ConflictsSpec struct {
	Apps []ConflictsAppItem `yaml:"apps,omitempty" json:"apps,omitempty"`
}

// ConflictsAppItem 单条应用级冲突项。
type ConflictsAppItem struct {
	ID     string `yaml:"id" json:"id"`
	Reason string `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// ParseAppSpec 从 YAML 字节解析 AppSpec
func ParseAppSpec(data []byte) (*AppSpec, error) {
	var m AppSpec
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("解析 manifest YAML 失败: %w", err)
	}
	return &m, nil
}

// Validate 校验 AppSpec 的必填字段和枚举值
func (m *AppSpec) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("manifest: id is required")
	}
	if m.Name == "" {
		return fmt.Errorf("manifest: name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest: version is required")
	}
	// id / version / compatibility 格式校验（格式在解析期，install gate/升级单调由平台后续阶段执行）。
	if !appIDRegex.MatchString(m.ID) {
		return fmt.Errorf("manifest: id %q 非法（要求 ^[a-z][a-z0-9-]*$：小写字母开头，仅含小写字母/数字/连字符）", m.ID)
	}
	if !semverRegex.MatchString(m.Version) {
		return fmt.Errorf("manifest: version %q 非 semver（要求三段 x.y.z，如 1.2.0）", m.Version)
	}
	if m.Compatibility.Keystone != "" && !compatRangeRegex.MatchString(m.Compatibility.Keystone) {
		return fmt.Errorf("manifest: compatibility.keystone %q 非合法区间（如 \">=1.5.0\" 或 \">=1.5.0 <2.0.0\"）", m.Compatibility.Keystone)
	}
	if !m.Type.Valid() {
		return fmt.Errorf("manifest: invalid type %q", m.Type)
	}
	if m.Pricing.Type != "" && !m.Pricing.Type.Valid() {
		return fmt.Errorf("manifest: invalid pricing type %q", m.Pricing.Type)
	}

	// RuntimeMode 校验
	if !m.Runtime.Mode.Valid() {
		return fmt.Errorf("manifest: invalid runtime mode %q", m.Runtime.Mode)
	}

	// runtime.resources 格式/边界校验（解析期；拉起时应用由平台后续阶段执行）
	if err := m.Runtime.Resources.Validate(); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}

	// Protection 校验
	if !m.Protection.Valid() {
		return fmt.Errorf("manifest: invalid protection %q", m.Protection)
	}

	// auth.mode 校验
	if !m.Auth.Mode.Valid() {
		return fmt.Errorf("manifest: invalid auth.mode %q (允许 none|keystone_jwks|static_bearer)", m.Auth.Mode)
	}

	if m.Store.Presentation != "" && !m.Store.Presentation.Valid() {
		return fmt.Errorf("manifest: store.presentation %q 不合法", m.Store.Presentation)
	}
	if m.Store.Presentation == StorePresentationExpertTeam {
		if m.Store.Team == nil || len(m.Store.Team.Members) == 0 {
			return fmt.Errorf("manifest: store.presentation=expert_team 时 store.team.members 至少需要 1 个成员")
		}
		for i, member := range m.Store.Team.Members {
			if strings.TrimSpace(member.Key) == "" || strings.TrimSpace(member.Name) == "" {
				return fmt.Errorf("manifest: store.team.members[%d].key/name 为必填", i)
			}
		}
	}

	// skill 类型不能有运行时进程
	if m.Type == AppTypeSkill {
		if m.Runtime.Mode != "" && m.Runtime.Mode != RuntimeModeNone {
			return fmt.Errorf("manifest: type=skill 时 runtime.mode 必须为空或 none")
		}
	}
	if m.Mount.Agent != nil && m.Mount.Agent.Profile != nil {
		if err := m.Mount.Agent.Profile.Validate("mount.agent.profile"); err != nil {
			return fmt.Errorf("manifest: %w", err)
		}
	}
	if m.Mount.Skill != nil && m.Mount.Skill.Profile != nil {
		if err := m.Mount.Skill.Profile.Validate("mount.skill.profile"); err != nil {
			return fmt.Errorf("manifest: %w", err)
		}
	}

	// capability mesh：provides / requires / conflicts 校验
	if err := m.Provides.Validate(m.ID, m.Runtime.Mode); err != nil {
		return fmt.Errorf("manifest: provides: %w", err)
	}
	if err := m.Requires.Validate(); err != nil {
		return fmt.Errorf("manifest: requires: %w", err)
	}
	if err := m.Conflicts.Validate(); err != nil {
		return fmt.Errorf("manifest: conflicts: %w", err)
	}

	if err := m.ManagedResources.Validate(); err != nil {
		return fmt.Errorf("manifest: managed_resources: %w", err)
	}
	if err := m.ManagedSecrets.Validate(); err != nil {
		return fmt.Errorf("manifest: managed_secrets: %w", err)
	}
	if err := m.PlatformServices.Validate(); err != nil {
		return fmt.Errorf("manifest: platform_services: %w", err)
	}
	if err := m.Readiness.Validate(); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}

	// task_templates 段校验
	for i, t := range m.TaskTemplates {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("manifest: task_templates[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate 校验托管资源声明。
func (r ManagedResourcesSpec) Validate() error {
	if r.MySQL != nil {
		if err := r.MySQL.Validate(); err != nil {
			return err
		}
	}
	if r.ObjectStorage != nil {
		if err := r.ObjectStorage.Validate(); err != nil {
			return err
		}
	}
	if r.VectorStore != nil {
		if err := r.VectorStore.Validate(); err != nil {
			return err
		}
	}
	if r.Storage != nil {
		if err := r.Storage.Validate(); err != nil {
			return err
		}
	}
	if r.Cache != nil {
		if err := r.Cache.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate 校验 MySQL 托管资源声明（razor 瘦身后仅校验 inject env 名映射必填）。
func (r ManagedMySQLResourceSpec) Validate() error {
	if r.Inject.Host == "" || r.Inject.Port == "" || r.Inject.Database == "" ||
		r.Inject.User == "" || r.Inject.Password == "" {
		return fmt.Errorf("mysql.inject.{host,port,database,user,password} 均为必填")
	}
	return nil
}

// Validate 校验对象存储托管资源声明（razor 瘦身后保留 access 枚举 + inject 必填）。
func (r ManagedObjectStorageResourceSpec) Validate() error {
	switch r.Access {
	case "", "private", "public":
	default:
		return fmt.Errorf("object_storage.access 仅支持 private 或 public")
	}
	if r.Inject.Endpoint == "" || r.Inject.Bucket == "" || r.Inject.Prefix == "" {
		return fmt.Errorf("object_storage.inject.{endpoint,bucket,prefix} 均为必填")
	}
	return nil
}

// Validate 校验向量存储托管资源声明。
func (r ManagedVectorStoreResourceSpec) Validate() error {
	// v0.22.0 起 dim/provider/collection_prefix/inject 全部由平台决定，无字段需要校验。
	return nil
}

// Validate 校验应用私有存储声明（razor 瘦身后仅校验 ≥1 scope + size_mb>=0）。
func (r ManagedStorageResourceSpec) Validate() error {
	scopes := map[string]*ManagedStorageScopeSpec{
		"files":  r.Files,
		"cache":  r.Cache,
		"tmp":    r.Tmp,
		"logs":   r.Logs,
		"config": r.Config,
	}
	declared := 0
	for name, scope := range scopes {
		if scope == nil {
			continue
		}
		declared++
		if scope.SizeMB < 0 {
			return fmt.Errorf("storage.%s.size_mb 必须 >= 0", name)
		}
	}
	if declared == 0 {
		return fmt.Errorf("storage 至少需要声明 files/cache/tmp/logs/config 中的一个 scope")
	}
	return nil
}

// Validate 校验缓存托管资源声明（razor 瘦身后仅校验 inject env 名映射必填）。
func (r ManagedCacheResourceSpec) Validate() error {
	if r.Inject.URL == "" || r.Inject.KeyPrefix == "" {
		return fmt.Errorf("cache.inject.{url,key_prefix} 均为必填")
	}
	return nil
}

// Validate 校验托管密钥声明。
func (s ManagedSecretsSpec) Validate() error {
	seen := map[string]struct{}{}
	for i, item := range s.Items {
		if item.Name == "" {
			return fmt.Errorf("items[%d].name 为必填项", i)
		}
		if _, ok := seen[item.Name]; ok {
			return fmt.Errorf("items[%d].name %q 重复", i, item.Name)
		}
		seen[item.Name] = struct{}{}
		switch item.Generate {
		case "random_hex_32", "random_base64_32":
		default:
			return fmt.Errorf("items[%d].generate 当前仅支持 random_hex_32 或 random_base64_32", i)
		}
		if item.Inject == "" {
			return fmt.Errorf("items[%d].inject 为必填项", i)
		}
		switch item.Rotate {
		case "", "manual":
		default:
			return fmt.Errorf("items[%d].rotate 当前仅支持 manual", i)
		}
	}
	return nil
}

// Validate 校验平台服务声明。
func (s PlatformServicesSpec) Validate() error {
	if s.Embedding == nil {
		return nil
	}
	return s.Embedding.Validate()
}

// Validate 校验 embedding 平台服务声明。
func (s PlatformEmbeddingServiceSpec) Validate() error {
	if !s.Required {
		// 应用未要求 embedding 平台服务，不强校验。
		return nil
	}
	if strings.TrimSpace(s.Model) == "" {
		return fmt.Errorf("platform_services.embedding.model 不能为空（v0.22.0 起 model 必填）")
	}
	if strings.TrimSpace(s.Inject.Model) == "" {
		return fmt.Errorf("platform_services.embedding.inject.model 不能为空")
	}
	if strings.TrimSpace(s.Inject.Dim) == "" {
		return fmt.Errorf("platform_services.embedding.inject.dim 不能为空")
	}
	return nil
}

// Validate 校验 CapabilityProfileSpec 的稳定标识字段。
func (p *CapabilityProfileSpec) Validate(path string) error {
	if p == nil {
		return nil
	}
	if strings.TrimSpace(p.CanonicalName) != "" && !canonicalNameRegex.MatchString(p.CanonicalName) {
		return fmt.Errorf("%s.canonical_name %q 不符合命名规范 (要求 ^[a-z][a-z0-9_-]*(\\.[a-z][a-z0-9_-]*)+$)", path, p.CanonicalName)
	}
	return nil
}

// Validate 校验 ProvidesSpec：裸名 name 正则（单段无点）/ 拒写全名 canonical_name /
// 同 manifest 内 name 不重复 / backend.kind 与 runtime.mode 一致性 / execution_mode 枚举 /
// side_effect_level 枚举。canonical_name 由 keystone 注册期从 name 派生。
//
// 跨应用 canonical_name 全局唯一性留给 keystone 注册期校验（依赖平台侧已注册能力清单）。
func (p ProvidesSpec) Validate(appID string, runtimeMode RuntimeMode) error {
	if len(p.Capabilities) == 0 {
		return nil
	}
	seenNames := make(map[string]struct{}, len(p.Capabilities))
	for i, cap := range p.Capabilities {
		// 去前缀：app_provided 作者写裸名 name，不写 canonical_name；
		// 全局 canonical_name 由 keystone 注册期派生 <app_id>.<name>。
		if cap.CanonicalName != "" {
			return fmt.Errorf("capabilities[%d]: app 能力请用裸名 name（如 web_search），不要写全名 canonical_name %q（前缀由注册期派生）", i, cap.CanonicalName)
		}
		if cap.Name == "" {
			return fmt.Errorf("capabilities[%d].name 为必填项（裸名，单段无点）", i)
		}
		if !bareNameRegex.MatchString(cap.Name) {
			return fmt.Errorf("capabilities[%d].name %q 不符合裸名规范 (要求 ^[a-z][a-z0-9_-]*$，单段无点)", i, cap.Name)
		}
		if _, dup := seenNames[cap.Name]; dup {
			return fmt.Errorf("capabilities[%d].name %q 在本 manifest 内重复", i, cap.Name)
		}
		seenNames[cap.Name] = struct{}{}

		// execution_mode 枚举
		if cap.ExecutionMode == "" {
			return fmt.Errorf("capabilities[%d].execution_mode 为必填项", i)
		}
		if !validExecutionModes[cap.ExecutionMode] {
			return fmt.Errorf("capabilities[%d].execution_mode %q 不合法 (允许 sync|long_running|hybrid)", i, cap.ExecutionMode)
		}

		// backend.kind 枚举与 runtime.mode 一致性
		if err := cap.Backend.Validate(runtimeMode); err != nil {
			return fmt.Errorf("capabilities[%d].backend: %w", i, err)
		}

		// side_effect_level 枚举（非空时校验；空值留给 keystone 入库时强制必填）
		if cap.SideEffectLevel != "" && !validSideEffectLevels[cap.SideEffectLevel] {
			return fmt.Errorf("capabilities[%d].side_effect_level %q 不合法 (允许 none|soft_reversible|hard_irreversible)", i, cap.SideEffectLevel)
		}

		// decision_mode 枚举（非空时校验）+ pre_authorize_duration 约束
		if cap.DecisionMode != "" && !cap.DecisionMode.IsValid() {
			return fmt.Errorf("capabilities[%d].decision_mode %q 不合法 (允许 user_only|user_authorized|agent_autonomous)", i, cap.DecisionMode)
		}
		if cap.PreAuthorizeDuration != "" {
			if !cap.PreAuthorizeDuration.IsValid() {
				return fmt.Errorf("capabilities[%d].pre_authorize_duration %q 不合法 (允许 5m|30m|2h|24h|forever)", i, cap.PreAuthorizeDuration)
			}
			if cap.DecisionMode != DecisionModeUserAuthorized {
				return fmt.Errorf("capabilities[%d].pre_authorize_duration 仅 decision_mode=user_authorized 时有效", i)
			}
		}

		// timeout_ms 边界校验（原无上下界，可写负数/超大）
		if cap.TimeoutMs != 0 && (cap.TimeoutMs < 100 || cap.TimeoutMs > 600000) {
			return fmt.Errorf("capabilities[%d].timeout_ms %d 越界（允许 100..600000）", i, cap.TimeoutMs)
		}
	}
	return nil
}

// Validate 校验 BackendSpec 字段与 runtime.mode 的一致性。
//
// 一致性规则：
//   - http_endpoint 必须 runtime.mode = container（dispatcher HTTP 路由进容器）
//   - mcp_tool 允许 container 或 extension（内置 MCP server 暴露 tool）
func (b BackendSpec) Validate(runtimeMode RuntimeMode) error {
	if b.Kind == "" {
		return fmt.Errorf("kind 为必填项")
	}
	if !validCapabilityKinds[b.Kind] {
		return fmt.Errorf("kind %q 不合法 (允许 mcp_tool|http_endpoint)", b.Kind)
	}
	switch b.Kind {
	case "mcp_tool":
		if b.ToolName == "" {
			return fmt.Errorf("kind=mcp_tool 时 tool_name 为必填项")
		}
		if runtimeMode != "" && runtimeMode != RuntimeModeContainer && runtimeMode != RuntimeModeExtension {
			return fmt.Errorf("kind=mcp_tool 要求 runtime.mode 为 container 或 extension，实际 %q", runtimeMode)
		}
	case "http_endpoint":
		if b.Path == "" {
			return fmt.Errorf("kind=http_endpoint 时 path 为必填项")
		}
		if b.Method == "" {
			return fmt.Errorf("kind=http_endpoint 时 method 为必填项")
		}
		if runtimeMode != "" && runtimeMode != RuntimeModeContainer {
			return fmt.Errorf("kind=http_endpoint 要求 runtime.mode = container，实际 %q", runtimeMode)
		}
	}
	return nil
}

// Validate 校验 RequiresSpec：canonical_name 正则 / mode 枚举。
func (r RequiresSpec) Validate() error {
	for i, req := range r.Capabilities {
		if req.CanonicalName == "" {
			return fmt.Errorf("capabilities[%d].canonical_name 为必填项", i)
		}
		if !canonicalNameRegex.MatchString(req.CanonicalName) {
			return fmt.Errorf("capabilities[%d].canonical_name %q 不符合命名规范", i, req.CanonicalName)
		}
		if req.Mode == "" {
			return fmt.Errorf("capabilities[%d].mode 为必填项 (允许 required|recommended|optional)", i)
		}
		if !validRequiresModes[req.Mode] {
			return fmt.Errorf("capabilities[%d].mode %q 不合法 (允许 required|recommended|optional)", i, req.Mode)
		}
	}
	return nil
}

// Validate 校验 ConflictsSpec：app id 非空。
func (c ConflictsSpec) Validate() error {
	for i, app := range c.Apps {
		if strings.TrimSpace(app.ID) == "" {
			return fmt.Errorf("apps[%d].id 为必填项", i)
		}
	}
	return nil
}
