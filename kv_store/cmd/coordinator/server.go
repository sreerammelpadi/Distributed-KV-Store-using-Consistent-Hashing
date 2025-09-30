package main

import (
	"context"
	"fmt"
	pb "kv_store/api/pb/coordinator"
	clpb "kv_store/api/pb/node"
	config "kv_store/config"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedCoordinatorServer
	data  map[string]string
	nodes map[int]clpb.KVStoreClient
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {
	node := GetNode(&pb.Key{Key: item.Key})
	suc, err := s.nodes[node].Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
	// time.Sleep(time.Duration(2 * time.Second))
	return &pb.Response{Success: suc.Success, Message: suc.Message}, err
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	node := GetNode(key)
	resp, err := s.nodes[node].Get(ctx, &clpb.Key{Key: key.Key})
	return &pb.Value{Value: resp.Value, Found: resp.Found}, err
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
	node := GetNode(key)
	resp, err := s.nodes[node].Delete(ctx, &clpb.Key{Key: key.Key})
	return &pb.Response{Success: resp.Success, Message: resp.Message}, err
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	grpcServer := grpc.NewServer()
	s := &server{
		data:  make(map[string]string),
		nodes: make(map[int]clpb.KVStoreClient),
	}

	for node, addr := range cfg.Nodes {
		conn, err := grpc.NewClient(
			"dns:///localhost"+addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			fmt.Printf("Error: Client connection fault for node: {%v : %v} : %v", node, addr, err)
		}

		s.nodes[node] = clpb.NewKVStoreClient(conn)
	}

	pb.RegisterCoordinatorServer(grpcServer, s)

	lis, _ := net.Listen("tcp", ":50055")

	fmt.Println("Coordinator Listening")
	err2 := grpcServer.Serve(lis)
	fmt.Print(err2)

}
