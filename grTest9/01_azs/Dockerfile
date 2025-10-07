# Dockerfile для сборки Go приложения

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev libx11-dev mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /fuelazs ./cmd/fuelazs/main.go

FROM alpine:latest

RUN apk add --no-cache libx11 mesa-gl libxi libxcursor libxrandr libxinerama

WORKDIR /app

COPY --from=builder /fuelazs /app/fuelazs
COPY config.yaml credentials.yaml /app/
COPY ./internal/gui/*.ttf /app/internal/gui/
COPY internal/usecase/config/kazs_activation.json /app/internal/usecase/config/

CMD ["/app/fuelazs"]