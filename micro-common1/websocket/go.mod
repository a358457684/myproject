module common/websocket

go 1.18

require (
	common/log v0.0.0
	common/util v0.0.0
	github.com/gorilla/websocket v1.4.2
)

replace (
	common/config => ../config
	common/log => ../log
	common/redis => ../redis
	common/util => ../util
)
