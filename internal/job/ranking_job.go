package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
	"vbook/internal/service"
)

type RankingJob struct {
	svc       service.RankingService
	timeOut   time.Duration
	client    *rlock.Client
	key       string
	localLock *sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(svc service.RankingService, timeOut time.Duration, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeOut:   timeOut,
		key:       "job:ranking",
		localLock: &sync.Mutex{},
		client:    client,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		//抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeOut, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		//续约机制
		go func() {
			err := lock.AutoRefresh(r.timeOut/2, r.timeOut)
			if err != nil {
				//续约失败
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	//拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeOut)
	defer cancel()
	return r.svc.TopN(ctx)
}
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

//func (r *RankingJob) Run() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeOut, &rlock.FixIntervalRetry{
//		Interval: time.Millisecond * 100,
//		Max:      3,
//	}, time.Second)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		err := lock.Unlock(ctx)
//		if err != nil {
//			return
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeOut)
//	defer cancel()
//	return r.svc.TopN(ctx)
//}
