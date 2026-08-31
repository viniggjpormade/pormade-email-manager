package domain

import "time"

type ConnectInfo struct {
	Address  string
	Port     int
	Username string
	Password string
}

type EmailDTO struct {
	SeqNum    uint32       `json:"seq_num"`
	Uid       uint32       `json:"uid"`
	Size      uint32       `json:"size"`
	Flags     []string     `json:"flags"`
	Subject   string       `json:"subject"`
	Date      time.Time    `json:"date"`
	MessageId string       `json:"message_id"`
	From      []AddressDTO `json:"from"`
	To        []AddressDTO `json:"to"`
}

type AddressDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type EmailRepository interface {
	GetAllUnreadEmails() ([]EmailDTO, error)
	Disconnect() error
}
