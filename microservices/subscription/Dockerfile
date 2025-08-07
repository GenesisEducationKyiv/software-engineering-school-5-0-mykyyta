# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

COPY subscription/go.mod subscription/go.sum ./subscription/
COPY pkg/ ./pkg/
RUN cd subscription && go mod download

COPY subscription/ ./subscription/
RUN cd subscription && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/subscription/main .
COPY --from=builder /app/subscription/logs ./logs
COPY --from=builder /app/subscription/templates templates/

RUN chmod +x /app/main
RUN adduser -D appuser
USER appuser

EXPOSE 8080
CMD ["/app/main"]