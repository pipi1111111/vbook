package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Cfg        string
	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status   int
	Version  int
	NextTime int64 `gorm:"index"`

	Utime int64
	Ctime int64
}

const (
	//没人抢
	jobStatusWaiting = iota
	//已经被人抢了
	jobStatusRunning
	//不在需要调度
	jobStatusPaused
)

type CornJobDao interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error
}

type CornJobDaoGorm struct {
	db *gorm.DB
}

func NewCornJobDaoGorm(db *gorm.DB) CornJobDao {
	return &CornJobDaoGorm{db: db}
}
func (c *CornJobDaoGorm) Preempt(ctx context.Context) (Job, error) {
	db := c.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		err := db.Where("status = ? AND next_time <?",
			jobStatusWaiting, now).
			First(&j).Error
		if err != nil {
			return j, err
		}
		res := db.WithContext(ctx).Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return j, err
	}
}
func (c *CornJobDaoGorm) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return c.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}
func (c *CornJobDaoGorm) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return c.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"utime": now,
	}).Error
}

func (c *CornJobDaoGorm) UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error {
	now := time.Now().UnixMilli()
	return c.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"utime":     now,
		"next_time": nextTime.UnixMilli(),
	}).Error
}
