package ratelimit

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"vbook/pkg/limiter"
)

type Builder struct {
	prefix  string
	limiter limiter.Limiter
}

func NewBuilder(limiter limiter.Limiter) *Builder {
	return &Builder{

		prefix:  "ip-limiter",
		limiter: limiter,
	}
}

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limiter.Limit(ctx, fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP()))
		if err != nil {
			log.Println(err)
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			//保守做法：因为借助Redis来做限流，那么Redis崩溃了，为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			//积极做法：虽然Redis崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
			//ctx.Next()
			return
		}
		if limited {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}