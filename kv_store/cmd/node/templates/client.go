//go:build ignore
// +build ignore

package main

import (
	"context"
	"log"
	"time"

	pb "kv_store/api/pb/Node" // replace with your module path

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Create the ClientConn without performing I/O; connection happens lazily on first RPC.
	conn, err := grpc.NewClient(
		"dns:///localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	c := pb.NewKVStoreClient(conn)

	// Per-RPC timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test Put operation
	log.Println("Testing Put operation...")
	putResp, err := c.Put(ctx, &pb.Item{Key: "test-key", Value: "test-value"})
	if err != nil {
		log.Fatalf("could not put: %v", err)
	}
	log.Printf("Put response: success=%v, message=%s", putResp.Success, putResp.Message)

	// Test Get operation
	log.Println("Testing Get operation...")
	getResp, err := c.Get(ctx, &pb.Key{Key: "test-key"})
	if err != nil {
		log.Fatalf("could not get: %v", err)
	}
	log.Printf("Get response: value=%s, found=%v", getResp.Value, getResp.Found)

	// Test Get operation for non-existent key
	log.Println("Testing Get operation for non-existent key...")
	getResp2, err := c.Get(ctx, &pb.Key{Key: "non-existent-key"})
	if err != nil {
		log.Fatalf("could not get: %v", err)
	}
	log.Printf("Get response: value=%s, found=%v", getResp2.Value, getResp2.Found)

	// Test Delete operation
	log.Println("Testing Delete operation...")
	deleteResp, err := c.Delete(ctx, &pb.Key{Key: "test-key"})
	if err != nil {
		log.Fatalf("could not delete: %v", err)
	}
	log.Printf("Delete response: success=%v, message=%s", deleteResp.Success, deleteResp.Message)

	// Test Get operation after delete
	log.Println("Testing Get operation after delete...")
	getResp3, err := c.Get(ctx, &pb.Key{Key: "test-key"})
	if err != nil {
		log.Fatalf("could not get: %v", err)
	}
	log.Printf("Get response: value=%s, found=%v", getResp3.Value, getResp3.Found)
}
