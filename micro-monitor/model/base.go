package model

import (
	"time"
)

type BaseModel struct {
	CreateBy   string    // 创建者
	CreateDate time.Time // 创建时间
	UpdateBy   string    // 更新者
	UpdateDate time.Time // 更新时间
	Remarks    string    // 备注信息
	DelFlag    string    // 删除标记
}

const (
	DelFlagNormal = "0"
)
