package server

import (
	"log"
	"net/http"
	"os"

	"payment/internal/service"
	"payment/server/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router         *chi.Mux
	paymentService *service.PaymentService
}

func NewServer(svc *service.PaymentService) *Server {
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

func (s *Server) Start() error {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}
