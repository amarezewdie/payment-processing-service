package model

import "fmt"

// Common domain errors
var (
	ErrPaymentNotFound   = fmt.Errorf("payment not found")
	ErrPaymentNotPending = fmt.Errorf("payment not in PENDING status")
)
