package kstypes

import "testing"

// TestSDUIPrimitiveSchemas_WarRoomRegistered 锁定作战室 wire 契约：war-room 必须在后端原语注册表里，
// 否则 keystone ValidateSDUITree 会以 "unknown primitive type" 拒掉 squad 返回的 ks:// SDUI 作战室节点。
func TestSDUIPrimitiveSchemas_WarRoomRegistered(t *testing.T) {
	if _, ok := SDUIPrimitiveSchemas[PrimitiveWarRoom]; !ok {
		t.Fatalf("war-room 必须注册进 SDUIPrimitiveSchemas（后端 ValidateSDUITree 据此接受作战室节点）")
	}
	// war-room 是叶子原语（data 驱动、无 children），不应允许 children。
	if ContainerPrimitives[PrimitiveWarRoom] {
		t.Fatalf("war-room 不应在 ContainerPrimitives（它无 children，子视图前端从实时流组装）")
	}
}
