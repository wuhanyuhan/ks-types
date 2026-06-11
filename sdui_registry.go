package kstypes

import "reflect"

// PrimitiveWarRoom 是作战室容器原语：node.data 绑定 team_progress_stream 实时数据源，
// 前端 war-room 原语订阅 SSE 后从 TeamState 组装协作子视图（connection-status / lead-narration /
// expert-roster / decision-gate / deliverable-panel 等）。这些子视图是**前端内部渲染**、由实时流
// 驱动，不作为独立 wire 原语下发——故后端只需注册 war-room 本身即可接受作战室节点。
//
// war-room 无 props（纯 data 驱动），props schema 为空 struct；是叶子原语（无 children）。
// 刻意定义在本文件（非 sdui_primitives.go）：war-room 是纯后端校验事实，无 TS props 派生需求。
const PrimitiveWarRoom = "war-room"

// SDUIWarRoomProps 是 war-room 的（空）props schema——war-room 由 node.data 驱动、无 props 字段。
type SDUIWarRoomProps struct{}

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
	// 协作（P2 作战室）：data 驱动、无 props；子视图前端从实时流组装，不下发为独立 wire 原语。
	PrimitiveWarRoom: reflect.TypeOf(SDUIWarRoomProps{}),
}

// ContainerPrimitives 标记哪些原语允许 children（容器）。校验器据此拒绝叶子原语带 children。
var ContainerPrimitives = map[string]bool{
	PrimitiveStack: true, PrimitiveGrid: true, PrimitiveCard: true,
	PrimitiveSection: true, PrimitiveTabs: true, PrimitiveSplit: true,
}
