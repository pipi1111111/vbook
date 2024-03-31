package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

type RankingService interface {
	//TopN 前100
	TopN(ctx context.Context) error
	topN(ctx context.Context) ([]domain.Article, error)
	GetTopN(ctx context.Context) ([]domain.Article, error)
}
type BatchRankingService struct {
	//用来去点赞数
	interSvc InteractiveService
	//用来查找文章
	artSvc    ArticleService
	batchSize int
	scoreFunc func(likeCnt int64, ut time.Time) float64
	n         int
	repo      repository.RankingRepository
}

func NewBatchRankingService(interSvc InteractiveService, artSvc ArticleService) RankingService {
	return &BatchRankingService{
		interSvc:  interSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, ut time.Time) float64 {
			duration := time.Since(ut).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}
func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	//最终要放到缓存中
	//存在缓存里面
	return b.repo.ReplaceTopN(ctx, arts)
}
func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)
	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue(b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})
	for {
		//取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return []domain.Article{}, err
		}
		if len(arts) == 0 {
			break
		}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		//取点赞数
		interMap, err := b.interSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return []domain.Article{}, err
		}
		for _, art := range arts {
			inter := interMap[art.Id]
			score := b.scoreFunc(inter.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err := topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				//满了 拿出最小的元素
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(ele)
				}
			}
		}
		offset = offset + len(arts)
		//没有取够一批 我们就直接中断执行 没有下一批了
		if len(arts) < b.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}
	//topN里面就是最终结果
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}
