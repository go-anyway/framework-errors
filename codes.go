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

// CodeDefinition 定义了错误码的详细信息
type CodeDefinition struct {
	Message           string // 错误消息
	IsAffectStability bool   // 是否影响系统稳定性，可用于告警分级
}

// 业务错误码（使用 int32 以兼容 gRPC）
const (
	CodeSuccess int32 = 200

	// 通用错误 1000-1999
	CodeInvalidParam   int32 = 1001
	CodeUnauthorized   int32 = 1002
	CodeForbidden      int32 = 1003
	CodeNotFound       int32 = 1004
	CodeAlreadyExists  int32 = 1005
	CodeInternalError  int32 = 1006
	CodeRequestTimeout int32 = 1007

	// 业务错误 2000-2999
	CodeUserNotFound      int32 = 2001
	CodeUserAlreadyExist  int32 = 2002
	CodeRateLimitExceeded int32 = 2003
	CodeTokenExpired      int32 = 2004
	// ... 更多业务错误码可以在这里添加
)

// CodeDefinitions 是预定义的错误码及其定义的映射
// 业务可以在自己的包中通过 init() 函数向此 map 添加自定义的错误码
var CodeDefinitions = map[int32]CodeDefinition{
	CodeSuccess: {
		Message:           "success",
		IsAffectStability: false,
	},
	CodeInvalidParam: {
		Message:           "参数无效",
		IsAffectStability: false,
	},
	CodeUnauthorized: {
		Message:           "未授权",
		IsAffectStability: false,
	},
	CodeForbidden: {
		Message:           "禁止访问",
		IsAffectStability: false,
	},
	CodeNotFound: {
		Message:           "资源未找到",
		IsAffectStability: false,
	},
	CodeAlreadyExists: {
		Message:           "资源已存在",
		IsAffectStability: false,
	},
	CodeInternalError: {
		Message:           "内部服务器错误",
		IsAffectStability: true,
	},
	CodeUserNotFound: {
		Message:           "用户不存在",
		IsAffectStability: false,
	},
	CodeUserAlreadyExist: {
		Message:           "用户已存在",
		IsAffectStability: false,
	},
	CodeRateLimitExceeded: {
		Message:           "请求过于频繁",
		IsAffectStability: false,
	},
	CodeTokenExpired: {
		Message:           "认证令牌已过期",
		IsAffectStability: false,
	},
	CodeRequestTimeout: {
		Message:           "请求超时",
		IsAffectStability: false,
	},
}
