package client

import (
	"context"
	"google.golang.org/grpc"
	interv1 "vbook/api/proto/gen/inter/v1"
	"vbook/interactive/domain"
	"vbook/interactive/service"
)

type InteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveServiceAdapter(svc service.InteractiveService) *InteractiveServiceAdapter {
	return &InteractiveServiceAdapter{svc: svc}
}

func (i *InteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceAdapter) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interv1.LikeResponse{}, err
}

func (i *InteractiveServiceAdapter) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceAdapter) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &interv1.CollectResponse{}, err
}

func (i *InteractiveServiceAdapter) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	inter, err := i.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &interv1.GetResponse{
		Inter: i.toDTO(inter),
	}, nil
}

func (i *InteractiveServiceAdapter) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, in.GetBiz(), in.GetIds())
	if err != nil {
		return nil, err
	}
	inters := make(map[int64]*interv1.Interactive, len(res))
	for k, v := range res {
		inters[k] = i.toDTO(v)
	}
	return &interv1.GetByIdsResponse{
		Inters: inters,
	}, nil
}
func (i *InteractiveServiceAdapter) toDTO(inter domain.Interactive) *interv1.Interactive {
	return &interv1.Interactive{
		Biz:        inter.Biz,
		BizId:      inter.BizId,
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
		Liked:      inter.Liked,
		Collected:  inter.Collected,
	}
}
