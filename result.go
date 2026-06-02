package kstypes

// Result 统一 HTTP JSON 响应结构
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 成功响应
func OK(data any) Result {
	return Result{Code: 0, Message: "ok", Data: data}
}

// Fail 失败响应
func Fail(code int, message string) Result {
	return Result{Code: code, Message: message}
}

// FailErr 从 BizError 构建失败响应
func FailErr(err *BizError) Result {
	return Result{Code: err.Code, Message: err.Message}
}

// PageResult 分页列表响应
type PageResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

// ListResult 非分页列表响应
type ListResult[T any] struct {
	Items []T `json:"items"`
}
