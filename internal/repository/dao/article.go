package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Title   string `gorm:"type=varchar(4096)"`
	Content string `gorm:"type=BLOB"`
	//根据用户Id 来查询
	AuthorId int64 `gorm:"index"`
	Ctime    int64
	Utime    int64
}
type ArticleDao interface {
	Insert(ctx context.Context, article Article) (int64, error)
}
type GormArticleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) ArticleDao {
	return &GormArticleDao{
		db: db,
	}
}
func (ad *GormArticleDao) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Utime = now
	article.Ctime = now
	err := ad.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}
