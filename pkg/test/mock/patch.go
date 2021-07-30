package mock

import (
	"reflect"

	"bou.ke/monkey"
)

func Patch(target, replacement interface{}) *monkey.PatchGuard {
	return monkey.Patch(target, replacement)
}

func PatchInstanceMethod(target reflect.Type, methodName string, replacement interface{}) *monkey.PatchGuard {
	return monkey.PatchInstanceMethod(target, methodName, replacement)
}

func Unpatch(target interface{}) bool {
	return monkey.Unpatch(target)
}

func UnpatchInstanceMethod(target reflect.Type, methodName string) bool {
	return monkey.UnpatchInstanceMethod(target, methodName)
}

func UnpatchAll() {
	monkey.UnpatchAll()
}
