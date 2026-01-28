package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"

	"payment/internal/adapters/rabbitmq"
	handler "payment/internal/handler"
	"payment/internal/pkg"
	repository "payment/internal/repository"
	service "payment/internal/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Recoverer)

	// Database connection
	dbpool, err := pgxpool.ConnectConfig(context.Background(), pkg.NewDBConfig(os.Getenv("DATABASE_URL")))
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer dbpool.Close()

	// RabbitMQ connection
	amqpConn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer amqpConn.Close()

	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		log.Fatal("Failed to open RabbitMQ channel: ", err)
	}
	defer amqpChannel.Close()

	exchangeName := "app_events_exchange"

	// Declare the queue
	queueName := "payment_processing"
	_, err = amqpChannel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Bind the queue to the exchange
	bindingKey := "payment.created"
	err = amqpChannel.QueueBind(
		queueName,    // queue name
		bindingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	// Initialize adapters
	paymentRepo := repository.NewPaymentRepository(dbpool)
	paymentNotifier := rabbitmq.NewPaymentNotifier(amqpChannel, exchangeName)

	// Initialize application services
	paymentService := service.NewPaymentService(paymentRepo, paymentNotifier)

	// Initialize HTTP handlers
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Routes
	r.Post("/payments", paymentHandler.CreatePayment)
	r.Get("/payments/{id}", paymentHandler.GetPayment)

	log.Printf("Server starting on port %s", os.Getenv("API_PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("API_PORT"), r))
}
