package jwt

import "github.com/gin-gonic/gin"

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
}
