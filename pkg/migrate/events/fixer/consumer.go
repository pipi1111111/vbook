package fixer

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log"
	"time"
	"vbook/pkg/migrate"
	"vbook/pkg/migrate/events"
	"vbook/pkg/migrate/fixer"
	"vbook/pkg/samarax"
)

type Consumer[T migrate.Entity] struct {
	client   sarama.Client
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrate.Entity](
	client sarama.Client,
	topic string,
	src *gorm.DB,
	dst *gorm.DB) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

// Start 这边就是自己启动 goroutine 了
func (r *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{r.topic},
			samarax.NewHandler[events.InconsistentEvent](r.Consume))
		if err != nil {
			log.Println("退出了消费循环异常", err)
		}
	}()
	return err
}

func (r *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return r.srcFirst.Fix(ctx, t.ID)
	case "DST":
		return r.dstFirst.Fix(ctx, t.ID)
	}
	return errors.New("未知的校验方向")
}
