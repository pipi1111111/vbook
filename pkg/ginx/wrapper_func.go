package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strconv"
)

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

// WrapBodyAndClaims bizFn 就是你的业务逻辑
func WrapBodyAndClaims[Req any, Claims jwt.Claims](
	bizFn func(ctx *gin.Context, req Req, uc Claims) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Println("输入错误", err)
			return
		}
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			log.Println("执行业务逻辑失败", err)
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[Req any](
	bizFn func(ctx *gin.Context, req Req) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Println("输入错误", err)
			return
		}

		res, err := bizFn(ctx, req)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			log.Println("执行业务逻辑失败", err)
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[Claims any](
	bizFn func(ctx *gin.Context, uc Claims) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			log.Println("执行业务逻辑失败", err)
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			log.Println("处理业务逻辑出错", err)
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}
