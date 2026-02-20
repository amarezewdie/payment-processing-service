package handler

import (
	"fmt"
	"net/http"

	"github.com/amarezewdie/payment-processing-service/internal/application/service"
	"github.com/amarezewdie/payment-processing-service/internal/domain"
	httpResponse "github.com/amarezewdie/payment-processing-service/internal/pkg/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PaymentHandler handles HTTP requests related to payments.
type PaymentHandler struct {
	paymentService service.IPaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentService service.IPaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePaymentRequest is the request body for creating a payment
type CreatePaymentRequest struct {
	Reference string  `json:"reference" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required,oneof=USD ETB"`
}

// Validate validates the request body
func (h *PaymentHandler) Validate(req *CreatePaymentRequest) error {
	if req.Reference == "" {
		return fmt.Errorf("reference is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.Currency != "USD" && req.Currency != "ETB" {
		return fmt.Errorf("currency must be USD or ETB")
	}
	return nil
}

// CreatePayment handles the creation of a new payment.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	req := new(CreatePaymentRequest)
	if err := httpResponse.DecodeRequestJSON(r, req); err != nil {
		httpResponse.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Validate(req); err != nil {
		httpResponse.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	payment, err := h.paymentService.CreatePayment(r.Context(), req.Reference, req.Amount, req.Currency)
	if err != nil {
		httpResponse.HTTPError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create payment: %v", err))
		return
	}

	httpResponse.ResponseJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        payment.ID,
		"reference": payment.Reference,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"status":    payment.Status,
	})
}

// GetPayment handles retrieving a payment by ID.
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		httpResponse.HTTPError(w, http.StatusBadRequest, "Invalid payment ID format")
		return
	}

	payment, err := h.paymentService.GetPayment(r.Context(), paymentID)
	if err != nil {
		if err.Error() == domain.ErrPaymentNotFound.Error() {
			httpResponse.HTTPError(w, http.StatusNotFound, "Payment not found")
			return
		}
		httpResponse.HTTPError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve payment: %v", err))
		return
	}

	httpResponse.ResponseJSON(w, http.StatusOK, payment)
}
