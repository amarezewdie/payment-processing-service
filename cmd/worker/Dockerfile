# -------- Build Stage --------
FROM golang:1.24-alpine3.20 AS builder

# Install build tools (only in builder)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies first (faster CI builds)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary (small + secure)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o worker ./cmd/worker


# -------- Runtime Stage --------
FROM alpine:3.20

# Install runtime deps only
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/worker .

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

# Expose if worker has metrics / health endpoint (optional)
EXPOSE 8080

# Add container healthcheck (important in production)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD pgrep worker || exit 1

# Run app
ENTRYPOINT ["./worker"]