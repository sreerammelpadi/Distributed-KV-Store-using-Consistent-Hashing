package main

import (
	"context"
	"fmt"
	"net"
	pb "node/pb"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKVStoreServer
	mu   sync.RWMutex
	data map[string]string
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {
	s.data[item.Key] = item.Value
	fmt.Printf("Put: %v\n", s.data)
	return &pb.Response{Success: true, Message: "success"}, nil
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	val, ok := s.data[key.Key]
	return &pb.Value{Value: val, Found: ok}, nil
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
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
	grpcServer.Serve(lis)

}
