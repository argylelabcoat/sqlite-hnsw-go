# SQLite-HNSW

A disk-backed, memory-efficient vector database in pure Go, combining HNSW approximate nearest neighbor search with SQLite persistence and FTS5 full-text search.

## Features

- **HNSW vector search** — custom from-scratch implementation, pure Go
- **SQLite persistence** — graph serialized to CBOR, payloads on disk
- **FTS5 BM25 search** — full-text search with phrase, prefix, and filtering
- **Hybrid search** — Reciprocal Rank Fusion (RRF) of vector + BM25 results
- **Lazy payload loading** — only loads document content on demand
- **Thread-safe** — `sync.RWMutex` for concurrent reads/writes

## Architecture

```
RAM:                              Disk (SQLite):
┌─────────────────┐               ┌──────────────────────────┐
│  HNSW Graph     │──rowid─────→│  vectors table           │
│  nodes + edges  │               │  rowid | vec(BLOB)       │
│  (stays in RAM) │               │  meta(JSON) | content    │
└─────────────────┘               └──────────────────────────┘
                                  ┌──────────────────────────┐
                                  │  fts5 virtual table      │
                                  │  (BM25 full-text search) │
                                  └──────────────────────────┘
                                  ┌──────────────────────────┐
                                  │  hnsw_graph table        │
                                  │  (CBOR-serialized graph) │
                                  └──────────────────────────┘
```

**Why HNSW in RAM?** Each search hop requires random access to node edges. Disk seeks would kill latency. At ~2KB per node (M=16, 384-dim), 1M vectors = ~2GB RAM — acceptable for a dedicated search node.

**Why payloads on disk?** The 79% RAM reduction vs govector comes from keeping document content (text, metadata) off the heap. Payloads load on demand via SQLite after search returns.

## Installation

```bash
go get github.com/macintosh-codex/sqlite-hnsw
```

## Quick Start

```go
package main

import (
    "fmt"
    sqlitehnsw "github.com/macintosh-codex/sqlite-hnsw"
)

func main() {
    store, err := sqlitehnsw.NewStore(sqlitehnsw.Config{
        DBPath:    "vectors.db",
        Dimension: 384,
        M:         16,
        EfSearch:  64,
    })
    if err != nil {
        panic(err)
    }
    defer store.Close()

    // Upsert vectors
    err = store.Upsert([]sqlitehnsw.Point{{
        Vector:  []float32{0.1, 0.3, ...}, // 384-dim
        Content: "Draw1Control creates a control",
        EntityName: "Draw1Control",
    }})
    if err != nil {
        panic(err)
    }

    // Vector search
    results, err := store.Search(queryVector, 10)
    for _, r := range results {
        fmt.Printf("rowid=%d score=%.3f\n", r.RowID, r.Score)
    }

    // BM25 text search
    bm25, err := store.BM25Search("Draw1Control", 10)

    // Hybrid search (RRF fusion)
    hybrid, err := store.HybridSearch("Draw1Control", queryVector,
        sqlitehnsw.HybridOptions{TopK: 10, Alpha: 0.5})
}
```

## API

### Core

| Method | Description |
|--------|-------------|
| `NewStore(cfg)` | Open/create persistent store |
| `Store.Close()` | Flush graph and close SQLite |
| `Store.Upsert(points)` | Insert or update vectors |
| `Store.Search(query, topK)` | HNSW vector similarity search |
| `Store.Delete(rowid)` | Remove from SQLite and HNSW |
| `Store.Count()` | Document count |

### Text Search

| Method | Description |
|--------|-------------|
| `Store.BM25Search(query, topK)` | FTS5 BM25 full-text search |
| `Store.BM25SearchOpts(query, opts)` | BM25 with phrase/prefix/book filter |
| `Store.HybridSearch(query, vec, opts)` | RRF fusion of vector + BM25 |

### Metadata

| Method | Description |
|--------|-------------|
| `Store.GetPayload(rowid)` | Lazy load document from SQLite |
| `Store.GetPayloads(rowids)` | Batch lazy load |
| `Store.LookupEntity(name)` | Exact entity name lookup |
| `Store.LookupManager(manager)` | Manager name lookup |
| `Store.LookupByFilter(where, args...)` | Arbitrary SQL filter |

### Admin

| Method | Description |
|--------|-------------|
| `Store.FlushGraph()` | Persist HNSW graph to CBOR blob |
| `Store.RebuildGraph()` | Rebuild graph from vectors table |
| `Store.Optimize()` | VACUUM + FTS5 optimize |
| `Store.Stats()` | Store statistics |

### Configuration

```go
type Config struct {
    DBPath         string   // SQLite file path
    Collection     string   // Collection name (default: "default")
    Dimension      int      // Vector dimension (required)
    Metric         Metric   // "cosine" | "euclidean" | "dot" (default: cosine)
    M              int      // Max connections per layer (default: 16)
    EfConstruction int      // Construction quality (default: 200)
    EfSearch       int      // Search quality (default: 64)
    FlushThreshold int      // Serialize after N inserts (default: 1000)
    CacheSize      int      // SQLite page cache KB (default: 10240)
}
```

## Benchmarks (M4, 1000 vectors, dim=384)

```
BenchmarkInsert      1,908 µs/op   226 KB/op   4,180 allocs/op
BenchmarkSearch      770  µs/op   176 KB/op   3,923 allocs/op
BenchmarkSerialize   8.5  ms/op   3.1 MB/op      34 allocs/op
```

## Why From-Scratch HNSW?

Existing pure-Go HNSW libraries (e.g., `coder/hnsw`) don't expose graph internals, making serialization impossible without fighting their internal APIs. sqlite-hnsw implements HNSW from scratch to have full control over the graph structure for CBOR serialization.

## Dependencies

- `modernc.org/sqlite` — pure Go SQLite (FTS5 support)
- `fxamacker/cbor/v2` — CBOR serialization for graph persistence
- `github.com/stretchr/testify` — test assertions

No external HNSW or vector libraries. Everything is from scratch.

## Testing

```bash
go test ./...        # Run all tests (85+)
go vet ./...         # Vet
go test -bench=. -benchmem -count=1 -run=^$ ./...  # Benchmarks
```
