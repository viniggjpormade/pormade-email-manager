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

func (repository *EmailRepository) Create(email *domain.Email) error {
	return repository.db.Create(email).Error
}

func (repository *EmailRepository) Update(email *domain.Email) error {
	return repository.db.Save(email).Error
}

func (repository *EmailRepository) GetPendingOutboundEmails() ([]domain.Email, error) {
	var emails []domain.Email
	err := repository.db.Where("status NOT IN (?, ?, ?, ?)",
		string(domain.EmailStatusSent),
		string(domain.EmailStatusBounced),
		string(domain.EmailStatusBlocked),
		string(domain.EmailStatusReceived),
	).Find(&emails).Error
	return emails, err
}

func (repository *EmailRepository) GetQueuedEmails() ([]domain.Email, error) {
	var emails []domain.Email
	err := repository.db.Preload("Account").Where("status = ?", string(domain.EmailStatusQueued)).Find(&emails).Error
	return emails, err
}

func (repository *EmailRepository) GetAttachmentById(id string) (*domain.Attachment, error) {
	var attachment domain.Attachment
	err := repository.db.Where("id = ?", id).First(&attachment).Error
	return &attachment, err
}
