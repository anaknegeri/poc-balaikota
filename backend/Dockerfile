FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install required dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o api ./cmd/api

# Create a lightweight production image
FROM alpine:3.21

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy the binary from builder
COPY --from=builder /build/api /app/api

# Create data directories
RUN mkdir -p /app/data/alert /app/data/people-count

# Set permissions for the data directories
RUN chmod -R 755 /app/data

# Set environment variables
ENV SERVER_PORT=3001
ENV DATA_ROOT=/app/data

# Expose the application port
EXPOSE 3001

# Run the application
CMD ["/app/api"]
