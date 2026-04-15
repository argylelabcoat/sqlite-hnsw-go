package hnsw

// Delete removes the node with the given id from the graph.
// If the node does not exist, Delete is a no-op.
func (g *Graph) Delete(id int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deleteLocked(id)
}

// deleteLocked removes a node without acquiring the mutex.
// Caller must hold g.mu.Lock().
func (g *Graph) deleteLocked(id int) {
	node, ok := g.nodes[id]
	if !ok {
		return
	}

	for l := 0; l <= node.Level; l++ {
		if l >= len(node.Edges) {
			continue
		}
		for _, neighborID := range node.Edges[l] {
			neighbor, ok := g.nodes[neighborID]
			if !ok {
				continue
			}
			if l < len(neighbor.Edges) {
				neighbor.Edges[l] = removeInt(neighbor.Edges[l], id)
			}
		}
	}

	delete(g.nodes, id)

	if g.entryPoint == id {
		newMaxLevel := -1
		for _, n := range g.nodes {
			if n.Level > newMaxLevel {
				newMaxLevel = n.Level
				g.entryPoint = n.ID
			}
		}
		if newMaxLevel >= 0 {
			g.maxLevel = newMaxLevel
		} else {
			g.entryPoint = 0
			g.maxLevel = 0
		}
	}
}

func removeInt(slice []int, val int) []int {
	for i, v := range slice {
		if v == val {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
