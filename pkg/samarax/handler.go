package samarax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
)

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, event T) error
}

func NewHandler[T any](fn func(msg *sarama.ConsumerMessage, event T) error) *Handler[T] {
	return &Handler[T]{fn: fn}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		//在这里调用业务处理逻辑
		var t T
		err := json.Unmarshal(msg.Value, t)
		if err != nil {
			//你也可以在这里引入重试的逻辑
			log.Println(err)
		}
		err = h.fn(msg, t)
		if err != nil {
			log.Println(err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
