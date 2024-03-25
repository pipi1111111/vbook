package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"vbook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Set(ctx context.Context, user domain.User) error
	Get(ctx context.Context, uid int64) (domain.User, error)
}
type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
func (r *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := r.Key(du.Id)
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return r.cmd.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := r.Key(uid)
	data, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}
func (r *RedisUserCache) Key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}
