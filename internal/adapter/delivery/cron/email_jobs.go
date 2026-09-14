package cron

import (
	"log"
	"strconv"
	"sync"

	emailRepository "github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/email"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	emailUseCase "github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
)

type EmailJobs struct {
	accountRepository         domain.AccountRepository
	sendEmailUseCase          emailUseCase.SendEmailUseCase
	syncEmailStatusUseCase    emailUseCase.SyncEmailStatusUseCase
	retryFailedEventsUseCase  emailUseCase.RetryFailedEventsUseCase
	verifyAndSaveInboxUseCase emailUseCase.VerifyAndSaveInboxUseCase
}

func NewEmailJobs(
	accountRepo domain.AccountRepository,
	sendEmailCase emailUseCase.SendEmailUseCase,
	syncUseCase emailUseCase.SyncEmailStatusUseCase,
	retryUseCase emailUseCase.RetryFailedEventsUseCase,
	verifyUseCase emailUseCase.VerifyAndSaveInboxUseCase,
) *EmailJobs {
	return &EmailJobs{
		accountRepository:         accountRepo,
		syncEmailStatusUseCase:    syncUseCase,
		retryFailedEventsUseCase:  retryUseCase,
		verifyAndSaveInboxUseCase: verifyUseCase,
		sendEmailUseCase:          sendEmailCase,
	}
}

func (job *EmailJobs) RunVerifyAndSaveInbox() {
	accounts, err := job.accountRepository.FindAll()
	if err != nil {
		log.Printf("[Cron] Erro ao buscar contas no banco: %v", err)
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, account := range accounts {
		wg.Add(1)
		go func(acc domain.Account) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			port, _ := strconv.Atoi(acc.ImapPort)
			if port == 0 {
				port = 993
			}

			user := ""
			if acc.User != nil {
				user = *acc.User
			}

			connectInfo := domain.ConnectInfo{
				IMAPHost:     acc.ImapHost,
				IMAPPort:     port,
				Username:     user,
				IMAPPassword: acc.ImapPassword,
			}

			imapProvider, err := emailRepository.NewImapRepository(connectInfo)
			if err != nil {
				log.Printf("[Conta %d] Erro ao conectar no IMAP (%s): %v", acc.ID, acc.ImapHost, err)
				return
			}
			defer imapProvider.Disconnect()

			err = job.verifyAndSaveInboxUseCase.Execute(acc, imapProvider)
			if err != nil {
				log.Printf("[Conta %d] Erro ao executar caso de uso: %v", acc.ID, err)
			}
		}(account)
	}

	wg.Wait()
}

func (job *EmailJobs) RunSyncEmailStatus() {
	err := job.syncEmailStatusUseCase.Execute()
	if err != nil {
		log.Printf("[Cron] Erro ao sincronizar status de emails pendentes: %v", err)
	}
}

func (job *EmailJobs) RunRetryFailedEvents() {
	err := job.retryFailedEventsUseCase.Execute()
	if err != nil {
		log.Printf("[Cron] Erro ao retentar eventos com falha: %v", err)
	}
}

func (job *EmailJobs) RunProcessQueuedEmails() {
	err := job.sendEmailUseCase.ProcessQueuedEmails()
	if err != nil {
		log.Printf("[Cron] Erro ao processar emails na fila: %v", err)
	}
}
