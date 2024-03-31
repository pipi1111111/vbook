package cache

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
	"vbook/internal/domain"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	r.topN.Store(arts)
	r.ddl.Store(time.Now().Add(r.expiration))
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	arts := r.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存失败了")
	}
	return arts, nil
}
func (r *RankingLocalCache) FalseGet(ctx context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	if len(arts) == 0 {
		return nil, errors.New("本地缓存失败了")
	}
	return arts, nil
}
