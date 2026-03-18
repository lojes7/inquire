package infra

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	RabbitMQConn    *amqp.Connection
	RabbitMQChannel *amqp.Channel
)

func InitRabbitMQ() error {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		log.Println("RABBITMQ_URL not set, skipping RabbitMQ init")
		return nil
	}

	var err error
	RabbitMQConn, err = amqp.Dial(url)
	if err != nil {
		return err
	}

	RabbitMQChannel, err = RabbitMQConn.Channel()
	if err != nil {
		return err
	}

	// Declare Queue (Durable = true for persistence)
	q, err := RabbitMQChannel.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	// Fair dispatch
	err = RabbitMQChannel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	log.Printf("Connected to RabbitMQ, queue declared: %s", q.Name)
	return nil
}

func CloseRabbitMQ() {
	if RabbitMQChannel != nil {
		RabbitMQChannel.Close()
	}
	if RabbitMQConn != nil {
		RabbitMQConn.Close()
	}
}

func GetRabbitMQChannel() *amqp.Channel {
	return RabbitMQChannel
}
