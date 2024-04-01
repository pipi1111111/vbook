package service

import (
	"context"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

type CornJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}
type cornJobService struct {
	repo            repository.CornJobRepository
	refreshInterval time.Duration
}

func NewCornJobService(repo repository.CornJobRepository) CornJobService {
	return &cornJobService{repo: repo, refreshInterval: time.Minute}
}

func (c *cornJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.Id)
		if err != nil {
			log.Println(err)
		}
	}
	return j, err
}
func (c *cornJobService) refresh(id int64) {
	//本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		log.Println(err)
	}
}

func (c *cornJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}
