package main

import (
	"fmt"
	"net"

	pb "kv_store/api/pb/coordinator"
	clpb "kv_store/api/pb/node"
	config "kv_store/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getServerObj(cfg *config.Config, ring *HashRing) *server {
	s := &server{
		nodes: make(map[string]clpb.KVStoreClient),
		ring:  ring,
	}

	for node, addr := range cfg.Nodes {
		fmt.Println(addr)
		ring.AddNode(addr)
		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			fmt.Printf("Error: Client connection fault for node: {%v : %v} : %v", node, addr, err)
		}

		s.nodes[addr] = clpb.NewKVStoreClient(conn)
	}
	return s
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	ring := NewHashRing(100)
	grpcServer := grpc.NewServer()
	s := getServerObj(cfg, ring)

	pb.RegisterCoordinatorServer(grpcServer, s)

	lis, _ := net.Listen("tcp", ":50051")

	fmt.Println("Coordinator Listening")
	err2 := grpcServer.Serve(lis)
	fmt.Print(err2)

}
