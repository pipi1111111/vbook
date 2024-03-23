package dao

import (
	"context"
	"errors"
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
	UpdateById(ctx context.Context, art Article) error
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
func (ad *GormArticleDao) UpdateById(ctx context.Context, art Article) error {
	res := ad.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"utime":   time.Now().UnixMilli(),
		"title":   art.Title,
		"content": art.Content,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败,Id不对 或者作者不对")
	}
	return nil
}
