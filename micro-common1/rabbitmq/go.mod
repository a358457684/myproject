module common/rabbitmq

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	common/util v0.0.0
	common/redis v0.0.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.0
	github.com/streadway/amqp v1.0.0
	github.com/suiyunonghen/DxCommonLib v0.2.9
	github.com/suiyunonghen/dxsvalue v0.2.1
)

replace (
	common/config => ../config
	common/log => ../log
	common/util => ../util
	common/redis => ../redis
)
