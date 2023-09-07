module common/wechat

go 1.18

replace common/log => ../log

replace common/config => ../config

require (
	common/config v0.0.0
	common/log v0.0.0-00010101000000-000000000000
	github.com/chanxuehong/wechat v0.0.0-20200409104612-0a1fd76d7a3a
)
