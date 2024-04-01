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
	ioc.InitRedis, ioc.InitDB, ioc.InitSaramaClient,
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
		ioc.InitConsumers,
		ioc.NewGrpcXServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
