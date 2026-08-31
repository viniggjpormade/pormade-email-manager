package account

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type ValidateTokenUseCase interface {
	Execute(rawToken string) bool
}

type validateTokenUseCase struct {
	repository domain.AccountRepository
}

func NewValidateTokenUseCase(repository domain.AccountRepository) ValidateTokenUseCase {
	return &validateTokenUseCase{
		repository: repository,
	}
}

func (useCase *validateTokenUseCase) Execute(rawToken string) bool {
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	_, err := useCase.repository.FindByToken(hashedToken)
	
	// Se err for nil, significa que encontrou a conta (token válido)
	return err == nil
}
