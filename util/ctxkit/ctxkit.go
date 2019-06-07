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
	// UserIDKey 用户 ID，未登录则为 0，类型：int64
	UserIDKey
	// UserIPKey 用户 IP，类型：string
	UserIPKey
	// PlatformKey 用户使用平台，ios, android, pc
	PlatformKey
	// BuildKey 客户端构建版本号
	BuildKey
	// VersionKey 客户端版本号
	VersionKey
	// AccessKeyKey 移动端支付令牌
	AccessKeyKey
	// DeviceKey 移动 app 设备标识，android, phone, pad
	DeviceKey
	// MobiAppKey 移动 app 标识，android, phone, pad
	MobiAppKey
	// UserPortKey 用户端口
	UserPortKey
	// ManageUserKey 管理后台用户名
	ManageUserKey
	// BuvidKey 非登录用户标识
	BuvidKey
)

// GetUserID 获取当前登录用户 ID
func GetUserID(ctx context.Context) int64 {
	uid, _ := ctx.Value(UserIDKey).(int64)
	return uid
}

// GetUserIP 获取用户 IP
func GetUserIP(ctx context.Context) string {
	ip, _ := ctx.Value(UserIPKey).(string)
	return ip
}

// GetUserPort 获取用户端口
func GetUserPort(ctx context.Context) string {
	port, _ := ctx.Value(UserPortKey).(string)
	return port
}

// GetPlatform 获取用户平台
func GetPlatform(ctx context.Context) string {
	platform, _ := ctx.Value(PlatformKey).(string)
	return platform
}

// IsIOSPlatform 判断是否为 IOS 平台
func IsIOSPlatform(ctx context.Context) bool {
	return GetPlatform(ctx) == "ios"
}

// GetTraceID 获取用户请求标识
func GetTraceID(ctx context.Context) string {
	id, _ := ctx.Value(TraceIDKey).(string)
	return id
}

// GetBuild 获取客户端构建版本号
func GetBuild(ctx context.Context) string {
	build, _ := ctx.Value(BuildKey).(string)
	return build
}

// GetDevice 获取用户设备，配合 GetPlatform 使用
func GetDevice(ctx context.Context) string {
	device, _ := ctx.Value(DeviceKey).(string)
	return device
}

// GetMobiApp 获取 APP 标识
func GetMobiApp(ctx context.Context) string {
	app, _ := ctx.Value(MobiAppKey).(string)
	return app
}

// GetVersion 获取客户端版本
func GetVersion(ctx context.Context) string {
	version, _ := ctx.Value(VersionKey).(string)
	return version
}

// GetAccessKey 获取客户端版本
func GetAccessKey(ctx context.Context) string {
	key, _ := ctx.Value(AccessKeyKey).(string)
	return key
}

// WithTraceID 注入 trace_id
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetManageUser 获取管理后台用户名
func GetManageUser(ctx context.Context) string {
	user, _ := ctx.Value(ManageUserKey).(string)
	return user
}

// GetBuvid 获取用户 buvid
func GetBuvid(ctx context.Context) string {
	buvid, _ := ctx.Value(BuvidKey).(string)
	return buvid
}
