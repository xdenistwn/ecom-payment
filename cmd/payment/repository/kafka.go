package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"payment/models"

	"github.com/segmentio/kafka-go"
)

type PaymentEventPublisher interface {
	PublishPaymentSuccess(ctx context.Context, orderID int64) error
	PublishEventPaymentStatus(ctx context.Context, orderID int64, status string, topic string) error
	PublishPaymentFailed(ctx context.Context, orderID int64) error
}

type kafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) PaymentEventPublisher {
	return &kafkaPublisher{
		writer: writer,
	}
}

func (k *kafkaPublisher) PublishEventPaymentStatus(ctx context.Context, orderID int64, status string, topic string) error {
	payload := models.PaymentStatusUpdateEvent{
		OrderID: orderID,
		Status:  status,
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
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

func (k *kafkaPublisher) PublishPaymentFailed(ctx context.Context, orderID int64) error {
	payload := map[string]interface{}{
		"order_id": orderID,
		"status":   "failed",
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
}
