package hnsw

import (
	"math"
)

// randomLevel returns a non-negative integer sampled from an exponential distribution.
//
// The level is computed as: level = -log(rand) / log(m)
// where m is theGraph's fanout parameter and rand is a uniform random value in (0, 1].
// This produces the standard skip-list style exponential distribution where
// approximately 1/m of elements reach level L or higher.
func (g *Graph) randomLevel() int {
	if g.m <= 1 {
		return 0
	}
	mL := 1.0 / math.Log(float64(g.m))
	return int(-math.Log(g.rng.Float64()) * mL)
}
