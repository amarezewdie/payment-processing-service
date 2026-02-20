package service

import (
	"context"
	"fmt"
	"time"

	"github.com/amarezewdie/payment-processing-service/internal/domain"
	"github.com/google/uuid"
)

// CreatePayment creates a new payment and notifies about its creation.
func (s *PaymentService) CreatePayment(ctx context.Context, reference string, amount float64, currency string) (*domain.Payment, error) {
	payment := domain.NewPayment(reference, amount, currency)

	if err := s.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	if err := s.notifier.NotifyPaymentCreated(ctx, payment.ID); err != nil {
		// Log the error but don't fail the payment creation, as the payment is already saved.
		fmt.Printf("failed to notify payment creation for payment ID %s: %v\n", payment.ID, err)
	}

	return payment, nil
}

// GetPayment retrieves a payment by its ID.
func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return payment, nil
}

// ProcessPayment simulates processing a payment and updates its status.
// It uses an idempotent update mechanism.
func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID uuid.UUID) error {
	// Simulate external payment processing delay
	time.Sleep(time.Duration(500+RandIntn(2000)) * time.Millisecond)

	var newStatus domain.PaymentStatus
	if RandIntn(100) < 80 { // 80% chance of success
		newStatus = domain.PaymentStatusSuccess
	} else {
		newStatus = domain.PaymentStatusFailed
	}

	updated, err := s.repo.UpdateStatusIfPending(ctx, paymentID, newStatus)
	if err != nil {
		return fmt.Errorf("failed to idempotently update payment status: %w", err)
	}
	if !updated {
		// This means the payment was not in PENDING status, so another worker or process
		// already handled it, or it was in a terminal state. This is not an error for idempotency.
		return domain.ErrPaymentNotPending
	}

	return nil
}

// RandIntn is a helper to generate random numbers.
func RandIntn(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}
