package hnsw

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert_EmptyGraph_SetsEntryPoint(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})

	g.mu.RLock()
	defer g.mu.RUnlock()

	assert.Equal(t, 1, len(g.nodes))
	assert.Equal(t, 0, g.entryPoint)
}

func TestInsert_TwoNodes_BidirectionalEdges(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	g.Insert(1, []float32{0, 1})

	g.mu.RLock()
	defer g.mu.RUnlock()

	require.Contains(t, g.nodes, 0)
	require.Contains(t, g.nodes, 1)
	assert.Contains(t, g.nodes[0].Edges[0], 1, "node 0 should connect to node 1")
	assert.Contains(t, g.nodes[1].Edges[0], 0, "node 1 should connect to node 0")
}

func TestInsert_50Nodes_AllReachable(t *testing.T) {
	g := testNewGraph()
	for i := range 50 {
		angle := float32(i) * 2 * math.Pi / 50
		vec := []float32{float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))}
		g.Insert(i, vec)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	assert.Equal(t, 50, len(g.nodes))
	for i := range 50 {
		require.Contains(t, g.nodes, i, "node %d should exist", i)
		assert.NotEmpty(t, g.nodes[i].Edges[0], "node %d should have level-0 edges", i)
	}
}

func testNewGraph() *Graph {
	g := NewGraph(16, 200, 64, CosineDistance)
	g.rng = rand.New(rand.NewSource(42))
	return g
}
