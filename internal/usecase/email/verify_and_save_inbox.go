package email

import (
	"log"

	"github.com/google/uuid"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type VerifyAndSaveInboxUseCase interface {
	Execute(account domain.Account, imapProvider domain.ImapEmailProvider) error
}

type verifyAndSaveInboxUseCase struct {
	repository       domain.EmailRepository
	storageProvider  domain.StorageProvider
	publishEventCase PublishEventUseCase
}

type emailPayload struct {
	email            domain.Email
	eventAttachments []map[string]interface{}
}

func NewVerifyAndSaveInboxUseCase(
	repository domain.EmailRepository,
	storageProvider domain.StorageProvider,
	publishEventCase PublishEventUseCase,
) VerifyAndSaveInboxUseCase {
	return &verifyAndSaveInboxUseCase{
		repository:       repository,
		storageProvider:  storageProvider,
		publishEventCase: publishEventCase,
	}
}

func (useCase *verifyAndSaveInboxUseCase) Execute(account domain.Account, imapProvider domain.ImapEmailProvider) error {
	emails, err := imapProvider.GetAllUnreadEmails()
	if err != nil {
		return err
	}

	var emailsToSave []emailPayload
	for _, emailDTO := range emails {
		content, err := imapProvider.GetEmailByUid(emailDTO.Uid)
		if err != nil {
			log.Printf("[Conta %d] Erro ao buscar corpo do email %s: %v", account.ID, emailDTO.MessageId, err)
			continue
		}

		emailEntity := useCase.buildEmailEntity(account, emailDTO, content.Body)
		eventAttachments := useCase.processAttachments(account, &emailEntity, content.Attachments)

		emailsToSave = append(emailsToSave, emailPayload{
			email:            emailEntity,
			eventAttachments: eventAttachments,
		})
	}

	for _, item := range emailsToSave {
		err := useCase.repository.Create(&item.email)
		if err != nil {
			log.Printf("[Conta %d] Erro ao salvar email %s no banco: %v", account.ID, item.email.ID, err)
			continue
		}

		useCase.publishEventCase.Execute(account, item.email, "EMAIL_RECEIVED", item.eventAttachments)
	}

	return nil
}

func (useCase *verifyAndSaveInboxUseCase) buildEmailEntity(account domain.Account, dto domain.EmailDTO, body string) domain.Email {
	var fromEmail, toEmail string
	if len(dto.From) > 0 {
		fromEmail = dto.From[0].Email
	}
	if len(dto.To) > 0 {
		toEmail = dto.To[0].Email
	}

	var inReplyTo *string
	if dto.InReplyTo != "" {
		inReplyTo = &dto.InReplyTo
	}

	var repliedTo *string
	if len(dto.ReplyTo) > 0 {
		repliedTo = &dto.ReplyTo[0].Email
	}

	var references *string
	if dto.References != "" {
		references = &dto.References
	}

	return domain.Email{
		ID:         dto.MessageId,
		From:       fromEmail,
		To:         []string{toEmail},
		Subject:    dto.Subject,
		Date:       dto.Date,
		Body:       &body,
		InReplyTo:  inReplyTo,
		References: references,
		RepliedTo:  repliedTo,
		IdAccounts: account.ID,
		Status:     string(domain.EmailStatusReceived),
	}
}

func (useCase *verifyAndSaveInboxUseCase) processAttachments(account domain.Account, emailEntity *domain.Email, attachments []domain.AttachmentsDTO) []map[string]interface{} {
	var eventAttachments []map[string]interface{}

	for _, attachmentDTO := range attachments {
		fileUrl, err := useCase.storageProvider.Upload(attachmentDTO.Filename, attachmentDTO.Data)
		if err != nil {
			log.Printf("[Conta %d] Erro ao salvar anexo %s: %v", account.ID, attachmentDTO.Filename, err)
			continue
		}

		attachmentRecord := domain.Attachment{
			ID:          uuid.New().String(),
			Filename:    attachmentDTO.Filename,
			ContentType: attachmentDTO.ContentType,
			FileUrl:     fileUrl,
			IdEmails:    emailEntity.ID,
		}
		emailEntity.Attachments = append(emailEntity.Attachments, attachmentRecord)

		eventAttachments = append(eventAttachments, map[string]interface{}{
			"id":           attachmentRecord.ID,
			"filename":     attachmentDTO.Filename,
			"content_type": attachmentDTO.ContentType,
		})
	}

	return eventAttachments
}
