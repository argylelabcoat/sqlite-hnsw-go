package hnsw

import (
	"math/rand"
	"sync"
	"time"
)

type Node struct {
	ID     int
	Vector []float32
	Level  int
	Edges  [][]int
}

type Neighbor struct {
	ID   int
	Dist float32
}

type Graph struct {
	nodes       map[int]*Node
	entryPoint  int
	maxLevel    int
	m           int
	efConstruct int
	efSearch    int
	dist        DistanceFunc
	mu          sync.RWMutex
	rng         *rand.Rand
}

func NewGraph(m, efConstruct, efSearch int, dist DistanceFunc) *Graph {
	return &Graph{
		nodes:       make(map[int]*Node),
		m:           m,
		efConstruct: efConstruct,
		efSearch:    efSearch,
		dist:        dist,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Graph) Len() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}
