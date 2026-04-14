package hnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGraph_ReturnsNonNil(t *testing.T) {
	g := NewGraph(16, 200, 64, CosineDistance)
	assert.NotNil(t, g)
}

func TestNewGraph_LenIsZero(t *testing.T) {
	g := NewGraph(16, 200, 64, CosineDistance)
	assert.Equal(t, 0, g.Len())
}
