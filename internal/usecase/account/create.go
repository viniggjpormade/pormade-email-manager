package account

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type CreateAccountDto struct {
	ImapHost     string  `json:"imap_host" form:"imap_host" binding:"required"`
	ImapPort     string  `json:"imap_port" form:"imap_port" binding:"required"`
	User         string  `json:"user" form:"user"`
	ImapPassword string  `json:"imap_password" form:"imap_password" binding:"required"`
	KafkaTopic   string  `json:"kafka_topic" form:"kafka_topic"`
	Webhook      string  `json:"webhook" form:"webhook"`
	SmtpPort     *string `json:"smtp_port" form:"smtp_port"`
	SmtpHost     *string `json:"smtp_host" form:"smtp_host"`
	SmtpPassword *string `json:"smtp_password" form:"smtp_password"`
}

type CreateUseCase interface {
	Execute(input CreateAccountDto) (*domain.Account, error)
}

type createAccountUseCase struct {
	repository domain.AccountRepository
}

func NewCreateUseCase(repository domain.AccountRepository) CreateUseCase {
	return &createAccountUseCase{
		repository: repository,
	}
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (useCase *createAccountUseCase) Execute(input CreateAccountDto) (*domain.Account, error) {
	rawToken, err := generateSecureToken()
	if err != nil {
		return nil, errors.New("falha ao gerar access token")
	}

	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	account := &domain.Account{
		ImapHost:     input.ImapHost,
		ImapPort:     input.ImapPort,
		ImapPassword: input.ImapPassword,
		SmtpHost:     input.SmtpHost,
		SmtpPort:     input.SmtpPort,
		SmtpPassword: input.SmtpPassword,
		AccessToken:  hashedToken,
	}

	hasKafka := input.KafkaTopic != ""
	hasWebhook := input.Webhook != ""
	if hasKafka == hasWebhook {
		return nil, errors.New("você deve preencher o kafka_topic ou o webhook (apenas um dos dois é permitido)")
	}

	if input.User != "" {
		account.User = &input.User
	}
	if input.KafkaTopic != "" {
		account.KafkaTopic = &input.KafkaTopic
	}
	if input.Webhook != "" {
		account.Webhook = &input.Webhook
	}

	if err := useCase.repository.Create(account); err != nil {
		return nil, err
	}

	account.AccessToken = rawToken

	return account, nil
}
