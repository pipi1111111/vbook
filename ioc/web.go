package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
	"vbook/internal/web"
	ijwt "vbook/internal/web/jwt"
	"vbook/internal/web/middlerware"
	"vbook/pkg/ginx/middleware/prometheus"
	"vbook/pkg/ginx/middleware/ratelimit"
	"vbook/pkg/limiter"
)

func InitWeb(mdls []gin.HandlerFunc, userHdl *web.UserHandler, artHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRouters(server)
	artHdl.RegisterRouters(server)
	return server
}
func InitGinMiddleware(redisClient redis.Cmdable, hdl ijwt.Handler) []gin.HandlerFunc {
	pd := prometheus.Builder{
		Namespace: "ahGy",
		Subsystem: "vbook",
		Name:      "git_http",
		Help:      "统计Gin的http接口",
	}
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, ".com")
			},
			MaxAge: 15 * time.Hour,
		}),
		//限流
		pd.BuildResponseTime(),
		pd.BuildActiveRequest(),
		otelgin.Middleware("vbook"),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowsLimiter(redisClient, time.Second, 500)).Build(),
		//middlerware.NewLoginMiddlewareBuilder().CheckLogin(),
		middlerware.NewLoginJwtMiddlewareBuilder(hdl).CheckLogin(),
	}
}
