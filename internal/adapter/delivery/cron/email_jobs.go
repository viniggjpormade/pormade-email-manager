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
	accountRepository         domain.AccountRepository
}

func NewEmailJobs(
	verifyUseCase emailUseCase.VerifyAndSaveInboxUseCase,
	syncUseCase emailUseCase.SyncEmailStatusUseCase,
	accountRepo domain.AccountRepository,
) *EmailJobs {
	return &EmailJobs{
		verifyAndSaveInboxUseCase: verifyUseCase,
		syncEmailStatusUseCase:    syncUseCase,
		accountRepository:         accountRepo,
	}
}

func (job *EmailJobs) RunVerifyAndSaveInbox() {
	accounts, err := job.accountRepository.FindAll()
	if err != nil {
		log.Printf("[Cron] Erro ao buscar contas no banco: %v", err)
		return
	}

	for _, acc := range accounts {
		port, _ := strconv.Atoi(acc.Port)
		if port == 0 {
			port = 993
		}

		user := ""
		if acc.User != nil {
			user = *acc.User
		}

		connectInfo := domain.ConnectInfo{
			Address:  acc.Host,
			Port:     port,
			Username: user,
			Password: acc.Password,
		}

		imapProvider, err := emailRepository.NewImapRepository(connectInfo)
		if err != nil {
			log.Printf("[Conta %d] Erro ao conectar no IMAP (%s): %v", acc.ID, acc.Host, err)
			continue
		}

		err = job.verifyAndSaveInboxUseCase.Execute(acc, imapProvider)
		if err != nil {
			log.Printf("[Conta %d] Erro ao executar caso de uso: %v", acc.ID, err)
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
