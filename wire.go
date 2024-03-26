//go:build wireinject

package main

import (
	"github.com/google/wire"
	"vbook/internal/events/article"
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

func InitWebServer() *App {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitSaramaClient, ioc.InitSyncProducer, ioc.InitConsumers,
		//dao部分
		dao.NewUserDao, dao.NewArticleDao, article.NewSaramaSyncProducer, article.NewInteractiveReadEventConsumer,
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
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
