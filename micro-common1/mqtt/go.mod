module common/mqtt

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.7.0
)

replace common/log => ../log

replace common/config => ../config
