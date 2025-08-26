package client

import (
	"context"
	"fmt"
	pb "github.com/lvg-erp/sandbox/grTest7/04_gRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

func RunClient(ctx context.Context, wg *sync.WaitGroup, result chan<- string) {
	defer wg.Done()

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		result <- fmt.Sprintf("Failed to connect: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewMessageExchangeServiceClient(conn)

	stream, err := client.StreamMessages(ctx)
	if err != nil {
		result <- fmt.Sprintf("Failed to create stream: %v", err)
		return
	}

	clientWg := sync.WaitGroup{}
	clientWg.Add(1)
	go func() {
		defer clientWg.Done()
		for {
			resp, err := stream.Recv()
			if err != nil {
				result <- fmt.Sprintf("Stream receive error :%v", err)
				return
			}
			result <- fmt.Sprintf("Received: ID %s, %s at %s", resp.Id, resp.Received, resp.ProcessedAt)
		}
	}()

	messages := []struct {
		id      string
		content string
	}{
		{"msg1", "Stream message 1"},
		{"msg2", "Stream message 2"},
		{"msg3", "Stream message 3"},
	}

	sendWg := sync.WaitGroup{}
	for _, msg := range messages {
		sendWg.Add(1)
		go func(id, content string) {
			defer sendWg.Done()
			if err := stream.Send(&pb.MessageRequest{Id: id, Content: content}); err != nil {
				result <- fmt.Sprintf("Failed to send message %s: %v", id, err)
				return
			}
			log.Printf("Sent message Id %s: %s", id, content)
			time.Sleep(200 * time.Millisecond)
		}(msg.id, msg.content)
	}

	sendWg.Wait()
	stream.CloseSend()
	clientWg.Wait()

}
