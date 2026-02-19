package main

import (
	"log"

	"payment/internal"
	"payment/worker"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	// 1. Initialize shared infrastructure
	app := internal.NewApp()
	defer app.Close()

	// 2. Start Worker with injected dependencies
	log.Println("Starting Application in WORKER mode")
	paymentWorker := worker.NewWorker(app.PaymentService, app.RabbitMQ)
	paymentWorker.Start()
}
