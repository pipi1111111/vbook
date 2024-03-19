package middlerware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/register" ||
			path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		now := time.Now()
		const updataTimeKey = "update_time"
		//拿出上一次的刷新时间
		val := sess.Get(updataTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			sess.Set(updataTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				log.Println(err)
			}
		}
	}
}
