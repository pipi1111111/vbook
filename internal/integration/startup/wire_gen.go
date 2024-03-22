// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"vbook/internal/repository"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
	"vbook/internal/service"
	"vbook/internal/web"
	"vbook/internal/web/jwt"
	"vbook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitGinMiddleware(cmdable, handler)
	db := ioc.InitDB()
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
	articleRepository := repository.NewArticleRepository(articleDao)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(articleService)
	engine := ioc.InitWeb(v, userHandler, articleHandler)
	return engine
}

func InitArticleHandler() *web.ArticleHandler {
	db := ioc.InitDB()
	articleDao := dao.NewArticleDao(db)
	articleRepository := repository.NewArticleRepository(articleDao)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(articleService)
	return articleHandler
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitDB, ioc.InitRedis)
