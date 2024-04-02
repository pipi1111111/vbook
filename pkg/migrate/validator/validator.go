package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"log"
	"time"
	"vbook/pkg/migrate"
	"vbook/pkg/migrate/events"
)

type Validator[T migrate.Entity] struct {
	// 你数据迁移，是不是肯定有
	base      *gorm.DB
	target    *gorm.DB
	producer  events.Producer
	direction string
	batchSize int
	utime     int64
	// <= 0 就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrate.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	p events.Producer) *Validator[T] {
	res := &Validator[T]{base: base, target: target, producer: p, direction: direction, batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		src, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 你增量校验，要考虑一直运行的
			// 这个就是咩有数据
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错了
			log.Println("base -> target 查询 base 失败", err)
			// 在这里，
			offset++
			continue
		}
		// 这边就是正常情况
		var dst T
		err = v.target.WithContext(ctx).
			Where("id = ?", src.ID()).
			First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target 没有
			// 丢一条消息到 Kafka 上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 要丢一条消息到 Kafka 上
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续
			// 做好监控
			log.Println("base -> target 查询 target 失败", err)
		}
		offset++
	}
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").
		Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Order("utime").
		Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).Select("id").
			//Where("utime > ?", v.utime).
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			log.Println("target => base 查询 target 失败", err)
			offset += len(ts)
			continue
		}
		// 在这里
		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 都代表。base 里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			log.Println("target => base 查询 base 失败", err)
			// 保守起见，我都认为 base 里面没有数据
			// v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		// 找差集，diff 里面的，就是 target 有，但是 base 没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 说明也没了
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		log.Println("发送不一致消息失败", err)
	}
}
