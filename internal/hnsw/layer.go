package hnsw

import (
	"math"
)

func (g *Graph) randomLevel() int {
	mL := 1.0 / math.Log(float64(g.m))
	return int(-math.Log(g.rng.Float64()) * mL)
}
