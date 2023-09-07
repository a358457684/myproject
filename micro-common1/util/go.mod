module common/util

require (
	common/config v0.0.0
	common/log v0.0.0
	common/redis v0.0.0
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/bwmarrin/snowflake v0.3.0
	github.com/disintegration/imaging v1.6.2
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.0
	github.com/suiyunonghen/DxCommonLib v0.2.9
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee
)

go 1.18

replace (
	common/config => ../config
	common/log => ../log
	common/redis => ../redis
)
