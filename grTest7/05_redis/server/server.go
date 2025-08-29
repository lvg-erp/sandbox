package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type Message struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type PublishRequest struct {
	Content string `json:"content"`
}

func main() {
	// Получаем адрес Redis из переменной окружения
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// Подключение к Redis
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Проверка подключения
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Метрики
	var sentMessages, receivedMessages uint64
	var mu sync.Mutex

	// Контекст для управления горутинами
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подписка на response_channel
	results := make(chan string, 10)
	go func() {
		pubsub := client.Subscribe(ctx, "response_channel")
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					log.Printf("Error receiving message: %v", err)
					continue
				}
				var message Message
				if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}
				mu.Lock()
				atomic.AddUint64(&receivedMessages, 1)
				results <- fmt.Sprintf("Server received at %s: ID=%s, Content=%s, Total received: %d",
					time.Now().Format(time.RFC3339), message.ID, message.Content, receivedMessages)
				mu.Unlock()
			}
		}
	}()

	// Горутина для вывода результатов
	go func() {
		for result := range results {
			fmt.Println(result)
		}
	}()

	// Публикация сообщений в exchange_channel
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				message := Message{
					ID:        fmt.Sprintf("server-%d", i),
					Content:   fmt.Sprintf("Message %d from server", i),
					Timestamp: time.Now().Format(time.RFC3339),
				}
				messageJSON, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling message: %v", err)
					continue
				}
				err = client.Publish(ctx, "exchange_channel", messageJSON).Err()
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
					continue
				}
				mu.Lock()
				atomic.AddUint64(&sentMessages, 1)
				log.Printf("Server published: ID=%s, Content=%s, Total sent: %d", message.ID, message.Content, sentMessages)
				mu.Unlock()
			}
		}
	}()

	// HTTP-эндпоинт для проверки
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status": "server is running"}`)
		log.Println("Received request on /")
	})

	// HTTP-эндпоинт для публикации сообщений
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req PublishRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		message := Message{
			ID:        fmt.Sprintf("postman-%d", time.Now().UnixNano()),
			Content:   req.Content,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		messageJSON, err := json.Marshal(message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error marshaling message: %v", err), http.StatusInternalServerError)
			return
		}

		err = client.Publish(context.Background(), "exchange_channel", messageJSON).Err()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to publish: %v", err), http.StatusInternalServerError)
			return
		}

		mu.Lock()
		atomic.AddUint64(&sentMessages, 1)
		log.Printf("Published via HTTP: ID=%s, Content=%s, Total sent: %d", message.ID, message.Content, sentMessages)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "published", "message": "%s"}`, message.Content)
	})

	// Запуск HTTP-сервера
	log.Println("HTTP server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
