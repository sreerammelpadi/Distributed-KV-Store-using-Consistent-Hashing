package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// HashRing implements a consistent hashing ring with virtual nodes.
type HashRing struct {
	virtualnodes int                 // number of virtual nodes per real node
	ring         []uint32            // sorted hash ring (virtual node hashes)
	vnodeMap     map[uint32]string   // hash -> real node
	nodes        map[string]struct{} // set of real nodes
	mu           sync.RWMutex
}

// NewHashRing creates a new ring. virtualnodes = number of virtual nodes per real node.
func NewHashRing(virtualnodes int) *HashRing {
	if virtualnodes <= 0 {
		virtualnodes = 100
	}
	return &HashRing{
		virtualnodes: virtualnodes,
		ring:         make([]uint32, 0),
		vnodeMap:     make(map[uint32]string),
		nodes:        make(map[string]struct{}),
	}
}

// md5ToUint32 returns the first 4 bytes of md5 hash as uint32 (big endian).
func md5ToUint32(key string) uint32 {
	sum := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(sum[:4])
}

// generateVNodeHash returns the hash for a given node + virtualnode index
func vnodeHash(node string, idx int) uint32 {
	return md5ToUint32(node + "#" + strconv.Itoa(idx))
}

// AddNode adds a real node and its virtual nodes to the ring.
func (hr *HashRing) AddNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[node]; exists {
		return // already present
	}
	hr.nodes[node] = struct{}{}

	inserted := make([]uint32, 0, hr.virtualnodes)
	for i := 0; i < hr.virtualnodes; i++ {
		h := vnodeHash(node, i)
		// map the vnode hash to the real node
		hr.vnodeMap[h] = node
		inserted = append(inserted, h)
	}
	// add to hr.ring and re-sort
	hr.ring = append(hr.ring, inserted...)
	sort.Slice(hr.ring, func(i, j int) bool { return hr.ring[i] < hr.ring[j] })
}

// RemoveNode removes a real node and its virtual nodes from the ring.
func (hr *HashRing) RemoveNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[node]; !exists {
		return // not present
	}
	delete(hr.nodes, node)

	// collect hashes to remove
	toRemove := make(map[uint32]struct{}, hr.virtualnodes)
	for i := 0; i < hr.virtualnodes; i++ {
		h := vnodeHash(node, i)
		toRemove[h] = struct{}{}
		delete(hr.vnodeMap, h)
	}

	// rebuild ring without removed hashes (efficient enough for typical sizes)
	newRing := make([]uint32, 0, len(hr.ring)-len(toRemove))
	for _, h := range hr.ring {
		if _, rem := toRemove[h]; !rem {
			newRing = append(newRing, h)
		}
	}
	hr.ring = newRing
}

// getHashOwner returns the real node responsible for the given hash
// assumes hr.mu read-locked or called from methods that have lock
func (hr *HashRing) getHashOwner(hash uint32) (string, bool) {
	if len(hr.ring) == 0 {
		return "", false
	}
	// binary search for the first vnode >= hash
	idx := sort.Search(len(hr.ring), func(i int) bool { return hr.ring[i] >= hash })
	if idx == len(hr.ring) {
		// wrap around to the first vnode
		idx = 0
	}
	owner, ok := hr.vnodeMap[hr.ring[idx]]
	return owner, ok
}

// GetNode returns the real node responsible for a key.
func (hr *HashRing) GetNode(key string) (string, bool) {
	h := md5ToUint32(key)
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.getHashOwner(h)
}

// GetNodesForKey returns up to 'count' distinct real nodes suitable for replication,
// by walking the ring clockwise and skipping duplicates. Use count=1 for single replica.
func (hr *HashRing) GetNodesForKey(key string, count int) []string {
	if count <= 0 {
		return nil
	}
	h := md5ToUint32(key)

	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.ring) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	results := make([]string, 0, count)
	// start at the first vnode >= hash
	idx := sort.Search(len(hr.ring), func(i int) bool { return hr.ring[i] >= h })
	if idx == len(hr.ring) {
		idx = 0
	}
	i := idx
	for len(results) < count && len(seen) < len(hr.nodes) {
		vhash := hr.ring[i]
		node := hr.vnodeMap[vhash]
		if _, ok := seen[node]; !ok {
			results = append(results, node)
			seen[node] = struct{}{}
		}
		i++
		if i >= len(hr.ring) {
			i = 0 // wrap
		}
	}
	return results
}

// Rebalance computes which keys must move given the current distribution `data`.
// `data` is a map: realNode -> map[key]value (we only examine keys here).
// Returns moves: map[srcNode]map[destNode][]keys
func (hr *HashRing) Rebalance(data map[string]map[string]string) map[string]map[string][]string {
	moves := make(map[string]map[string][]string)

	hr.mu.RLock()
	defer hr.mu.RUnlock()

	// for each key present on each node, compute the correct owner and record move if needed
	for srcNode, kvs := range data {
		for key := range kvs {
			h := md5ToUint32(key)
			destNode, ok := hr.getHashOwner(h)
			if !ok {
				// no nodes in ring; nothing to do
				continue
			}
			if destNode == srcNode {
				continue // key is already on correct node
			}
			// record move from srcNode -> destNode
			if _, ok := moves[srcNode]; !ok {
				moves[srcNode] = make(map[string][]string)
			}
			moves[srcNode][destNode] = append(moves[srcNode][destNode], key)
		}
	}
	return moves
}

// Helper to pretty-print moves
func printMoves(prefix string, moves map[string]map[string][]string) {
	if len(moves) == 0 {
		fmt.Println(prefix, "no moves")
		return
	}
	fmt.Println(prefix, "moves:")
	for src, dests := range moves {
		for dest, keys := range dests {
			fmt.Printf("  %s -> %s : %d keys -> %s\n", src, dest, len(keys), strings.Join(keys, ", "))
		}
	}
}

// Demo main: create ring, create sample data, add node, compute moves, remove node, compute moves.
func consistenthashing() {
	// Create a ring with 100 virtual nodes per real node
	ring := NewHashRing(100)

	// Add initial nodes
	ring.AddNode("node-A")
	ring.AddNode("node-B")
	ring.AddNode("node-C")

	// Create sample data distribution: node -> (key -> value)
	data := map[string]map[string]string{
		"node-A": {"apple": "1", "banana": "2", "cherry": "3"},
		"node-B": {"date": "4", "elderberry": "5", "fig": "6"},
		"node-C": {"grape": "7", "honeydew": "8", "kiwi": "9"},
	}

	fmt.Println("Initial ring nodes: node-A, node-B, node-C")
	// Check which node currently owns each key (per current ring)
	fmt.Println("\nCurrent owners (before changes):")
	for node, kvs := range data {
		for k := range kvs {
			owner, _ := ring.GetNode(k)
			fmt.Printf("  key=%s on %s  -> owner=%s\n", k, node, owner)
		}
	}

	// Add a new node and compute moves
	fmt.Println("\nAdding node-D to the ring")
	ring.AddNode("node-D")
	movesOnAdd := ring.Rebalance(data)
	printMoves("After add", movesOnAdd)

	// Simulate applying moves: move keys from src to dest in data map
	applyMoves(data, movesOnAdd)

	// Now remove a node and compute moves
	fmt.Println("\nRemoving node-B from the ring")
	ring.RemoveNode("node-B")
	movesOnRemove := ring.Rebalance(data)
	printMoves("After remove", movesOnRemove)
}

// applyMoves applies the computed moves to the data map (simple simulation)
func applyMoves(data map[string]map[string]string, moves map[string]map[string][]string) {
	for src, dests := range moves {
		for dest, keys := range dests {
			// ensure dest map exists
			if _, ok := data[dest]; !ok {
				data[dest] = make(map[string]string)
			}
			for _, k := range keys {
				// move value if present
				if val, ok := data[src][k]; ok {
					data[dest][k] = val
					delete(data[src], k)
				} else {
					// key missing on src - ignore or log
				}
			}
		}
	}
}
