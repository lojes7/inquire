package mq

import (
	"log"
	"time"

	"github.com/lojes7/inquire/pkg/infra"
)

// StartWorker starts a blocking consumer
func StartWorker() {
	ch := infra.GetRabbitMQChannel()
	if ch == nil {
		log.Fatalln("RabbitMQ channel is not initialized!")
	}

	msgs, err := ch.Consume(
		"task_queue", // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("[Worker] Received a message: %s", d.Body)

			// Simulate processing time
			// In real world, use dots count or actual varied workload
			time.Sleep(1 * time.Second)

			log.Printf("[Worker] Done processing: %s", d.Body)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Worker is waiting for messages. To exit press CTRL+C")
	<-forever
}
