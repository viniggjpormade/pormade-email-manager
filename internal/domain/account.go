package domain

import (
	"time"
)

type Account struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;type:bigserial"`
	Port        string    `gorm:"type:varchar;not null;default:'993'"`
	Host        string    `gorm:"type:varchar;not null"`
	Secure      bool      `gorm:"type:boolean;not null;default:false"`
	User        *string   `gorm:"type:varchar"`
	Password    string    `gorm:"type:varchar;not null"`
	KafkaTopic  *string   `gorm:"type:varchar"`
	Webhook     *string   `gorm:"type:varchar"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now()"`
	AccessToken string    `gorm:"type:varchar;not null"`
	Emails      []Email   `gorm:"foreignKey:IdAccounts"`
}

type AccountRepository interface {
	Create(account *Account) error
	Update(account *Account) error
	FindByToken(token string) (*Account, error)
	FindById(id int64) (*Account, error)
	FindAll() ([]Account, error)
}
