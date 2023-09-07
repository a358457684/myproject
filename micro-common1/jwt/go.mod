module common/jwt

go 1.18

require (
	common/log v0.0.0
	common/config v0.0.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.3
)

replace common/log => ../log
replace common/config => ../config
