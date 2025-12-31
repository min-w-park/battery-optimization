# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy pkg/events first (required dependency)
COPY pkg/ ./pkg/

# Copy service go.mod and go.sum for dependency caching
COPY services/bidding/go.mod services/bidding/go.sum ./services/bidding/

# Download dependencies
WORKDIR /app/services/bidding
RUN go mod download

# Copy service source code
COPY services/bidding/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bidding-service ./cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bidding-service .

# Expose port (for future health endpoint)
EXPOSE 8084

# Run the service
CMD ["./bidding-service"]
