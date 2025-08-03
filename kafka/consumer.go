package kafka

import (
	"context"
	"encoding/json"
	"log"
	"payment/models"

	"github.com/segmentio/kafka-go"
)

func StartOrderConsumer(broker string, topic string, handler func(models.OrderCreatedEvent)) {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "payment",
	})

	// listen with go r
	go func(r *kafka.Reader) {
		for {
			message, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Println("Error while read Kafka Message: ", err.Error())
				// store data to database
				continue
			}

			var event models.OrderCreatedEvent
			err = json.Unmarshal(message.Value, &event)
			if err != nil {
				log.Println("Error while unmarshal kafka message: ", err.Error())
				continue
			}

			log.Printf("Received Event order_created: %+v", event)
			handler(event)
		}
	}(consumer)
}
