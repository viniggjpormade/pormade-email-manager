package main

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/app"
)

// @title           Pormade Email Manager API
// @version         1.0
// @description     API Para o gerenciamento de emails.
// @host            localhost:8000
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	application := app.NewApp()
	application.Run()
}
