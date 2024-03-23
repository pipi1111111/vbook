package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"vbook/internal/domain"
	"vbook/internal/service"
	ijwt "vbook/internal/web/jwt"
)

type ArticleHandler struct {
	as service.ArticleService
}

func NewArticleHandler(as service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		as: as,
	}
}
func (ah *ArticleHandler) RegisterRouters(server *gin.Engine) {
	a := server.Group("/articles")
	a.POST("/edit", ah.Edit)
	a.POST("/publish", ah.Publish)
}

// Edit 接受一个Article输入 返回一个文章的Id
func (ah *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	id, err := ah.as.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: id})
}

func (ah *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	id, err := ah.as.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: id})
}
