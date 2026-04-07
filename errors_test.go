package kstypes

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBizErrorError(t *testing.T) {
	err := &BizError{Code: 40001, Message: "参数校验失败"}
	got := err.Error()
	if got != "[40001] 参数校验失败" {
		t.Errorf("got %q", got)
	}
}

func TestBizErrorHTTPStatus(t *testing.T) {
	cases := []struct {
		err        *BizError
		wantStatus int
	}{
		// 400xx → 400
		{ErrInvalidParams, 400},
		// 401xx → 401
		{ErrTokenExpired, 401},
		{ErrTokenInvalid, 401},
		{ErrInstanceInvalid, 401},
		{ErrInstanceRevoked, 401},
		// 403xx → 403
		{ErrForbidden, 403},
		// 404xx → 404
		{ErrNotFound, 404},
		{ErrAppNotFound, 404},
		{ErrVersionNotFound, 404},
		// 409xx → 409
		{ErrDuplicate, 409},
		{ErrManifestInvalid, 409},
		{ErrManifestIDMismatch, 409},
		{ErrChecksumMismatch, 409},
		{ErrPermissionInvalid, 409},
		{ErrPermissionUnknown, 409},
		// 500xx → 500
		{ErrInternalServer, 500},
		// 503xx → 503
		{ErrStorageUploadFailed, 503},
		{ErrStorageDownloadFailed, 503},
		// 未知码 → 500 兜底
		{&BizError{99999, "未知错误"}, 500},
	}
	for _, c := range cases {
		got := c.err.HTTPStatus()
		if got != c.wantStatus {
			t.Errorf("BizError{%d}.HTTPStatus() = %d, want %d", c.err.Code, got, c.wantStatus)
		}
	}
}

func TestPredefinedErrors(t *testing.T) {
	cases := []struct {
		err      *BizError
		wantCode int
	}{
		{ErrInvalidParams, 40001},
		{ErrNotFound, 40401},
		{ErrDuplicate, 40901},
		{ErrTokenExpired, 40102},
		{ErrTokenInvalid, 40103},
		{ErrInstanceInvalid, 40104},
		{ErrInstanceRevoked, 40105},
		{ErrForbidden, 40301},
		{ErrAppNotFound, 40450},
		{ErrVersionNotFound, 40451},
		{ErrManifestInvalid, 40953},
		{ErrManifestIDMismatch, 40954},
		{ErrChecksumMismatch, 40955},
		{ErrPermissionInvalid, 40970},
		{ErrPermissionUnknown, 40971},
		{ErrStorageUploadFailed, 50301},
		{ErrStorageDownloadFailed, 50302},
		{ErrInternalServer, 50000},
	}
	for _, c := range cases {
		if c.err.Code != c.wantCode {
			t.Errorf("%s: got code %d, want %d", c.err.Message, c.err.Code, c.wantCode)
		}
	}
}

func TestNewf(t *testing.T) {
	err := Newf(40001, "字段 %s 不能为空", "name")
	if err.Code != 40001 {
		t.Errorf("code: got %d", err.Code)
	}
	if err.Message != "字段 name 不能为空" {
		t.Errorf("message: got %q", err.Message)
	}
	if err.Error() != "[40001] 字段 name 不能为空" {
		t.Errorf("Error(): got %q", err.Error())
	}
}

func TestBizErrorIs(t *testing.T) {
	a := &BizError{Code: 40001, Message: "参数校验失败"}
	b := &BizError{Code: 40001, Message: "不同的消息但同码"}
	c := &BizError{Code: 40401, Message: "资源不存在"}

	if !a.Is(b) {
		t.Error("同码 BizError 应匹配")
	}
	if a.Is(c) {
		t.Error("异码 BizError 不应匹配")
	}

	// 非 BizError 类型
	if a.Is(fmt.Errorf("普通错误")) {
		t.Error("非 BizError 不应匹配")
	}
}

func TestBizErrorJSON(t *testing.T) {
	err := &BizError{Code: 40001, Message: "参数校验失败"}
	data, _ := json.Marshal(err)
	got := string(data)
	want := `{"code":40001,"message":"参数校验失败"}`
	if got != want {
		t.Errorf("json: got %s, want %s", got, want)
	}
}
