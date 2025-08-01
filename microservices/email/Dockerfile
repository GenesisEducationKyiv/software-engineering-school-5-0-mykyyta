# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

COPY email/go.mod email/go.sum ./email/
COPY pkg/ ./pkg/
RUN cd email && go mod download

COPY email/ ./email/
RUN cd email && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/email/main .
COPY --from=builder /app/email/logs ./logs
COPY --from=builder /app/email/internal/template ./template

RUN chmod +x /app/main
RUN adduser -D appuser
USER appuser

EXPOSE 8081
CMD ["/app/main"]