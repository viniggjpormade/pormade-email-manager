package email

import (
	"encoding/json"
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
	broker          domain.MessageBroker
}

func NewVerifyAndSaveInboxUseCase(repository domain.EmailRepository, storageProvider domain.StorageProvider, broker domain.MessageBroker) VerifyAndSaveInboxUseCase {
	return &verifyAndSaveInboxUseCase{
		repository:      repository,
		storageProvider: storageProvider,
		broker:          broker,
	}
}

func (useCase *verifyAndSaveInboxUseCase) Execute(account domain.Account, imapProvider domain.ImapEmailProvider) error {
	emails, err := imapProvider.GetAllUnreadEmails()
	if err != nil {
		return err
	}

	type emailPayload struct {
		email            domain.Email
		kafkaAttachments []map[string]interface{}
	}

	var emailsToSave []emailPayload
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

		var kafkaAttachments []map[string]interface{}

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

			kafkaAttachments = append(kafkaAttachments, map[string]interface{}{
				"id":           attachmentRecord.ID,
				"filename":     att.Filename,
				"content_type": att.ContentType,
			})
		}

		emailsToSave = append(emailsToSave, emailPayload{email: emailToSave, kafkaAttachments: kafkaAttachments})
	}

	for _, item := range emailsToSave {
		emailToSave := item.email
		err := useCase.repository.Create(&emailToSave)
		if err != nil {
			log.Printf("[Conta %d] Erro ao salvar email %s no banco: %v", account.ID, emailToSave.ID, err)
			continue
		}

		if account.KafkaTopic != nil && useCase.broker != nil {
			payload := map[string]interface{}{
				"id":          emailToSave.ID,
				"subject":     emailToSave.Subject,
				"from":        emailToSave.From,
				"to":          emailToSave.To,
				"body":        emailToSave.Body,
				"date":        emailToSave.Date,
				"in_reply_to": emailToSave.InReplyTo,
				"replied_to":  emailToSave.RepliedTo,
				"id_account":  emailToSave.IdAccounts,
				"attachments": item.kafkaAttachments,
			}

			emailBytes, _ := json.Marshal(payload)
			err = useCase.broker.SendMessage(*account.KafkaTopic, emailToSave.ID, emailBytes)
			if err != nil {
				log.Printf("[Conta %d] Erro ao publicar email %s no Kafka: %v", account.ID, emailToSave.ID, err)
			}
		}
	}

	return nil
}
