package utils

// Error Code接口
type IErrorCode interface {
	// 转换为int
	ToCode() int
	// 转换为int
	ToStatus() int
	// 方便与int比较
	Equals(errCode int) bool
	// 获取错误信息
	ErrMsg(args ...interface{}) string

	GetDesc() string
}
