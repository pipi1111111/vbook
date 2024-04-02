package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"vbook/interactive/repository/dao"
	"vbook/pkg/ginx"
	"vbook/pkg/gormx/connpool"
	"vbook/pkg/migrate/events"
	"vbook/pkg/migrate/events/fixer"
	"vbook/pkg/migrate/scheduler"
)

// InitGinxServer 管理后台的 server
func InitGinxServer(
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer) *ginx.Server {
	engine := gin.Default()
	group := engine.Group("/migrator")
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(client sarama.Client,
	src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
