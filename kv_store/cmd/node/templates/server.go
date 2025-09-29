// package main

// import (
// 	"context"
// 	"log"
// 	"net"
// 	"sync"

// 	pb "node/pb" // replace with your module path
// 	"google.golang.org/grpc"
// )

// type server struct {
// 	pb.UnimplementedKVStoreServer
// 	mu   sync.RWMutex
// 	data map[string]string
// }

// func (s *server) Put(ctx context.Context, in *pb.Item) (*pb.Response, error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	s.data[in.Key] = in.Value
// 	log.Printf("Put: key=%s, value=%s", in.Key, in.Value)
// 	return &pb.Response{Success: true, Message: "Key stored successfully"}, nil
// }

// func (s *server) Get(ctx context.Context, in *pb.Key) (*pb.Value, error) {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()

// 	value, exists := s.data[in.Key]
// 	log.Printf("Get: key=%s, found=%v", in.Key, exists)
// 	return &pb.Value{Value: value, Found: exists}, nil
// }

// func (s *server) Delete(ctx context.Context, in *pb.Key) (*pb.Response, error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	_, exists := s.data[in.Key]
// 	if exists {
// 		delete(s.data, in.Key)
// 		log.Printf("Delete: key=%s, success=true", in.Key)
// 		return &pb.Response{Success: true, Message: "Key deleted successfully"}, nil
// 	}
// 	log.Printf("Delete: key=%s, success=false", in.Key)
// 	return &pb.Response{Success: false, Message: "Key not found"}, nil
// }

// func main() {
// 	lis, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}

// 	grpcServer := grpc.NewServer()
// 	s := &server{
// 		data: make(map[string]string),
// 	}
// 	pb.RegisterKVStoreServer(grpcServer, s)

// 	log.Println("KV Store server listening on :50051")
// 	if err := grpcServer.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
