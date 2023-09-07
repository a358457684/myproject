module micro-monitor

go 1.18

require (
	eps_common/huping/net_resuse v0.0.0
	micro-common1/biz v0.0.0
	micro-common1/config v0.0.0
	micro-common1/elasticsearch v0.0.0
	micro-common1/log v0.0.0
	micro-common1/mqtt v0.0.0
	micro-common1/orm v0.0.0
	micro-common1/rabbitmq v0.0.0
	micro-common1/redis v0.0.0
	micro-common1/util v0.0.0
	micro-common1/websocket v0.0.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-openapi/spec v0.19.11 // indirect
	github.com/go-openapi/swag v0.19.11 // indirect
	github.com/go-redis/redis/v8 v8.1.3
	github.com/gorilla/websocket v1.4.2
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.0
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.6.9
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392
	golang.org/x/net v0.0.0-20201029055024-942e2f445f3c // indirect
	golang.org/x/tools v0.0.0-20201028224754-2c115999a7f0 // indirect
	gorm.io/gorm v1.20.12
)

replace (
	eps_common/huping/net_resuse => ../eps_common/huping/net_resuse
	micro-common1/biz => ../micro-common1/biz
	micro-common1/config => ../micro-common1/config
	micro-common1/elasticsearch => ../micro-common1/elasticsearch
	micro-common1/log => ../micro-common1/log
	micro-common1/mqtt => ../micro-common1/mqtt
	micro-common1/orm => ../micro-common1/orm
	micro-common1/rabbitmq => ../micro-common1/rabbitmq
	micro-common1/redis => ../micro-common1/redis
	micro-common1/util => ../micro-common1/util
	micro-common1/websocket => ../micro-common1/websocket
)
