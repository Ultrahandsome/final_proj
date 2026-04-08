# Start from a Go base image
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files (if they exist)
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application (targeting the specific file)
RUN go build -o server ./cmd/grpctest

# Start a new stage from scratch
FROM alpine:latest

# Add necessary runtime dependencies
RUN apk add --no-cache ca-certificates

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the gRPC port
EXPOSE 50051

# Command to run the executable
CMD ["./server"]
