package service

import (
	"context"
	"errors"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/events/article"
	"vbook/internal/repository"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}
type articleService struct {
	ar repository.ArticleRepository
	// V1 写法专用
	authorRepo repository.AuthorRepository
	readerRepo repository.ReaderRepository
	producer   article.Producer
}

func NewArticleServiceV1(readerRepo repository.ReaderRepository, authorRepo repository.AuthorRepository) *articleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
	}
}
func NewArticleService(ar repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		ar:       ar,
		producer: producer,
	}
}
func (as *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnPublish
	if article.Id > 0 {
		err := as.ar.Update(ctx, article)
		return article.Id, err
	} else {
		return as.ar.Create(ctx, article)
	}
}
func (as *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return as.ar.Sync(ctx, article)
}

func (as *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		id  = article.Id
		err error
	)
	if article.Id > 0 {
		err = as.authorRepo.Update(ctx, article)
	} else {
		id, err = as.authorRepo.Create(ctx, article)
	}
	if err != nil {
		return 0, err
	}
	article.Id = id
	for i := 0; i < 3; i++ {
		err = as.readerRepo.Save(ctx, article)
		if err != nil {
			log.Println("保存到制作库成功，但是到线上库失败")
		} else {
			return id, nil
		}
	}
	log.Println(err)
	return id, errors.New("保存到线上库失败,重试次数耗尽")
}
func (as *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return as.ar.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}
func (as *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return as.ar.GetByAuthor(ctx, uid, offset, limit)
}
func (as *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return as.ar.GetById(ctx, id)
}
func (as *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := as.ar.GetPubById(ctx, id)
	go func() {
		if err == nil {
			//在这里发一个消息
			er := as.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				log.Println("发送ReadEvent失败", er)
			}
		}
	}()
	return res, err
}
func (as *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return as.ar.ListPub(ctx, start, offset, limit)
}
