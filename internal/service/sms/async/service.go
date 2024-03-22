package async

import (
	"context"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository"
	"vbook/internal/service/sms"
)

type Service struct {
	svc sms.Service
	//转异步 存储发短信的请求 repository
	repo repository.AsyncSmsRepository
}

func NewService(svc sms.Service, repo repository.AsyncSmsRepository) *Service {
	return &Service{
		svc:  svc,
		repo: repo,
	}
}

// StartAsyncCycle 异步发消息 抢占式调度
func (s *Service) StartAsyncCycle() {
	time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}
}
func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			log.Println(err)
		}
		res := err == nil
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			log.Println("执行异步发送消息成功 但是标记书记库失败")
		}
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		log.Println("抢占异步发送短信任务失败")
		time.Sleep(time.Second)
	}
}
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if s.needAsync(ctx, tplId, args, numbers...) {
		err := s.repo.Add(ctx, domain.AsyncSms{
			TplId:    tplId,
			Args:     args,
			Numbers:  numbers,
			RetryMax: 3,
		})
		return err
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}

func (s *Service) needAsync(ctx context.Context, tplId string, args []string, numbers ...string) bool {
	var errors []error
	err := s.svc.Send(ctx, tplId, args, numbers...)
	errors = append(errors, err)
	total := len(errors)
	errCount := 0
	for _, err := range errors {
		if err != nil {
			errCount++
		}
	}
	errRate := float64(errCount) / float64(total)
	if errRate > 0.3 {
		return true
	}
	return false
}
