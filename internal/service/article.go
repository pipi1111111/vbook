package service

import (
	"context"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
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
	return as.ar.Create(ctx, article)
}
