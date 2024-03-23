package repository

import (
	"context"
	"vbook/internal/domain"
	"vbook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}
type articleRepository struct {
	ad dao.ArticleDao
}

func NewArticleRepository(ad dao.ArticleDao) ArticleRepository {
	return &articleRepository{
		ad: ad,
	}
}
func (ar *articleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return ar.ad.Insert(ctx, ar.toDao(article))
}
func (ar *articleRepository) toDao(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}
func (ar *articleRepository) Update(ctx context.Context, article domain.Article) error {
	return ar.ad.UpdateById(ctx, ar.toDao(article))
}
