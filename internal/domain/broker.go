package domain

type MessageBroker interface {
	SendEmailMessage(topic string, key string, message []byte) error
	Disconnect()
}
