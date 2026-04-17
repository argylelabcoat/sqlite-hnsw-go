package sqlitehnsw

import (
	"fmt"
	"sort"
)

const rrfK = 60.0

// HybridSearchResult is a fused result from vector + BM25 search at the
// document (content) level.
type HybridSearchResult struct {
	ContentID   int // ID from content table
	BestChunkID int // Best matching chunk rowid (0 if only BM25)
	RRFScore    float64
	VectorRank  int // 1-based, 0 if absent from vector results
	BM25Rank    int // 1-based, 0 if absent from BM25 results
}

func (s *Store) HybridSearch(query string, queryVec []float32, opts HybridOptions) ([]HybridSearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	if opts.TopK <= 0 {
		opts.TopK = 10
	}
	if opts.Alpha < 0 {
		opts.Alpha = 0.5
	}

	fetchK := opts.TopK * 2

	// 1. Vector search (chunk-level).
	neighbors := s.graph.Search(queryVec, fetchK)
	vecResults := make([]SearchResult, len(neighbors))
	for i, n := range neighbors {
		vecResults[i] = SearchResult{RowID: n.ID, Score: float64(1 - n.Dist)}
	}

	// 2. Map vector rowids → content IDs.
	vecContentRanks := make(map[int]int) // contentID → best rank
	vecContentChunk := make(map[int]int) // contentID → best chunk rowid
	if len(vecResults) > 0 {
		rowids := make([]int, len(vecResults))
		for i, r := range vecResults {
			rowids[i] = r.RowID
		}
		// Batch query content_id for all vector results.
		placeholders := make([]string, len(rowids))
		args := make([]any, len(rowids))
		for i, id := range rowids {
			placeholders[i] = "?"
			args[i] = id
		}
		query := fmt.Sprintf(
			"SELECT rowid, content_id FROM vectors WHERE rowid IN (%s)",
			joinStrings(placeholders, ","))
		rows, err := s.db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("hybrid search map rowids: %w", err)
		}
		for rows.Next() {
			var rowid, contentID int
			rows.Scan(&rowid, &contentID)
			// Find the rank for this rowid.
			for rank, r := range vecResults {
				if r.RowID == rowid {
					if existing, ok := vecContentRanks[contentID]; !ok || rank+1 < existing {
						vecContentRanks[contentID] = rank + 1
						vecContentChunk[contentID] = rowid
					}
					break
				}
			}
		}
		rows.Close()
	}

	// 3. BM25 search (document-level via fts_chapters).
	bm25Rows, err := s.db.Query(`
		SELECT rowid, bm25(fts_chapters) AS score
		FROM fts_chapters
		WHERE fts_chapters MATCH ?
		ORDER BY score
		LIMIT ?`, query, fetchK)
	if err != nil {
		return nil, fmt.Errorf("hybrid search bm25: %w", err)
	}
	defer bm25Rows.Close()

	bm25Ranks := make(map[int]int)
	for bm25Rows.Next() {
		var contentID int
		var score float64
		bm25Rows.Scan(&contentID, &score)
		bm25Ranks[contentID] = len(bm25Ranks) + 1
	}

	// 4. RRF fuse at content ID level.
	allIDs := make(map[int]bool)
	for id := range vecContentRanks {
		allIDs[id] = true
	}
	for id := range bm25Ranks {
		allIDs[id] = true
	}

	fused := make([]HybridSearchResult, 0, len(allIDs))
	for id := range allIDs {
		score := 0.0
		vr := 0
		br := 0

		if rank, ok := vecContentRanks[id]; ok {
			vr = rank
			score += opts.Alpha / (rrfK + float64(rank))
		}
		if rank, ok := bm25Ranks[id]; ok {
			br = rank
			score += (1 - opts.Alpha) / (rrfK + float64(rank))
		}

		fused = append(fused, HybridSearchResult{
			ContentID:   id,
			BestChunkID: vecContentChunk[id],
			RRFScore:    score,
			VectorRank:  vr,
			BM25Rank:    br,
		})
	}

	sort.Slice(fused, func(i, j int) bool {
		return fused[i].RRFScore > fused[j].RRFScore
	})

	if opts.TopK < len(fused) {
		fused = fused[:opts.TopK]
	}
	return fused, nil
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
