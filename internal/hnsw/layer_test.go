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
