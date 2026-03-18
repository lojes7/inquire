package mq

import (
	"context"
	"log"
	"time"

	"github.com/lojes7/inquire/pkg/infra"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SendTask publishes a message to the "task_queue"
func SendTask(message string) error {
	ch := infra.GetRabbitMQChannel()
	if ch == nil {
		log.Println("RabbitMQ channel not initialized, skipping publish")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		"",           // exchange
		"task_queue", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // persistent
			ContentType:  "text/plain",
			Body:         []byte(message),
			Timestamp:    time.Now(),
		})

	if err != nil {
		log.Printf("[MQ] Failed to publish a message: %s", err)
		return err
	}

	log.Printf("[MQ] Sent %s", message)
	return nil
}
