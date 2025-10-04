// Usage
// go run client_self.go --put 1 one
// # Get example:
// go run client_self.go --get 1
// # Delete example:
// go run client_self.go --del 1
// # Use a custom coordinator address (e.g. docker-compose mapping to host 50055)
// go run client_self.go --put 1 one --addr localhost:50055
// # Adjust timeout
// go run client_self.go --put 1 one --timeout 5s

package main

import (
	"context"
	"flag"
	"fmt"
	pb "kv_store/api/pb/coordinator"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func usageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  --put <key> <value>    Put a key/value pair\n")
	fmt.Fprintf(os.Stderr, "  --get <key>            Get a key\n")
	fmt.Fprintf(os.Stderr, "  --del <key>            Delete a key\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	putFlag := flag.Bool("put", false, "run a Put operation (requires two positional args: key value)")
	getFlag := flag.Bool("get", false, "run a Get operation (requires one positional arg: key)")
	delFlag := flag.Bool("del", false, "run a Delete operation (requires one positional arg: key)")
	addr := flag.String("addr", "localhost:50051", "coordinator address (host:port)")
	timeout := flag.Duration("timeout", 3*time.Second, "per-request timeout")

	flag.Parse()
	args := flag.Args()

	modes := 0
	if *putFlag {
		modes++
	}
	if *getFlag {
		modes++
	}
	if *delFlag {
		modes++
	}
	if modes != 1 {
		fmt.Fprintln(os.Stderr, "please supply exactly one of --put, --get or --del")
		usageAndExit()
	}

	// Validate args for selected mode
	if *putFlag {
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "--put requires two positional args: <key> <value>")
			usageAndExit()
		}
	} else {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "--get/--del requires one positional arg: <key>")
			usageAndExit()
		}
	}

	// Dial coordinator with timeout and WithBlock so we know connection status
	// dialCtx, dialCancel := context.WithTimeout(context.Background(), *timeout)
	// defer dialCancel()
	conn, err := grpc.NewClient(
		*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to dial coordinator at %s: %v", *addr, err)
	}
	defer conn.Close()

	c := pb.NewCoordinatorClient(conn)

	if *putFlag {
		key := args[0]
		val := args[1]
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		start := time.Now()
		resp, err := c.Put(ctx, &pb.Item{Key: key, Value: val})
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("Put Error: %v; elapsed=%v\n", err, elapsed)
			os.Exit(1)
		}
		fmt.Printf("Put Response: %+v; elapsed=%v\n", resp, elapsed)
		return
	}

	if *getFlag {
		key := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		start := time.Now()
		resp, err := c.Get(ctx, &pb.Key{Key: key})
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("Get Error: %v; elapsed=%v\n", err, elapsed)
			os.Exit(1)
		}
		if !resp.Found {
			fmt.Printf("Get: key=%s not found; elapsed=%v\n", key, elapsed)
		} else {
			fmt.Printf("Get Response: value=%s; elapsed=%v\n", resp.Value, elapsed)
		}
		return
	}

	if *delFlag {
		key := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		start := time.Now()
		resp, err := c.Delete(ctx, &pb.Key{Key: key})
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("Delete Error: %v; elapsed=%v\n", err, elapsed)
			os.Exit(1)
		}
		fmt.Printf("Delete Response: %+v; elapsed=%v\n", resp, elapsed)
		return
	}
}
