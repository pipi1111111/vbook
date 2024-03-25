package repository

import (
	"context"
	"vbook/internal/domain"
)

type ReaderRepository interface {
	// Save 有就更新，没有就插入
	Save(ctx context.Context, article domain.Article) error
}
type readerRepository struct {
}

func (r readerRepository) Save(ctx context.Context, article domain.Article) error {
	//TODO implement me
	panic("implement me")
}

func NewReaderRepository() ReaderRepository {
	return &readerRepository{}
}
