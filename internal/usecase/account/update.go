package account

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type UpdateAccountDto struct {
	ImapHost     *string `json:"imap_host" form:"imap_host"`
	ImapPort     *string `json:"imap_port" form:"imap_port"`
	User         *string `json:"user" form:"user"`
	ImapPassword *string `json:"imap_password" form:"imap_password"`
	KafkaTopic   *string `json:"kafka_topic" form:"kafka_topic"`
	Webhook      *string `json:"webhook" form:"webhook"`
	SmtpPort     *string `json:"smtp_port" form:"smtp_port"`
	SmtpHost     *string `json:"smtp_host" form:"smtp_host"`
	SmtpPassword *string `json:"smtp_password" form:"smtp_password"`
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
	if input.ImapHost != nil {
		account.ImapHost = *input.ImapHost
	}
	if input.ImapPort != nil {
		account.ImapPort = *input.ImapPort
	}
	if input.User != nil {
		account.User = input.User
	}
	if input.ImapPassword != nil {
		account.ImapPassword = *input.ImapPassword
	}
	if input.SmtpHost != nil {
		account.SmtpHost = input.SmtpHost
	}
	if input.SmtpPort != nil {
		account.SmtpPort = input.SmtpPort
	}
	if input.SmtpPassword != nil {
		account.SmtpPassword = input.SmtpPassword
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
