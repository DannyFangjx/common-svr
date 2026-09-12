package router

import (
	"log/slog"

	healthhandler "common-svr/internal/handler/health"
	userhandler "common-svr/internal/handler/user"
	routermiddleware "common-svr/internal/router/middleware"

	"github.com/gin-gonic/gin"
)

func New(logger *slog.Logger, healthHandler *healthhandler.Handler, userHandler *userhandler.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		routermiddleware.TraceID(),
		routermiddleware.AccessLog(logger),
		routermiddleware.Recovery(logger),
		routermiddleware.CORS(),
	)

	registerHealthRoutes(engine, healthHandler)
	registerUserRoutes(engine, userHandler)
	return engine
}
