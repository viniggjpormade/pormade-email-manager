package kafka

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaBroker struct {
	kafkaProducer *kafka.Producer
}

func NewKafkaBroker(config *kafka.ConfigMap) (*KafkaBroker, error) {
	kafkaProducer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente kafka %w", err)
	}
	return &KafkaBroker{
		kafkaProducer: kafkaProducer,
	}, nil
}

func (broker *KafkaBroker) ProducerEventLoop() {
	defer log.Println("")

	for event := range broker.kafkaProducer.Events() {
		switch ev := event.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Falha ao entregar mensagem: %v", ev.TopicPartition.Error)
			} else {
				log.Printf("Mensagem entregue em %s[%d]@%d",
					*ev.TopicPartition.Topic,
					ev.TopicPartition.Partition,
					ev.TopicPartition.Offset)
			}
		case kafka.Error:
			log.Printf("Erro no produtor: %v", ev)
		}
	}
}

func (broker *KafkaBroker) SendEmailMessage(topic string, key string, message []byte) error {
	err := broker.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            []byte(key),
	}, nil)

	if err != nil {
		return fmt.Errorf("Falha ao produzir: %w", err)
	}

	return nil
}

func (broker *KafkaBroker) Disconnect() {
	broker.kafkaProducer.Close()
}
