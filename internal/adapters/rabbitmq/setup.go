package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Delivery is an alias for the messaging delivery type
type Delivery = amqp.Delivery

// RabbitMQConfig holds the configuration for setting up RabbitMQ
type RabbitMQConfig struct {
	URL          string
	ExchangeName string
	QueueName    string
	BindingKey   string
}

// RabbitMQDetails returns the connection and channel initialized with exchange and queue
type RabbitMQDetails struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

// Setup initializes RabbitMQ connection, channel, exchange and queue binding
func Setup(config RabbitMQConfig) (*RabbitMQDetails, error) {
	log.Printf("Connecting to RabbitMQ at %s...", config.URL)
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ. Opening channel...")
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	log.Printf("Declaring exchange '%s'...", config.ExchangeName)
	err = ch.ExchangeDeclare(
		config.ExchangeName, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("Declaring queue '%s'...", config.QueueName)
	q, err := ch.QueueDeclare(
		config.QueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if config.BindingKey != "" {
		log.Printf("Binding queue '%s' to exchange '%s' with routing key '%s'...", config.QueueName, config.ExchangeName, config.BindingKey)
		err = ch.QueueBind(
			q.Name,              // queue name
			config.BindingKey,   // routing key
			config.ExchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return &RabbitMQDetails{
		Conn:    conn,
		Channel: ch,
		Queue:   q,
	}, nil
}
