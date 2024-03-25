package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"vbook/internal/domain"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DeleteLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectionIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, id int64, res domain.Interactive) error
}
type RedisInteractiveCache struct {
	cmd redis.Cmdable
}

func NewRedisInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		cmd: cmd,
	}
}
func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := r.key(biz, bizId)
	_, err := r.cmd.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Int()
	return err
}
func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
func (r *RedisInteractiveCache) DeleteLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := r.key(biz, id)
	_, err := r.cmd.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Int()
	return err
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := r.key(biz, id)
	return r.cmd.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Err()
}
func (r *RedisInteractiveCache) IncrCollectionIfPresent(ctx context.Context, biz string, id int64) error {
	key := r.key(biz, id)
	_, err := r.cmd.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Int()
	return err
}

func (r *RedisInteractiveCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	key := r.key(biz, id)
	res, err := r.cmd.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	var inter domain.Interactive
	inter.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	inter.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	inter.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	return inter, nil
}
func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, id int64, res domain.Interactive) error {
	key := r.key(biz, id)
	err := r.cmd.HSet(ctx, key, fieldCollectCnt, res.CollectCnt, fieldReadCnt, res.ReadCnt, fieldLikeCnt, res.LikeCnt).Err()
	if err != nil {
		return err
	}
	return r.cmd.Expire(ctx, key, time.Minute*20).Err()
}
