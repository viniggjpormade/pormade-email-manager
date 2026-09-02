package router

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/viniggjpormade/pormade-email-manager/docs"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/handlers"
)

func Logger() gin.HandlerFunc {
	return func(context *gin.Context) {
		t := time.Now()

		fmt.Println("Início da request:", context.Request.Method, context.Request.URL.Path)

		context.Next()
		latencia := time.Since(t)
		status := context.Writer.Status()
		fmt.Println("Fim da request:", status, " | tempo:", latencia)
	}
}

func NewRouter(
	accountHandler *handlers.AccountHandler,
	emailHandler *handlers.EmailHandler,
	authMiddleware gin.HandlerFunc,
) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 8388608
	router.Use(Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8000", "http://127.0.0.1:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
	}))

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	accountApi := router.Group("/accounts")
	{
		accountApi.POST("", accountHandler.Create)
		accountApi.PUT("", authMiddleware, accountHandler.Update)
	}

	emailApi := router.Group("/emails")
	{
		emailApi.POST("", authMiddleware, emailHandler.SendEmail)
	}

	return router
}
