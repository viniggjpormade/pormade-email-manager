package email

import (
	"log"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type RetryFailedEventsUseCase interface {
	Execute() error
}

type retryFailedEventsUseCase struct {
	webhookLogRepository domain.EmailWebhookLogRepository
	accountRepository    domain.AccountRepository
	broker               domain.MessageBroker
}

func NewRetryFailedEventsUseCase(
	webhookLogRepository domain.EmailWebhookLogRepository,
	accountRepository domain.AccountRepository,
	broker domain.MessageBroker,
) RetryFailedEventsUseCase {
	return &retryFailedEventsUseCase{
		webhookLogRepository: webhookLogRepository,
		accountRepository:    accountRepository,
		broker:               broker,
	}
}

func (useCase *retryFailedEventsUseCase) Execute() error {
	failedLogs, err := useCase.webhookLogRepository.FindAllFailed()
	if err != nil {
		return err
	}

	for _, webhookLog := range failedLogs {
		useCase.retryEvent(webhookLog)
	}

	return nil
}

func (useCase *retryFailedEventsUseCase) retryEvent(webhookLog domain.EmailWebhookLog) {
	account, err := useCase.accountRepository.FindById(webhookLog.Email.IdAccounts)
	if err != nil {
		log.Printf("[Retry] Erro ao buscar conta %d: %v", webhookLog.Email.IdAccounts, err)
		return
	}

	hasKafkaTopic := account.KafkaTopic != nil && *account.KafkaTopic != ""
	hasWebhook := account.Webhook != nil && *account.Webhook != ""

	if !hasKafkaTopic && !hasWebhook {
		return
	}

	payloadBytes := buildEventPayload(webhookLog.Email, "EMAIL_RETRY", nil)

	var retryError error

	if hasKafkaTopic && useCase.broker != nil {
		retryError = useCase.broker.SendEmailMessage(*account.KafkaTopic, webhookLog.Email.ID, payloadBytes)
	} else if hasWebhook {
		retryError = sendWebhookRequest(*account.Webhook, payloadBytes)
	}

	if retryError != nil {
		errorMessage := retryError.Error()
		webhookLog.Error = &errorMessage
		log.Printf("[Retry] Falha ao reenviar evento do email %s: %v", webhookLog.Email.ID, retryError)
	} else {
		webhookLog.Success = true
		webhookLog.Error = nil
		log.Printf("[Retry] Evento do email %s reenviado com sucesso", webhookLog.Email.ID)
	}

	if updateError := useCase.webhookLogRepository.Update(&webhookLog); updateError != nil {
		log.Printf("[Retry] Erro ao atualizar log do email %s: %v", webhookLog.Email.ID, updateError)
	}
}
