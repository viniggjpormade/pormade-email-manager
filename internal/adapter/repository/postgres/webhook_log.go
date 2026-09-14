package postgres

import (
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	"gorm.io/gorm"
)

type WebhookLogRepository struct {
	db *gorm.DB
}

func NewWebhookLogRepository(db *gorm.DB) domain.EmailWebhookLogRepository {
	return &WebhookLogRepository{db: db}
}

func (repository *WebhookLogRepository) Create(webhookLog *domain.EmailWebhookLog) error {
	return repository.db.Create(webhookLog).Error
}

func (repository *WebhookLogRepository) Update(webhookLog *domain.EmailWebhookLog) error {
	return repository.db.Save(webhookLog).Error
}

func (repository *WebhookLogRepository) FindAllFailed() ([]domain.EmailWebhookLog, error) {
	var webhookLogs []domain.EmailWebhookLog
	err := repository.db.Preload("Email").Preload("Email.Account").Where("success = ?", false).Find(&webhookLogs).Error
	return webhookLogs, err
}
