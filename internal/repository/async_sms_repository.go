package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"vbook/internal/domain"
	"vbook/internal/repository/dao"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSmsRepository interface {
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	Add(ctx context.Context, sms domain.AsyncSms) error
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}
type asyncSmsRepository struct {
	dao dao.AsyncSmsDao
}

func NewAsyncSmsRepository(dao dao.AsyncSmsDao) AsyncSmsRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}
func (a *asyncSmsRepository) Add(ctx context.Context, sms domain.AsyncSms) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   sms.TplId,
				Args:    sms.Args,
				Numbers: sms.Numbers,
			},
			Valid: true,
		},
		RetryMax: sms.RetryMax,
	})
}
func (a *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return domain.AsyncSms{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Numbers:  as.Config.Val.Numbers,
		Args:     as.Config.Val.Args,
		RetryMax: as.RetryMax,
	}, nil
}
func (a *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
