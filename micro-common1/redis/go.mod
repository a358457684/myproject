module common/redis

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	github.com/go-redis/redis/v8 v8.1.3
	github.com/kr/pretty v0.1.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/suiyunonghen/DxCommonLib v0.2.9
    github.com/suiyunonghen/dxsvalue v0.2.1
	golang.org/x/sys v0.0.0-20200615200032-f1bc736245b1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace common/config => ../config

replace common/log => ../log
