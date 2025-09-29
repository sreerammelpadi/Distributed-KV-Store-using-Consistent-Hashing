package main

import (
	"context"
	"fmt"
	pb "kv_store/api/pb/coordinator"
	clpb "kv_store/api/pb/node"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedCoordinatorServer
	mu     sync.RWMutex
	data   map[string]string
	client clpb.KVStoreClient
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	suc, err := s.client.Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
	// time.Sleep(time.Duration(2 * time.Second))
	return &pb.Response{Success: suc.Success, Message: suc.Message}, err
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp, err := s.client.Get(ctx, &clpb.Key{Key: key.Key})
	return &pb.Value{Value: resp.Value, Found: resp.Found}, err
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.Delete(ctx, &clpb.Key{Key: key.Key})
	return &pb.Response{Success: resp.Success, Message: resp.Message}, err
}

func main() {
	grpcServer := grpc.NewServer()
	conn, err := grpc.NewClient(
		"dns:///localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("Error: Client connection fault: %v", err)
		return
	}
	defer conn.Close()

	s := &server{
		data:   make(map[string]string),
		client: clpb.NewKVStoreClient(conn),
	}
	pb.RegisterCoordinatorServer(grpcServer, s)

	lis, _ := net.Listen("tcp", ":50052")

	fmt.Println("Coordinator Listening")
	err2 := grpcServer.Serve(lis)
	fmt.Print(err2)

}
