package main

import (
	"fmt"
	"time"

	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/config"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/database"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/router"
)

// @title           Pormade Email Manager API
// @version         1.0
// @description     API Para o gerenciamento de emails.
// @host            localhost:8000
// @BasePath        /
// @schemes         http
// @tag.name        Auth
// @tag.description Operações relacionadas a autenticação
// @tag.name        User
// @tag.description Operações relacionadas a usuários
func main() {
	fmt.Println("\033[32mIniciado em:\033[39m", "\033[33m", time.Now().Format("2006-01-02 15:04:05"), "\033[39m")
	// gin.SetMode(gin.ReleaseMode)
	config.LoadConfig()
	db := database.DBConnect()
	router.RouterInit(db)
}
