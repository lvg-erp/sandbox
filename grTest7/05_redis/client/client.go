package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	//коннектим клиента
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connect to Redis")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan string, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pubsub := client.Subscribe(ctx, "exchange channel")
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

				result := fmt.Sprintf("Received at %s: %s", time.Now().Format(time.RFC3339), msg.Payload)
				log.Printf("Client processed: %s", result)
				results <- result
			}
		}
	}()

	//Выводим результаты
	go func() {
		for rslt := range results {
			fmt.Println(rslt)
		}
	}()

	<-ctx.Done()
}
