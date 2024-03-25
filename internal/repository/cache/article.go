package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"vbook/internal/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
}
type ArticleRedis struct {
	cmd redis.Cmdable
}

func NewArticleCache(cmd redis.Cmdable) ArticleCache {
	return &ArticleRedis{
		cmd: cmd,
	}
}
func (a *ArticleRedis) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	//val,err :=a.cmd.Get(ctx,key).Result()
	val, err := a.cmd.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
func (a *ArticleRedis) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := a.firstKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.cmd.Set(ctx, key, val, time.Minute*10).Err()
}
func (a *ArticleRedis) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
func (a *ArticleRedis) DelFirstPage(ctx context.Context, uid int64) error {
	return a.cmd.Del(ctx, a.firstKey(uid)).Err()
}
func (a *ArticleRedis) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.cmd.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
func (a *ArticleRedis) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.cmd.Set(ctx, a.key(art.Id), val, time.Minute*10).Err()
}
func (a *ArticleRedis) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}
func (a *ArticleRedis) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}
func (a *ArticleRedis) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.cmd.Get(ctx, a.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
func (a *ArticleRedis) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.cmd.Set(ctx, a.pubKey(art.Id), val, time.Minute*10).Err()
}
