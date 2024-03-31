package job

import (
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/service"
)

// Executor 执行器，任务执行器
type Executor interface {
	Name() string
	// Exec ctx 这个是全局控制 Executor的实现者要正确处理ctx超时或者取消
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地方法
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: map[string]func(ctx context.Context, j domain.Job) error{}}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}
func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}
func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return errors.New("未注册")
	}
	return fn(ctx, j)
}

type Scheduler struct {
	dbTimeOut time.Duration
	svc       service.CornJobService
	executors map[string]Executor
	limiter   *semaphore.Weighted
}

func NewScheduler(svc service.CornJobService) *Scheduler {
	return &Scheduler{
		svc:       svc,
		dbTimeOut: time.Second,
		executors: map[string]Executor{},
		limiter:   semaphore.NewWeighted(100),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}
func (s *Scheduler) Scheduler(ctx context.Context) {
	for {
		//放弃调度了
		if ctx.Err() != nil {
			return
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return
		}
		dbCtx, cancel := context.WithTimeout(context.Background(), s.dbTimeOut)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			//小一轮
			continue
		}
		//调度执行j
		exec, ok := s.executors[j.Executor]
		if !ok {
			//可以直接中断，也可以下一轮
			log.Println("找不到执行器")
			continue
		}
		go func() {
			defer func() {
				//释放
				s.limiter.Release(1)
				j.CancelFunc()
			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				log.Println("调度任务执行失败")
				return
			}
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				log.Println("更新下一次的执行失败")
			}
		}()

	}
}
