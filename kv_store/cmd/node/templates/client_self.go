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
		"dns:///localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer conn.Close()
	c := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := c.Get(ctx, &pb.Key{Key: "1"})

	fmt.Printf("Put Response: %v, Error: %v", resp, err)
}
