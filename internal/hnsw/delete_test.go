package hnsw

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete_NodeExists_Removed(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	g.Insert(1, []float32{0, 1})

	g.Delete(1)

	assert.Equal(t, 1, g.Len())

	g.mu.RLock()
	_, exists := g.nodes[1]
	g.mu.RUnlock()
	assert.False(t, exists, "node 1 should be removed")
}

func TestDelete_NodeHasNeighbors_EdgesCleaned(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	g.Insert(1, []float32{0, 1})
	g.Insert(2, []float32{-1, 0})

	g.Delete(2)

	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, neighborID := range g.nodes[0].Edges[0] {
		assert.NotEqual(t, 2, neighborID, "node 0 should not reference deleted node 2")
	}
	for _, neighborID := range g.nodes[1].Edges[0] {
		assert.NotEqual(t, 2, neighborID, "node 1 should not reference deleted node 2")
	}
}

func TestDelete_SearchNoLongerReturnsDeletedNode(t *testing.T) {
	g := testPopulatedGraph()
	g.Delete(0)

	results := g.Search([]float32{1, 0}, 5)
	for _, r := range results {
		assert.NotEqual(t, 0, r.ID, "deleted node should not appear in results")
	}
}

func TestDelete_Nonexistent_Noop(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	g.Delete(999)
	assert.Equal(t, 1, g.Len())
}

func testPopulatedGraph() *Graph {
	g := testNewGraph()
	for i := 0; i < 20; i++ {
		angle := float32(i) * 2 * math.Pi / 20
		vec := []float32{float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))}
		g.Insert(i, vec)
	}
	return g
}
