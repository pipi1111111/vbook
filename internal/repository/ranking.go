package repository

import (
	"context"
	"vbook/internal/domain"
	"vbook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}
type CachedRankingRepository struct {
	cache      cache.RankingCache
	redisCache *cache.RankingRedis
	localCache *cache.RankingLocalCache
}

func NewCachedRankingRepositoryV1(redisCache *cache.RankingRedis, localCache *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: redisCache, localCache: localCache}
}

func NewRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}
func (c *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}
func (c *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := c.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = c.redisCache.Get(ctx)
	if err != nil {
		return c.localCache.FalseGet(ctx)
	}
	_ = c.localCache.Set(ctx, res)
	return res, nil
}
