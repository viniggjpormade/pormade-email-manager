package router

import (
	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/handlers"
)

func RegisterAccountRoutes(router *gin.Engine, handler *handlers.AccountHandler, authMiddleware gin.HandlerFunc) {
	group := router.Group("/accounts")
	{
		group.POST("", handler.Create)
		group.PATCH("", authMiddleware, handler.Update)
	}
}
