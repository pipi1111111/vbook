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

var thirdPartySet = wire.NewSet(
	ioc.InitDB, ioc.InitRedis,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		//dao部分
		dao.NewUserDao, dao.NewArticleDao,
		//cache部分
		cache.NewUserCache, cache.NewCodeCache,
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
	)
	return gin.Default()
}
func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewArticleDao,
		repository.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
