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
	accountRepository         domain.AccountRepository
}

func NewEmailJobs(usecase emailUseCase.VerifyAndSaveInboxUseCase, accountRepo domain.AccountRepository) *EmailJobs {
	return &EmailJobs{
		verifyAndSaveInboxUseCase: usecase,
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
