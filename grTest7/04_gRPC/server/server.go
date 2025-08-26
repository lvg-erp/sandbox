package server

import (
	"context"
	pb "github.com/lvg-erp/sandbox/grTest7/04_gRPC/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type messageExchangeServer struct {
	pb.UnimplementedMessageExchangeServiceServer
	mu                sync.Mutex
	processedMessages []string
}

func (s *messageExchangeServer) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	time.Sleep(100 * (time.Millisecond))
	processed := "Processed: " + req.Content
	s.mu.Lock()
	s.processedMessages = append(s.processedMessages, processed)
	s.mu.Unlock()

	log.Printf("Processed message ID %s: %s", req.Id, processed)

	return &pb.MessageResponse{
		Id:          req.Id,
		Received:    processed,
		ProcessedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *messageExchangeServer) StreamMessages(stream pb.MessageExchangeService_StreamMessageServer) error {
	var wg sync.WaitGroup
	resultCh := make(chan *pb.MessageResponse, 10)

	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				close(resultCh)
				return
			}
			if err != nil {
				log.Printf("Stream error: %v", err)
				return
			}

			wg.Add(1)
			go func(req *pb.MessageRequest) {
				defer wg.Done()
				time.Sleep(100 * time.Millisecond)
				processed := "Stream processed: " + req.Content
				s.mu.Lock()
				s.processedMessages = append(s.processedMessages, processed)
				s.mu.Unlock()

				resultCh <- &pb.MessageResponse{
					Id:          req.Id,
					Received:    processed,
					ProcessedAt: time.Now().Format(time.RFC3339),
				}
			}(req)
		}
	}()

	for resp := range resultCh {
		if err := stream.Send(resp); err != nil {
			log.Printf("Failed to send response: %v", err)
			return err
		}
		log.Printf("Sent response for message ID %s: %s", resp.Id, resp.Received)
	}

	wg.Wait()
	return nil

}

func RunServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageExchangeServiceServer(s, &messageExchangeServer{})
	log.Println("Server is running on port 50051.....")

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Printf("Server stopped: %v", err)
	}

}
