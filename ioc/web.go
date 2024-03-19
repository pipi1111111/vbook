package ioc

import (
	"github.com/gin-gonic/gin"
	"vbook/internal/web"
)

func InitWeb(userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	userHdl.RegisterRouters(server)
	return server
}
