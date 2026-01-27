# Payment Gateway Module

A robust, production-ready payment gateway service responsible for creating payments, processing them asynchronously, and reliably tracking their status. Built with Go, PostgreSQL, RabbitMQ, and Docker.

## Features

- **Payment Creation**: Validates and creates payments with `PENDING` status.
- **Asynchronous Processing**: Background workers process payments via RabbitMQ.
- **Idempotency**: Guarantees that payments are processed exactly once, handling redelivery and concurrency safely.
- **Reliability**: Uses PostgreSQL transactions and atomic updates to ensure data consistency.
- **Status Tracking**: API to query payment status.

## Tech Stack

- **Language**: Go (Golang)
- **Web Framework**: Echo
- **Database**: PostgreSQL (pgx driver)
- **Messaging**: RabbitMQ (amqp)
- **Infrastructure**: Docker & Docker Compose

## Architecture

The system consists of two main services:

1.  **API Service**: Handles HTTP requests, validates input, persists payment data, and publishes events to RabbitMQ.
2.  **Worker Service**: Consumes events from RabbitMQ, simulates processing (delay + random success/failure), and updates payment status using optimistic locking for idempotency.

### Idempotency & Concurrency

To handle concurrent workers and message redelivery, the system uses an atomic database update strategy:

```sql
UPDATE payments 
SET status = $1, updated_at = NOW() 
WHERE id = $2 AND status = 'PENDING'
```

If multiple workers attempt to process the same payment, only one will successfully match the `status = 'PENDING'` condition. The others will detect no rows were affected and discard the message safely.

## getting Started

### Prerequisites

- Docker
- Docker Compose

### Running the Application

1.  Clone the repository.
2.  Start the services:

    ```bash
    docker-compose up -d --build
    ```

    This will start:
    - Postgres (Port 5432)
    - RabbitMQ (Management UI on Port 15672)
    - Payment API (Port 8080)
    - Payment Worker

### Testing

#### 1. Create a Payment

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "reference": "REF-12345",
    "amount": 100.50,
    "currency": "USD"
  }'
```

**Response:**
```json
{
  "amount": 100.5,
  "currency": "USD",
  "id": "a1b2c3d4-...",
  "reference": "REF-12345",
  "status": "PENDING"
}
```

#### 2. Check Payment Status

Wait a few seconds for the worker to process it, then check the status:

```bash
# Replace with the ID from the creation response
curl http://localhost:8080/payments/a1b2c3d4-...
```

**Response:**
```json
{
  "amount": 100.5,
  "currency": "USD",
  "id": "a1b2c3d4-...",
  "reference": "REF-12345",
  "status": "SUCCESS"  // or FAILED
}
```

## Project Structure

```
├── cmd
│   ├── api          # API entrypoint
│   └── worker       # Worker entrypoint
├── internal
│   ├── adapters     # Implementations (HTTP, Postgres, RabbitMQ)
│   ├── application  # Business logic (Service, Ports)
│   ├── domain       # Domain models
│   └── pkg          # Shared utilities
├── db               # Database migrations
└── docker-compose.yml
```
