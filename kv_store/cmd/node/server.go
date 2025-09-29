package main

import (
	"context"
	"fmt"
	pb "kv_store/api/pb"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKVStoreServer
	mu   sync.RWMutex
	data map[string]string
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[item.Key] = item.Value
	fmt.Printf("Put: %v\n", s.data)
	// time.Sleep(time.Duration(2 * time.Second))
	return &pb.Response{Success: true, Message: "success"}, nil
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key.Key]
	return &pb.Value{Value: val, Found: ok}, nil
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key.Key)
	fmt.Printf("Delete: %v\n", s.data)
	return &pb.Response{Success: true, Message: "success"}, nil
}

func main() {
	grpcServer := grpc.NewServer()
	s := &server{
		data: make(map[string]string),
	}
	pb.RegisterKVStoreServer(grpcServer, s)

	lis, _ := net.Listen("tcp", ":50051")

	fmt.Println("Server Listening")
	err := grpcServer.Serve(lis)
	fmt.Print(err)

}
