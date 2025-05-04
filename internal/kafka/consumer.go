package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// MessageHandler is a function that processes Kafka messages
type MessageHandler func(topic string, key []byte, value []byte) error

// Consumer defines the interface for consuming messages from Kafka
type Consumer interface {
	Subscribe(topics []string, handler MessageHandler) error
	Start(ctx context.Context) error
	Close() error
}

// KafkaConsumer is an implementation of Consumer using Confluent Kafka
type KafkaConsumer struct {
	consumer *kafka.Consumer
	handlers map[string]MessageHandler
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// NewKafkaConsumer creates a new instance of KafkaConsumer
func NewKafkaConsumer(bootstrapServers, groupID string) (Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       bootstrapServers,
		"group.id":                groupID,
		"auto.offset.reset":       "earliest",
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 5000,
		"session.timeout.ms":      30000,
		"max.poll.interval.ms":    300000,
		"heartbeat.interval.ms":   3000,
		"statistics.interval.ms":  0,
		"enable.partition.eof":    false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer: c,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// Subscribe subscribes to the specified Kafka topics
func (c *KafkaConsumer) Subscribe(topics []string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	// Register the handler for each topic
	for _, topic := range topics {
		c.handlers[topic] = handler
	}

	return nil
}

// Start starts consuming messages from Kafka
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}
	c.running = true
	c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consume(ctx)
	}()

	return nil
}

// consume consumes messages from Kafka until the context is cancelled
func (c *KafkaConsumer) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping Kafka consumer")
			return
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Ignore timeout errors
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Error reading message: %v", err)
				continue
			}

			// Get the topic
			topic := *msg.TopicPartition.Topic

			// Get the handler for the topic
			c.mu.Lock()
			handler, ok := c.handlers[topic]
			c.mu.Unlock()

			if !ok {
				log.Printf("No handler registered for topic %s", topic)
				continue
			}

			// Handle the message
			err = handler(topic, msg.Key, msg.Value)
			if err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	c.wg.Wait()
	return c.consumer.Close()
}

// HandleNotification is a utility function to handle payment notifications
func HandleNotification(topic string, key []byte, value []byte) error {
	var notification map[string]interface{}
	err := json.Unmarshal(value, &notification)
	if err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	// Log the notification
	log.Printf("Received notification on topic %s: %+v", topic, notification)

	// Here you would implement specific handling based on the notification type
	// For example, sending emails, updating other systems, etc.
	switch topic {
	case "payment.initiated":
		// Handle payment initiated
		log.Printf("Payment initiated for invoice %s", notification["invoice_id"])
	case "payment.completed":
		// Handle payment completed
		log.Printf("Payment completed for invoice %s", notification["invoice_id"])
	case "payment.failed":
		// Handle payment failed
		log.Printf("Payment failed for invoice %s", notification["invoice_id"])
	default:
		log.Printf("Unknown notification topic: %s", topic)
	}

	return nil
}
