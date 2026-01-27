package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"payment/internal/domain"
)

// PaymentRepository implements application.PaymentRepository for PostgreSQL.
type PaymentRepository struct {
	db *pgxpool.Pool
}

// NewPaymentRepository creates a new PostgreSQL PaymentRepository.
func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Save inserts a new payment into the database.
func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO payments (id, reference, amount, currency, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		payment.ID, payment.Reference, payment.Amount, payment.Currency, payment.Status, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	return nil
}

// FindByID retrieves a payment by its ID from the database.
func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.db.QueryRow(ctx,
		"SELECT id, reference, amount, currency, status, created_at, updated_at FROM payments WHERE id = $1",
		id).Scan(&payment.ID, &payment.Reference, &payment.Amount, &payment.Currency, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to query payment by ID: %w", err)
	}
	return payment, nil
}

// Update updates an existing payment in the database.
func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	_, err := r.db.Exec(ctx,
		"UPDATE payments SET reference = $1, amount = $2, currency = $3, status = $4, updated_at = $5 WHERE id = $6",
		payment.Reference, payment.Amount, payment.Currency, payment.Status, payment.UpdatedAt, payment.ID)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}

// UpdateStatusIfPending atomically updates the payment status from PENDING to newStatus.
// It returns true if the update was successful, false if the status was not PENDING.
func (r *PaymentRepository) UpdateStatusIfPending(ctx context.Context, id uuid.UUID, newStatus domain.PaymentStatus) (bool, error) {
	result, err := r.db.Exec(ctx,
		`UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2 AND status = $3`,
		newStatus, id, domain.PaymentStatusPending)
	if err != nil {
		return false, fmt.Errorf("failed to update payment status idempotently: %w", err)
	}

	return result.RowsAffected() > 0, nil
}
