package main

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	kafka_broker "github.com/viniggjpormade/pormade-email-manager/internal/adapter/broker/kafka"
	cron_adapter "github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/cron"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/handlers"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/middlewares"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/gateway"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/postgres"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/storage"

	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/config"
	cron_infra "github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/cron"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/database"
	"github.com/viniggjpormade/pormade-email-manager/internal/infrastructure/router"

	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
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
	// Config & Database
	config.LoadConfig()
	db := database.DBConnect()

	// Repositories
	accountRepo := postgres.NewAccountRepository(db)
	emailRepo := postgres.NewEmailRepository(db)
	webhookLogRepo := postgres.NewWebhookLogRepository(db)

	// Providers
	storageProvider := storage.NewLocalStorage("./uploads")
	emailProvider := gateway.NewExternalEmailProvider(os.Getenv("EMAIL_API_TOKEN"))

	// Kafka Broker
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_SERVERS"),
		"client.id":         os.Getenv("KAFKA_CLIENT_ID"),
		"group.id":          "pormade-email-manager-group",
		"auto.offset.reset": "earliest",
	}
	kafkaBroker, err := kafka_broker.NewKafkaBroker(kafkaConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Kafka broker: %v", err))
	}
	defer kafkaBroker.Disconnect()

	// UseCases
	createAccountUseCase := account.NewCreateUseCase(accountRepo)
	updateAccountUseCase := account.NewUpdateUseCase(accountRepo)
	validateTokenUseCase := account.NewValidateTokenUseCase(accountRepo)

	publishEventUseCase := email.NewPublishEventUseCase(webhookLogRepo, kafkaBroker)
	sendEmailUseCase := email.NewSendEmailUseCase(emailRepo, emailProvider, storageProvider, publishEventUseCase)
	verifyInboxUseCase := email.NewVerifyAndSaveInboxUseCase(emailRepo, storageProvider, publishEventUseCase)
	syncEmailStatusUseCase := email.NewSyncEmailStatusUseCase(emailRepo, accountRepo, emailProvider, kafkaBroker)
	retryFailedEventsUseCase := email.NewRetryFailedEventsUseCase(webhookLogRepo, accountRepo, kafkaBroker)

	// Handlers & Middleware
	accountHandler := handlers.NewAccountHandler(createAccountUseCase, updateAccountUseCase)
	emailHandler := handlers.NewEmailHandler(sendEmailUseCase, emailRepo)
	authMiddleware := middlewares.EnsureAuth(validateTokenUseCase)

	// Router
	engine := router.NewRouter()
	router.RegisterAccountRoutes(engine, accountHandler, authMiddleware)
	router.RegisterEmailRoutes(engine, emailHandler, authMiddleware)

	emailJobs := cron_adapter.NewEmailJobs(accountRepo, sendEmailUseCase, syncEmailStatusUseCase, retryFailedEventsUseCase, verifyInboxUseCase)
	cronScheduler := cron_infra.InitScheduler(emailJobs)
	cronScheduler.Start()
	defer cronScheduler.Stop()

	// Start Server
	fmt.Println("\033[32mIniciado em:\033[39m", "\033[33m", time.Now().Format("2006-01-02 15:04:05"), "\033[39m")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	engine.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
