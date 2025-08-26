package server

//import (
//	"context"
//	pb "github.com/lvg-erp/sandbox/grTest7/04_gRPC/proto"
//	"log"
//	"sync"
//	"time"
//)
//
//type messageExchangeServer struct {
//	pb.UnimplementedMessageExchangeServiceServer
//	mu                sync.Mutex
//	processedMessages []string
//}
//
//func (s *messageExchangeServer) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
//	time.Sleep(100 * (time.Millisecond))
//	processed := "Processed: " + req.Content
//	s.mu.Lock()
//	s.processedMessages = append(s.processedMessages, processed)
//	s.mu.Unlock()
//
//	log.Printf("Processed message ID %s: %s", req.Id, processed)
//
//	return &pb.MessageResponse{
//		Id:          req.Id,
//		Received:    processed,
//		ProcessedAt: time.Now().Format(time.RFC3339),
//	}, nil
//}
//
//func (s *messageExchangeServer) StreamMessages(stream pb.MessageExchangeService_StreamMessageServer) error {
//	var wg sync.WaitGroup
//	resultCh := make(chan *pb.MessageRequest, 10)
//
//	go func() {
//		for {
//			req, err := stream.Recv()
//		}
//	}
//}
