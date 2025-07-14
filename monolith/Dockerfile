# Build stage (x86_64)
FROM golang:1.24 AS builder
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build binary
COPY .. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

# Runtime stage (Alpine)
FROM alpine:latest
WORKDIR /app

# Install CA certificates (needed for HTTPS requests)
RUN apk --no-cache add ca-certificates

# Copy binary and templates
COPY --from=builder /app/main .
COPY templates templates/

# Ensure executable permissions (redundant if already set, but safe)
RUN chmod +x /app/main

EXPOSE 8080

# Run as non-root user (optional and safe)
RUN adduser -D appuser
USER appuser

CMD ["/app/main"]

