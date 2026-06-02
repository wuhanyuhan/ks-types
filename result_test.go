package kstypes

import (
	"encoding/json"
	"testing"
)

func TestOK(t *testing.T) {
	r := OK(map[string]string{"key": "value"})
	if r.Code != 0 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "ok" {
		t.Errorf("message: got %q", r.Message)
	}
}

func TestOK_JSON(t *testing.T) {
	r := OK(nil)
	data, _ := json.Marshal(r)
	got := string(data)
	want := `{"code":0,"message":"ok","data":null}`
	if got != want {
		t.Errorf("json: got %s", got)
	}
}

func TestFail(t *testing.T) {
	r := Fail(40001, "参数校验失败")
	if r.Code != 40001 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "参数校验失败" {
		t.Errorf("message: got %q", r.Message)
	}
	if r.Data != nil {
		t.Errorf("data: got %v, want nil", r.Data)
	}
}

func TestFailErr(t *testing.T) {
	err := &BizError{Code: 40401, Message: "资源不存在"}
	r := FailErr(err)
	if r.Code != 40401 {
		t.Errorf("code: got %d", r.Code)
	}
	if r.Message != "资源不存在" {
		t.Errorf("message: got %q", r.Message)
	}
}

func TestPageResult_JSON(t *testing.T) {
	pr := PageResult[string]{
		Items: []string{"a", "b"},
		Total: 10,
		Page:  1,
		Size:  2,
	}
	data, _ := json.Marshal(pr)
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["total"].(float64) != 10 {
		t.Errorf("total: got %v", m["total"])
	}
	if m["page"].(float64) != 1 {
		t.Errorf("page: got %v", m["page"])
	}
}

func TestListResult_JSON(t *testing.T) {
	lr := ListResult[int]{Items: []int{1, 2, 3}}
	data, _ := json.Marshal(lr)
	var m map[string]any
	json.Unmarshal(data, &m)
	items := m["items"].([]any)
	if len(items) != 3 {
		t.Errorf("items len: got %d", len(items))
	}
}
