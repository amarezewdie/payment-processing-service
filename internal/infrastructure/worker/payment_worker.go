package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/amarezewdie/payment-processing-service/internal/application/service"
	"github.com/amarezewdie/payment-processing-service/internal/domain"
	"github.com/amarezewdie/payment-processing-service/internal/infrastructure/messaging"
	pkgMessaging "github.com/amarezewdie/payment-processing-service/internal/pkg/messaging"

	"github.com/google/uuid"
)

// Worker handles the consumption and processing of payment events.
type Worker struct {
	paymentService service.IPaymentService
	rmq            *messaging.RabbitMQDetails
}

// NewWorker creates a new Worker instance with its dependencies injected.
func NewWorker(svc service.IPaymentService, rmq *messaging.RabbitMQDetails) *Worker {
	return &Worker{
		paymentService: svc,
		rmq:            rmq,
	}
}

// Start begins consuming messages from RabbitMQ. This is a blocking call.
func (w *Worker) Start() {
	log.Println("Worker: Registering consumer...")

	msgs, err := w.rmq.Channel.Consume(
		w.rmq.Queue.Name, // queue
		"payment-worker", // consumer name
		false,            // auto-ack (we manual ack)
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		log.Fatalf("Worker: Failed to register consumer: %v", err)
	}

	log.Println("Worker: Consumer active. Waiting for messages.")

	// Process messages in a loop
	for d := range msgs {
		var event pkgMessaging.EventMessage
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("Worker: Error parsing message: %v", err)
			d.Nack(false, false)
			continue
		}

		log.Printf("Worker: Received '%s' event", event.Type)

		if event.Type == "payment.created" {
			w.handlePaymentCreated(d, event)
		} else {
			// Unknown message types are acknowledged to clear the queue
			d.Ack(false)
		}
	}
}

func (w *Worker) handlePaymentCreated(d messaging.Delivery, event pkgMessaging.EventMessage) {
	dataMap, ok := event.Data.(map[string]interface{})
	if !ok {
		log.Println("Worker: Invalid event data format")
		d.Nack(false, false)
		return
	}

	paymentIDStr, _ := dataMap["payment_id"].(string)
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		log.Printf("Worker: Invalid payment ID: %v", err)
		d.Nack(false, false)
		return
	}

	// 1. Process the payment using the injected service
	err = w.paymentService.ProcessPayment(context.Background(), paymentID)

	// 2. Handle Acknowledgement
	if err != nil {
		if err == domain.ErrPaymentNotPending {
			log.Printf("Worker: Payment %s already processed. Acking.", paymentID)
			d.Ack(false)
		} else {
			log.Printf("Worker: Processing error for %s: %v. Requeuing.", paymentID, err)
			d.Nack(false, true) // Requeue for another attempt
		}
	} else {
		log.Printf("Worker: Successfully processed payment %s", paymentID)
		d.Ack(false)
	}
}
