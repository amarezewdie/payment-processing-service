# Use the official Golang image as a build base
FROM golang:1.25.5-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the worker service (Note: building from cmd/worker)
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Use a minimal image for the final stage
FROM alpine:latest

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Set the current working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /worker .

# Command to run the executable
CMD ["./worker"]
