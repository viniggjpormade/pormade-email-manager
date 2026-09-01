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
	repository      domain.EmailRepository
	storageProvider domain.StorageProvider
}

func NewVerifyAndSaveInboxUseCase(repository domain.EmailRepository, storageProvider domain.StorageProvider) VerifyAndSaveInboxUseCase {
	return &verifyAndSaveInboxUseCase{
		repository:      repository,
		storageProvider: storageProvider,
	}
}

func (useCase *verifyAndSaveInboxUseCase) Execute(account domain.Account, imapProvider domain.ImapEmailProvider) error {
	emails, err := imapProvider.GetAllUnreadEmails()
	if err != nil {
		return err
	}

	var emailsToSave []domain.Email
	for _, email := range emails {
		content, err := imapProvider.GetEmailByUid(email.Uid)
		if err != nil {
			log.Printf("[Conta %d] Erro ao buscar corpo do email %s: %v", account.ID, email.MessageId, err)
			continue
		}

		var fromEmail, toEmail string
		if len(email.From) > 0 {
			fromEmail = email.From[0].Email
		}
		if len(email.To) > 0 {
			toEmail = email.To[0].Email
		}

		var inReplyTo *string
		if email.InReplyTo != "" {
			inReplyTo = &email.InReplyTo
		}

		var repliedTo *string
		if len(email.ReplyTo) > 0 {
			repliedTo = &email.ReplyTo[0].Email
		}

		bodyStr := content.Body
		emailToSave := domain.Email{
			ID:         email.MessageId,
			From:       fromEmail,
			To:         toEmail,
			Subject:    email.Subject,
			Date:       email.Date,
			Body:       &bodyStr,
			InReplyTo:  inReplyTo,
			RepliedTo:  repliedTo,
			IdAccounts: account.ID,
		}

		for _, att := range content.Attachments {
			fileUrl, err := useCase.storageProvider.Upload(att.Filename, att.Data)
			if err != nil {
				log.Printf("[Conta %d] Erro ao salvar anexo %s: %v", account.ID, att.Filename, err)
				continue
			}

			attachmentRecord := domain.Attachment{
				ID:          uuid.New().String(),
				Filename:    att.Filename,
				ContentType: att.ContentType,
				FileUrl:     fileUrl,
				IdEmails:    emailToSave.ID,
			}
			emailToSave.Attachments = append(emailToSave.Attachments, attachmentRecord)
		}

		emailsToSave = append(emailsToSave, emailToSave)
	}

	for _, emailToSave := range emailsToSave {
		err := useCase.repository.Create(&emailToSave)
		if err != nil {
			log.Printf("[Conta %d] Erro ao salvar email %s no banco: %v", account.ID, emailToSave.ID, err)
		}
	}

	return nil
}
