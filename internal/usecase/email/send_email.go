package email

import (
	"time"

	"github.com/google/uuid"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type SendEmailDTO struct {
	To          string                 `form:"to" json:"to"`
	Subject     string                 `form:"subject" json:"subject"`
	Body        string                 `form:"body" json:"body"`
	Attachments []domain.AttachmentDTO `form:"-" json:"-"`
}

type SendEmailUseCase interface {
	Execute(account domain.Account, input SendEmailDTO) error
}

type sendEmailUseCase struct {
	repository      domain.EmailRepository
	emailProvider   domain.OutboundEmailProvider
	storageProvider domain.StorageProvider
}

func NewSendEmailUseCase(
	repository domain.EmailRepository,
	emailProvider domain.OutboundEmailProvider,
	storageProvider domain.StorageProvider,
) SendEmailUseCase {
	return &sendEmailUseCase{
		repository:      repository,
		emailProvider:   emailProvider,
		storageProvider: storageProvider,
	}
}

func (useCase *sendEmailUseCase) Execute(account domain.Account, input SendEmailDTO) error {
	fromAddress := "no-reply@pormade.com.br"
	if account.User != nil && *account.User != "" {
		fromAddress = *account.User
	}

	response, err := useCase.emailProvider.SendEmail(fromAddress, input.To, input.Subject, input.Body, input.Attachments)
	if err != nil {
		return err
	}

	var attachmentsEntities []domain.Attachment
	for _, attDTO := range input.Attachments {
		fileUrl, err := useCase.storageProvider.Upload(attDTO.Filename, attDTO.Data)
		if err == nil {
			attachmentsEntities = append(attachmentsEntities, domain.Attachment{
				ID:          uuid.New().String(),
				Filename:    attDTO.Filename,
				ContentType: attDTO.ContentType,
				FileUrl:     fileUrl,
			})
		}
	}

	emailEntity := domain.Email{
		ID:          response.ID,
		To:          input.To,
		From:        fromAddress,
		Subject:     input.Subject,
		Body:        &input.Body,
		Date:        time.Now(),
		Status:      response.Status,
		IdAccounts:  account.ID,
		Attachments: attachmentsEntities,
	}

	err = useCase.repository.Create(&emailEntity)
	if err != nil {
		return err
	}

	return nil
}
