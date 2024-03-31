package repository

import (
	"context"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository/dao"
)

type CornJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error
}
type PreemptJobRepository struct {
	dao dao.CornJobDao
}

func NewCornJobRepository(dao dao.CornJobDao) CornJobRepository {
	return &PreemptJobRepository{dao: dao}
}

func (c *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

func (c *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return c.dao.Release(ctx, jid)
}

func (c *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return c.dao.UpdateUtime(ctx, id)
}

func (c *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error {
	return c.dao.UpdateNextTime(ctx, id, nextTime)
}
