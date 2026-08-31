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
	Host       string `json:"host" form:"host" binding:"required"`
	Port       string `json:"port" form:"port" binding:"required"`
	User       string `json:"user" form:"user"`
	Password   string `json:"password" form:"password" binding:"required"`
	KafkaTopic string `json:"kafka_topic" form:"kafka_topic"`
	Webhook    string `json:"webhook" form:"webhook"`
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
		Host:        input.Host,
		Port:        input.Port,
		Password:    input.Password,
		AccessToken: hashedToken,
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
