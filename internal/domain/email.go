package domain

import "time"

type ConnectInfo struct {
	Address  string
	Port     int
	Username string
	Password string
}

type EmailStatus string

const (
	EmailStatusQueued     EmailStatus = "queued"
	EmailStatusReceived   EmailStatus = "received"
	EmailStatusSent       EmailStatus = "sent"
	EmailStatusDeferred   EmailStatus = "deferred"
	EmailStatusBounced    EmailStatus = "bounced"
	EmailStatusBlocked    EmailStatus = "blocked"
	EmailStatusQuarantine EmailStatus = "quarantine"
)

type Email struct {
	ID          string       `gorm:"type:varchar;primaryKey"`
	From        string       `gorm:"type:varchar;not null"`
	To          string       `gorm:"type:varchar;not null"`
	RepliedTo   *string      `gorm:"type:varchar"`
	Body        *string      `gorm:"type:text"`
	Date        time.Time    `gorm:"type:timestamptz"`
	Subject     string       `gorm:"type:varchar;not null"`
	InReplyTo   *string      `gorm:"type:varchar"`
	IdAccounts  int64        `gorm:"column:id_accounts;type:bigint;not null"`
	Status      string       `gorm:"type:email_manager.email_status;not null;default:'received'"`
	Account     Account      `gorm:"foreignKey:IdAccounts"`
	Attachments []Attachment `gorm:"foreignKey:IdEmails"`
}

type Attachment struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	Filename    string `gorm:"type:varchar;not null"`
	IdEmails    string `gorm:"column:id_emails;type:varchar;not null"`
	ContentType string `gorm:"type:varchar;not null"`
	FileUrl     string `gorm:"type:varchar;not null"`
}

type EmailDTO struct {
	SeqNum    uint32       `json:"seq_num"`
	Uid       uint32       `json:"uid"`
	Size      uint32       `json:"size"`
	Flags     []string     `json:"flags"`
	Subject   string       `json:"subject"`
	Date      time.Time    `json:"date"`
	MessageId string       `json:"message_id"`
	InReplyTo string       `json:"in_reply_to"`
	From      []AddressDTO `json:"from"`
	To        []AddressDTO `json:"to"`
	ReplyTo   []AddressDTO `json:"reply_to"`
}

type AddressDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AttachmentDTO struct {
	Filename    string
	ContentType string
	Data        []byte
}

type EmailBodyAndAttachments struct {
	Body        string
	Attachments []AttachmentDTO
}

type ImapEmailProvider interface {
	GetAllUnreadEmails() ([]EmailDTO, error)
	GetEmailByUid(uid uint32) (*EmailBodyAndAttachments, error)
	Disconnect() error
}

type EmailRepository interface {
	Create(email *Email) error
	Update(email *Email) error
	GetPendingOutboundEmails() ([]Email, error)
}

type SendEmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type OutboundEmailProvider interface {
	SendEmail(from string, to string, subject string, body string, inReplyTo string, attachments []AttachmentDTO) (*SendEmailResponse, error)
	VerifyEmailStatus(id string) (*SendEmailResponse, error)
}
