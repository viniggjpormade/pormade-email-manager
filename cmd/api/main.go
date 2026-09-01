package main

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kafka_broker "github.com/viniggjpormade/pormade-email-manager/internal/adapter/broker/kafka"
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
func main() {
	fmt.Println("\033[32mIniciado em:\033[39m", "\033[33m", time.Now().Format("2006-01-02 15:04:05"), "\033[39m")
	config.LoadConfig()
	db := database.DBConnect()
	accountRepo := postgres.NewAccountRepository(db)
	emailRepo := postgres.NewEmailRepository(db)

	storageProvider := storage.NewLocalStorage("./uploads")

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_SERVERS"),
		"client.id":         os.Getenv("KAFKA_CLIENT_ID"),
		"acks":              "1",
		// "enable.idempotence":                    true,
		"retries":                               2147483647,
		"reconnect.backoff.ms":                  30000,
		"max.in.flight.requests.per.connection": 1,
		"linger.ms":                             5,
		"batch.num.messages":                    10000,
		"reconnect.backoff.max.ms":              30000,
	}
	kafkaBroker, err := kafka_broker.NewKafkaBroker(kafkaConfig)
	if err != nil {
		fmt.Printf("Aviso: Falha ao iniciar Kafka Broker: %v\n", err)
	}

	verifyInboxUseCase := email.NewVerifyAndSaveInboxUseCase(emailRepo, storageProvider, kafkaBroker)

	emailJobs := cron_adapter.NewEmailJobs(verifyInboxUseCase, accountRepo)

	scheduler := cron_infra.InitScheduler(emailJobs)
	scheduler.Start()

	router.RouterInit(db)
}
