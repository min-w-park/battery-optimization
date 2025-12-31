# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY services/market-data/go.mod services/market-data/go.sum* ./
RUN go mod download

# Copy source code
COPY services/market-data/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY --from=builder /app/internal/adapters/postgres/migrations ./internal/adapters/postgres/migrations

EXPOSE 8080

CMD ["./server"]
