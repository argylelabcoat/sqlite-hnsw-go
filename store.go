package sqlitehnsw

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/macintosh-codex/sqlite-hnsw/internal/hnsw"
	_ "modernc.org/sqlite"
)

type Store struct {
	db             *sql.DB
	graph          *hnsw.Graph
	cfg            Config
	pendingFlushes int
	closed         bool
	mu             sync.RWMutex
}

func NewStore(cfg Config) (*Store, error) {
	if cfg.Dimension <= 0 {
		return nil, fmt.Errorf("new store: dimension must be > 0, got %d", cfg.Dimension)
	}
	cfg.applyDefaults()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("new store: open sqlite: %w", err)
	}

	db.Exec(fmt.Sprintf("PRAGMA cache_size = %d", cfg.CacheSize))

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("new store: init schema: %w", err)
	}

	distFunc := newDistanceFunc(cfg.Metric)
	graph := hnsw.NewGraph(cfg.M, cfg.EfConstruction, cfg.EfSearch, distFunc)

	s := &Store{
		db:    db,
		graph: graph,
		cfg:   cfg,
	}

	if err := s.loadGraph(); err != nil {
		db.Close()
		return nil, fmt.Errorf("new store: load graph: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	if err := s.flushGraphLocked(); err != nil {
		return fmt.Errorf("close: flush graph: %w", err)
	}
	return s.db.Close()
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM vectors").Scan(&count)
	return count
}

func newDistanceFunc(m Metric) hnsw.DistanceFunc {
	switch m {
	case Euclidean:
		return hnsw.EuclideanDistance
	case Dot:
		return hnsw.DotDistance
	default:
		return hnsw.CosineDistance
	}
}

func (s *Store) loadGraph() error {
	var blob []byte
	err := s.db.QueryRow(
		"SELECT blob FROM hnsw_graph WHERE collection = ?", s.cfg.Collection,
	).Scan(&blob)

	if err == nil && len(blob) > 0 {
		if err := s.graph.Deserialize(blob); err != nil {
			return fmt.Errorf("%w: %v", ErrGraphCorrupt, err)
		}
		return nil
	}

	if err != sql.ErrNoRows {
		return nil
	}

	return s.rebuildGraphFromTable()
}

func (s *Store) rebuildGraphFromTable() error {
	rows, err := s.db.Query("SELECT rowid, vec FROM vectors ORDER BY rowid")
	if err != nil {
		return fmt.Errorf("rebuild graph: query vectors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			return fmt.Errorf("rebuild graph: scan: %w", err)
		}
		vec := decodeVector(data)
		if vec == nil {
			continue
		}
		s.graph.Insert(id, vec)
	}

	return s.flushGraphLocked()
}

func (s *Store) flushGraphLocked() error {
	data, err := s.graph.Serialize()
	if err != nil {
		return fmt.Errorf("flush graph: serialize: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO hnsw_graph
			(collection, dim, metric, m, ef_construction, blob, version)
		VALUES (?, ?, ?, ?, ?, ?, 1)`,
		s.cfg.Collection, s.cfg.Dimension, string(s.cfg.Metric),
		s.cfg.M, s.cfg.EfConstruction, data,
	)
	if err != nil {
		return fmt.Errorf("flush graph: write: %w", err)
	}

	s.pendingFlushes = 0
	return nil
}
