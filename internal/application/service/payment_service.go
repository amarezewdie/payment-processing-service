package service

import (
	"context"

	"github.com/amarezewdie/payment-processing-service/internal/domain"
	"github.com/google/uuid"
)

// IPaymentService defines the interface for payment-related business logic.
type IPaymentService interface {
	CreatePayment(ctx context.Context, reference string, amount float64, currency string) (*domain.Payment, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	ProcessPayment(ctx context.Context, paymentID uuid.UUID) error
}

// PaymentService coordinates tasks and delegates work to the domain and infrastructure.
type PaymentService struct {
	repo     domain.PaymentRepository
	notifier domain.PaymentNotifier
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(repo domain.PaymentRepository, notifier domain.PaymentNotifier) *PaymentService {
	return &PaymentService{
		repo:     repo,
		notifier: notifier,
	}
}
