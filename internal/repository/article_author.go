package repository

import (
	"context"
	"vbook/internal/domain"
)

type AuthorRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}
type authorRepository struct {
}

func (ar *authorRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	panic("implement me")
}
func (a *authorRepository) Update(ctx context.Context, article domain.Article) error {
	//TODO implement me
	panic("implement me")
}
