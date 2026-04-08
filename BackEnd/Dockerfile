# Start from golang base image
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Set GOPROXY to boost build speed
RUN go env -w GOPROXY=https://goproxy.io,direct

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

# Add ca certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy the env file
COPY init.env .

# Expose port 38080
EXPOSE 38080

# Set environment variable for MongoDB connection
#ENV MONGODB_URI="mongodb://mongodb:27017"

# Command to run the executable
CMD ["./main"]
