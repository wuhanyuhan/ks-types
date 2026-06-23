package kstypes

import (
	"encoding/json"
	"testing"
)

func TestMountedRouteMethodConstants(t *testing.T) {
	if PMMethodMountedRouteChanged != "keystone.mounted.route.changed" {
		t.Fatalf("unexpected route changed method: %s", PMMethodMountedRouteChanged)
	}
	if PMMethodMountedRouteRestore != "keystone.mounted.route.restore" {
		t.Fatalf("unexpected route restore method: %s", PMMethodMountedRouteRestore)
	}
}

func TestMountedRouteChangedMessageJSON(t *testing.T) {
	msg := MountedRouteChangedMessage{
		Type:    PMMethodMountedRouteChanged,
		Version: 1,
		AppID:   "document",
		Path:    "/templates",
		Hash:    "#/templates",
		Title:   "模板管理",
		Replace: true,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal mounted route changed message: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal mounted route changed message: %v", err)
	}

	if got["type"] != PMMethodMountedRouteChanged {
		t.Fatalf("type mismatch: %#v", got["type"])
	}
	if got["appId"] != "document" {
		t.Fatalf("appId mismatch: %#v", got["appId"])
	}
	if got["path"] != "/templates" {
		t.Fatalf("path mismatch: %#v", got["path"])
	}
	if got["hash"] != "#/templates" {
		t.Fatalf("hash mismatch: %#v", got["hash"])
	}
	if got["replace"] != true {
		t.Fatalf("replace mismatch: %#v", got["replace"])
	}
}

func TestMountedRouteRestoreMessageJSON(t *testing.T) {
	msg := MountedRouteRestoreMessage{
		Type:    PMMethodMountedRouteRestore,
		Version: 1,
		Path:    "/ocr",
		Hash:    "#/ocr",
		Replace: true,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal mounted route restore message: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal mounted route restore message: %v", err)
	}

	if got["type"] != PMMethodMountedRouteRestore {
		t.Fatalf("type mismatch: %#v", got["type"])
	}
	if got["path"] != "/ocr" {
		t.Fatalf("path mismatch: %#v", got["path"])
	}
	if got["hash"] != "#/ocr" {
		t.Fatalf("hash mismatch: %#v", got["hash"])
	}
}
