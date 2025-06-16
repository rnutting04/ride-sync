# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder

WORKDIR /app/backend

# Copy go.mod and go.sum first
COPY backend/go.mod ./

# Download dependencies early for cache benefits
RUN go mod download

# Copy the rest of the source code
COPY backend/ .

# Build the Go binary
RUN go build -o server

# Final slim image
FROM debian:bookworm-slim

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/backend/server .

# Expose port and run
EXPOSE 8080
CMD ["./server"]