package middlerware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	ijwt "vbook/internal/web/jwt"
)

type LoginJwtMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJwtMiddlewareBuilder(hdl ijwt.Handler) *LoginJwtMiddlewareBuilder {
	return &LoginJwtMiddlewareBuilder{
		Handler: hdl,
	}
}
func (m *LoginJwtMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/register" ||
			path == "/users/login" ||
			path == "/users/loginSms" ||
			path == "/users/sendSms" ||
			path == "/oauth2wechat/authurl" ||
			path == "/oauth2wechat/callback" {
			return
		}
		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			//进来这里大概率是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token 无效或者 redis 有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//expireTime := uc.ExpiresAt
		//if expireTime.Sub(time.Now()) < time.Minute*20 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
		//	newToken, err := token.SignedString(web.JWTKey)
		//	if err != nil {
		//		log.Println(err)
		//	} else {
		//		ctx.Header("x-jwt-token", newToken)
		//	}
		//}
		ctx.Set("user", uc)
	}
}
