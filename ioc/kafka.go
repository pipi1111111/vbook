package ioc

import (
	"github.com/IBM/sarama"
	"vbook/internal/events"
	"vbook/internal/events/article"
)

func InitSaramaClient() sarama.Client {
	//type Config struct {
	//	Addr []string `yaml:"addr"`
	//}
	//var cfg Config
	//err := viper.UnmarshalKey("kafka", &cfg)
	//if err != nil {
	//	panic(err)
	//}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, scfg)
	if err != nil {
		panic(err)
	}
	return client
}
func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}
func InitConsumers(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
