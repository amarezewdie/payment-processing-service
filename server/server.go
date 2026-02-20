package server

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/amarezewdie/payment-processing-service/internal/application/service"
	"github.com/amarezewdie/payment-processing-service/internal/infrastructure/messaging"
	"github.com/amarezewdie/payment-processing-service/internal/infrastructure/persistence"
	"github.com/amarezewdie/payment-processing-service/internal/pkg"
	"github.com/amarezewdie/payment-processing-service/server/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v4/pgxpool"
)

// App holds the shared dependencies for both API and Worker.
type App struct {
	DBPool         *pgxpool.Pool
	RabbitMQ       *messaging.RabbitMQDetails
	PaymentService service.IPaymentService
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
	rmq, err := messaging.Setup(messaging.RabbitMQConfig{
		URL:          os.Getenv("RABBITMQ_URL"),
		ExchangeName: exchangeName,
		QueueName:    queueName,
		BindingKey:   "payment.created",
	})
	if err != nil {
		log.Fatalf("Critical: Could not connect to RabbitMQ: %v", err)
	}

	// Repository and Service Layers
	paymentRepo := persistence.NewPaymentRepository(dbpool)
	paymentNotifier := messaging.NewPaymentNotifier(rmq.Channel, exchangeName)
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

// Server handles the HTTP server and routing.
type Server struct {
	router         *chi.Mux
	paymentService service.IPaymentService
}

// NewServer creates a new HTTP server instance.
func NewServer(svc service.IPaymentService) *Server {
	s := &Server{
		router:         chi.NewRouter(),
		paymentService: svc,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	paymentHandler := handler.NewPaymentHandler(s.paymentService)

	s.router.Post("/payments", paymentHandler.CreatePayment)
	s.router.Get("/payments/{id}", paymentHandler.GetPayment)
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}
