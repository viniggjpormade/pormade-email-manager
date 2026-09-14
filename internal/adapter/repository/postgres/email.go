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

func (r *EmailRepository) Update(email *domain.Email) error {
	return r.db.Save(email).Error
}

func (r *EmailRepository) GetPendingOutboundEmails() ([]domain.Email, error) {
	var emails []domain.Email
	err := r.db.Where("status NOT IN (?, ?, ?, ?)",
		string(domain.EmailStatusSent),
		string(domain.EmailStatusBounced),
		string(domain.EmailStatusBlocked),
		string(domain.EmailStatusReceived),
	).Find(&emails).Error
	return emails, err
}
