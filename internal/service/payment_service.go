package service

import (
	"context"

	"payment/internal/model"

	"github.com/google/uuid"
)

// PaymentRepository defines the interface for interacting with payment data storage.
type PaymentRepository interface {
	Save(ctx context.Context, payment *model.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	UpdateStatusIfPending(ctx context.Context, id uuid.UUID, newStatus model.PaymentStatus) (bool, error)
}

// PaymentNotifier defines the interface for notifying about payment events.
type PaymentNotifier interface {
	NotifyPaymentCreated(ctx context.Context, paymentID uuid.UUID) error
}
