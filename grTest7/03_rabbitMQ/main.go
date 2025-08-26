package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Result struct {
	MessageID string
	Processed string
}

type Error struct {
	WorkerID int
	Error    error
}

func connectToRabbitMQ(url, queueName string) (*amqp091.Connection, *amqp091.Channel, amqp091.Queue, error) {
	log.Printf("Connecting to RabbitMQ at %s", url)
	for i := 0; i < 10; i++ {
		conn, err := amqp091.Dial(url)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				conn.Close()
				return nil, nil, amqp091.Queue{}, fmt.Errorf("failed to open a channel: %v", err)
			}

			log.Printf("Declaring queue %s", queueName)
			q, err := ch.QueueDeclare(
				queueName,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				ch.Close()
				conn.Close()
				return nil, nil, amqp091.Queue{}, fmt.Errorf("failed to declare a queue: %v", err)
			}
			return conn, ch, q, nil
		}
		log.Printf("Retrying connection (%d/10): %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return nil, nil, amqp091.Queue{}, fmt.Errorf("failed to connect to RabbitMQ after retries")
}

func main() {

	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/" // Для локального запуска
	}
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "task_queue"
	}

	conn, ch, q, err := connectToRabbitMQ(url, queueName)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	err = ch.Publish(
		"amq.direct", // Используем amq.direct
		q.Name,       // Routing key = имя очереди
		false,        // Mandatory
		false,        // Immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte("test message"),
			MessageId:   "test-1",
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Println("Test message published successfully")

	// Контекст без таймаута
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Каналы для результатов и ошибок
	resultCh := make(chan Result, 100)
	errCh := make(chan Error, 100)
	var wg sync.WaitGroup
	workers := 5
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(ctx, i, ch, q.Name, resultCh, errCh, &wg)
	}

	var results []Result
	var errResults []Error

	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	done := make(chan struct{})
	go func() {
		for {
			select {
			case r, ok := <-resultCh:
				if !ok {
					resultCh = nil
				} else {
					results = append(results, r)
					log.Printf("Received result: %+v", r)
				}
			case e, ok := <-errCh:
				if !ok {
					errCh = nil
				} else {
					errResults = append(errResults, e)
					log.Printf("Received error: %+v", e)
				}
			}
			if resultCh == nil && errCh == nil {
				close(done)
				return
			}
		}
	}()

	<-ctx.Done()
}

func worker(ctx context.Context, id int, ch *amqp091.Channel, queueName string, resultCh chan<- Result, errCh chan<- Error, wg *sync.WaitGroup) {
	defer wg.Done()

	msgs, err := ch.Consume(
		queueName,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		errCh <- Error{WorkerID: 0, Error: fmt.Errorf("worker %d failed to consume: %v", id, err)}
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			result := Result{
				MessageID: msg.MessageId,
				Processed: fmt.Sprintf("Processed: %s", string(msg.Body)),
			}
			log.Printf("Worker %d processed message %s: %s", id, msg.MessageId, result.Processed)
			resultCh <- result
			msg.Ack(false)
		}
	}
}

func processMessage(body []byte) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("Processed: %s", string(body)), nil
}
