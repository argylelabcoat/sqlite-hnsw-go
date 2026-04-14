package hnsw

import "math"

// DistanceFunc defines a function that computes the distance between two vectors.
type DistanceFunc func(a, b []float32) float32

// CosineDistance computes the cosine distance between two vectors.
func CosineDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		ai, bi := float64(a[i]), float64(b[i])
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	if normA == 0 || normB == 0 {
		return 1
	}
	return float32(1 - dot/(math.Sqrt(normA)*math.Sqrt(normB)))
}

// EuclideanDistance computes the Euclidean distance between two vectors.
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var sum float64
	for i := range a {
		d := float64(a[i]) - float64(b[i])
		sum += d * d
	}
	return float32(math.Sqrt(sum))
}

// DotDistance computes the negative dot product between two vectors.
func DotDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return -dot
}
