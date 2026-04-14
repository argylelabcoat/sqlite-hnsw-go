package hnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosineDistance_SameVector_ReturnsZero(t *testing.T) {
	v := []float32{1, 0, 0}
	assert.InDelta(t, float32(0), CosineDistance(v, v), 1e-6)
}

func TestCosineDistance_Orthogonal_ReturnsOne(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	assert.InDelta(t, float32(1), CosineDistance(a, b), 1e-6)
}

func TestCosineDistance_Opposite_ReturnsTwo(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{-1, 0, 0}
	assert.InDelta(t, float32(2), CosineDistance(a, b), 1e-6)
}

func TestEuclideanDistance_SameVector_ReturnsZero(t *testing.T) {
	v := []float32{3, 4}
	assert.InDelta(t, float32(0), EuclideanDistance(v, v), 1e-6)
}

func TestEuclideanDistance_KnownPythagorean(t *testing.T) {
	a := []float32{0, 0}
	b := []float32{3, 4}
	assert.InDelta(t, float32(5), EuclideanDistance(a, b), 1e-6)
}

func TestDotDistance_SameUnitVector_ReturnsNegOne(t *testing.T) {
	v := []float32{1, 0, 0}
	assert.InDelta(t, float32(-1), DotDistance(v, v), 1e-6)
}

func TestDotDistance_Orthogonal_ReturnsZero(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	assert.InDelta(t, float32(0), DotDistance(a, b), 1e-6)
}
