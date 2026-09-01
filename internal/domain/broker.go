package domain

type MessageBroker interface {
	SendMessage(topic string, key string, message []byte) error
	Disconnect()
}

