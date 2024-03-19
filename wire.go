//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"vbook/internal/repository"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/web"
	"vbook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		dao.NewUserDao,
		repository.NewUserRepository,
		service.NewUserService,
		web.NewUserHandler,
		ioc.InitWeb,
		ioc.InitGinMiddleware,
	)
	return gin.Default()
}
