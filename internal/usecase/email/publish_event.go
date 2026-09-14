package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type PublishEventUseCase interface {
	Execute(account domain.Account, emailEntity domain.Email, eventType string, attachments []map[string]interface{})
}

type publishEventUseCase struct {
	webhookLogRepository domain.EmailWebhookLogRepository
	broker               domain.MessageBroker
}

func NewPublishEventUseCase(
	webhookLogRepository domain.EmailWebhookLogRepository,
	broker domain.MessageBroker,
) PublishEventUseCase {
	return &publishEventUseCase{
		webhookLogRepository: webhookLogRepository,
		broker:               broker,
	}
}

func (useCase *publishEventUseCase) Execute(account domain.Account, emailEntity domain.Email, eventType string, attachments []map[string]interface{}) {
	hasKafkaTopic := account.KafkaTopic != nil && *account.KafkaTopic != ""
	hasWebhook := account.Webhook != nil && *account.Webhook != ""

	if !hasKafkaTopic && !hasWebhook {
		return
	}

	payloadBytes := buildEventPayload(emailEntity, eventType, attachments)

	var publishError error

	if hasKafkaTopic && useCase.broker != nil {
		publishError = useCase.broker.SendEmailMessage(*account.KafkaTopic, emailEntity.ID, payloadBytes)
	} else if hasWebhook {
		publishError = sendWebhookRequest(*account.Webhook, payloadBytes)
	}

	useCase.saveWebhookLog(emailEntity.ID, publishError)
}

func buildEventPayload(emailEntity domain.Email, eventType string, attachments []map[string]interface{}) []byte {
	payloadContent := map[string]interface{}{
		"id":          emailEntity.ID,
		"subject":     emailEntity.Subject,
		"from":        emailEntity.From,
		"to":          emailEntity.To,
		"date":        emailEntity.Date.Format(time.RFC3339),
		"in_reply_to": emailEntity.InReplyTo,
		"replied_to":  emailEntity.RepliedTo,
		"status":      emailEntity.Status,
		"attachments": attachments,
	}

	if emailEntity.Body != nil {
		payloadContent["body"] = *emailEntity.Body
	}

	payload := map[string]interface{}{
		"event_type": eventType,
		"email_id":   emailEntity.ID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"payload":    payloadContent,
	}

	payloadBytes, _ := json.Marshal(payload)
	return payloadBytes
}

func sendWebhookRequest(webhookUrl string, payloadBytes []byte) error {
	request, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição webhook: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("erro ao enviar webhook: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("webhook retornou status %d", response.StatusCode)
	}

	return nil
}

func (useCase *publishEventUseCase) saveWebhookLog(emailId string, publishError error) {
	webhookLog := domain.EmailWebhookLog{
		IdEmails: emailId,
		Success:  publishError == nil,
	}

	if publishError != nil {
		errorMessage := publishError.Error()
		webhookLog.Error = &errorMessage
	}

	useCase.webhookLogRepository.Create(&webhookLog)
}
