package utils

import (
	"fmt"
	"strconv"
)

// 计算两个时间之间差，保留小数点后两位
func GetTimeDif(startTime, endTime int64) float64 {
	if endTime == 0 || startTime == 0 || endTime < startTime {
		return 0
	}
	timeLag := float64(endTime-startTime) / (1000 * 60)
	timeLag, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", timeLag), 64)
	return timeLag
}
