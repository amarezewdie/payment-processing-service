package persistence

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// PaymentRepository implements domain.PaymentRepository for PostgreSQL.
type PaymentRepository struct {
	db *pgxpool.Pool
}

// NewPaymentRepository creates a new PostgreSQL PaymentRepository.
func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}
