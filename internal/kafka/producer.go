package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Producer defines the interface for sending messages to Kafka
type Producer interface {
	Send(topic string, value interface{}) error
	Close()
}

// KafkaProducer is an implementation of Producer using Confluent Kafka
type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(bootstrapServers string, topic string) (Producer, error) {
	// 1. Khởi tạo admin client để kiểm tra và tạo topic nếu cần
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}
	defer admin.Close()

	// 2. Tạo topic nếu chưa có (nếu đã có thì không sao)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := admin.CreateTopics(ctx, []kafka.TopicSpecification{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	// Kiểm tra kết quả tạo topic
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			return nil, fmt.Errorf("failed to create topic %s: %v", result.Topic, result.Error)
		}
		log.Printf("Topic check: %s -> %v", result.Topic, result.Error)
	}

	// 3. Tạo producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "payment_service",
		"acks":              "all",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// 4. Goroutine xử lý phản hồi delivery
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("Successfully delivered message to %v", ev.TopicPartition)
				}
			}
		}
	}()

	return &KafkaProducer{
		producer: p,
	}, nil
}

// Send sends a message to the specified Kafka topic
func (p *KafkaProducer) Send(topic string, value interface{}) error {
	// Marshal the value to JSON
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	// Create and send the Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          valueBytes,
	}

	return p.producer.Produce(message, nil)
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() {
	p.producer.Flush(15 * 1000) // Wait for up to 15 seconds for any outstanding messages to be delivered
	p.producer.Close()
}
