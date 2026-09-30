package email

import (
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
	Execute(account domain.Account, input SendEmailDTO) (string, error)
	ProcessQueuedEmails() error
}

type sendEmailUseCase struct {
	repository       domain.EmailRepository
	emailProvider    domain.OutboundEmailProvider
	storageProvider  domain.StorageProvider
	publishEventCase PublishEventUseCase
	semaphore        chan struct{}
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
		semaphore:        make(chan struct{}, 10),
	}
}

func (useCase *sendEmailUseCase) Execute(account domain.Account, input SendEmailDTO) (string, error) {
	emailId := uuid.New().String()

	var attachmentsEntities []domain.Attachment
	var eventAttachments []map[string]interface{}

	for _, attachmentDTO := range input.Attachments {
		fileUrl, err := useCase.storageProvider.Upload(attachmentDTO.Filename, attachmentDTO.Data)
		if err == nil {
			attachmentId := uuid.New().String()
			attachmentsEntities = append(attachmentsEntities, domain.Attachment{
				ID:          attachmentId,
				Filename:    attachmentDTO.Filename,
				ContentType: attachmentDTO.ContentType,
				FileUrl:     fileUrl,
			})
			eventAttachments = append(eventAttachments, map[string]interface{}{
				"id":           attachmentId,
				"filename":     attachmentDTO.Filename,
				"content_type": attachmentDTO.ContentType,
			})
		}
	}

	var inReplyTo *string
	if input.InReplyTo != nil {
		inReplyTo = input.InReplyTo
	}

	emailEntity := domain.Email{
		ID:          emailId,
		To:          input.To,
		From:        *account.User,
		Subject:     input.Subject,
		Body:        &input.Body,
		Date:        time.Now(),
		Status:      string(domain.EmailStatusQueued),
		InReplyTo:   inReplyTo,
		IdAccounts:  account.ID,
		Attachments: attachmentsEntities,
	}

	err := useCase.repository.Create(&emailEntity)
	if err != nil {
		return emailId, err
	}

	useCase.publishEventCase.Execute(account, emailEntity, "EMAIL_QUEUED", eventAttachments)

	params := domain.EmailParams{
		From:        *account.User,
		To:          input.To,
		Subject:     input.Subject,
		Body:        input.Body,
		InReplyTo:   input.InReplyTo,
		References:  input.References,
		Attachments: input.Attachments,
	}

	go useCase.processSend(emailEntity, params, account)

	return nil
}

func (useCase *sendEmailUseCase) processSend(emailEntity domain.Email, params domain.EmailParams, account domain.Account) {
	useCase.semaphore <- struct{}{}
	defer func() { <-useCase.semaphore }()

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
		if err == nil {
			response, err = smtpProvider.SendEmail(params)
		}
	} else {
		response, err = useCase.emailProvider.SendEmail(params)
	}

	if err != nil {
		emailEntity.Status = string(domain.EmailStatusDeferred)
	} else {
		emailEntity.Status = string(domain.EmailStatusSent)
		if response != nil && response.ID != "" {
			emailEntity.References = &response.ID
		}
	}

	useCase.repository.Update(&emailEntity)

	statusPayload := []map[string]interface{}{}
	useCase.publishEventCase.Execute(account, emailEntity, "EMAIL_STATUS_UPDATED", statusPayload)
}

func (useCase *sendEmailUseCase) ProcessQueuedEmails() error {
	queuedEmails, err := useCase.repository.GetQueuedEmails()
	if err != nil {
		return err
	}

	for _, emailEntity := range queuedEmails {
		if time.Since(emailEntity.Date) > 5*time.Minute {
			params := domain.EmailParams{
				From:    emailEntity.From,
				To:      emailEntity.To,
				Subject: emailEntity.Subject,
			}
			if emailEntity.Body != nil {
				params.Body = *emailEntity.Body
			}
			if emailEntity.InReplyTo != nil {
				params.InReplyTo = emailEntity.InReplyTo
			}
			if emailEntity.References != nil {
				params.References = emailEntity.References
			}

			go useCase.processSend(emailEntity, params, emailEntity.Account)
		}
	}
	return nil
}
