package sqlitehnsw

import (
	"fmt"
)

// BM25Search performs full-text search over the content table.
// Returns content IDs (from the content table), not vector rowids.
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
		SELECT rowid, bm25(fts_chapters) AS score
		FROM fts_chapters
		WHERE fts_chapters MATCH ?
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

// BM25SearchOpts performs filtered BM25 search over the content table.
func (s *Store) BM25SearchOpts(query string, opts SearchOptions) ([]BM25Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	if opts.TopK <= 0 {
		opts.TopK = 10
	}

	ftsQuery := query
	if opts.Phrase {
		ftsQuery = fmt.Sprintf(`"%s"`, query)
	} else if opts.Prefix {
		ftsQuery = query + "*"
	}

	sqlQuery := `
		SELECT f.rowid, bm25(fts_chapters) AS score
		FROM fts_chapters f
		JOIN content c ON f.rowid = c.id
		WHERE f.fts_chapters MATCH ?`
	args := []any{ftsQuery}

	if opts.BookID != "" {
		sqlQuery += " AND c.book_id = ?"
		args = append(args, opts.BookID)
	}

	sqlQuery += " ORDER BY score LIMIT ?"
	args = append(args, opts.TopK)

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("bm25 search opts: %w", err)
	}
	defer rows.Close()

	var results []BM25Result
	for rows.Next() {
		var r BM25Result
		if err := rows.Scan(&r.RowID, &r.Score); err != nil {
			return nil, fmt.Errorf("bm25 search opts: scan: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}
