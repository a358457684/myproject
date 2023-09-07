package manager

import (
	"common/biz/enum"
	"fmt"
	"testing"
)

func TestFunctionNum_IsFuncIndexOpen(t *testing.T) {
	f := RobotFunction(0)
	f.Open(enum.RfTest)
	if f.IsOpen(enum.RfTest) {
		fmt.Println(enum.RfTest.String() + "打开了")
		f.Close(enum.RfTest)
		if !f.IsOpen(enum.RfTest) {
			fmt.Println(enum.RfTest.String() + "关闭了")
		}
	}
	f.Open(1)
	f.Open(2)
	fmt.Println(f)
}
