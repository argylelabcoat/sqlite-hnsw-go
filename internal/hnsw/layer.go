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

// searchLayer performs a best-first search on a single graph layer.
//
// It starts from entryPoints and explores neighbors at the given level,
// maintaining up to ef candidates in the result set ordered by distance.
// The returned slice contains up to ef nearest neighbors sorted by distance.
func (g *Graph) searchLayer(query []float32, entryPoints []int, ef int, level int) []candidate {
	if ef <= 0 {
		return nil
	}

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

// selectNeighbors selects up to m neighbors from candidates using distance-based
// ordering and a diversity heuristic.
//
// Candidates are sorted by distance to query, then evaluated in order. A candidate
// is selected only if it is sufficiently far from all already-selected neighbors
// (distToSelected >= candidate's own distance to query). This enforces that selected
// neighbors are approximately equidistant from the query point, improving recall
// by avoiding redundant coverage of the same region.
//
// If fewer than m candidates satisfy the diversity heuristic, the remaining slots
// are filled from the discarded candidates (closest rejects) to ensure we return
// exactly m neighbors when possible.
func (g *Graph) selectNeighbors(query []float32, candidates []candidate, m int) []candidate {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})

	var selected []candidate
	var discarded []candidate

	for _, c := range candidates {
		if len(selected) >= m {
			break
		}
		node, ok := g.nodes[c.id]
		if !ok {
			continue
		}
		good := true
		for _, s := range selected {
			distToSelected := g.dist(node.Vector, g.nodes[s.id].Vector)
			if distToSelected < c.dist {
				good = false
				break
			}
		}
		if good {
			selected = append(selected, c)
		} else {
			discarded = append(discarded, c)
		}
	}

	// Fallback: if we have fewer than m selected, fill from discarded candidates.
	// These are the closest rejects and are the best available alternatives.
	for _, d := range discarded {
		if len(selected) >= m {
			break
		}
		selected = append(selected, d)
	}

	return selected
}
