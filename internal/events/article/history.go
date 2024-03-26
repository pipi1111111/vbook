package article

import (
	"context"
	"github.com/IBM/sarama"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository"
	"vbook/pkg/samarax"
)

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
}

func (h *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", h.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, samarax.NewHandler[ReadEvent](h.Consume))
		if er != nil {
			log.Println(err)
		}
	}()
	return err
}
func (h *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return h.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})
}
