# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder

WORKDIR /app

COPY backend/go.mod ./backend/
RUN cd backend && go mod download

COPY backend/ ./backend/
COPY graph/ ./graph/

WORKDIR /app/backend
RUN go build -o server


FROM debian:bookworm-slim

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/backend/server .

# Copy graph folder from builder
COPY --from=builder /app/graph ./graph

EXPOSE 8080

CMD ["./server"]
