package sqlitehnsw

import (
	"fmt"
	"strings"
)

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

// SearchFiltered performs vector search and post-filters results by book_id.
// When bookID is empty it behaves identically to Search.
func (s *Store) SearchFiltered(query []float32, topK int, bookID string) ([]SearchResult, error) {
	if bookID == "" {
		return s.Search(query, topK)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	// Over-fetch from HNSW to account for filtering loss.
	fetchK := topK * 10
	if fetchK < 1000 {
		fetchK = 1000
	}
	neighbors := s.graph.Search(query, fetchK)
	if len(neighbors) == 0 {
		return nil, nil
	}

	// Batch-validate book_id against the denormalized column.
	placeholders := strings.Repeat("?,", len(neighbors))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(neighbors)+1)
	for i, n := range neighbors {
		args[i] = n.ID
	}
	args[len(neighbors)] = bookID

	rows, err := s.db.Query(
		fmt.Sprintf("SELECT rowid FROM vectors WHERE rowid IN (%s) AND book_id = ?", placeholders),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("search filtered: %w", err)
	}
	defer rows.Close()

	valid := make(map[int]struct{}, topK)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		valid[id] = struct{}{}
	}

	var results []SearchResult
	for _, n := range neighbors {
		if _, ok := valid[n.ID]; ok {
			results = append(results, SearchResult{RowID: n.ID, Score: float64(1 - n.Dist)})
			if len(results) >= topK {
				break
			}
		}
	}
	return results, nil
}
