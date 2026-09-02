package account

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type ValidateTokenUseCase interface {
	Execute(rawToken string) (*domain.Account, error)
}

type validateTokenUseCase struct {
	repository domain.AccountRepository
}

func NewValidateTokenUseCase(repository domain.AccountRepository) ValidateTokenUseCase {
	return &validateTokenUseCase{
		repository: repository,
	}
}

func (useCase *validateTokenUseCase) Execute(rawToken string) (*domain.Account, error) {
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	return useCase.repository.FindByToken(hashedToken)
}
