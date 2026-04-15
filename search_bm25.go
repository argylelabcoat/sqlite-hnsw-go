package sqlitehnsw

import (
	"fmt"
)

func (s *Store) BM25Search(query string, topK int) ([]BM25Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	if topK <= 0 {
		topK = 10
	}

	rows, err := s.db.Query(`
		SELECT rowid, bm25(fts_content) AS score
		FROM fts_content
		WHERE fts_content MATCH ?
		ORDER BY score
		LIMIT ?`, query, topK)
	if err != nil {
		return nil, fmt.Errorf("bm25 search: %w", err)
	}
	defer rows.Close()

	var results []BM25Result
	for rows.Next() {
		var r BM25Result
		if err := rows.Scan(&r.RowID, &r.Score); err != nil {
			return nil, fmt.Errorf("bm25 search: scan: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}
