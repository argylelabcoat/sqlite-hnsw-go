package hnsw

import (
	"container/heap"
	"math"
	"sort"
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

func (g *Graph) searchLayer(query []float32, entryPoints []int, ef int, level int) []candidate {
	visited := make(map[int]bool)
	candidates := &candidateMinHeap{}
	results := &candidateMaxHeap{}
	heap.Init(candidates)
	heap.Init(results)

	for _, ep := range entryPoints {
		node, ok := g.nodes[ep]
		if !ok {
			continue
		}
		dist := g.dist(query, node.Vector)
		heap.Push(candidates, candidate{id: ep, dist: dist})
		heap.Push(results, candidate{id: ep, dist: dist})
		visited[ep] = true
	}

	for candidates.Len() > 0 {
		c := heap.Pop(candidates).(candidate)
		if results.Len() > 0 && c.dist > (*results)[0].dist {
			break
		}

		node, ok := g.nodes[c.id]
		if !ok || level >= len(node.Edges) {
			continue
		}

		for _, neighborID := range node.Edges[level] {
			if visited[neighborID] {
				continue
			}
			neighbor, ok := g.nodes[neighborID]
			if !ok {
				continue
			}
			visited[neighborID] = true
			dist := g.dist(query, neighbor.Vector)

			if results.Len() < ef || dist < (*results)[0].dist {
				heap.Push(candidates, candidate{id: neighborID, dist: dist})
				heap.Push(results, candidate{id: neighborID, dist: dist})
				if results.Len() > ef {
					heap.Pop(results)
				}
			}
		}
	}

	out := make([]candidate, len(*results))
	copy(out, *results)
	sort.Slice(out, func(i, j int) bool { return out[i].dist < out[j].dist })
	return out
}
