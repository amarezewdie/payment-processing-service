package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment/internal/pkg/messaging"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// PaymentNotifier implements application.PaymentNotifier for RabbitMQ.
type PaymentNotifier struct {
	channel      *amqp.Channel
	exchangeName string
}

// NewPaymentNotifier creates a new RabbitMQ PaymentNotifier.
func NewPaymentNotifier(channel *amqp.Channel, exchangeName string) *PaymentNotifier {
	return &PaymentNotifier{
		channel:      channel,
		exchangeName: exchangeName,
	}
}

// NotifyPaymentCreated publishes a message to RabbitMQ when a payment is created.
func (n *PaymentNotifier) NotifyPaymentCreated(ctx context.Context, paymentID uuid.UUID) error {
	// Create the specific payload for this event
	payload := messaging.PaymentCreatedPayload{
		PaymentID: paymentID,
	}

	// Create the standardized event message (the "envelope")
	event := messaging.NewEvent("payment.created", "payment-api", payload)

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event message: %w", err)
	}

	// The routing key describes the event, e.g., "payment.created"
	routingKey := "payment.created"

	err = n.channel.Publish(
		n.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	log.Printf("Published '%s' event for payment ID: %s", routingKey, paymentID.String())
	return nil
}
