package application

import (
	"context"

	"github.com/google/uuid"
	"payment/internal/domain"
)

// PaymentRepository defines the interface for interacting with payment data storage.
type PaymentRepository interface {
	Save(ctx context.Context, payment *domain.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	UpdateStatusIfPending(ctx context.Context, id uuid.UUID, newStatus domain.PaymentStatus) (bool, error)
}

// PaymentNotifier defines the interface for notifying about payment events.
type PaymentNotifier interface {
	NotifyPaymentCreated(ctx context.Context, paymentID uuid.UUID) error
}
