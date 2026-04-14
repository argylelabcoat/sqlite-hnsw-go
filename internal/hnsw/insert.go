package hnsw

func (g *Graph) deleteLocked(id int) {
	delete(g.nodes, id)
}

func (g *Graph) Insert(id int, vec []float32) {
	g.mu.Lock()
	defer g.mu.Unlock()

	level := g.randomLevel()
	node := &Node{
		ID:     id,
		Vector: vec,
		Level:  level,
		Edges:  make([][]int, level+1),
	}

	if _, exists := g.nodes[id]; exists {
		g.deleteLocked(id)
	}

	if len(g.nodes) == 0 {
		g.nodes[id] = node
		g.entryPoint = id
		g.maxLevel = level
		return
	}

	cur := g.entryPoint
	curDist := g.dist(vec, g.nodes[cur].Vector)

	for l := g.maxLevel; l > level; l-- {
		changed := true
		for changed {
			changed = false
			n := g.nodes[cur]
			if l >= len(n.Edges) || len(n.Edges[l]) == 0 {
				break
			}
			for _, neighborID := range n.Edges[l] {
				neighbor, ok := g.nodes[neighborID]
				if !ok {
					continue
				}
				d := g.dist(vec, neighbor.Vector)
				if d < curDist {
					cur = neighborID
					curDist = d
					changed = true
				}
			}
		}
	}

	epList := []int{cur}
	for l := min(level, g.maxLevel); l >= 0; l-- {
		results := g.searchLayer(vec, epList, g.efConstruct, l)
		neighbors := g.selectNeighbors(vec, results, g.m)

		node.Edges[l] = make([]int, len(neighbors))
		for i, n := range neighbors {
			node.Edges[l][i] = n.id
		}

		for _, n := range neighbors {
			neighbor := g.nodes[n.id]
			if neighbor == nil {
				continue
			}
			for len(neighbor.Edges) <= l {
				neighbor.Edges = append(neighbor.Edges, nil)
			}
			neighbor.Edges[l] = append(neighbor.Edges[l], id)

			maxConn := g.m
			if l == 0 {
				maxConn = g.m * 2
			}
			if len(neighbor.Edges[l]) > maxConn {
				cands := make([]candidate, 0, len(neighbor.Edges[l]))
				for _, nid := range neighbor.Edges[l] {
					if g.nodes[nid] == nil {
						continue
					}
					cands = append(cands, candidate{
						id:   nid,
						dist: g.dist(neighbor.Vector, g.nodes[nid].Vector),
					})
				}
				if len(cands) == 0 {
					continue
				}
				pruned := g.selectNeighbors(neighbor.Vector, cands, maxConn)
				neighbor.Edges[l] = make([]int, len(pruned))
				for i, p := range pruned {
					neighbor.Edges[l][i] = p.id
				}
			}
		}

		epList = make([]int, len(results))
		for i, r := range results {
			epList[i] = r.id
		}
	}

	g.nodes[id] = node

	if level > g.maxLevel {
		g.maxLevel = level
		g.entryPoint = id
	}
}
