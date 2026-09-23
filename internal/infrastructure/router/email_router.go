package router

import (
	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/handlers"
)

func RegisterEmailRoutes(router *gin.Engine, handler *handlers.EmailHandler, authMiddleware gin.HandlerFunc) {
	group := router.Group("/emails")
	{
		group.POST("", authMiddleware, handler.SendEmail)
		group.GET("/attachments/:id/file", authMiddleware, handler.DownloadAttachment)
	}
}
