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
