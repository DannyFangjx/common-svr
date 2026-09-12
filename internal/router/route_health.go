package router

import (
	healthhandler "common-svr/internal/handler/health"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(engine *gin.Engine, handler *healthhandler.Handler) {
	engine.GET("/ping", handler.Ping)
	engine.GET("/test", handler.Test)
	engine.GET("/health", handler.Health)
	engine.GET("/ready", handler.Ready)
}
