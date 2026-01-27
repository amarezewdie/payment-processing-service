package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/streadway/amqp"

	paymenthttp "payment/internal/adapters/http"
	"payment/internal/adapters/postgres"
	"payment/internal/adapters/rabbitmq"
	"payment/internal/application"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Database connection
	dbpool, err := pgxpool.ConnectConfig(context.Background(), newDBConfig(os.Getenv("DATABASE_URL")))
	if err != nil {
		e.Logger.Fatal("Failed to connect to database: ", err)
	}
	defer dbpool.Close()

	// RabbitMQ connection
	amqpConn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		e.Logger.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer amqpConn.Close()

	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		e.Logger.Fatal("Failed to open RabbitMQ channel: ", err)
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
		e.Logger.Fatalf("Failed to declare a queue: %v", err)
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
		e.Logger.Fatalf("Failed to bind a queue: %v", err)
	}

	// Initialize adapters
	paymentRepo := postgres.NewPaymentRepository(dbpool)
	paymentNotifier := rabbitmq.NewPaymentNotifier(amqpChannel, exchangeName)

	// Initialize application services
	paymentService := application.NewPaymentService(paymentRepo, paymentNotifier)

	// Initialize HTTP handlers
	paymentHandler := paymenthttp.NewPaymentHandler(paymentService)

	e.Validator = paymentHandler.NewCustomValidator()

	// Routes
	e.POST("/payments", paymentHandler.CreatePayment)
	e.GET("/payments/:id", paymentHandler.GetPayment)

	e.Logger.Fatal(e.Start(":8080"))
}

func newDBConfig(databaseURL string) *pgxpool.Config {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	return config
}
