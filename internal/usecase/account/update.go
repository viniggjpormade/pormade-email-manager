package account

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type UpdateAccountDto struct {
	Host       *string `json:"host" form:"host"`
	Port       *string `json:"port" form:"port"`
	User       *string `json:"user" form:"user"`
	Password   *string `json:"password" form:"password"`
	KafkaTopic *string `json:"kafka_topic" form:"kafka_topic"`
	Webhook    *string `json:"webhook" form:"webhook"`
}

type UpdateUseCase interface {
	Execute(account domain.Account, input UpdateAccountDto) (*domain.Account, error)
}

type updateUseCase struct {
	repository domain.AccountRepository
}

func NewUpdateUseCase(repository domain.AccountRepository) UpdateUseCase {
	return &updateUseCase{
		repository: repository,
	}
}

func (useCase *updateUseCase) Execute(account domain.Account, input UpdateAccountDto) (*domain.Account, error) {
	if input.Host != nil {
		account.Host = *input.Host
	}
	if input.Port != nil {
		account.Port = *input.Port
	}
	if input.User != nil {
		account.User = input.User
	}
	if input.Password != nil {
		account.Password = *input.Password
	}
	if input.KafkaTopic != nil {
		account.KafkaTopic = input.KafkaTopic
	}
	if input.Webhook != nil {
		account.Webhook = input.Webhook
	}

	err := useCase.repository.Update(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
