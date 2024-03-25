package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDao interface {
	Save(ctx context.Context, art Article) error
	SaveV2(ctx context.Context, art PublishedArticle) error
}
type GormArticleReaderDao struct {
	db *gorm.DB
}

func (g *GormArticleReaderDao) SaveV2(ctx context.Context, art PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

func (g *GormArticleReaderDao) Save(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func NewGormArticleReaderDao(db *gorm.DB) ArticleReaderDao {
	return &GormArticleReaderDao{
		db: db,
	}
}
