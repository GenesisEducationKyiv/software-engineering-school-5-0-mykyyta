# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

COPY weather/go.mod weather/go.sum ./weather/
COPY pkg/ ./pkg/
RUN cd weather && go mod download

COPY weather/ ./weather/
RUN cd weather && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/weather/main .
COPY --from=builder /app/weather/logs ./logs

RUN chmod +x /app/main
RUN adduser -D appuser
USER appuser

EXPOSE 8082
CMD ["/app/main"]