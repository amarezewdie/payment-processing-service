package main

import (
	"log"

	"github.com/amarezewdie/payment-processing-service/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	// 1. Initialize shared infrastructure
	app := server.NewApp()
	defer app.Close()

	// 2. Start API Server with injected PaymentService
	log.Println("Starting Application in API mode")
	srv := server.NewServer(app.PaymentService)
	if err := srv.Start(); err != nil {
		log.Fatalf("Critical: Server failed: %v", err)
	}
}
