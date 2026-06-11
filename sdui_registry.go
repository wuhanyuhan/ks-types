package kstypes

import "reflect"

// SDUIPrimitiveSchemas 是原语 type → props schema 类型的注册表。
// 后端递归校验器按 node.type 查表、reflect 解码 props + 调 Validate()。
// 复合原语（list-actions 等）复用 widgets_data.go 现有 WidgetXxxV1 类型。
//
// 注意：本文件刻意不进 tygo.yaml 的 include_files（reflect 注册表是纯后端事实源，
// 不派生前端 TS 类型），与 widgets_registry.go / widgets_data.go 的拆分一致。
var SDUIPrimitiveSchemas = map[string]reflect.Type{
	PrimitiveStack:       reflect.TypeOf(SDUIStackProps{}),
	PrimitiveGrid:        reflect.TypeOf(SDUIGridProps{}),
	PrimitiveCard:        reflect.TypeOf(SDUICardProps{}),
	PrimitiveSection:     reflect.TypeOf(SDUISectionProps{}),
	PrimitiveTabs:        reflect.TypeOf(SDUITabsProps{}),
	PrimitiveSplit:       reflect.TypeOf(SDUISplitProps{}),
	PrimitiveText:        reflect.TypeOf(SDUITextProps{}),
	PrimitiveMarkdown:    reflect.TypeOf(SDUIMarkdownProps{}),
	PrimitiveFieldGroup:  reflect.TypeOf(SDUIFieldGroupProps{}),
	PrimitiveTable:       reflect.TypeOf(SDUITableProps{}),
	PrimitiveStatusBadge: reflect.TypeOf(SDUIStatusBadgeProps{}),
	PrimitiveMetric:      reflect.TypeOf(SDUIMetricProps{}),
	PrimitiveEmptyState:  reflect.TypeOf(SDUIEmptyStateProps{}),
	PrimitiveButton:      reflect.TypeOf(SDUIButtonProps{}),
	PrimitiveForm:        reflect.TypeOf(SDUIFormProps{}),
	PrimitiveLink:        reflect.TypeOf(SDUILinkProps{}),
	// 复合（遗留降级）
	PrimitiveListActions:   reflect.TypeOf(WidgetListActionsV1{}),
	PrimitiveDiffReview:    reflect.TypeOf(WidgetDiffReviewV1{}),
	PrimitiveTimeline:      reflect.TypeOf(WidgetTimelineV1{}),
	PrimitiveCardGrid:      reflect.TypeOf(WidgetCardGridV1{}),
	PrimitiveImageVariants: reflect.TypeOf(WidgetImageVariantsV1{}),
}

// ContainerPrimitives 标记哪些原语允许 children（容器）。校验器据此拒绝叶子原语带 children。
var ContainerPrimitives = map[string]bool{
	PrimitiveStack: true, PrimitiveGrid: true, PrimitiveCard: true,
	PrimitiveSection: true, PrimitiveTabs: true, PrimitiveSplit: true,
}
