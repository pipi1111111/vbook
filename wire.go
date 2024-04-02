//go:build wireinject

package main

import (
	"github.com/google/wire"
	"vbook/interactive/events"
	repository2 "vbook/interactive/repository"
	cache2 "vbook/interactive/repository/cache"
	dao2 "vbook/interactive/repository/dao"
	service2 "vbook/interactive/service"
	"vbook/internal/events/article"
	"vbook/internal/repository"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/web"
	ijwt "vbook/internal/web/jwt"
	"vbook/ioc"
)

var interactiveSvcSet = wire.NewSet(dao2.NewGormInteractiveDao,
	cache2.NewRedisInteractiveCache,
	repository2.NewCacheInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedis,
	repository.NewRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,
		// DAO 部分
		dao.NewUserDao,
		dao.NewArticleDao,

		interactiveSvcSet,
		ioc.InitIntrClient,
		rankingSvcSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		article.NewSaramaSyncProducer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewArticleRepository,

		// Service 部分
		ioc.InitSmsService,
		//ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		//web.NewOAuth2WechatHandler,
		ioc.InitGinMiddleware,
		ioc.InitWeb,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
