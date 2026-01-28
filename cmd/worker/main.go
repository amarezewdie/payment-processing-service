package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"payment/internal/adapters/rabbitmq"
	"payment/internal/repository"
	"payment/internal/pkg"
	"payment/internal/pkg/messaging"
	"payment/internal/service"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	// Database connection
	dbpool, err := pgxpool.ConnectConfig(context.Background(), pkg.NewDBConfig(os.Getenv("DATABASE_URL")))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbpool.Close()

	log.Println("Connecting to RabbitMQ...")
	amqpConn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer amqpConn.Close()
	log.Println("Successfully connected to RabbitMQ")

	log.Println("Opening a channel...")
	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer amqpChannel.Close()
	log.Println("Successfully opened a channel")

	// 1. Declare the Exchange
	exchangeName := "app_events_exchange"
	log.Printf("Declaring exchange '%s'...", exchangeName)
	err = amqpChannel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}
	log.Printf("Successfully declared exchange '%s'", exchangeName)

	// 2. Declare the Queue
	queueName := "payment_processing"
	log.Printf("Declaring queue '%s'...", queueName)
	q, err := amqpChannel.QueueDeclare(
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
	log.Printf("Successfully declared queue '%s'", queueName)

	// 3. Bind the Queue to the Exchange
	bindingKey := "payment.created"
	log.Printf("Binding queue '%s' to exchange '%s' with routing key '%s'...", queueName, exchangeName, bindingKey)
	err = amqpChannel.QueueBind(
		q.Name,       // queue name
		bindingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}
	log.Printf("Successfully bound queue")

	// Initialize adapters
	paymentRepo := repository.NewPaymentRepository(dbpool)

	// Use the real RabbitMQ notifier
	paymentNotifier := rabbitmq.NewPaymentNotifier(amqpChannel, exchangeName)

	// Initialize application services
	paymentService := payment.NewPaymentService(paymentRepo, paymentNotifier)

	// 4. Start Consuming
	log.Println("Starting consumer...")
	msgs, err := amqpChannel.Consume(
		q.Name,            // queue
		"worker-consumer", // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	log.Println("Consumer started. Waiting for messages...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var event messaging.EventMessage
			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				d.Nack(false, false) // Dead letter or discard if unparseable
				continue
			}
			log.Printf("Received a message of type '%s' from '%s'", event.Type, event.Source)

			// 5. Identify message and route to handler
			switch event.Type {
			case "payment.created":
				// The data is decoded as map[string]interface{}. We need to parse it.
				dataMap, ok := event.Data.(map[string]interface{})
				if !ok {
					log.Printf("Error: could not assert Data to map[string]interface{}")
					d.Nack(false, false)
					continue
				}

				paymentIDStr, ok := dataMap["payment_id"].(string)
				if !ok {
					log.Printf("Error: could not assert payment_id to string")
					d.Nack(false, false)
					continue
				}

				paymentID, err := uuid.Parse(paymentIDStr)
				if err != nil {
					log.Printf("Error parsing payment ID: %v", err)
					d.Nack(false, false)
					continue
				}

				err = paymentService.ProcessPayment(context.Background(), paymentID)
				if err != nil {
					if err == payment.ErrPaymentNotPending {
						log.Printf("Payment %s already processed or not pending. Acking.", paymentID.String())
						d.Ack(false)
					} else {
						log.Printf("Error processing payment %s: %v", paymentID.String(), err)
						d.Nack(false, true) // Nack and requeue
					}
				} else {
					log.Printf("Successfully processed payment %s", paymentID.String())
					d.Ack(false) // Ack
				}
			default:
				log.Printf("Unknown message type: %s", event.Type)
				d.Ack(false) // Acknowledge to remove from queue
			}
		}
	}()

	<-forever
}
