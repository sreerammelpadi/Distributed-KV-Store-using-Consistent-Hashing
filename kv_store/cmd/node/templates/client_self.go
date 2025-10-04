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

	// start := time.Now()
	// resp, err := c.Put(ctx, &pb.Item{Key: "3", Value: "three"})
	// elapsed := time.Since(start)

	// if err != nil {
	// 	fmt.Printf("Put Error: %v; elapsed=%v\n", err, elapsed)
	// 	return
	// }

	// fmt.Printf("Put Response: %v; elapsed=%v\n", resp, elapsed)

	start := time.Now()
	resp, err := c.Get(ctx, &pb.Key{Key: "4"})
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Get Error: %v; elapsed=%v\n", err, elapsed)
		return
	}

	if !resp.Found {
		fmt.Printf("Get Response: key not found\n")
	} else {
		fmt.Printf("Get Response: %v; elapsed=%v\n", resp, elapsed)
	}

}
