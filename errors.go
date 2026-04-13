package kstypes

import "fmt"

// BizError 业务错误
type BizError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BizError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// 通用错误 (400xx)
var (
	ErrInvalidParams = &BizError{40001, "参数校验失败"}
	ErrNotFound      = &BizError{40401, "资源不存在"}
	ErrDuplicate     = &BizError{40901, "资源已存在"}
)

// 认证错误 (401xx)
var (
	ErrTokenExpired    = &BizError{40102, "Token 已过期"}
	ErrTokenInvalid    = &BizError{40103, "无效的 Token"}
	ErrInstanceInvalid = &BizError{40104, "无效的实例令牌"}
	ErrInstanceRevoked = &BizError{40105, "实例令牌已吊销"}
)

// 权限错误 (403xx)
var (
	ErrForbidden = &BizError{40301, "无权访问"}
)

// 应用错误 (4045x/4095x/4097x)
var (
	ErrAppNotFound        = &BizError{40450, "应用不存在"}
	ErrVersionNotFound    = &BizError{40451, "版本不存在"}
	ErrManifestInvalid    = &BizError{40953, "Manifest 格式无效"}
	ErrManifestIDMismatch = &BizError{40954, "Manifest 中的 app_id 与路径不匹配"}
	ErrChecksumMismatch   = &BizError{40955, "文件校验失败"}
	ErrPermissionInvalid  = &BizError{40970, "权限声明无效"}
	ErrPermissionUnknown  = &BizError{40971, "未知的权限维度"}
)

// 存储错误 (503xx)
var (
	ErrStorageUploadFailed   = &BizError{50301, "文件上传失败"}
	ErrStorageDownloadFailed = &BizError{50302, "文件下载失败"}
)

// 服务器错误 (500xx)
var (
	ErrInternalServer = &BizError{50000, "服务器内部错误"}
)

// Newf 构造带格式化 message 的 BizError
func Newf(code int, format string, args ...any) *BizError {
	return &BizError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Is 判断当前错误是否与目标 BizError 同码（满足 errors.Is 协议）
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	return ok && e.Code == t.Code
}

// HTTPStatus 根据错误码前缀推导 HTTP 状态码
func (e *BizError) HTTPStatus() int {
	prefix := e.Code / 100
	switch prefix {
	case 400:
		return 400
	case 401:
		return 401
	case 403:
		return 403
	case 404:
		return 404
	case 409:
		return 409
	case 500:
		return 500
	case 503:
		return 503
	default:
		return 500
	}
}
