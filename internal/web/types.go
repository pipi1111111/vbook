package web

import "github.com/gin-gonic/gin"

type Handler interface {
	RegisterRouters(server *gin.Engine)
}
type Page struct {
	Limit  int
	Offset int
}
