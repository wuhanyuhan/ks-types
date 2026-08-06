package kstypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestIsKnownErrorCategory(t *testing.T) {
	t.Parallel()
	for _, c := range []ErrorCategory{
		ErrorCategoryPermission, ErrorCategoryNotFound, ErrorCategoryValidation,
		ErrorCategoryDependency, ErrorCategoryTimeout, ErrorCategoryRateLimited,
		ErrorCategoryInternal, ErrorCategoryUpstream,
	} {
		if !IsKnownErrorCategory(c) {
			t.Errorf("IsKnownErrorCategory(%q) = false, want true", c)
		}
	}
	for _, c := range []ErrorCategory{"", "bogus", "DEPENDENCY"} {
		if IsKnownErrorCategory(c) {
			t.Errorf("IsKnownErrorCategory(%q) = true, want false", c)
		}
	}
}

func TestNormalizeErrorCategory(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   ErrorCategory
		want ErrorCategory
	}{
		{ErrorCategoryUpstream, ErrorCategoryUpstream},
		{ErrorCategoryTimeout, ErrorCategoryTimeout},
		{"", ErrorCategoryInternal},
		{"unknown_thing", ErrorCategoryInternal},
	}
	for _, tc := range cases {
		if got := NormalizeErrorCategory(tc.in); got != tc.want {
			t.Errorf("NormalizeErrorCategory(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestToolErrorErrorsAs(t *testing.T) {
	t.Parallel()
	var wrapped error = fmt.Errorf("handler: %w", NewUpstreamError(521, "origin down"))
	var te *ToolError
	if !errors.As(wrapped, &te) {
		t.Fatal("errors.As 应命中 *ToolError")
	}
	if te.Category != ErrorCategoryUpstream || te.UpstreamStatus != 521 {
		t.Errorf("got %+v, want upstream/521", te)
	}
}

// TestToolErrorJSONWireCompat 守护 wire 兼容：可选维缺省不出 wire（旧 keystone 收到的
// payload 与下沉前完全一致），显式设置时按约定字段名序列化。
func TestToolErrorJSONWireCompat(t *testing.T) {
	t.Parallel()
	plain, err := json.Marshal(NewToolError(ErrorCategoryValidation, "bad args"))
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != `{"category":"validation","message":"bad args"}` {
		t.Errorf("缺省序列化 = %s，可选维不得出 wire", plain)
	}

	retryable := true
	full, err := json.Marshal(&ToolError{
		Category: ErrorCategoryUpstream, Message: "origin 521",
		Retryable: &retryable, UpstreamStatus: 521,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"category":"upstream","message":"origin 521","retryable":true,"upstream_status":521}`
	if string(full) != want {
		t.Errorf("完整序列化 = %s, want %s", full, want)
	}

	var back ToolError
	if err := json.Unmarshal(full, &back); err != nil {
		t.Fatal(err)
	}
	if back.Retryable == nil || !*back.Retryable || back.UpstreamStatus != 521 {
		t.Errorf("round-trip 丢维：%+v", back)
	}
}

func TestToolErrorErrorFormat(t *testing.T) {
	t.Parallel()
	e := NewToolError(ErrorCategoryTimeout, "took too long")
	if got, want := e.Error(), "[timeout] took too long"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
