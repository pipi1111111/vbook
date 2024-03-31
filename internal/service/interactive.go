package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=./mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, id, cid, uid int64) error
	Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error)
}
type interactiveService struct {
	repoInter repository.InteractiveRepository
}

func NewInteractiveService(repoInter repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repoInter: repoInter,
	}
}
func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repoInter.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repoInter.IncrLike(ctx, biz, id, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repoInter.DecrLike(ctx, biz, id, uid)
}
func (i *interactiveService) Collect(ctx context.Context, biz string, id, cid, uid int64) error {
	return i.repoInter.AddCollectionItem(ctx, biz, id, cid, uid)
}

func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	inter, err := i.repoInter.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		inter.Liked, er = i.repoInter.Liked(ctx, biz, id, uid)
		return er
	})
	eg.Go(func() error {
		var er error
		inter.Collected, er = i.repoInter.Collected(ctx, biz, id, uid)
		return er
	})
	return inter, eg.Wait()
}
func (i *interactiveService) GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error) {
	inters, err := i.repoInter.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(inters))
	for _, inter := range inters {
		res[inter.BizId] = inter
	}
	return res, nil
}
