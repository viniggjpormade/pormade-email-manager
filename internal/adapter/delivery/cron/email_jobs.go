package cron

import (
	"log"
	"strconv"

	emailRepository "github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/email"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	emailUseCase "github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
)

type EmailJobs struct {
	verifyAndSaveInboxUseCase emailUseCase.VerifyAndSaveInboxUseCase
	syncEmailStatusUseCase    emailUseCase.SyncEmailStatusUseCase
	retryFailedEventsUseCase  emailUseCase.RetryFailedEventsUseCase
	accountRepository         domain.AccountRepository
}

func NewEmailJobs(
	verifyUseCase emailUseCase.VerifyAndSaveInboxUseCase,
	syncUseCase emailUseCase.SyncEmailStatusUseCase,
	retryUseCase emailUseCase.RetryFailedEventsUseCase,
	accountRepo domain.AccountRepository,
) *EmailJobs {
	return &EmailJobs{
		verifyAndSaveInboxUseCase: verifyUseCase,
		syncEmailStatusUseCase:    syncUseCase,
		retryFailedEventsUseCase:  retryUseCase,
		accountRepository:         accountRepo,
	}
}

func (job *EmailJobs) RunVerifyAndSaveInbox() {
	accounts, err := job.accountRepository.FindAll()
	if err != nil {
		log.Printf("[Cron] Erro ao buscar contas no banco: %v", err)
		return
	}

	for _, account := range accounts {
		port, _ := strconv.Atoi(account.ImapPort)
		if port == 0 {
			port = 993
		}

		user := ""
		if account.User != nil {
			user = *account.User
		}

		connectInfo := domain.ConnectInfo{
			IMAPHost:     account.ImapHost,
			IMAPPort:     port,
			Username:     user,
			IMAPPassword: account.ImapPassword,
		}

		imapProvider, err := emailRepository.NewImapRepository(connectInfo)
		if err != nil {
			log.Printf("[Conta %d] Erro ao conectar no IMAP (%s): %v", account.ID, account.ImapHost, err)
			continue
		}

		err = job.verifyAndSaveInboxUseCase.Execute(account, imapProvider)
		if err != nil {
			log.Printf("[Conta %d] Erro ao executar caso de uso: %v", account.ID, err)
		}

		imapProvider.Disconnect()
	}
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
