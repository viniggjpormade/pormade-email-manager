package email

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	emailRepo "github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/email"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type SendEmailDTO struct {
	To          []string                `form:"to" json:"to"`
	Subject     string                  `form:"subject" json:"subject"`
	Body        string                  `form:"body" json:"body"`
	InReplyTo   *string                 `form:"inReplyTo" json:"inReplyTo"`
	References  *string                 `form:"references" json:"references"`
	Attachments []domain.AttachmentsDTO `form:"-" json:"-"`
}

type SendEmailUseCase interface {
	Execute(account domain.Account, input SendEmailDTO) error
}

type sendEmailUseCase struct {
	repository       domain.EmailRepository
	emailProvider    domain.OutboundEmailProvider
	storageProvider  domain.StorageProvider
	publishEventCase PublishEventUseCase
}

func NewSendEmailUseCase(
	repository domain.EmailRepository,
	emailProvider domain.OutboundEmailProvider,
	storageProvider domain.StorageProvider,
	publishEventCase PublishEventUseCase,
) SendEmailUseCase {
	return &sendEmailUseCase{
		repository:       repository,
		emailProvider:    emailProvider,
		storageProvider:  storageProvider,
		publishEventCase: publishEventCase,
	}
}

func (useCase *sendEmailUseCase) Execute(account domain.Account, input SendEmailDTO) error {

	params := domain.EmailParams{
		From:        *account.User,
		To:          input.To,
		Subject:     input.Subject,
		Body:        input.Body,
		InReplyTo:   input.InReplyTo,
		References:  input.References,
		Attachments: input.Attachments,
	}
	var response *domain.SendEmailResponse
	var err error

	if account.SmtpHost != nil && account.SmtpPassword != nil && account.SmtpPort != nil {
		username := ""
		if account.User != nil {
			username = *account.User
		}

		connectInfo := domain.ConnectInfo{
			SMTPHost:     account.SmtpHost,
			SMTPPort:     account.SmtpPort,
			SMTPPassword: account.SmtpPassword,
			Username:     username,
		}

		var smtpProvider *emailRepo.SmtpRepository
		smtpProvider, err = emailRepo.NewSmtpRepository(connectInfo)
		if err != nil {
			return fmt.Errorf("erro ao iniciar SMTP: %w", err)
		}

		response, err = smtpProvider.SendEmail(params)
		if err != nil {
			return err
		}
	} else {
		response, err = useCase.emailProvider.SendEmail(params)
		if err != nil {
			return err
		}
	}

	var attachmentsEntities []domain.Attachment
	for _, attachmentDTO := range input.Attachments {
		fileUrl, err := useCase.storageProvider.Upload(attachmentDTO.Filename, attachmentDTO.Data)
		if err == nil {
			attachmentsEntities = append(attachmentsEntities, domain.Attachment{
				ID:          uuid.New().String(),
				Filename:    attachmentDTO.Filename,
				ContentType: attachmentDTO.ContentType,
				FileUrl:     fileUrl,
			})
		}
	}

	var inReplyTo *string
	if input.InReplyTo != nil {
		inReplyTo = input.InReplyTo
	}

	emailEntity := domain.Email{
		ID:          response.ID,
		To:          input.To,
		From:        *account.User,
		Subject:     input.Subject,
		Body:        &input.Body,
		Date:        time.Now(),
		Status:      response.Status,
		InReplyTo:   inReplyTo,
		IdAccounts:  account.ID,
		Attachments: attachmentsEntities,
	}

	err = useCase.repository.Create(&emailEntity)
	if err != nil {
		return err
	}

	var eventAttachments []map[string]interface{}
	for _, attachment := range attachmentsEntities {
		eventAttachments = append(eventAttachments, map[string]interface{}{
			"id":           attachment.ID,
			"filename":     attachment.Filename,
			"content_type": attachment.ContentType,
		})
	}

	useCase.publishEventCase.Execute(account, emailEntity, "EMAIL_QUEUED", eventAttachments)

	return nil
}
