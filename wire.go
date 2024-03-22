//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"vbook/internal/repository"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/web"
	"vbook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		dao.NewUserDao,
		cache.NewUserCache, cache.NewCodeCache,
		repository.NewUserRepository, repository.NewCodeRepository,
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService, service.NewCodeService,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitWeb,

		ioc.InitGinMiddleware,
	)
	return gin.Default()
}
