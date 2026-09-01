package postgres

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) domain.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *domain.Account) error {
	return r.db.Create(account).Error
}

func (r *AccountRepository) FindByToken(token string) (*domain.Account, error) {
	var account domain.Account
	err := r.db.Where("access_token = ?", token).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) FindAll() ([]domain.Account, error) {
	var accounts []domain.Account
	err := r.db.Find(&accounts).Error
	return accounts, err
}
