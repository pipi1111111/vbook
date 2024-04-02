package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"log"
	"vbook/interactive/domain"
	"vbook/interactive/repository/cache"
	"vbook/interactive/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	//BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}
type CacheInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
}

func NewCacheInteractiveRepository(dao dao.InteractiveDao, cache cache.InteractiveCache) InteractiveRepository {
	return &CacheInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
func (c *CacheInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//跟新缓存
	//部分失败问题 1.数据不一致
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

//func (c *CacheInteractiveRepository) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
//	err := c.dao.BatchIncrReadCnt(ctx, biz, bizId)
//	if err != nil {
//		return err
//	}
//	go func() {
//		for i := 0; i < len(biz); i++ {
//			er := c.cache.IncrReadCntIfPresent(ctx, biz[i], bizId[i])
//			if er != nil {
//				log.Println(err)
//			}
//		}
//	}()
//	return nil
//}

func (c *CacheInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DeleteLikeCntIfPresent(ctx, biz, id)
}

func (c *CacheInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}
func (c *CacheInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectionIfPresent(ctx, biz, id)
}

func (c *CacheInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	res, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return res, nil
	}
	resDao, err := c.dao.Get(ctx, biz, id)
	if err == nil {
		resDomain := c.toDomain(resDao)
		err = c.cache.Set(ctx, biz, id, resDomain)
		if err != nil {
			log.Println("回写缓存失败")
		}
		return resDomain, nil
	}
	return res, err
}

func (c *CacheInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CacheInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectedInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CacheInteractiveRepository) toDomain(resDao dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      resDao.BizId,
		ReadCnt:    resDao.ReadCnt,
		LikeCnt:    resDao.LikeCnt,
		CollectCnt: resDao.CollectCnt,
	}
}
func (c *CacheInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	inters, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(inters, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}
