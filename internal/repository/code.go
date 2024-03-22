package repository

import (
	"context"
	"vbook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany = cache.ErrCodeSendTooMany
)

type CodeRepository interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
type codeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &codeRepository{
		cache: cache,
	}
}
func (c *codeRepository) Set(ctx context.Context, biz string, phone string, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}
func (c *codeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, code)
}
