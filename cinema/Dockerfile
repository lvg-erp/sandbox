# Этап сборки (builder)
FROM golang:1.24.5 AS builder

# Установка необходимых зависимостей для сборки
RUN apt-get update && apt-get install -y \
    gcc \
    make \
    libgl1-mesa-dev \
    libegl1-mesa-dev \
    libgles2-mesa-dev \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем go.mod для зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем проект
COPY . .

# Собираем приложение
RUN go build -o /cinema-app -ldflags="-s -w" cmd/app/main.go


# Итоговая стадия (исполняемый образ)
FROM debian:bookworm-slim

# Установка необходимых библиотек для запуска
RUN apt-get update && apt-get install -y \
    libgl1 \
    libegl1 \
    libgles2 \
    libx11-6 \
    libxcursor1 \
    libxrandr2 \
    libxinerama1 \
    libxi6 \
    libxxf86vm1 \
    && rm -rf /var/lib/apt/lists/*

# Копируем собранное приложение
COPY --from=builder /cinema-app /cinema-app

WORKDIR /

CMD ["./cinema-app"]

#FROM golang:1.25-alpine AS builder
#
#WORKDIR /app
#COPY go.mod go.sum ./
#RUN go mod download
#
#COPY . .
#RUN go build -o /cinema-app cmd/app/main.go
#
#FROM alpine:latest
#RUN apk --no-cache add ca-certificates
#WORKDIR /root/
#COPY --from=builder /cinema-app .
#
#EXPOSE 8080
#CMD ["./cinema-app"]