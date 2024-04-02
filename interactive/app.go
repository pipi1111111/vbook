package main

import (
	"vbook/internal/events"
	"vbook/pkg/ginx"
	"vbook/pkg/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
}
