package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"vbook/internal/domain"
	svcmocks "vbook/internal/service/mocks"
)

func TestBatchRankingService_TopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	testCase := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (InteractiveService, ArticleService)
		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (InteractiveService, ArticleService) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				//先模拟批量获取数据
				//先模拟第一批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 2).Return([]domain.Article{
					{Id: 1, Utime: now},
					{Id: 2, Utime: now},
				}, nil)
				//模拟第二批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, 2).Return([]domain.Article{
					{Id: 3, Utime: now},
					{Id: 4, Utime: now},
				}, nil)
				//模拟第三批 没数据了
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, 2).Return([]domain.Article{}, nil)
				//第一批的点赞数据
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: domain.Interactive{LikeCnt: 1},
					2: domain.Interactive{LikeCnt: 2},
				}, nil)
				//第二批的点赞数据
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{
					3: domain.Interactive{LikeCnt: 3},
					4: domain.Interactive{LikeCnt: 4},
				}, nil)
				return interSvc, artSvc
			},
			wantErr: nil,
			wantArts: []domain.Article{
				{Id: 4, Utime: now},
				{Id: 3, Utime: now},
				{Id: 2, Utime: now},
			},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			interSvc, artSvc := tc.mock(ctrl)
			svc := &BatchRankingService{
				interSvc:  interSvc,
				artSvc:    artSvc,
				batchSize: batchSize,
				n:         3,
				scoreFunc: func(likeCnt int64, ut time.Time) float64 {
					return float64(likeCnt)
				},
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
