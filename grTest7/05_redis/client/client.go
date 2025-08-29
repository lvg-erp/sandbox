package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	var sentMessages, receivedMessages uint64
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для результатов
	results := make(chan string, 10)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pubsub := client.Subscribe(ctx, "exchange_channel")
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
				//можно было и так????
				//receivedMessages++
				atomic.AddUint64(&receivedMessages, 1)
				results <- fmt.Sprintf("Client received at %s: ID=%s, Content=%s, Total received: %d",
					time.Now().Format(time.RFC3339), message.ID, message.Content, receivedMessages)
				mu.Unlock()

				response := Message{
					ID:        fmt.Sprintf("client-response-%d", time.Now().UnixNano()),
					Content:   fmt.Sprintf("Response to %s", message.ID),
					Timestamp: time.Now().Format(time.RFC3339),
				}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error marshaling response: %v", err)
					continue
				}
				err = client.Publish(ctx, "response_channel", responseJSON).Err()
				if err != nil {
					log.Printf("Failed to publish response: %v", err)
					continue
				}
				mu.Lock()
				atomic.AddUint64(&sentMessages, 1)
				log.Printf("Client published: ID=%s, Content=%s, Total sent: %d", response.ID, response.Content, sentMessages)
				mu.Unlock()
			}
		}
	}()

	go func() {
		for result := range results {
			fmt.Println(result)
		}
	}()

	wg.Wait()
}
