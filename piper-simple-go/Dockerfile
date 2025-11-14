# Multi-stage Dockerfile for piper-simple-go
# Build stage
FROM golang:1.24-alpine AS builder

# Install git (needed for go mod download)
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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o piper-simple-go .

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/piper-simple-go .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./piper-simple-go"]
