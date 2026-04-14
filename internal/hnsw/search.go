package hnsw

// Search returns the k nearest neighbors to query, sorted by distance ascending.
func (g *Graph) Search(query []float32, k int) []Neighbor {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.nodes) == 0 || k <= 0 {
		return nil
	}

	cur := g.entryPoint
	curDist := g.dist(query, g.nodes[cur].Vector)

	for l := g.maxLevel; l > 0; l-- {
		changed := true
		for changed {
			changed = false
			node := g.nodes[cur]
			if l >= len(node.Edges) || len(node.Edges[l]) == 0 {
				continue
			}
			for _, neighborID := range node.Edges[l] {
				neighbor, ok := g.nodes[neighborID]
				if !ok {
					continue
				}
				d := g.dist(query, neighbor.Vector)
				if d < curDist {
					cur = neighborID
					curDist = d
					changed = true
				}
			}
		}
	}

	results := g.searchLayer(query, []int{cur}, g.efSearch, 0)

	if k > len(results) {
		k = len(results)
	}

	out := make([]Neighbor, k)
	for i := 0; i < k; i++ {
		out[i] = Neighbor{ID: results[i].id, Dist: results[i].dist}
	}
	return out
}
