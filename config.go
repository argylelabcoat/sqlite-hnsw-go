package sqlitehnsw

import "errors"

// Metric names the distance metric used for vector comparison.
type Metric string

const (
	// Cosine uses 1 minus cosine similarity. Range [0, 2].
	Cosine Metric = "cosine"
	// Euclidean uses L2 distance.
	Euclidean Metric = "euclidean"
	// Dot uses negated dot product. Lower is more similar.
	Dot Metric = "dot"
)

// Config controls Store behavior.
type Config struct {
	DBPath         string // Path to SQLite file
	Collection     string // Collection name (default: "default")
	Dimension      int    // Vector dimension (e.g., 384)
	Metric         Metric // Distance metric (default: cosine)
	M              int    // HNSW max connections per layer (default: 16)
	EfConstruction int    // HNSW construction quality (default: 200)
	EfSearch       int    // HNSW search quality (default: 64)
	FlushThreshold int    // Serialize graph after N inserts (default: 1000)
	CacheSize      int    // SQLite page cache in KB (default: 10240)
}

// applyDefaults sets default values for Config fields that are zero-valued.
func (c *Config) applyDefaults() {
	if c.Collection == "" {
		c.Collection = "default"
	}
	if c.Metric == "" {
		c.Metric = Cosine
	}
	if c.M == 0 {
		c.M = 16
	}
	if c.EfConstruction == 0 {
		c.EfConstruction = 200
	}
	if c.EfSearch == 0 {
		c.EfSearch = 64
	}
	if c.FlushThreshold == 0 {
		c.FlushThreshold = 1000
	}
	if c.CacheSize == 0 {
		c.CacheSize = 10240
	}
}

var (
	// ErrNotFound is returned when a document does not exist.
	ErrNotFound = errors.New("document not found")
	// ErrDimensionMismatch is returned when a vector has the wrong dimension.
	ErrDimensionMismatch = errors.New("vector dimension mismatch")
	// ErrStoreClosed is returned when operating on a closed store.
	ErrStoreClosed = errors.New("store is closed")
	// ErrGraphCorrupt is returned when the serialized graph cannot be deserialized.
	ErrGraphCorrupt = errors.New("HNSW graph is corrupt")
)
