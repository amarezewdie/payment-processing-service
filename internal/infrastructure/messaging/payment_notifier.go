package messaging

import (
	"github.com/streadway/amqp"
)

// PaymentNotifier implements domain.PaymentNotifier for RabbitMQ.
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
