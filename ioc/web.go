package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"vbook/internal/web"
	"vbook/internal/web/middlerware"
)

func InitWeb(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRouters(server)
	return server
}
func InitGinMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, ".com")
			},
			MaxAge: 15 * time.Hour,
		}),
		//middlerware.NewLoginMiddlewareBuilder().CheckLogin(),
		middlerware.NewLoginJwtMiddlewareBuilder().CheckLogin(),
	}
}
