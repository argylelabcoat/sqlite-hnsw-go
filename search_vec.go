package sqlitehnsw

func (s *Store) Search(query []float32, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	neighbors := s.graph.Search(query, topK)

	results := make([]SearchResult, len(neighbors))
	for i, n := range neighbors {
		results[i] = SearchResult{
			RowID: n.ID,
			Score: float64(1 - n.Dist),
		}
	}
	return results, nil
}
