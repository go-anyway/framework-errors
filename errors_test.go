package errors_test

import (
	errstd "errors"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/go-anyway/framework-errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCodeDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		code     int32
		expected errors.CodeDefinition
	}{
		{
			name:     "成功码",
			code:     errors.CodeSuccess,
			expected: errors.CodeDefinition{Message: "success", IsAffectStability: false},
		},
		{
			name:     "无效参数",
			code:     errors.CodeInvalidParam,
			expected: errors.CodeDefinition{Message: "参数无效", IsAffectStability: false},
		},
		{
			name:     "未授权",
			code:     errors.CodeUnauthorized,
			expected: errors.CodeDefinition{Message: "未授权", IsAffectStability: false},
		},
		{
			name:     "内部错误",
			code:     errors.CodeInternalError,
			expected: errors.CodeDefinition{Message: "内部服务器错误", IsAffectStability: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := errors.GetCodeDefinition(tt.code)
			if def.Message != tt.expected.Message {
				t.Errorf("GetCodeDefinition(%d).Message = %s, want %s", tt.code, def.Message, tt.expected.Message)
			}
			if def.IsAffectStability != tt.expected.IsAffectStability {
				t.Errorf("GetCodeDefinition(%d).IsAffectStability = %v, want %v", tt.code, def.IsAffectStability, tt.expected.IsAffectStability)
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name          string
		code          int32
		customMessage string
		expected      string
	}{
		{
			name:          "使用自定义消息",
			code:          errors.CodeNotFound,
			customMessage: "自定义消息",
			expected:      "自定义消息",
		},
		{
			name:          "使用默认消息",
			code:          errors.CodeNotFound,
			customMessage: "",
			expected:      "资源未找到",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := errors.GetMessage(tt.code, tt.customMessage)
			if msg != tt.expected {
				t.Errorf("GetMessage(%d, %s) = %s, want %s", tt.code, tt.customMessage, msg, tt.expected)
			}
		})
	}
}

func TestNewStatusError(t *testing.T) {
	err := errors.NewStatusError(errors.CodeNotFound, "资源不存在", nil)

	if err.Code() != errors.CodeNotFound {
		t.Errorf("Code() = %d, want %d", err.Code(), errors.CodeNotFound)
	}
	if err.Msg() != "资源不存在" {
		t.Errorf("Msg() = %s, want 资源不存在", err.Msg())
	}
	if err.IsAffectStability() {
		t.Error("IsAffectStability() = true, want false")
	}
}

func TestNewStatusErrorWithExtra(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	err := errors.NewStatusError(errors.CodeNotFound, "", data)

	extra := err.Extra()
	if extra["key1"] != "value1" {
		t.Errorf("Extra[key1] = %s, want value1", extra["key1"])
	}
	if extra["key2"] != "123" {
		t.Errorf("Extra[key2] = %s, want 123", extra["key2"])
	}
}

func TestStatusErrorInterface(t *testing.T) {
	err := errors.NewStatusError(errors.CodeInternalError, "内部错误", nil)

	var statusErr errors.StatusError
	if !errstd.As(err, &statusErr) {
		t.Error("应能转换为 StatusError 接口")
	}

	if statusErr.Code() != errors.CodeInternalError {
		t.Errorf("Code() = %d, want %d", statusErr.Code(), errors.CodeInternalError)
	}
}

func TestStatusErrorExtraNil(t *testing.T) {
	err := errors.NewStatusError(errors.CodeSuccess, "成功", nil)
	extra := err.Extra()

	if extra == nil {
		t.Error("Extra() 不应返回 nil")
	}
	if len(extra) != 0 {
		t.Errorf("Extra() length = %d, want 0", len(extra))
	}
}

func TestToGRPCStatus(t *testing.T) {
	tests := []struct {
		name             string
		err              errors.StatusError
		expectedGRPCCode codes.Code
	}{
		{
			name:             "无效参数",
			err:              errors.NewStatusError(errors.CodeInvalidParam, "参数错误", nil),
			expectedGRPCCode: codes.InvalidArgument,
		},
		{
			name:             "未授权",
			err:              errors.NewStatusError(errors.CodeUnauthorized, "未授权", nil),
			expectedGRPCCode: codes.Unauthenticated,
		},
		{
			name:             "禁止访问",
			err:              errors.NewStatusError(errors.CodeForbidden, "禁止访问", nil),
			expectedGRPCCode: codes.PermissionDenied,
		},
		{
			name:             "未找到",
			err:              errors.NewStatusError(errors.CodeNotFound, "资源未找到", nil),
			expectedGRPCCode: codes.NotFound,
		},
		{
			name:             "已存在",
			err:              errors.NewStatusError(errors.CodeAlreadyExists, "资源已存在", nil),
			expectedGRPCCode: codes.AlreadyExists,
		},
		{
			name:             "内部错误",
			err:              errors.NewStatusError(errors.CodeInternalError, "内部错误", nil),
			expectedGRPCCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := errors.ToGRPCStatus(tt.err)
			if st.Code() != tt.expectedGRPCCode {
				t.Errorf("gRPC status code = %v, want %v", st.Code(), tt.expectedGRPCCode)
			}
		})
	}
}

func TestToGRPCStatusWithNil(t *testing.T) {
	st := errors.ToGRPCStatus(nil)
	if st.Code() != codes.Internal {
		t.Errorf("nil 错误应返回 Internal code, got %v", st.Code())
	}
}

func TestFromGRPCStatus(t *testing.T) {
	st := status.New(codes.NotFound, "资源未找到")
	err := errors.FromGRPCStatus(st)

	if err.Code() != errors.CodeNotFound {
		t.Errorf("Code() = %d, want %d", err.Code(), errors.CodeNotFound)
	}
	if err.Msg() != "资源未找到" {
		t.Errorf("Msg() = %s, want 资源未找到", err.Msg())
	}
}

func TestFromGRPCStatusWithDetails(t *testing.T) {
	st := status.New(codes.InvalidArgument, "参数错误")
	anyVal, _ := anypb.New(&structpb.Struct{})
	st, _ = st.WithDetails(anyVal)

	err := errors.FromGRPCStatus(st)
	if err.Code() != errors.CodeInvalidParam {
		t.Errorf("Code() = %d, want %d", err.Code(), errors.CodeInvalidParam)
	}
}

func TestWithStatus(t *testing.T) {
	baseErr := errors.NewStatusError(errors.CodeNotFound, "资源未找到", nil)
	wrappedErr := errors.WithStack(baseErr)

	if wrappedErr.Code() != errors.CodeNotFound {
		t.Errorf("Code() = %d, want %d", wrappedErr.Code(), errors.CodeNotFound)
	}
	if wrappedErr.Msg() != "资源未找到" {
		t.Errorf("Msg() = %s, want 资源未找到", wrappedErr.Msg())
	}
}

func TestWithStackNil(t *testing.T) {
	err := errors.WithStack(nil)
	if err != nil {
		t.Error("WithStack(nil) 应返回 nil")
	}
}

func TestWithStackAlreadyWrapped(t *testing.T) {
	baseErr := errors.NewStatusError(errors.CodeNotFound, "资源未找到", nil)
	wrapped1 := errors.WithStack(baseErr)
	wrapped2 := errors.WithStack(wrapped1)

	if wrapped1.Code() != wrapped2.Code() {
		t.Error("重复包装应保持相同的错误码")
	}
}

func TestNewWithStatus(t *testing.T) {
	err := errors.NewWithStatus(errors.CodeUserNotFound, "用户不存在")

	if err.Code() != errors.CodeUserNotFound {
		t.Errorf("Code() = %d, want %d", err.Code(), errors.CodeUserNotFound)
	}
	if err.Msg() != "用户不存在" {
		t.Errorf("Msg() = %s, want 用户不存在", err.Msg())
	}
}

func TestNewWithStatusWithOptions(t *testing.T) {
	err := errors.NewWithStatus(
		errors.CodeNotFound,
		"user {name} not found",
		errors.Param("name", "zampo"),
		errors.Extra("key", "value"),
	)

	if err.Msg() != "user zampo not found" {
		t.Errorf("Msg() = %s, want user zampo not found", err.Msg())
	}

	extra := err.Extra()
	if extra["key"] != "value" {
		t.Errorf("Extra[key] = %s, want value", extra["key"])
	}
}

func TestWrapWithStatus(t *testing.T) {
	originalErr := errstd.New("原始错误")
	wrappedErr := errors.WrapWithStatus(originalErr, errors.CodeInternalError, "包装错误", nil)

	if wrappedErr.Code() != errors.CodeInternalError {
		t.Errorf("Code() = %d, want %d", wrappedErr.Code(), errors.CodeInternalError)
	}
	if wrappedErr.Msg() != "包装错误" {
		t.Errorf("Msg() = %s, want 包装错误", wrappedErr.Msg())
	}
}

func TestWrapWithStatusNil(t *testing.T) {
	err := errors.WrapWithStatus(nil, errors.CodeInternalError, "错误", nil)
	if err != nil {
		t.Error("WrapWithStatus(nil) 应返回 nil")
	}
}

func TestWrapWithStatusOptions(t *testing.T) {
	originalErr := errstd.New("原始错误")
	wrappedErr := errors.WrapWithStatusOptions(
		originalErr,
		errors.CodeNotFound,
		"user {id} not found",
		errors.Param("id", "123"),
	)

	if wrappedErr.Code() != errors.CodeNotFound {
		t.Errorf("Code() = %d, want %d", wrappedErr.Code(), errors.CodeNotFound)
	}
	if wrappedErr.Msg() != "user 123 not found" {
		t.Errorf("Msg() = %s, want user 123 not found", wrappedErr.Msg())
	}
}

func TestWithStatusUnwrap(t *testing.T) {
	originalErr := errstd.New("原始错误")
	wrappedErr := errors.WrapWithStatus(originalErr, errors.CodeInternalError, "包装错误", nil)

	unwrapped := errstd.Unwrap(wrappedErr)
	if unwrapped == nil || unwrapped.Error() != "原始错误" {
		t.Error("Unwrap 应返回原始错误")
	}
}

func TestWithStatusExtra(t *testing.T) {
	err := errors.NewWithStatus(
		errors.CodeNotFound,
		"未找到",
		errors.Extra("field", "value"),
	)

	extra := err.Extra()
	if extra["field"] != "value" {
		t.Errorf("Extra[field] = %s, want value", extra["field"])
	}
	if extra["stack"] == "" {
		t.Error("Extra 应包含堆栈信息")
	}
}

func TestStatusErrorIsAffectStability(t *testing.T) {
	tests := []struct {
		name     string
		code     int32
		expected bool
	}{
		{"成功", errors.CodeSuccess, false},
		{"内部错误", errors.CodeInternalError, true},
		{"参数无效", errors.CodeInvalidParam, false},
		{"用户未找到", errors.CodeUserNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.NewStatusError(tt.code, "", nil)
			if err.IsAffectStability() != tt.expected {
				t.Errorf("IsAffectStability() = %v, want %v", err.IsAffectStability(), tt.expected)
			}
		})
	}
}

func TestGetCodeDefinitionUnknown(t *testing.T) {
	def := errors.GetCodeDefinition(99999)
	if def.Message == "" {
		t.Error("未知错误码应有默认消息")
	}
	if !def.IsAffectStability {
		t.Error("未知错误应默认为影响稳定性")
	}
}

func TestParamOption(t *testing.T) {
	err := errors.NewWithStatus(
		errors.CodeNotFound,
		"user {name} not found",
		errors.Param("name", "testuser"),
	)

	if err.Msg() != "user testuser not found" {
		t.Errorf("Msg() = %s, want user testuser not found", err.Msg())
	}
}

func TestExtraOption(t *testing.T) {
	err := errors.NewWithStatus(
		errors.CodeNotFound,
		"未找到",
		errors.Extra("key1", "value1"),
		errors.Extra("key2", "value2"),
	)

	extra := err.Extra()
	if extra["key1"] != "value1" {
		t.Errorf("Extra[key1] = %s, want value1", extra["key1"])
	}
	if extra["key2"] != "value2" {
		t.Errorf("Extra[key2] = %s, want value2", extra["key2"])
	}
}

func TestErrorInterface(t *testing.T) {
	err := errors.NewStatusError(errors.CodeNotFound, "资源未找到", nil)

	if err.Error() != "资源未找到" {
		t.Errorf("Error() = %s, want 资源未找到", err.Error())
	}
}

func TestWithStatusErrorMessage(t *testing.T) {
	err := errors.NewStatusError(errors.CodeNotFound, "未找到", nil)

	if err.Error() != "未找到" {
		t.Errorf("Error() = %s, want 未找到", err.Error())
	}
}

func TestWithStatusCauseErrorMessage(t *testing.T) {
	originalErr := errstd.New("原始错误")
	wrappedErr := errors.WrapWithStatus(originalErr, errors.CodeInternalError, "包装错误", nil)

	expectedMsg := "包装错误: 原始错误"
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("Error() = %s, want %s", wrappedErr.Error(), expectedMsg)
	}
}
