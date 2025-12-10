package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Producer{writer: writer}, nil
}

func (p *Producer) Emit(topic string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: jsonData,
	})

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	log.Println("✅ Message produced to", topic)
	return nil
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("Error closing kafka writer: %v", err)
	}
}
