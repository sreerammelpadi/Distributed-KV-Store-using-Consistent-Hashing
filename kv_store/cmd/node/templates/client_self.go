package main

import (
	"context"
	"fmt"
	pb "kv_store/api/pb/coordinator"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, _ := grpc.NewClient(
		"dns:///localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer conn.Close()
	c := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Put(ctx, &pb.Item{Key: "1", Value: "one"})
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Put Response: %v, Error: %v", resp, err)
}
