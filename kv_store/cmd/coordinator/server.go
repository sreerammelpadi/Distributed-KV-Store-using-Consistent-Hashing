package main

import (
	"context"
	"errors"
	pb "kv_store/api/pb/coordinator"
	clpb "kv_store/api/pb/node"
	config "kv_store/config"
	"time"
)

type server struct {
	pb.UnimplementedCoordinatorServer
	nodes map[string]clpb.KVStoreClient
	ring  *HashRing
	cfg   *config.Config
}

type response struct {
	Success bool
	Value   string
	Node    string
	Err     error
}

type QuorumFunc func(ctx context.Context, client clpb.KVStoreClient) (bool, string, error)

func (s *server) doQuorum(ctx context.Context, key string, quorum int, fn QuorumFunc) (bool, string, error) {
	nodes := s.ring.GetNodesForKey(key, s.cfg.Consistency.N)

	ackCh := make(chan response, len(nodes))

	// Launch RPCs to all nodes. The first W successful acks will allow us to
	// return to the client; remaining node RPCs will continue in background
	// using a separate background context so they are not canceled when the
	// incoming request ctx is done.
	for _, node := range nodes {
		go func(n string) {
			client, ok := s.nodes[n]
			if !ok {
				// non-blocking send to ackCh to avoid blocking if nobody is
				// reading it anymore; ackCh is buffered so the first few will succeed
				select {
				case ackCh <- response{Success: false, Value: "", Node: n, Err: errors.New("node doesn't exist")}:
				default:
				}
				return
			}

			// Use the incoming ctx for fast-path synchronous calls so that a
			// slow node doesn't hold up getting the quorum. For background
			// replication we will issue a separate call using a background ctx.
			suc, val, err := fn(ctx, client)
			select {
			case ackCh <- response{Success: suc, Value: val, Node: n, Err: err}:
			default:
				// If ackCh is full (we returned early and nobody is draining it),
				// drop the diagnostic ack to avoid blocking.
			}
		}(node)
	}

	totalresp := 0
	acks := 0
	found := false
	value := ""
	timeout := time.After(1 * time.Second)

	// drain until we reach quorum or timeout; after quorum is reached we
	// return but continue background replication below
	acked := make(map[string]bool, len(nodes))
outloop:
	for acks < quorum && totalresp < len(nodes) {
		select {
		case resp := <-ackCh:
			totalresp++
			if resp.Err == nil {
				acks++
				if resp.Success {
					found = true
					value = resp.Value // Use last value for now
					acked[resp.Node] = true
				}
			}
		case <-timeout:
			break outloop
		}
	}

	if acks < quorum {
		return false, "", errors.New("quorum failed")
	}

	// We have quorum. Spawn background replication for the nodes that either
	// didn't ack yet or were slower — use a background context with timeout so
	// background operations are independent of the request ctx.
	go func(key string, nodes []string) {
		bgTimeout := 5 * time.Second

		// If the original request context is still alive, the original
		// in-flight RPC may still complete for the remaining nodes. Only
		// perform background replication if the original context is already
		// cancelled/expired to avoid duplicate calls.
		if ctx.Err() == nil {
			return
		}

		for _, n := range nodes {
			// Skip nodes that already succeeded
			if acked[n] {
				continue
			}
			// For each remaining node, issue a background call with its own context.
			go func(n string) {
				client, ok := s.nodes[n]
				if !ok {
					return
				}
				bgCtx, cancel := context.WithTimeout(context.Background(), bgTimeout)
				defer cancel()
				// Use fn but with bgCtx; ignore return values and errors here.
				_, _, _ = fn(bgCtx, client)
			}(n)
		}
	}(key, nodes)

	return found, value, nil
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {

	putquorum := func(ctx context.Context, client clpb.KVStoreClient) (bool, string, error) {
		resp, err := client.Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
		if err != nil {
			return false, "", err
		}
		return resp.Success, "", nil
	}

	suc, _, err := s.doQuorum(ctx, item.Key, s.cfg.Consistency.W, putquorum)
	if !suc {
		return &pb.Response{Success: false, Message: "Put Failed"}, err
	}

	// if !ok {
	// 	msg := "error in getting node from HashRing"
	// 	return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	// }
	// client, ok := s.nodes[node]
	// if !ok || client == nil {
	// 	msg := fmt.Sprintf("no client available for node %v", node)
	// 	return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	// }
	// suc, err := client.Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
	// if err != nil {
	// 	return &pb.Response{Success: false, Message: err.Error()}, err
	// }
	// if suc == nil {
	// 	msg := "nil response from node Put"
	// 	return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	// }

	return &pb.Response{Success: true, Message: "Put Success"}, nil
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {

	getquorum := func(ctx context.Context, client clpb.KVStoreClient) (bool, string, error) {
		resp, err := client.Get(ctx, &clpb.Key{Key: key.Key})
		if err != nil {
			return false, "", err
		}
		return resp.Found, resp.Value, nil
	}

	suc, val, err := s.doQuorum(ctx, key.Key, s.cfg.Consistency.R, getquorum)
	if err != nil {
		return &pb.Value{Found: false, Value: ""}, err
	}
	if !suc {
		return &pb.Value{Found: false, Value: ""}, nil
	}

	return &pb.Value{Value: val, Found: suc}, nil

}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {

	delquorum := func(ctx context.Context, client clpb.KVStoreClient) (bool, string, error) {
		_, err := client.Delete(ctx, &clpb.Key{Key: key.Key})
		return true, "", err
	}

	suc, _, err := s.doQuorum(ctx, key.Key, s.cfg.Consistency.W, delquorum)
	if !suc {
		return &pb.Response{Success: false, Message: "Delete Failed"}, err
	}

	return &pb.Response{Success: suc, Message: "Delete Sucess"}, nil
}
