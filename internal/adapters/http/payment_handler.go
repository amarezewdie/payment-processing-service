package http

import (
	"fmt"
	"net/http"

	"payment/internal/application"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// PaymentHandler handles HTTP requests related to payments.
type PaymentHandler struct {
	paymentService *application.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentService *application.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePaymentRequest is the request body for creating a payment
type CreatePaymentRequest struct {
	Reference string  `json:"reference" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required,oneof=USD ETB"`
}

// CustomValidator holds the echo validator
type CustomValidator struct {
	validator *Validator
}

// NewCustomValidator creates a new CustomValidator
func (h *PaymentHandler) NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: &Validator{}}
}

// Validate validates the request body
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Validate(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// Validator is a simple validator
type Validator struct {
}

// Validate validates the given struct
func (v *Validator) Validate(i interface{}) error {
	switch req := i.(type) {
	case *CreatePaymentRequest:
		if req.Reference == "" {
			return fmt.Errorf("reference is required")
		}
		if req.Amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
		if req.Currency != "USD" && req.Currency != "ETB" {
			return fmt.Errorf("currency must be USD or ETB")
		}
	}
	return nil
}

// CreatePayment handles the creation of a new payment.
func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	req := new(CreatePaymentRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err // c.Validate already returns HTTPError
	}

	payment, err := h.paymentService.CreatePayment(c.Request().Context(), req.Reference, req.Amount, req.Currency)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create payment: %v", err))
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"id":        payment.ID,
		"reference": payment.Reference,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"status":    payment.Status,
	})
}

// GetPayment handles retrieving a payment by ID.
func (h *PaymentHandler) GetPayment(c echo.Context) error {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment ID format")
	}

	payment, err := h.paymentService.GetPayment(c.Request().Context(), paymentID)
	if err != nil {
		if err.Error() == "payment not found" {
			return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve payment: %v", err))
	}

	return c.JSON(http.StatusOK, payment)
}
