FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY server/ ./server/
COPY client/ ./client/

RUN go build -o /app/redis-server ./server/server.go
RUN go build -o /app/redis-client ./client/client.go

CMD ["/app/redis-server"]