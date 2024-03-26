package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDao interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cd UserCollectionBiz) error
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectedInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
}
type GormInteractiveDao struct {
	db *gorm.DB
}

func NewGormInteractiveDao(db *gorm.DB) InteractiveDao {
	return &GormInteractiveDao{
		db: db,
	}
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}
type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Utime  int64
	Ctime  int64
	Status int
}
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//收藏夹的Id
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

func (g *GormInteractiveDao) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (g *GormInteractiveDao) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDao := NewGormInteractiveDao(tx)
		for i := 0; i < len(biz); i++ {
			err := txDao.IncrReadCnt(ctx, biz[i], bizId[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GormInteractiveDao) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).Where("uid = ? AND biz_id = ? ADN biz = ?", uid, id, biz).Updates(map[string]interface{}{
			"status": 0,
			"utime":  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz = ? ADN biz_id = ?", biz, id).Updates(map[string]interface{}{
			"like_cnt": gorm.Expr("like_cnt - 1"),
			"utime":    now,
		}).Error
	})
}

func (g *GormInteractiveDao) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status": 1,
				"utime":  now,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Ctime:  now,
			Utime:  now,
			Biz:    biz,
			BizId:  id,
			Status: 1,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (g *GormInteractiveDao) InsertCollectionBiz(ctx context.Context, cd UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cd.Ctime = now
	cd.Utime = now
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cd).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("collect_cnt + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cd.Biz,
			BizId:      cd.BizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

func (g *GormInteractiveDao) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := g.db.WithContext(ctx).Where("biz = ? AND biz_id =? AND uid = ? AND status =?", biz, id, uid, 1).First(&res).Error
	return res, err
}

func (g *GormInteractiveDao) GetCollectedInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := g.db.WithContext(ctx).Where("biz = ? AND biz_id =? AND uid = ?", biz, id, uid).First(&res).Error
	return res, err
}

func (g *GormInteractiveDao) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := g.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, id).First(&res).Error
	return res, err
}
