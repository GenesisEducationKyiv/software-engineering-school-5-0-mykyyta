# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main .
COPY ./logs ./logs

RUN chmod +x /app/main
RUN adduser -D appuser
USER appuser

EXPOSE 8082
CMD ["/app/main"]