module common/biz

go 1.18

require (
	common/log v0.0.0
	common/mqtt v0.0.0
	common/orm v0.0.0
	common/rabbitmq v0.0.0
	common/redis v0.0.0
	common/util v0.0.0
	github.com/go-redis/redis/v8 v8.1.3
	github.com/jmoiron/sqlx v1.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/suiyunonghen/DxCommonLib v0.2.9
	github.com/suiyunonghen/dxsvalue v0.2.1
	gorm.io/gorm v1.20.12
)

replace (
	//eps_common/huping/net_resuse => ../../eps_common/huping/net_resuse
	common/config => ../config
	common/log => ../log
	common/mqtt => ../mqtt
	common/orm => ../orm
	common/rabbitmq => ../rabbitmq
	common/redis => ../redis
	common/util => ../util
)
