package failover

import (
	"context"
	"sync/atomic"
	"vbook/internal/service/sms"
)

type TimeOutFailOverSMSService struct {
	svcs []sms.Service
	//下标
	idx int32
	//计数
	cnt int32
	//阈值
	threshold int32
}

func (t *TimeOutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			//重置 计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		atomic.AddInt32(&t.cnt, 1)
	}
	return err
}
