package sqlitehnsw

import (
	"context"
	"fmt"

	"github.com/macintosh-codex/sqlite-hnsw/internal/hnsw"
)

type StoreStats struct {
	TotalDocuments   int
	HNSWGraphVersion int
	HNSWGraphEntries int
	SQLiteFileSize   int64
	FTS5DocCount     int
}

func (s *Store) FlushGraph() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}
	return s.flushGraphLocked()
}

func (s *Store) RebuildGraph() error {
	return s.RebuildGraphWithContext(context.Background())
}

func (s *Store) RebuildGraphWithContext(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	distFunc := newDistanceFunc(s.cfg.Metric)
	s.graph = hnsw.NewGraph(s.cfg.M, s.cfg.EfConstruction, s.cfg.EfSearch, distFunc)

	return s.rebuildGraphFromTable()
}

func (s *Store) Stats() (StoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return StoreStats{}, ErrStoreClosed
	}

	stats := StoreStats{
		HNSWGraphEntries: s.graph.Len(),
	}

	s.db.QueryRow("SELECT COUNT(*) FROM vectors").Scan(&stats.TotalDocuments)
	s.db.QueryRow("SELECT COALESCE(version, 0) FROM hnsw_graph WHERE collection = ?",
		s.cfg.Collection).Scan(&stats.HNSWGraphVersion)

	return stats, nil
}

func (s *Store) Optimize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	if _, err := s.db.Exec("INSERT INTO fts_content(fts_content) VALUES('optimize')"); err != nil {
		return fmt.Errorf("optimize fts5: %w", err)
	}

	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("optimize vacuum: %w", err)
	}

	return nil
}
