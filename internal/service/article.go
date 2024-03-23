package service

import (
	"context"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
}
type articleService struct {
	ar repository.ArticleRepository
}

func NewArticleService(ar repository.ArticleRepository) ArticleService {
	return &articleService{
		ar: ar,
	}
}
func (as *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	if article.Id > 0 {
		err := as.ar.Update(ctx, article)
		return article.Id, err
	} else {
		return as.ar.Create(ctx, article)
	}
}
func (as *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}
