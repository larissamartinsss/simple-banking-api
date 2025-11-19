# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o simple-banking-api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/simple-banking-api .

# Create data directory for SQLite
RUN mkdir -p /root/data

# Expose port
EXPOSE 8080

# Set environment variables
ENV SERVER_ADDRESS=:8080
ENV DATABASE_PATH=/root/data/banking.db

# Run the application
CMD ["./simple-banking-api"]
