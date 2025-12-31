# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy pkg/ first (required dependencies for replace directives)
COPY pkg/ ./pkg/

# Copy go mod files
COPY services/telemetry/go.mod services/telemetry/go.sum* ./services/telemetry/
WORKDIR /app/services/telemetry
RUN go mod download

# Copy source code
COPY services/telemetry/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/telemetry-service ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/telemetry-service ./server

# Copy migrations
COPY --from=builder /app/services/telemetry/internal/adapters/postgres/migrations ./internal/adapters/postgres/migrations

# Expose port
EXPOSE 8082

# Run the binary
CMD ["./server"]
