package domain

import (
	"time"
)

type Account struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;type:bigserial"`
	ImapPort     string    `gorm:"column:imap_port;type:varchar;not null;default:'993'"`
	ImapHost     string    `gorm:"column:imap_host;type:varchar;not null"`
	Secure       bool      `gorm:"type:boolean;not null;default:false"`
	User         *string   `gorm:"type:varchar"`
	ImapPassword string    `gorm:"column:imap_password;type:varchar;not null"`
	KafkaTopic   *string   `gorm:"type:varchar"`
	Webhook      *string   `gorm:"type:varchar"`
	CreatedAt    time.Time `gorm:"type:timestamptz;default:now()"`
	AccessToken  string    `gorm:"type:varchar;not null"`
	SmtpPort     *string   `gorm:"column:smtp_port;type:varchar"`
	SmtpHost     *string   `gorm:"column:smtp_host;type:varchar"`
	SmtpPassword *string   `gorm:"column:smtp_password;type:varchar"`
	Emails       []Email   `gorm:"foreignKey:IdAccounts"`
}

type AccountRepository interface {
	Create(account *Account) error
	Update(account *Account) error
	FindByToken(token string) (*Account, error)
	FindById(id int64) (*Account, error)
	FindAll() ([]Account, error)
}
