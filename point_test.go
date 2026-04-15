package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoint_VectorDimension(t *testing.T) {
	p := Point{Vector: []float32{1, 2, 3}}
	assert.Len(t, p.Vector, 3)
}
