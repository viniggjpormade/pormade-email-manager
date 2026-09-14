package email

import (
	"encoding/json"
	"log"
	"time"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type SyncEmailStatusUseCase interface {
	Execute() error
}

type syncEmailStatusUseCase struct {
	emailRepo     domain.EmailRepository
	accountRepo   domain.AccountRepository
	emailProvider domain.OutboundEmailProvider
	broker        domain.MessageBroker
}

func NewSyncEmailStatusUseCase(
	emailRepo domain.EmailRepository,
	accountRepo domain.AccountRepository,
	emailProvider domain.OutboundEmailProvider,
	broker domain.MessageBroker,
) SyncEmailStatusUseCase {
	return &syncEmailStatusUseCase{
		emailRepo:     emailRepo,
		accountRepo:   accountRepo,
		emailProvider: emailProvider,
		broker:        broker,
	}
}

func (useCase *syncEmailStatusUseCase) Execute() error {
	emails, err := useCase.emailRepo.GetPendingOutboundEmails()
	if err != nil {
		return err
	}

	for _, emailEntity := range emails {
		useCase.processSingleEmail(emailEntity)
	}

	return nil
}

func (useCase *syncEmailStatusUseCase) processSingleEmail(emailEntity domain.Email) {
	res, err := useCase.emailProvider.VerifyEmailStatus(emailEntity.ID)
	if err != nil {
		log.Printf("Erro ao verificar email %s: %v", emailEntity.ID, err)
		return
	}

	if res.Status == emailEntity.Status {
		return
	}

	emailEntity.Status = res.Status
	if err := useCase.emailRepo.Update(&emailEntity); err != nil {
		log.Printf("Erro ao atualizar email %s no banco: %v", emailEntity.ID, err)
		return
	}

	useCase.notifyKafka(emailEntity)
}

func (useCase *syncEmailStatusUseCase) notifyKafka(emailEntity domain.Email) {
	account, err := useCase.accountRepo.FindById(emailEntity.IdAccounts)
	if err != nil {
		log.Printf("Erro ao buscar conta %d: %v", emailEntity.IdAccounts, err)
		return
	}

	if account.Webhook == nil || *account.Webhook == "" {
		return
	}
	topic := *account.Webhook

	payload := map[string]interface{}{
		"event_type": "EMAIL_STATUS_UPDATED",
		"email_id":   emailEntity.ID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"payload": map[string]interface{}{
			"status": emailEntity.Status,
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	err = useCase.broker.SendEmailMessage(topic, emailEntity.ID, payloadBytes)
	if err != nil {
		log.Printf("Erro ao enviar mensagem pro Kafka no topico %s: %v", topic, err)
	}
}
