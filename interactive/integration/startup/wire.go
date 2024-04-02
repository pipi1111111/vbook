//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"vbook/interactive/grpc"
	repository2 "vbook/interactive/repository"
	cache2 "vbook/interactive/repository/cache"
	dao2 "vbook/interactive/repository/dao"
	service2 "vbook/interactive/service"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	//InitSaramaClient,
	//InitSyncProducer,
)

var interactiveSvcSet = wire.NewSet(dao2.NewGormInteractiveDao,
	cache2.NewRedisInteractiveCache,
	repository2.NewCacheInteractiveRepository,
	service2.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
