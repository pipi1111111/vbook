package events

import (
	"context"
	"github.com/IBM/sarama"
	"log"
	"time"
	"vbook/interactive/repository"
	"vbook/pkg/samarax"
)

type ReadEvent struct {
	Aid int64
	Uid int64
}

const TopicReadEvent = "article_read"

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, samarax.NewHandler[ReadEvent](i.Consume))
		if er != nil {
			log.Println(err)
		}
	}()
	return err

}

//func (i *InteractiveReadEventConsumer) StartV1() error {
//	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
//	if err != nil {
//		return err
//	}
//	go func() {
//		er := cg.Consume(context.Background(), []string{TopicReadEvent}, samarax.NewHandler[ReadEvent](i.Consume))
//		if er != nil {
//			log.Println(err)
//		}
//	}()
//	return err
//
//}
//
//func (i *InteractiveReadEventConsumer) BatchConsume(msg []*sarama.ConsumerMessage, event []article.ReadEvent) error {
//	bizs := make([]string, 0, len(event))
//	bizId := make([]int64, 0, len(event))
//	for _, evt := range event {
//		bizs = append(bizs, "article")
//		bizId = append(bizId, evt.Aid)
//	}
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//	return i.repo.BatchIncrReadCnt(ctx, bizs, bizId)
//
//}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}
