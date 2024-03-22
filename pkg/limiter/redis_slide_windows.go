package limiter

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaScript string

type RedisSlidingWindowsLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisSlidingWindowsLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlidingWindowsLimiter {
	return &RedisSlidingWindowsLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}
func (b *RedisSlidingWindowsLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
