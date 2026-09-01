package main

import (
	"fmt"
	"os"
	"time"

	cron_adapter "github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/cron"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/postgres"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/storage"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/config"
	cron_infra "github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/cron"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/database"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/router"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
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
	config.LoadConfig()
	db := database.DBConnect()
	accountRepo := postgres.NewAccountRepository(db)
	emailRepo := postgres.NewEmailRepository(db)

	storagePath := os.Getenv("STORAGE_PATH")
	storageProvider := storage.NewLocalStorage(storagePath)
	verifyInboxUseCase := email.NewVerifyAndSaveInboxUseCase(emailRepo, storageProvider)

	emailJobs := cron_adapter.NewEmailJobs(verifyInboxUseCase, accountRepo)

	scheduler := cron_infra.InitScheduler(emailJobs)
	scheduler.Start()

	router.RouterInit(db)
}
