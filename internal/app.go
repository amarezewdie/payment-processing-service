package internal

import (
	"context"
	"log"
	"os"

	"payment/internal/adapters/rabbitmq"
	"payment/internal/pkg"
	"payment/internal/repository"
	"payment/internal/service"

	"github.com/jackc/pgx/v4/pgxpool"
)

// App holds the shared dependencies for both API and Worker.
type App struct {
	DBPool         *pgxpool.Pool
	RabbitMQ       *rabbitmq.RabbitMQDetails
	PaymentService *service.PaymentService
}

// NewApp initializes all shared infrastructure and services.
func NewApp() *App {
	// Database
	dbpool, err := pgxpool.ConnectConfig(context.Background(), pkg.NewDBConfig(os.Getenv("DATABASE_URL")))
	if err != nil {
		log.Fatalf("Critical: Could not connect to Database: %v", err)
	}

	exchangeName := "app_events_exchange"
	queueName := "payment_processing"

	// RabbitMQ
	rmq, err := rabbitmq.Setup(rabbitmq.RabbitMQConfig{
		URL:          os.Getenv("RABBITMQ_URL"),
		ExchangeName: exchangeName,
		QueueName:    queueName,
		BindingKey:   "payment.created",
	})
	if err != nil {
		log.Fatalf("Critical: Could not connect to RabbitMQ: %v", err)
	}

	// Repository and Service Layers
	paymentRepo := repository.NewPaymentRepository(dbpool)
	paymentNotifier := rabbitmq.NewPaymentNotifier(rmq.Channel, exchangeName)
	paymentService := service.NewPaymentService(paymentRepo, paymentNotifier)

	return &App{
		DBPool:         dbpool,
		RabbitMQ:       rmq,
		PaymentService: paymentService,
	}
}

// Close ensures all connections are shut down gracefully.
func (a *App) Close() {
	if a.DBPool != nil {
		a.DBPool.Close()
	}
	if a.RabbitMQ != nil {
		a.RabbitMQ.Channel.Close()
		a.RabbitMQ.Conn.Close()
	}
}
