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
	"context"

	"github.com/go-anyway/framework-log"

	"go.uber.org/zap"
)

// ToGRPCError 将 StatusError 转换为 gRPC error
// 这是一个便捷函数，用于在服务层直接返回 gRPC 错误
func ToGRPCError(err StatusError) error {
	if err == nil {
		return nil
	}
	return ToGRPCStatus(err).Err()
}

// LogAndReturnError 记录错误日志并返回 gRPC error
// 如果 err 是 StatusError，会自动记录包含错误码、消息和扩展信息的日志
// logger 从 ctx 中通过 log.FromContext 获取
func LogAndReturnError(ctx context.Context, err StatusError) error {
	if err == nil {
		return nil
	}

	// 从 context 中获取 logger
	logger := log.FromContext(ctx)

	// 构建日志字段
	fields := []zap.Field{
		zap.Int32("error_code", err.Code()),
		zap.String("error_msg", err.Msg()),
		zap.Bool("affect_stability", err.IsAffectStability()),
	}

	// 添加扩展信息
	extra := err.Extra()
	if len(extra) > 0 {
		fields = append(fields, zap.Any("extra", extra))
	}

	// 根据是否影响稳定性选择日志级别
	if err.IsAffectStability() {
		logger.Error("业务错误（影响稳定性）", fields...)
	} else {
		logger.Warn("业务错误", fields...)
	}

	// 转换为 gRPC error
	return ToGRPCError(err)
}

// WrapAndLogError 包装普通 error 为 StatusError，记录日志并返回 gRPC error
func WrapAndLogError(ctx context.Context, err error, code int32, message string, opts ...Option) error {
	if err == nil {
		return nil
	}

	// 使用 WrapWithStatusOptions 包装错误
	statusErr := WrapWithStatusOptions(err, code, message, opts...)

	// 记录日志并返回
	return LogAndReturnError(ctx, statusErr)
}

// NewAndLogError 创建新的 StatusError，记录日志并返回 gRPC error
func NewAndLogError(ctx context.Context, code int32, message string, opts ...Option) error {
	// 使用 NewWithStatus 创建错误
	statusErr := NewWithStatus(code, message, opts...)

	// 记录日志并返回
	return LogAndReturnError(ctx, statusErr)
}
