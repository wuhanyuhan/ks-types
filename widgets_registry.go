package kstypes

import "reflect"

// SharedWidgetSchemas 是共享 widget URI → 数据 schema 类型的注册表。
// keystone 后端 NormalizeToolResult 用此表查找并 schema 校验。
//
// 加新 widget：往本表加条目；ks-types 升 minor；前端消费方同步实现 React 组件。
// 改既有 widget：仅允许加可选字段（同 URI）；改字段类型必须升 major（@v2 共存）。
var SharedWidgetSchemas = map[string]reflect.Type{
	"ks://widgets/list-actions@v1":   reflect.TypeOf(WidgetListActionsV1{}),
	"ks://widgets/diff-review@v1":    reflect.TypeOf(WidgetDiffReviewV1{}),
	"ks://widgets/timeline@v1":       reflect.TypeOf(WidgetTimelineV1{}),
	"ks://widgets/card-grid@v1":      reflect.TypeOf(WidgetCardGridV1{}),
	"ks://widgets/image-variants@v1": reflect.TypeOf(WidgetImageVariantsV1{}),
}

// DeprecationInfo 记录将下线 widget 的元信息。
type DeprecationInfo struct {
	DeprecatedSince string
	EOLDate         string
	Replacement     string
	MigrationGuide  string
}

// DeprecatedWidgets 当前为空；future 升 v2 时填。
var DeprecatedWidgets = map[string]DeprecationInfo{}
