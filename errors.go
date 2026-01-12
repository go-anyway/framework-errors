package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// StatusError 状态错误接口
type StatusError interface {
	error
	Code() int32
	IsAffectStability() bool
	Msg() string
	Extra() map[string]string
}

// Extension 包含了错误的扩展信息
type Extension struct {
	IsAffectStability bool
	Extra             map[string]string
}

// statusError 存储了错误的核心信息
type statusError struct {
	statusCode int32
	message    string
	ext        Extension
}

// GetCodeDefinition 获取错误码定义，如果不存在则返回默认定义
func GetCodeDefinition(code int32) CodeDefinition {
	if def, ok := CodeDefinitions[code]; ok {
		return def
	}
	// 返回默认定义
	return CodeDefinition{
		Message:           "未知错误",
		IsAffectStability: true, // 未知错误默认影响稳定性
	}
}

// GetMessage 获取错误码对应的消息，如果提供了自定义消息则优先使用
func GetMessage(code int32, customMessage string) string {
	if customMessage != "" {
		return customMessage
	}
	return GetCodeDefinition(code).Message
}

// Error 实现 error 接口
func (e *statusError) Error() string {
	return e.message
}

// Code 返回错误码
func (e *statusError) Code() int32 {
	return e.statusCode
}

// IsAffectStability 返回是否影响系统稳定性
func (e *statusError) IsAffectStability() bool {
	return e.ext.IsAffectStability
}

// Msg 返回错误消息
func (e *statusError) Msg() string {
	return e.message
}

// Extra 返回扩展信息
func (e *statusError) Extra() map[string]string {
	if e.ext.Extra == nil {
		return make(map[string]string)
	}
	return e.ext.Extra
}

// NewStatusError 创建状态错误
// 如果 message 为空，则使用 CodeDefinitions 中定义的默认消息
func NewStatusError(code int32, message string, data interface{}) StatusError {
	if message == "" {
		message = GetMessage(code, "")
	}

	// 获取错误码定义
	def := GetCodeDefinition(code)

	// 转换 data 为 map[string]string
	extra := make(map[string]string)
	if data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			for k, v := range dataMap {
				extra[k] = fmt.Sprintf("%v", v)
			}
		} else if dataMap, ok := data.(map[string]string); ok {
			extra = dataMap
		}
	}

	return &statusError{
		statusCode: code,
		message:    message,
		ext: Extension{
			IsAffectStability: def.IsAffectStability,
			Extra:             extra,
		},
	}
}

// ToGRPCStatus 将 StatusError 转换为 gRPC status，使用 details 传递状态错误信息
func ToGRPCStatus(err StatusError) *status.Status {
	if err == nil {
		return status.New(codes.Internal, "unknown error")
	}

	// 根据业务错误码映射到 gRPC codes
	var grpcCode codes.Code
	switch err.Code() {
	case CodeInvalidParam:
		grpcCode = codes.InvalidArgument
	case CodeUnauthorized:
		grpcCode = codes.Unauthenticated
	case CodeForbidden:
		grpcCode = codes.PermissionDenied
	case CodeNotFound:
		grpcCode = codes.NotFound
	case CodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	default:
		grpcCode = codes.Internal
	}

	// 创建包含业务错误信息的 struct
	st := status.New(grpcCode, err.Msg())

	// 将扩展信息放入 details
	extra := err.Extra()
	if len(extra) > 0 {
		// 转换为 map[string]interface{} 以便使用 structpb
		extraMap := make(map[string]interface{})
		for k, v := range extra {
			extraMap[k] = v
		}
		if structValue, err := structpb.NewStruct(extraMap); err == nil {
			anyValue, _ := anypb.New(structValue)
			st, _ = st.WithDetails(anyValue)
		}
	}

	// 将业务错误码也放入 details（使用自定义字段）
	errorInfo := map[string]interface{}{
		"business_code": err.Code(),
		"business_msg":  err.Msg(),
	}
	if structValue, err := structpb.NewStruct(errorInfo); err == nil {
		anyValue, _ := anypb.New(structValue)
		st, _ = st.WithDetails(anyValue)
	}

	return st
}

// FromGRPCStatus 从 gRPC status 解析状态错误
func FromGRPCStatus(st *status.Status) StatusError {
	code := CodeInternalError
	message := st.Message()
	var extraData map[string]string

	// 从 details 中提取业务错误信息
	details := st.Details()
	for _, detail := range details {
		if anyValue, ok := detail.(*anypb.Any); ok {
			var structValue structpb.Struct
			if err := anyValue.UnmarshalTo(&structValue); err == nil {
				structMap := structValue.AsMap()

				// 检查是否是业务错误信息
				if bizCode, ok := structMap["business_code"].(float64); ok {
					code = int32(bizCode)
					if bizMsg, ok := structMap["business_msg"].(string); ok {
						message = bizMsg
					}
				} else {
					// 否则作为扩展数据
					if extraData == nil {
						extraData = make(map[string]string)
					}
					for k, v := range structMap {
						extraData[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}
	}

	// 如果没有从 details 中提取到业务错误码，根据 gRPC code 映射
	if code == CodeInternalError {
		switch st.Code() {
		case codes.InvalidArgument:
			code = CodeInvalidParam
		case codes.Unauthenticated:
			code = CodeUnauthorized
		case codes.PermissionDenied:
			code = CodeForbidden
		case codes.NotFound:
			code = CodeNotFound
		case codes.AlreadyExists:
			code = CodeAlreadyExists
		default:
			code = CodeInternalError
		}
	}

	return NewStatusError(code, message, extraData)
}
