package hnsw

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_EmptyGraph_ReturnsEmpty(t *testing.T) {
	g := testNewGraph()
	results := g.Search([]float32{1, 0}, 5)
	assert.Empty(t, results)
}

func TestSearch_SingleNode_ReturnsIt(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	results := g.Search([]float32{1, 0}, 5)
	require.Len(t, results, 1)
	assert.Equal(t, 0, results[0].ID)
	assert.InDelta(t, float32(0), results[0].Dist, 1e-6)
}

func TestSearch_ClosestNodeFirst(t *testing.T) {
	g := testNewGraph()
	g.Insert(0, []float32{1, 0})
	g.Insert(1, []float32{0, 1})
	g.Insert(2, []float32{-1, 0})

	results := g.Search([]float32{1, 0}, 3)
	require.Len(t, results, 3)
	assert.Equal(t, 0, results[0].ID, "node 0 ([1,0]) should be closest to query [1,0]")
}

func TestSearch_TopK(t *testing.T) {
	g := testNewGraph()
	for i := 0; i < 20; i++ {
		angle := float32(i) * 2 * math.Pi / 20
		vec := []float32{float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))}
		g.Insert(i, vec)
	}

	results := g.Search([]float32{1, 0}, 5)
	assert.Len(t, results, 5)
	assert.Equal(t, 0, results[0].ID, "node 0 (angle=0) should be closest to [1,0]")
}
