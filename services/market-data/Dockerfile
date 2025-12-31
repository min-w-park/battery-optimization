# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy pkg/events first (required dependency for replace directive)
COPY pkg/ ./pkg/

# Copy go mod files
COPY services/market-data/go.mod services/market-data/go.sum* ./services/market-data/
WORKDIR /app/services/market-data
RUN go mod download

# Copy source code
COPY services/market-data/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/market-data-service ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/market-data-service ./server

# Copy migrations
COPY --from=builder /app/services/market-data/internal/adapters/postgres/migrations ./internal/adapters/postgres/migrations

EXPOSE 8080

CMD ["./server"]
