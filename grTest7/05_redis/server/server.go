package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"time"
)

type PublishedRequest struct {
	Content string `json:"content"`
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	log.Printf("Connected to Reddis at %s", redisAddr)

	http.HandleFunc("/published", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req PublishedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Failed to parse request: "+err.Error(), http.StatusBadRequest)
			return
		}
		message := fmt.Sprintf("Message from Postman: %s", req.Content)
		err = client.Publish(context.Background(), "exchange channel", message).Err()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to publish: %v", err), http.StatusInternalServerError)
			return
		}
	})

	//Сервер старт
	go func() {
		log.Println("HTTP server running on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	//Пихаем сообщения в горутину
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				message := fmt.Sprintf("Message %d from server", i)
				err := client.Publish(ctx, "exchange channel", message).Err()
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
					continue
				}
				log.Printf("Publishsd: %s", message)
			}
		}
	}()

	<-ctx.Done()
}
