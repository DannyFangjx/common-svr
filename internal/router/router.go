package router

import (
	"log/slog"

	healthhandler "common-svr/internal/handler/health"
	userhandler "common-svr/internal/handler/user"
	routermiddleware "common-svr/internal/router/middleware"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func New(serviceName string, logger *slog.Logger, healthHandler *healthhandler.Handler, userHandler *userhandler.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		routermiddleware.RequestID(),
		otelgin.Middleware(serviceName),
		routermiddleware.AccessLog(logger),
		routermiddleware.Recovery(logger),
		routermiddleware.CORS(),
	)

	registerHealthRoutes(engine, healthHandler)
	registerUserRoutes(engine, userHandler)
	return engine
}
