# -------- Build Stage --------
FROM golang:1.24-alpine3.20 AS builder

# Install only what we need to build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies (faster CI/CD)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a static, small binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o api ./cmd/api


# -------- Runtime Stage --------
FROM alpine:3.20

# Install runtime dependencies only
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/api .

# Use non-root user (important security)
RUN chown -R appuser:appgroup /app
USER appuser

# Your API port
EXPOSE 8080

# Healthcheck for auto-restart / monitoring
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget --spider -q http://localhost:8080/health || exit 1

ENTRYPOINT ["./api"]