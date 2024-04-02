package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"log"
)

var errUnknownPattern = errors.New("未知的双写模式")

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

type DoubleWriteTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
}

func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err1 := d.dst.Commit()
			if err1 != nil {
				log.Println("目标表提交事务失败", err1)
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err1 := d.src.Commit()
			if err1 != nil {
				log.Println("原表提交事务失败", err1)
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err1 := d.dst.Rollback()
			if err1 != nil {
				log.Println("目标表提交事务失败", err1)
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err1 := d.src.Rollback()
			if err1 != nil {
				log.Println("原表提交事务失败", err1)
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	//不是合法的pattern
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst, PatternDstOnly, PatternDstFirst:
		d.pattern.Store(pattern)
		return nil
	default:
		return errUnknownPattern
	}
}
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//这个方法没办法改写 没办法返回一个双写的 sql.Stmt
	panic("双写模式不支持")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				log.Println("双写dst失败", err)
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				log.Println("双写src失败", err)
			}
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		//没有带上错误信息
		//return &sql.Row{}
		panic(errUnknownPattern)
	}
}
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	//要返回的是一个代表双写的事务
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, pattern: pattern}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err1 := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err1 != nil {
			log.Println("双写目标表开启事务失败", err1)
		}
		return &DoubleWriteTx{src: src, dst: dst, pattern: pattern}, nil
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst, pattern: pattern}, err
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err1 := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err1 != nil {
			log.Println("双写原表开启事务失败", err1)
		}
		return &DoubleWriteTx{src: src, dst: dst, pattern: pattern}, nil
	default:
		return nil, errUnknownPattern
	}

}
func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//这个方法没办法改写 没办法返回一个双写的 sql.Stmt
	panic("双写模式不支持")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil || d.dst != nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				log.Println("双写dst失败", err)
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil || d.src != nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				log.Println("双写src失败", err)
			}
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		//没有带上错误信息
		//return &sql.Row{}
		panic(errUnknownPattern)
	}
}
