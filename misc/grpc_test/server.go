package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "grpc_test/pb"

	"google.golang.org/grpc"
)

// GreeterServer implements the generated interface
type GreeterServer struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements the SayHello RPC
func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + req.Name}, nil
}

func main() {
	// Listen on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &GreeterServer{})

	fmt.Println("Server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
