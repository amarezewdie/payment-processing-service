package messaging

import (
	"time"

	"github.com/google/uuid"
)

// EventMessage is a standardized structure for all asynchronous messages.
type EventMessage struct {
	// Type identifies the nature of the event, e.g., "payment.created".
	// Consumers use this to know what the message is about.
	Type string `json:"type"`

	// Source identifies which service published the message, e.g., "api-service".
	Source string `json:"source"`

	// Timestamp is when the message was created.
	Timestamp time.Time `json:"timestamp"`

	// Data contains the actual message payload, which can be any structure.
	Data interface{} `json:"data"`
}

// NewEvent creates a new EventMessage.
func NewEvent(eventType, sourceService string, payload interface{}) EventMessage {
	return EventMessage{
		Type:      eventType,
		Source:    sourceService,
		Timestamp: time.Now().UTC(),
		Data:      payload,
	}
}

// Example Payload for a created payment.
type PaymentCreatedPayload struct {
	PaymentID uuid.UUID `json:"payment_id"`
}
