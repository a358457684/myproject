module common/mail

go 1.18

replace common/log => ../log

replace common/config => ../config

require (
	common/config v0.0.0
	common/log v0.0.0-00010101000000-000000000000
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)
