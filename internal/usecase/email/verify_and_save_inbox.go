package email

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

type emailPayload struct {
	email            domain.Email
	kafkaAttachments []map[string]interface{}
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

	var emailsToSave []emailPayload
	for _, emailDTO := range emails {
		content, err := imapProvider.GetEmailByUid(emailDTO.Uid)
		if err != nil {
			log.Printf("[Conta %d] Erro ao buscar corpo do email %s: %v", account.ID, emailDTO.MessageId, err)
			continue
		}

		emailEntity := useCase.buildEmailEntity(account, emailDTO, content.Body)
		kafkaAttachments := useCase.processAttachments(account, &emailEntity, content.Attachments)

		emailsToSave = append(emailsToSave, emailPayload{
			email:            emailEntity,
			kafkaAttachments: kafkaAttachments,
		})
	}

	for _, item := range emailsToSave {
		err := useCase.repository.Create(&item.email)
		if err != nil {
			log.Printf("[Conta %d] Erro ao salvar email %s no banco: %v", account.ID, item.email.ID, err)
			continue
		}

		useCase.publishEvent(account, item)
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
	var kafkaAttachments []map[string]interface{}

	for _, att := range attachments {
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
			IdEmails:    emailEntity.ID,
		}
		emailEntity.Attachments = append(emailEntity.Attachments, attachmentRecord)

		kafkaAttachments = append(kafkaAttachments, map[string]interface{}{
			"id":           attachmentRecord.ID,
			"filename":     att.Filename,
			"content_type": att.ContentType,
		})
	}

	return kafkaAttachments
}

func (useCase *verifyAndSaveInboxUseCase) publishEvent(account domain.Account, item emailPayload) {
	if (account.KafkaTopic == nil || *account.KafkaTopic == "") && (account.Webhook == nil || *account.Webhook == "") {
		return
	}

	payloadContent := map[string]interface{}{
		"id":          item.email.ID,
		"subject":     item.email.Subject,
		"from":        item.email.From,
		"to":          item.email.To,
		"body":        item.email.Body,
		"date":        item.email.Date,
		"in_reply_to": item.email.InReplyTo,
		"replied_to":  item.email.RepliedTo,
		"status":      item.email.Status,
		"attachments": item.kafkaAttachments,
	}

	payload := map[string]interface{}{
		"event_type": "EMAIL_RECEIVED",
		"email_id":   item.email.ID,
		"timestamp":  item.email.Date.Format(time.RFC3339),
		"payload":    payloadContent,
	}

	emailBytes, _ := json.Marshal(payload)

	if account.KafkaTopic != nil && *account.KafkaTopic != "" && useCase.broker != nil {
		err := useCase.broker.SendEmailMessage(*account.KafkaTopic, item.email.ID, emailBytes)
		if err != nil {
			log.Printf("[Conta %d] Erro ao publicar email %s no Kafka: %v", account.ID, item.email.ID, err)
		}
	} else if account.Webhook != nil && *account.Webhook != "" {
		go func(url string, data []byte) {
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 10 * time.Second}
			_, _ = client.Do(req)
		}(*account.Webhook, emailBytes)
	}
}
