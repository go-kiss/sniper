// Package ctxkit 操作请求 ctx 信息
package ctxkit

import (
	"context"
)

type key int

const (
	// TraceIDKey 请求唯一标识，类型：string
	TraceIDKey key = iota
	// StartTimeKey 请求开始时间，类型：time.Time
	StartTimeKey
	// UserIPKey 用户 IP，类型：string
	UserIPKey
	// UserIDKey 用户 ID，未登录则为 0，类型：int64
	UserIDKey
)

// GetTraceID 获取用户请求标识
func GetTraceID(ctx context.Context) string {
	id, _ := ctx.Value(TraceIDKey).(string)
	return id
}

// WithTraceID 注入 trace_id
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetUserIP 获取用户 IP
func GetUserIP(ctx context.Context) string {
	ip, _ := ctx.Value(UserIPKey).(string)
	return ip
}

// GetUserID 获取当前登录用户 ID
func GetUserID(ctx context.Context) int64 {
	uid, _ := ctx.Value(UserIDKey).(int64)
	return uid
}
