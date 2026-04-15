package sqlitehnsw

// Point represents a document to be stored and indexed.
type Point struct {
	ID           int            // Leave 0 for auto-assign, or set explicitly
	Vector       []float32      // Embedding vector (required)
	Content      string         // Full text for FTS5 indexing
	Meta         map[string]any // Flexible metadata stored as JSON
	EntityName   string         // Denormalized filter column
	EntityKind   string         // Denormalized filter column
	Manager      string         // Denormalized filter column
	ReturnType   string         // Denormalized filter column
	TrapID       string         // Denormalized filter column
	HeaderFile   string         // Denormalized filter column
	Availability string         // Denormalized filter column
	BookID       string         // Denormalized filter column
	ChapterFile  string         // Denormalized filter column
	Title        string         // Denormalized filter column
}

// SearchResult is returned by Store.Search.
type SearchResult struct {
	RowID int
	Score float64 // 0-1 cosine similarity (higher = more similar)
}

// BM25Result is returned by Store.BM25Search.
type BM25Result struct {
	RowID int
	Score float64 // BM25 score (more negative = better per SQLite convention)
}

// HybridSearchResult is a fused result from vector + BM25 search.
type HybridSearchResult struct {
	RowID      int
	RRFScore   float64
	VectorRank int // 1-based, 0 if absent from vector results
	BM25Rank   int // 1-based, 0 if absent from BM25 results
}

// SearchOptions controls BM25 query behavior.
type SearchOptions struct {
	TopK   int
	BookID string
	Phrase bool
	Prefix bool
}

// HybridOptions controls hybrid search behavior.
type HybridOptions struct {
	TopK   int
	Alpha  float64 // 0=pure BM25, 1=pure vector
	BookID string
}
