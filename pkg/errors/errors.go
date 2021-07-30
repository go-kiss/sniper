package errors

import (
	"sniper/pkg/twirp"

	"github.com/pkg/errors"
)

// NotLoginError 错误未登录
var NotLoginError = twirp.NewError(twirp.Unauthenticated, "must login")

// PermissionDeniedError 权限不够
var PermissionDeniedError = twirp.NewError(twirp.PermissionDenied, "permission denied")

// Wrap 包装错误信息，附加调用栈
// 第二个参数只能是 string，也可以不传，大部分情况不用传
func Wrap(err error, args ...interface{}) error {
	if len(args) >= 1 {
		if msg, ok := args[0].(string); ok {
			return errors.Wrap(err, msg)
		}
	}

	return errors.Wrap(err, "")
}

// Cause 获取原始错误对象
func Cause(err error) error {
	return errors.Cause(err)
}

// Errorf 创建新错误
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// InvalidArgumentError 参数错误，400
func InvalidArgumentError(argument string, validationMsg string) error {
	return twirp.InvalidArgumentError(argument, validationMsg)
}

type codeError struct {
	code int32
	err  string
}

func (c codeError) Error() string {
	return c.err
}

// CodeError 新建业务错误，附带错误码
func CodeError(code int32, err string) error {
	return codeError{code: code, err: err}
}

// Code 提取错误码，codeError 返回 code 和 true，其他返回 0 和 false
func Code(err error) (int32, bool) {
	err = errors.Cause(err)

	if err == nil {
		return 0, false
	}

	if ce, ok := err.(codeError); ok {
		return ce.code, true
	}

	return 0, false
}
