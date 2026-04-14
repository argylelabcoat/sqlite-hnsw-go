package hnsw

import (
	"fmt"

	cbor "github.com/fxamacker/cbor/v2"
)

// graphSnapshot is the CBOR-serializable representation of a Graph.
type graphSnapshot struct {
	EntryPoint  int
	MaxLevel    int
	M           int
	EfConstruct int
	EfSearch    int
	Nodes       []nodeSnapshot
}

// nodeSnapshot is the CBOR-serializable representation of a Node.
type nodeSnapshot struct {
	ID     int
	Vector []float32
	Level  int
	Edges  [][]int
}

// Serialize encodes the graph to a CBOR byte slice.
func (g *Graph) Serialize() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	snap := graphSnapshot{
		EntryPoint:  g.entryPoint,
		MaxLevel:    g.maxLevel,
		M:           g.m,
		EfConstruct: g.efConstruct,
		EfSearch:    g.efSearch,
		Nodes:       make([]nodeSnapshot, 0, len(g.nodes)),
	}

	for _, n := range g.nodes {
		snap.Nodes = append(snap.Nodes, nodeSnapshot{
			ID:     n.ID,
			Vector: n.Vector,
			Level:  n.Level,
			Edges:  n.Edges,
		})
	}

	data, err := cbor.Marshal(snap)
	if err != nil {
		return nil, fmt.Errorf("serialize HNSW graph: %w", err)
	}
	return data, nil
}
