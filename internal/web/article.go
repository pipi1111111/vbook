package web

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"strconv"
	"time"
	domain2 "vbook/interactive/domain"
	service2 "vbook/interactive/service"
	"vbook/internal/domain"
	"vbook/internal/service"
	ijwt "vbook/internal/web/jwt"
)

type ArticleHandler struct {
	as       service.ArticleService
	interSvc service2.InteractiveService
	biz      string
}

func NewArticleHandler(as service.ArticleService, interSvc service2.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		as:       as,
		interSvc: interSvc,
		biz:      "article",
	}
}
func (ah *ArticleHandler) RegisterRouters(server *gin.Engine) {
	a := server.Group("/articles")
	a.POST("/edit", ah.Edit)
	a.POST("/publish", ah.Publish)
	a.POST("/withdraw", ah.Withdraw)
	//创作者接口
	a.GET("/detail/:id", ah.Detail)
	a.POST("/list", ah.list)
	pub := a.Group("/pub")
	pub.GET("/:id", ah.PubDetail)
	//传入一个参数 true 就是点赞 false就是取消点赞
	pub.POST("/like", ah.like)
	pub.POST("/collect", ah.Collect)
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
		log.Println("发表文章失败")
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: id})
}

func (ah *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	err := ah.as.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("撤回文章失败")
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (ah *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		return
	}
	art, err := ah.as.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("查询文章失败", err)
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	if art.Author.Id != uc.Uid {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("非法查询")
		return
	}
	vo := ArticleVo{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
		Ctime:    art.Ctime.Format(time.DateTime),
		Utime:    art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{Data: vo})
}

func (ah *ArticleHandler) list(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	arts, err := ah.as.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("查找文章列表失败")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				AuthorId: src.Author.Id,
				Status:   uint8(src.Status),
				Ctime:    src.Ctime.Format(time.DateTime),
				Utime:    src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (ah *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		log.Println(err)
		return
	}
	var (
		eg    errgroup.Group
		art   domain.Article
		inter domain2.Interactive
	)
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	eg.Go(func() error {
		var er error
		art, err = ah.as.GetPubById(ctx, id, uc.Uid)
		return er
	})
	eg.Go(func() error {
		uc := ctx.MustGet("user").(ijwt.UserClaims)
		var er error
		inter, err = ah.interSvc.Get(ctx, ah.biz, id, uc.Uid)
		return er
	})
	//等待结果
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		log.Println(err)
		return
	}
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := ah.interSvc.IncrReadCnt(newCtx, ah.biz, art.Id)
		if er != nil {
			log.Println("更新阅读数失败", er)
		}
	}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			ReadCnt:    inter.ReadCnt,
			CollectCnt: inter.CollectCnt,
			LikeCnt:    inter.LikeCnt,
			Liked:      inter.Liked,
			Collected:  inter.Collected,
			Status:     uint8(art.Status),
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
		},
	})

}

func (ah *ArticleHandler) like(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		//true 点赞 false 不点赞
		Like bool `json:"like"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	var err error
	if req.Like {
		//点赞
		err = ah.interSvc.Like(ctx, ah.biz, req.Id, uc.Uid)
	} else {
		//取消点赞
		err = ah.interSvc.CancelLike(ctx, ah.biz, req.Id, uc.Uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("点赞/取消点赞失败", err)
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (ah *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	err := ah.interSvc.Collect(ctx, ah.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println("收藏失败", err)
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}
