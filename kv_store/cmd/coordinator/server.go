package main

import (
	"context"
	"errors"
	"fmt"
	pb "kv_store/api/pb/coordinator"
	clpb "kv_store/api/pb/node"
)

type server struct {
	pb.UnimplementedCoordinatorServer
	nodes map[string]clpb.KVStoreClient
	ring  *HashRing
}

func (s *server) Put(ctx context.Context, item *pb.Item) (*pb.Response, error) {
	node, ok := s.ring.GetNode(item.Key)
	if !ok {
		msg := "error in getting node from HashRing"
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	client, ok := s.nodes[node]
	if !ok || client == nil {
		msg := fmt.Sprintf("no client available for node %d", node)
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	suc, err := client.Put(ctx, &clpb.Item{Key: item.Key, Value: item.Value})
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, err
	}
	if suc == nil {
		msg := "nil response from node Put"
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	return &pb.Response{Success: suc.Success, Message: suc.Message}, nil
}

func (s *server) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	node, ok := s.ring.GetNode(key.Key)
	if !ok {
		msg := fmt.Sprintf("Error in getting node from HashRing")
		return &pb.Value{Value: "", Found: false}, errors.New(msg)
	}
	client, ok := s.nodes[node]
	if !ok || client == nil {
		return &pb.Value{Value: "", Found: false}, fmt.Errorf("no client available for node %d", node)
	}
	resp, err := client.Get(ctx, &clpb.Key{Key: key.Key})
	if err != nil {
		return &pb.Value{Value: "", Found: false}, err
	}
	if resp == nil {
		return &pb.Value{Value: "", Found: false}, errors.New("nil response from node Get")
	}
	return &pb.Value{Value: resp.Value, Found: resp.Found}, nil
}

func (s *server) Delete(ctx context.Context, key *pb.Key) (*pb.Response, error) {
	node, ok := s.ring.GetNode(key.Key)
	if !ok {
		msg := fmt.Sprintf("Error in getting node from HashRing")
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	client, ok := s.nodes[node]
	if !ok || client == nil {
		msg := fmt.Sprintf("no client available for node %d", node)
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	resp, err := client.Delete(ctx, &clpb.Key{Key: key.Key})
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, err
	}
	if resp == nil {
		msg := "nil response from node Delete"
		return &pb.Response{Success: false, Message: msg}, errors.New(msg)
	}
	return &pb.Response{Success: resp.Success, Message: resp.Message}, nil
}
