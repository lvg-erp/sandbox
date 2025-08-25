# Этап сборки приложения
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN ls -l /app || (echo "Directory /app is empty" && exit 1)
RUN ls -l /app/main.go || (echo "main.go not found" && exit 1)
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o rabbitmq-example main.go || (echo "Build failed" && exit 1)
RUN ls -l /app/rabbitmq-example || (echo "Binary not found" && exit 1)

# Финальный образ для приложения
FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /app
COPY --from=builder /app/rabbitmq-example /app/rabbitmq-example
RUN chmod +x /app/rabbitmq-example
RUN ls -l /app/rabbitmq-example || (echo "Binary not copied" && exit 1)
CMD ["/app/rabbitmq-example"]