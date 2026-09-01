package postgres

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	"gorm.io/gorm"
)

type EmailRepository struct {
	db *gorm.DB
}

func NewEmailRepository(db *gorm.DB) domain.EmailRepository {
	return &EmailRepository{db: db}
}

func (r *EmailRepository) Create(email *domain.Email) error {
	return r.db.Create(email).Error
}
