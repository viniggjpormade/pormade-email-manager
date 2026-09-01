package router

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/viniggjpormade/pormade-email-manager/docs"
	"gorm.io/gorm"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		fmt.Println("Início da request:", c.Request.Method, c.Request.URL.Path)

		c.Next()
		latencia := time.Since(t)
		status := c.Writer.Status()
		fmt.Println("Fim da request:", status, " | tempo:", latencia)
	}
}

func RouterInit(db *gorm.DB) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.Default()
	router.MaxMultipartMemory = 8388608
	router.Use(Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
	}))

	// Registra o swagger
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	SetupAccountRoutes(router, db)

	router.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
