module miniapp

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	github.com/chanxuehong/wechat v0.0.0-20200409104612-0a1fd76d7a3a
)

replace (
	common/config => ../config
	common/log => ../log
)
