package main

import (
	"context"
	"errors"
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
	client, ok := s.nodes[node]
	if !ok || client == nil {
		msg := fmt.Sprintf("no client available for node %d", node)
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}

	suc, err := client.Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, err
	}
	if suc == nil {
		msg := "nil response from node Put"
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	return &pb.Response{Success: suc.Success, Message: suc.Message}, nil
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	node := GetNode(key)
	client, ok := s.nodes[node]
	if !ok || client == nil {
		return &pb.Value{Value: "", Found: false}, fmt.Errorf("no client available for node %d", node)
	}
	resp, err := client.Get(ctx, &clpb.Key{Key: key.Key})
	if err != nil {
		return &pb.Value{Value: "", Found: false}, err
	}
	if resp == nil {
		return &pb.Value{Value: "", Found: false}, errors.New("nil response from node Get")
	}
	return &pb.Value{Value: resp.Value, Found: resp.Found}, nil
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
	node := GetNode(key)
	client, ok := s.nodes[node]
	if !ok || client == nil {
		msg := fmt.Sprintf("no client available for node %d", node)
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	resp, err := client.Delete(ctx, &clpb.Key{Key: key.Key})
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, err
	}
	if resp == nil {
		msg := "nil response from node Delete"
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	return &pb.Response{Success: resp.Success, Message: resp.Message}, nil
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
		fmt.Println(addr)
		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			fmt.Printf("Error: Client connection fault for node: {%v : %v} : %v", node, addr, err)
		}

		s.nodes[node] = clpb.NewKVStoreClient(conn)
	}

	pb.RegisterCoordinatorServer(grpcServer, s)

	lis, _ := net.Listen("tcp", ":50051")

	fmt.Println("Coordinator Listening")
	err2 := grpcServer.Serve(lis)
	fmt.Print(err2)

}
