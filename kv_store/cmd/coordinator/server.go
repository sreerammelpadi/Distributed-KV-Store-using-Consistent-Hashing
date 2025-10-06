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
	Err     error
}

type QuorumFunc func(ctx context.Context, client clpb.KVStoreClient) (bool, string, error)

func (s *server) doQuorum(ctx context.Context, key string, quorum int, fn QuorumFunc) (bool, string, error) {
	nodes := s.ring.GetNodesForKey(key, s.cfg.Consistency.N)

	ackCh := make(chan response, len(nodes))
	for _, node := range nodes {
		go func(n string) {
			client, ok := s.nodes[n]
			if !ok {
				ackCh <- response{
					Success: false,
					Value:   "",
					Err:     errors.New("node doesn't exist"),
				}
				return
			}
			suc, val, err := fn(ctx, client)
			ackCh <- response{
				Success: suc,
				Value:   val,
				Err:     err,
			}
		}(node)
	}

	totalresp := 0
	acks := 0
	found := false
	value := ""
	timeout := time.After(1 * time.Second)
outloop:
	for acks < s.cfg.Consistency.W && totalresp < s.cfg.Consistency.N {
		select {
		case resp := <-ackCh:
			totalresp++
			if resp.Err == nil {
				acks++
				if resp.Success {
					found = true
					value = resp.Value // choose last successful value observed
				}
			}
		case <-timeout:
			break outloop
		}
	}

	if acks < quorum {
		msg := "quorum failed"
		return false, "", errors.New(msg)
	}
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
