//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "vbook/interactive/repository"
	cache2 "vbook/interactive/repository/cache"
	dao2 "vbook/interactive/repository/dao"
	service2 "vbook/interactive/service"
	"vbook/internal/events/article"
	"vbook/internal/job"
	"vbook/internal/repository"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/service/sms"
	"vbook/internal/service/sms/async"
	"vbook/internal/web"
	ijwt "vbook/internal/web/jwt"
	"vbook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	ioc.InitRedis, InitDB,
	ioc.InitSaramaClient,
	ioc.InitSyncProducer,
)

var jobProviderSet = wire.NewSet(
	service.NewCornJobService,
	repository.NewCornJobRepository,
	dao.NewCornJobDaoGorm)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	repository.NewArticleRepository,
	cache.NewArticleCache,
	dao.NewArticleDao,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao2.NewGormInteractiveDao,
	cache2.NewRedisInteractiveCache,
	repository2.NewCacheInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,
		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCodeRepository,

		article.NewSaramaSyncProducer,

		// Service 部分
		ioc.InitSmsService,
		service.NewCodeService,
		//InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		//web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddleware,
		ioc.InitWeb,
	)
	return gin.Default()
}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	wire.Build(thirdPartySet, repository.NewAsyncSmsRepository,
		dao.NewGormAsyncSmsDao,
		async.NewService,
	)
	return &async.Service{}
}

func InitArticleHandler(dao dao.ArticleDao) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		//userSvcProvider,
		interactiveSvcSet,
		repository.NewArticleRepository,
		cache.NewArticleCache,
		service.NewArticleService,
		article.NewSaramaSyncProducer,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service2.NewInteractiveService(nil)
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}
