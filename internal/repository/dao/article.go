package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty" `
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	//根据用户Id 来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}
type PublishedArticle Article
type ArticleDao interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, u uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
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
		"status":  art.Status,
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
func (ad *GormArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	tx := ad.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	//防止后面业务panic
	defer tx.Rollback()
	var (
		id  = art.Id
		err error
	)
	dao := NewArticleDao(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	pubArt := PublishedArticle(art)
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   now,
			"status":  pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}
func (ad *GormArticleDao) SyncV1(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := ad.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dao := NewArticleDao(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   now,
			}),
		}).Create(&pubArt).Error
		return err
	})
	return id, err
}
func (ad *GormArticleDao) SyncStatus(ctx context.Context, uid int64, id int64, u uint8) error {
	now := time.Now().UnixMilli()
	return ad.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? and author = ?", uid, id).Updates(map[string]any{
			"utime":  now,
			"status": u,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PublishedArticle{}).Where("id = ?", uid).Updates(map[string]any{
			"utime":  now,
			"status": u,
		}).Error
	})
}

func (ad *GormArticleDao) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := ad.db.WithContext(ctx).Where("author_id = ?", uid).
		Offset(offset).Limit(limit).Order("utime DESC").
		Find(&arts).Error
	return arts, err
}
func (ad *GormArticleDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := ad.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}
func (ad *GormArticleDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubArt PublishedArticle
	err := ad.db.WithContext(ctx).Where("id = ?", id).First(&pubArt).Error
	return pubArt, err
}
func (ad *GormArticleDao) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	err := ad.db.WithContext(ctx).Where("utime<? status = ?", start.UnixMilli(), 2).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
