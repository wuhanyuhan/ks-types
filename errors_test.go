package kstypes

import "testing"

func TestBizErrorError(t *testing.T) {
	err := &BizError{Code: 40001, Message: "参数校验失败"}
	got := err.Error()
	if got != "[40001] 参数校验失败" {
		t.Errorf("got %q", got)
	}
}

func TestPredefinedErrors(t *testing.T) {
	if ErrTokenInvalid.Code != 40103 {
		t.Error("ErrTokenInvalid code mismatch")
	}
	if ErrManifestInvalid.Code != 40953 {
		t.Error("ErrManifestInvalid code mismatch")
	}
}
