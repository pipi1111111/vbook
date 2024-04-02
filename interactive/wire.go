//go:build wireinject

package main

import (
	"github.com/google/wire"
	"vbook/interactive/events"
	"vbook/interactive/grpc"
	"vbook/interactive/ioc"
	"vbook/interactive/repository"
	"vbook/interactive/repository/cache"
	"vbook/interactive/repository/dao"
	"vbook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitRedis, ioc.InitSaramaClient,
	ioc.InitBizDB,
	ioc.InitDstDB,
	ioc.InitSrcDB,
	ioc.InitDoubleWritePool,
	ioc.InitSaramaSyncProducer,
)
var interactiveSvcSet = wire.NewSet(dao.NewGormInteractiveDao,
	cache.NewRedisInteractiveCache,
	repository.NewCacheInteractiveRepository,
	service.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcXServer,
		ioc.InitGinxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
