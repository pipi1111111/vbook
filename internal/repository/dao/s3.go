package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type ArticleS3Dao struct {
	GormArticleDao
	oss *s3.S3
}

type PublishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty" `
	//根据用户Id 来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

func NewArticleS3Dao(db *gorm.DB, oss *s3.S3) *ArticleS3Dao {
	return &ArticleS3Dao{GormArticleDao: GormArticleDao{
		db: db,
	}, oss: oss}
}
func (ad *ArticleS3Dao) Sync(ctx context.Context, art Article) (int64, error) {
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
	pubArt := PublishedArticleV2{
		Id:       art.Id,
		Title:    art.Title,
		AuthorId: art.AuthorId,
		Ctime:    art.Ctime,
		Utime:    art.Utime,
		Status:   art.Status,
	}
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":  pubArt.Title,
			"utime":  now,
			"status": pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	_, err = ad.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("vbook-1314583317"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}
func (ad *ArticleS3Dao) SyncStatus(ctx context.Context, uid int64, id int64, u uint8) error {
	now := time.Now().UnixMilli()
	err := ad.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		return tx.Model(&PublishedArticleV2{}).Where("id = ?", uid).Updates(map[string]any{
			"utime":  now,
			"status": u,
		}).Error
	})
	if err != nil {
		return err
	}
	const statusPrivate = 3
	if u == statusPrivate {
		_, err = ad.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("vbook-1314583317"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}
