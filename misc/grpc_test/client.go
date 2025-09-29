package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "grpc_test/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Dial the server
	// conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Call SayHello
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Sreeram"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println("Greeting:", r.Message)
}
