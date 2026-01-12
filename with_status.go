// Copyright 2025 zampo.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @contact  zampo3380@gmail.com

package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// withStatus 是一个包装器，它包含了 statusError、调用堆栈和底层的 cause error.
type withStatus struct {
	status *statusError
	stack  string
	cause  error
}

// Option 是一个用于修改 withStatus 错误的函数.
// withStatus 是我们内部用来包装 statusError 的结构体，将在 errors.go 中定义.
type Option func(ws *withStatus)

// Param 用于替换错误消息中的占位符.
// 例如，如果消息是 "user {name} not found"，使用 Param("name", "zampo")
// 将会把消息格式化为 "user zampo not found".
func Param(k, v string) Option {
	return func(ws *withStatus) {
		if ws == nil || ws.status == nil {
			return
		}
		ws.status.message = strings.ReplaceAll(ws.status.message, fmt.Sprintf("{%s}", k), v)
	}
}

// Extra 用于向错误添加额外的键值对元数据.
// 这些数据可以用于日志记录、监控或调试.
func Extra(k, v string) Option {
	return func(ws *withStatus) {
		if ws == nil || ws.status == nil {
			return
		}
		if ws.status.ext.Extra == nil {
			ws.status.ext.Extra = make(map[string]string)
		}
		ws.status.ext.Extra[k] = v
	}
}

// Error 实现 error 接口
func (w *withStatus) Error() string {
	if w.cause != nil {
		return fmt.Sprintf("%s: %v", w.status.message, w.cause)
	}
	return w.status.message
}

// Code 返回错误码
func (w *withStatus) Code() int32 {
	return w.status.statusCode
}

// IsAffectStability 返回是否影响系统稳定性
func (w *withStatus) IsAffectStability() bool {
	return w.status.ext.IsAffectStability
}

// Msg 返回错误消息
func (w *withStatus) Msg() string {
	return w.status.message
}

// Extra 返回扩展信息
func (w *withStatus) Extra() map[string]string {
	if w.status.ext.Extra == nil {
		return make(map[string]string)
	}
	// 复制扩展信息，并添加堆栈信息
	extra := make(map[string]string)
	for k, v := range w.status.ext.Extra {
		extra[k] = v
	}
	if w.stack != "" {
		extra["stack"] = w.stack
	}
	return extra
}

// Unwrap 返回底层的 cause error，用于 errors.Unwrap()
func (w *withStatus) Unwrap() error {
	return w.cause
}

// Stack 返回调用堆栈
func (w *withStatus) Stack() string {
	return w.stack
}

// Cause 返回底层的 cause error
func (w *withStatus) Cause() error {
	return w.cause
}

// WithStack 为 StatusError 添加调用堆栈信息
func WithStack(err StatusError) StatusError {
	if err == nil {
		return nil
	}

	// 如果已经是 withStatus，直接返回
	var ws *withStatus
	if errors.As(err, &ws) {
		return ws
	}

	// 获取调用堆栈
	stack := captureStack(2) // 跳过当前函数和调用者

	// 如果是 statusError，包装为 withStatus
	var se *statusError
	if errors.As(err, &se) {
		return &withStatus{
			status: se,
			stack:  stack,
			cause:  nil,
		}
	}

	// 其他情况，尝试提取底层错误
	return &withStatus{
		status: &statusError{
			statusCode: err.Code(),
			message:    err.Msg(),
			ext: Extension{
				IsAffectStability: err.IsAffectStability(),
				Extra:             err.Extra(),
			},
		},
		stack: stack,
		cause: err,
	}
}

// WrapWithStatus 将一个普通 error 包装为带状态的 StatusError
func WrapWithStatus(err error, code int32, message string, data interface{}) StatusError {
	if err == nil {
		return nil
	}

	// 创建 statusError
	statusErr := NewStatusError(code, message, data)
	stack := captureStack(2) // 跳过当前函数和调用者

	// 尝试提取内部的 statusError
	var se *statusError
	var ws *withStatus
	if errors.As(statusErr, &ws) {
		se = ws.status
	} else if errors.As(statusErr, &se) {
		// se 已经通过 errors.As 赋值
	} else {
		// 如果无法提取，创建一个新的
		def := GetCodeDefinition(code)
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
		se = &statusError{
			statusCode: code,
			message:    statusErr.Msg(),
			ext: Extension{
				IsAffectStability: def.IsAffectStability,
				Extra:             extra,
			},
		}
	}

	return &withStatus{
		status: se,
		stack:  stack,
		cause:  err,
	}
}

// NewWithStatus 创建一个带堆栈的 StatusError，支持 Option 模式
func NewWithStatus(code int32, message string, opts ...Option) StatusError {
	if message == "" {
		message = GetMessage(code, "")
	}

	// 获取错误码定义
	def := GetCodeDefinition(code)

	// 创建 statusError
	se := &statusError{
		statusCode: code,
		message:    message,
		ext: Extension{
			IsAffectStability: def.IsAffectStability,
			Extra:             make(map[string]string),
		},
	}

	// 创建 withStatus
	ws := &withStatus{
		status: se,
		stack:  captureStack(2), // 跳过当前函数和调用者
		cause:  nil,
	}

	// 应用所有 Option
	for _, opt := range opts {
		opt(ws)
	}

	return ws
}

// WrapWithStatusOptions 将一个普通 error 包装为带状态的 StatusError，支持 Option 模式
func WrapWithStatusOptions(err error, code int32, message string, opts ...Option) StatusError {
	if err == nil {
		return nil
	}

	if message == "" {
		message = GetMessage(code, "")
	}

	// 获取错误码定义
	def := GetCodeDefinition(code)

	// 创建 statusError
	se := &statusError{
		statusCode: code,
		message:    message,
		ext: Extension{
			IsAffectStability: def.IsAffectStability,
			Extra:             make(map[string]string),
		},
	}

	// 创建 withStatus
	ws := &withStatus{
		status: se,
		stack:  captureStack(2), // 跳过当前函数和调用者
		cause:  err,
	}

	// 应用所有 Option
	for _, opt := range opts {
		opt(ws)
	}

	return ws
}

// captureStack 捕获调用堆栈
func captureStack(skip int) string {
	var pcs [32]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var lines []string
	for {
		frame, more := frames.Next()
		lines = append(lines, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return strings.Join(lines, "\n")
}
