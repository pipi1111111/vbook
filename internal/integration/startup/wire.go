//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"vbook/internal/repository"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/web"
	ijwt "vbook/internal/web/jwt"
	"vbook/ioc"
)

var interactiveSvsSet = wire.NewSet(dao.NewGormInteractiveDao,
	cache.NewRedisInteractiveCache,
	repository.NewCacheInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		//dao部分
		dao.NewUserDao, dao.NewArticleDao,
		//cache部分
		cache.NewUserCache, cache.NewCodeCache, cache.NewArticleCache,
		//repository部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewArticleRepository,
		//service部分
		ioc.InitSmsService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		//handler部分
		ijwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewArticleHandler,

		ioc.InitWeb,
		ioc.InitGinMiddleware,
		interactiveSvsSet,
	)
	return gin.Default()
}
