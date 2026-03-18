package main

import (
	"log"

	"github.com/lojes7/inquire/internal/mq"
	"github.com/lojes7/inquire/pkg/infra"
)

func main() {
	log.Println("Initializing Worker...")

	infra.Init()
	// Note: infra.Init() initializes DB and MQ.

	defer infra.CloseRabbitMQ()

	mq.StartWorker()
}
