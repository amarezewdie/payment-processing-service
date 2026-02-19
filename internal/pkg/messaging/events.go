package messaging

import (
	"time"

	"github.com/google/uuid"
)

// EventMessage is a standardized structure for all asynchronous messages.
type EventMessage struct {
	Type      string      `json:"type"`      // Type identifies the nature of the event, e.g., "payment.created".  	// Consumers use this to know what the message is about.
	Source    string      `json:"source"`    // Source identifies which service published the message, e.g., "api-service".
	Timestamp time.Time   `json:"timestamp"` // Timestamp is when the message was created.
	Data      interface{} `json:"data"`      // Data contains the actual message payload, which can be any structure.
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
