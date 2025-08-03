package kafka

import (
	"github.com/segmentio/kafka-go"
)

func NewWriter(broker string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}
