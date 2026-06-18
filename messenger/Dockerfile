# Используем официальный образ Go 1.26 на Alpine Linux
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые системные зависимости
RUN apk add --no-cache git ca-certificates

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарный файл. Важно: CGO_ENABLED=0 для статической сборки.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server ./cmd/server

# Финальный минимальный образ
FROM alpine:latest

# Устанавливаем корневые сертификаты для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл и статические файлы (HTML) из этапа сборки
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web

# Открываем порт, который слушает приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./server"]