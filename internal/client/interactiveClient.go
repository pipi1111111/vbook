package client

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
	interv1 "vbook/api/proto/gen/inter/v1"
)

type InteractiveClient struct {
	remote    interv1.InteractiveServiceClient
	local     interv1.InteractiveServiceClient
	threshold *atomicx.Value[int32]
}

func (i *InteractiveClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}
func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}
func (i *InteractiveClient) selectClient() interv1.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}
func NewInteractiveClient(remote interv1.InteractiveServiceClient, local interv1.InteractiveServiceClient) *InteractiveClient {
	return &InteractiveClient{remote: remote, local: local, threshold: atomicx.NewValue[int32]()}
}
