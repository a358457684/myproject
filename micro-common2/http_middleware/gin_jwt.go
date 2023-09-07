package http_middleware

import (
	"net/http"
	"strconv"

	"pp/common-golang/ginx"

	"pp/common-golang/contextx"

	"github.com/gin-gonic/gin"
	"pp/common-golang/jws"
	"pp/common-golang/utils"
)

// JWT中间件，检查token，校验不通过报错
func NeedLogin(jwt *jws.JWT, skippers ...SkipperFunc) gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		if SkipHandler(ctxt, skippers...) {
			ctxt.Next()
			return
		}
		token := jwt.GetToken(ctxt)
		if token == "" {
			res := utils.E2(http.StatusUnauthorized, 101, "无权限访问", "")
			panic(res)
			return
		}

		// parseToken 解析token包含的信息
		claims, err := jwt.ParseToken(ctxt, token)
		if err != nil {
			res := utils.E2(http.StatusUnauthorized, 102, "token错误", "")
			panic(res)
			return

		}

		claims.Token = token
		// 继续交由下一个路由处理,并将解析出的信息传递下去
		userId := strconv.FormatInt(claims.UserId, 10)

		ctxt.Set(jws.Claims, claims)
		ginx.SetUserID(ctxt, userId)
		contextx.NewUserID(ctxt.Request.Context(), userId)
		jwt.CreateGroupInfo(ctxt)
		ctxt.Next()
	}
}

// JWT中间件，仅通过token获取载荷信息，校验不通过也不报错
//func GetClaims(jwt *jws.JWT) gin.HandlerFunc {
//	return func(ctxt *gin.Context) {
//		token := jwt.GetToken(ctxt)
//		if token != "" {
//			j := jws.NewJWT()
//			// parseToken 解析token包含的信息
//			claims, err := j.ParseToken(token)
//			if err == nil {
//				claims.Token = token
//				// 继续交由下一个路由处理,并将解析出的信息传递下去
//				ctxt.Set(jws.Claims, claims)
//			} else {
//				claims = &jws.TokenClaims{Token: token}
//			}
//		}
//		ctxt.Next()
//	}
//}
