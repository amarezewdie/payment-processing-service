package domain

import "errors"

// Common domain errors
var (
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrPaymentNotPending = errors.New("payment not in PENDING status")
)
