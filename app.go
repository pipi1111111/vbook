package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"vbook/internal/events"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
	corn      *cron.Cron
}
