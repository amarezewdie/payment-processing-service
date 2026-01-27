package domain

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus defines the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusSuccess  PaymentStatus = "SUCCESS"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusCanceled PaymentStatus = "CANCELED"
)

// Payment represents the payment model in the domain
type Payment struct {
	ID        uuid.UUID     `json:"id"`
	Reference string        `json:"reference"`
	Amount    float64       `json:"amount"`
	Currency  string        `json:"currency"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewPayment creates a new payment with a pending status
func NewPayment(reference string, amount float64, currency string) *Payment {
	return &Payment{
		ID:        uuid.New(),
		Reference: reference,
		Amount:    amount,
		Currency:  currency,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// MarkAsSuccess updates the payment status to SUCCESS
func (p *Payment) MarkAsSuccess() {
	p.Status = PaymentStatusSuccess
	p.UpdatedAt = time.Now()
}

// MarkAsFailed updates the payment status to FAILED
func (p *Payment) MarkAsFailed() {
	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now()
}

// MarkAsCanceled updates the payment status to CANCELED
func (p *Payment) MarkAsCanceled() {
	p.Status = PaymentStatusCanceled
	p.UpdatedAt = time.Now()
}
