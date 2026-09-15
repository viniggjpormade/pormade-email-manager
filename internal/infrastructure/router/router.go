package router

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/viniggjpormade/pormade-email-manager/docs"
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

func NewRouter() *gin.Engine {
	isProduction, _ := strconv.ParseBool(os.Getenv("PRODUCTION"))
	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	if rawOrigins != "" {
		origins := strings.Split(rawOrigins, ",")
		for _, origin := range origins {
			allowedOrigins = append(allowedOrigins, strings.TrimSpace(origin))
		}
	}

	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 8388608
	router.Use(Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
	}))
	if !isProduction {
		router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}
