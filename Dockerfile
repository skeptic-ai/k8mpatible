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
RUN CGO_ENABLED=0 GOOS=linux go build -o worker .

# Final stage
FROM alpine:3.19

# Install CA certificates for HTTPS connections
RUN apk add --no-cache ca-certificates curl

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/worker .

# Copy any necessary configuration files
COPY compatibility compatibility

# Switch to non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["./worker"]