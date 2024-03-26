package main

import (
	"github.com/gin-gonic/gin"
	"vbook/internal/events"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}
