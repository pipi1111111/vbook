package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"vbook/internal/web"
	"vbook/internal/web/middlerware"
	"vbook/pkg/limiter"
	"vbook/pkg/middleware/ratelimit"
)

func InitWeb(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRouters(server)
	return server
}
func InitGinMiddleware(redisClient redis.Cmdable) []gin.HandlerFunc {
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
		//限流
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowsLimiter(redisClient, time.Second, 500)).Build(),
		//middlerware.NewLoginMiddlewareBuilder().CheckLogin(),
		middlerware.NewLoginJwtMiddlewareBuilder().CheckLogin(),
	}
}
