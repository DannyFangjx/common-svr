package router

import (
	userhandler "common-svr/internal/handler/user"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(engine *gin.Engine, handler *userhandler.Handler) {
	users := engine.Group("/api/v1/users")
	users.GET("/list", handler.List)
	users.GET("/get", handler.Get)
	users.POST("/create", handler.Create)
	users.POST("/update", handler.Update)
	users.POST("/delete", handler.Delete)
}
