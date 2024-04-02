// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"vbook/internal/web/jwt"
	"vbook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitGinMiddleware(cmdable, handler)
	db := InitDB()
	userDao := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	articleDao := dao.NewArticleDao(db)
	articleCache := cache.NewArticleCache(cmdable)
	articleRepository := repository.NewArticleRepository(articleDao, articleCache)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDao := dao2.NewGormInteractiveDao(db)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCacheInteractiveRepository(interactiveDao, interactiveCache)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, interactiveService)
	engine := ioc.InitWeb(v, userHandler, articleHandler)
	return engine
}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	db := InitDB()
	asyncSmsDao := dao.NewGormAsyncSmsDao(db)
	asyncSmsRepository := repository.NewAsyncSmsRepository(asyncSmsDao)
	asyncService := async.NewService(svc, asyncSmsRepository)
	return asyncService
}

func InitArticleHandler(dao3 dao.ArticleDao) *web.ArticleHandler {
	cmdable := ioc.InitRedis()
	articleCache := cache.NewArticleCache(cmdable)
	articleRepository := repository.NewArticleRepository(dao3, articleCache)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	db := InitDB()
	interactiveDao := dao2.NewGormInteractiveDao(db)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCacheInteractiveRepository(interactiveDao, interactiveCache)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, interactiveService)
	return articleHandler
}

func InitInteractiveService() service2.InteractiveService {
	db := InitDB()
	interactiveDao := dao2.NewGormInteractiveDao(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCacheInteractiveRepository(interactiveDao, interactiveCache)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	return interactiveService
}

func InitJobScheduler() *job.Scheduler {
	db := InitDB()
	cornJobDao := dao.NewCornJobDaoGorm(db)
	cornJobRepository := repository.NewCornJobRepository(cornJobDao)
	cornJobService := service.NewCornJobService(cornJobRepository)
	scheduler := job.NewScheduler(cornJobService)
	return scheduler
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitRedis, InitDB, ioc.InitSaramaClient, ioc.InitSyncProducer)

var jobProviderSet = wire.NewSet(service.NewCornJobService, repository.NewCornJobRepository, dao.NewCornJobDaoGorm)

var userSvcProvider = wire.NewSet(dao.NewUserDao, cache.NewUserCache, repository.NewUserRepository, service.NewUserService)

var articleSvcProvider = wire.NewSet(repository.NewArticleRepository, cache.NewArticleCache, dao.NewArticleDao, service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao2.NewGormInteractiveDao, cache2.NewRedisInteractiveCache, repository2.NewCacheInteractiveRepository, service2.NewInteractiveService)
