package log

import "context"

func Trace(ctx context.Context, args ...interface{}) {
	Get(ctx).Trace(args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	Get(ctx).Debug(args...)
}

func Info(ctx context.Context, args ...interface{}) {
	Get(ctx).Info(args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	Get(ctx).Warn(args...)
}

func Error(ctx context.Context, args ...interface{}) {
	Get(ctx).Error(args...)
}

func Fatal(ctx context.Context, args ...interface{}) {
	Get(ctx).Fatal(args...)
}

func Panic(ctx context.Context, args ...interface{}) {
	Get(ctx).Panic(args...)
}

func Tracef(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Tracef(format, args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Debugf(format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Infof(format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Errorf(format, args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Fatalf(format, args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	Get(ctx).Panicf(format, args...)
}
