package hnsw

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomLevel_ExponentialDecay(t *testing.T) {
	g := NewGraph(16, 200, 64, CosineDistance)
	g.rng = rand.New(rand.NewSource(42))

	levels := make(map[int]int)
	for range 10000 {
		levels[g.randomLevel()]++
	}

	assert.Greater(t, levels[0], levels[1], "level 0 should be most common")
	assert.Greater(t, levels[1], levels[2], "higher levels should be rarer")
}

func TestRandomLevel_DeterministicWithSeed(t *testing.T) {
	g1 := NewGraph(16, 200, 64, CosineDistance)
	g1.rng = rand.New(rand.NewSource(99))
	results1 := make([]int, 10)
	for i := range results1 {
		results1[i] = g1.randomLevel()
	}

	g2 := NewGraph(16, 200, 64, CosineDistance)
	g2.rng = rand.New(rand.NewSource(99))
	results2 := make([]int, 10)
	for i := range results2 {
		results2[i] = g2.randomLevel()
	}

	assert.Equal(t, results1, results2, "same seed must produce same levels")
}

func TestSearchLayer_FindsClosestNodes(t *testing.T) {
	g := testGraphTriangle()

	query := []float32{1, 0}
	results := g.searchLayer(query, []int{0}, 3, 0)

	assert.Len(t, results, 3)
	assert.Equal(t, 0, results[0].id, "node 0 (same as query) should be closest")
	assert.InDelta(t, float32(0), results[0].dist, 1e-6)
}

func TestSearchLayer_StartsFromEntryPoint(t *testing.T) {
	g := testGraphTriangle()

	query := []float32{-1, 0}
	results := g.searchLayer(query, []int{0}, 3, 0)

	assert.Len(t, results, 3)
	assert.Equal(t, 2, results[0].id, "node 2 ([-1,0]) should be closest to [-1,0] query")
}

func testGraphTriangle() *Graph {
	g := NewGraph(16, 200, 64, CosineDistance)
	g.nodes[0] = &Node{ID: 0, Vector: []float32{1, 0}, Level: 0, Edges: [][]int{{1, 2}}}
	g.nodes[1] = &Node{ID: 1, Vector: []float32{0, 1}, Level: 0, Edges: [][]int{{0, 2}}}
	g.nodes[2] = &Node{ID: 2, Vector: []float32{-1, 0}, Level: 0, Edges: [][]int{{0, 1}}}
	g.entryPoint = 0
	g.maxLevel = 0
	return g
}
