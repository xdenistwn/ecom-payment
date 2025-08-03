package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type PaymentEventPublisher interface {
	PublishPaymentSuccess(ctx context.Context, orderID int64) error
}

type kafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) PaymentEventPublisher {
	return &kafkaPublisher{
		writer: writer,
	}
}

// publish payment success
func (k *kafkaPublisher) PublishPaymentSuccess(ctx context.Context, orderID int64) error {
	payload := map[string]interface{}{
		"order_id": orderID,
		"status":   "paid",
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
}
