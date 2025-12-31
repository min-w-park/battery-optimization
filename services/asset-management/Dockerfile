# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy pkg/events first (required dependency for replace directive)
COPY pkg/ ./pkg/

# Copy go mod files
COPY services/asset-management/go.mod services/asset-management/go.sum* ./services/asset-management/
WORKDIR /app/services/asset-management
RUN go mod download

# Copy source code
COPY services/asset-management/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/asset-management-service ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/asset-management-service ./server

# Copy migrations
COPY --from=builder /app/services/asset-management/internal/adapters/postgres/migrations ./internal/adapters/postgres/migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]
