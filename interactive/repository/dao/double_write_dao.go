package dao

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"log"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWriteDao struct {
	src     InteractiveDao
	dst     InteractiveDao
	pattern *atomicx.Value[string]
}

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

func (d *DoubleWriteDao) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}
func (d *DoubleWriteDao) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 要不要 return？
			// 正常来说，我们认为双写阶段，src 成功了就算业务上成功了
			log.Println("双写写入dst 失败", err)
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err == nil {
			err1 := d.src.IncrReadCnt(ctx, biz, bizId)
			if err1 != nil {
				log.Println("双写写入src 失败", err1)
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDao) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		err := d.src.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = d.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			// 要不要 return？
			// 正常来说，我们认为双写阶段，src 成功了就算业务上成功了
			log.Println("双写写入dst 失败", err)
		}
		return nil
	case PatternDstFirst:
		err := d.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err == nil {
			err1 := d.src.DeleteLikeInfo(ctx, biz, id, uid)
			if err1 != nil {
				log.Println("双写写入src 失败", err1)
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, id, uid)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDao) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDao) InsertCollectionBiz(ctx context.Context, cd UserCollectionBiz) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDao) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDao) GetCollectedInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDao) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.Get(ctx, biz, id)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDao) GetV1(ctx context.Context, biz string, id int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intr, err := d.src.Get(ctx, biz, id)
		if err == nil {
			go func() {
				intrDst, err1 := d.dst.Get(ctx, biz, id)
				if err1 != nil {
					if intr != intrDst {
						log.Println(err1)
					}
				}
			}()
		}
		return intr, err
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDao) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDao) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}
