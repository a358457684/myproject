package util

import "runtime"

// 获取正在运行的函数名
func RunFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

// 获取上一个运行的函数名
func FuncName(skip int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}
