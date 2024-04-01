package grpc

import (
	"context"
	"google.golang.org/grpc"
	interv1 "vbook/api/proto/gen/inter/v1"
	"vbook/interactive/domain"
	"vbook/interactive/service"
)

type InteractiveServiceServer struct {
	interv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *interv1.IncrReadCntRequest) (*interv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &interv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *interv1.LikeRequest) (*interv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *interv1.CancelLikeRequest) (*interv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *interv1.CollectRequest) (*interv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &interv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *interv1.GetRequest) (*interv1.GetResponse, error) {
	inter, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interv1.GetResponse{
		Inter: i.toDTO(inter),
	}, nil
}
func (i *InteractiveServiceServer) Register(s *grpc.Server) {
	interv1.RegisterInteractiveServiceServer(s, i)
}
func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *interv1.GetByIdsRequest) (*interv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetIds())
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
func (i *InteractiveServiceServer) toDTO(inter domain.Interactive) *interv1.Interactive {
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
