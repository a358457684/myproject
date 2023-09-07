package utils

// 分页数据对象
type PagingResult struct {
	TotalCount int64       `json:"totalCount" xml:"TotalCount"` // 总大小
	PageSize   int64       `json:"pageSize" xml:"PageSize"`     // 分页大小
	PageIndex  int64       `json:"pageIndex" xml:"PageIndex"`   // 页面索引
	Data       interface{} `json:"data" xml:"Result"`           // 分页结果
}
