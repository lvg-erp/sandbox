# Используем официальный образ Go 1.26 на Alpine Linux
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые системные зависимости
RUN apk add --no-cache git ca-certificates

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Финальный минимальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./server"]