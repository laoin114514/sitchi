package model

import (
	"fmt"
	"net/http"
)

// AppError 全局业务错误定义。
type AppError struct {
	Code    int    `json:"code"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
	I18nKey string `json:"i18n_key,omitempty"`
	Detail  any    `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("code=%d type=%s message=%s", e.Code, e.Type, e.Message)
}

func (e *AppError) WithDetail(detail any) *AppError {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Detail = detail
	return &cp
}

// HTTPStatus 将业务错误码映射为 HTTP 状态码。
func (e *AppError) HTTPStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}

	switch {
	case e.Code >= 400000 && e.Code < 401000:
		return http.StatusBadRequest
	case e.Code >= 401000 && e.Code < 402000:
		return http.StatusUnauthorized
	case e.Code >= 403000 && e.Code < 404000:
		return http.StatusForbidden
	case e.Code >= 404000 && e.Code < 405000:
		return http.StatusNotFound
	case e.Code >= 409000 && e.Code < 410000:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// HTTPStatusFromError 从任意错误推导 HTTP 状态码。
func HTTPStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if ae, ok := err.(*AppError); ok {
		return ae.HTTPStatus()
	}
	return http.StatusInternalServerError
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewAppTypedError(code int, errType, message string) *AppError {
	return &AppError{Code: code, Type: errType, Message: message}
}

var (
	ErrInvalidParams = &AppError{Code: 400001, Type: "INVALID_PARAMS", Message: "invalid params", I18nKey: "error.invalid_params"}
	ErrUnauthorized  = &AppError{Code: 401001, Type: "UNAUTHORIZED", Message: "unauthorized", I18nKey: "error.unauthorized"}
	ErrForbidden     = &AppError{Code: 403001, Type: "FORBIDDEN", Message: "forbidden", I18nKey: "error.forbidden"}
	ErrNotFound      = &AppError{Code: 404001, Type: "NOT_FOUND", Message: "resource not found", I18nKey: "error.not_found"}
	ErrConflict      = &AppError{Code: 409001, Type: "CONFLICT", Message: "resource conflict", I18nKey: "error.conflict"}
	ErrInternal      = &AppError{Code: 500001, Type: "INTERNAL_ERROR", Message: "internal server error", I18nKey: "error.internal"}
)
