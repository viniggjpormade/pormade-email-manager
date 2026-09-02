package app

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"

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

	"github.com/robfig/cron/v3"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
)

type App struct {
	engine      *gin.Engine
	cron        *cron.Cron
	kafkaBroker *kafka_broker.KafkaBroker
}

func NewApp() *App {
	config.LoadConfig()
	db := database.DBConnect()

	// Repositories
	accountRepo := postgres.NewAccountRepository(db)
	emailRepo := postgres.NewEmailRepository(db)

	// Providers
	storageProvider := storage.NewLocalStorage("./uploads")
	apiKey := os.Getenv("EMAIL_API_TOKEN")
	emailProvider := gateway.NewExternalEmailProvider(apiKey)

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

	// UseCases
	createAccountUseCase := account.NewCreateUseCase(accountRepo)
	updateAccountUseCase := account.NewUpdateUseCase(accountRepo)
	validateTokenUseCase := account.NewValidateTokenUseCase(accountRepo)

	sendEmailUseCase := email.NewSendEmailUseCase(emailRepo, emailProvider, storageProvider)
	verifyInboxUseCase := email.NewVerifyAndSaveInboxUseCase(emailRepo, storageProvider, kafkaBroker)
	syncEmailStatusUseCase := email.NewSyncEmailStatusUseCase(emailRepo, accountRepo, emailProvider, kafkaBroker)

	// Handlers
	accountHandler := handlers.NewAccountHandler(createAccountUseCase, updateAccountUseCase)
	emailHandler := handlers.NewEmailHandler(sendEmailUseCase)
	authMiddleware := middlewares.EnsureAuth(validateTokenUseCase)

	// Router
	engine := router.NewRouter(accountHandler, emailHandler, authMiddleware)

	emailJobs := cron_adapter.NewEmailJobs(verifyInboxUseCase, syncEmailStatusUseCase, accountRepo)
	cronScheduler := cron_infra.InitScheduler(emailJobs)

	return &App{
		engine:      engine,
		cron:        cronScheduler,
		kafkaBroker: kafkaBroker,
	}
}

func (app *App) Run() {
	fmt.Println("\033[32mIniciado em:\033[39m", "\033[33m", time.Now().Format("2006-01-02 15:04:05"), "\033[39m")

	app.cron.Start()
	defer app.cron.Stop()

	defer app.kafkaBroker.Disconnect()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app.engine.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
