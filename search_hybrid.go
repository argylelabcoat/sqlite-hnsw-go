package sqlitehnsw

import (
	"fmt"
	"sort"
)

const rrfK = 60.0

func rrfFuse(vecResults []SearchResult, bm25Results []BM25Result, alpha float64, topK int) []HybridSearchResult {
	vecRank := make(map[int]int)
	for i, r := range vecResults {
		vecRank[r.RowID] = i + 1
	}

	bm25Rank := make(map[int]int)
	for i, r := range bm25Results {
		bm25Rank[r.RowID] = i + 1
	}

	allIDs := make(map[int]bool)
	for id := range vecRank {
		allIDs[id] = true
	}
	for id := range bm25Rank {
		allIDs[id] = true
	}

	fused := make([]HybridSearchResult, 0, len(allIDs))
	for id := range allIDs {
		score := 0.0
		vr := 0
		br := 0

		if rank, ok := vecRank[id]; ok {
			vr = rank
			score += alpha / (rrfK + float64(rank))
		}
		if rank, ok := bm25Rank[id]; ok {
			br = rank
			score += (1 - alpha) / (rrfK + float64(rank))
		}

		fused = append(fused, HybridSearchResult{
			RowID:      id,
			RRFScore:   score,
			VectorRank: vr,
			BM25Rank:   br,
		})
	}

	sort.Slice(fused, func(i, j int) bool {
		return fused[i].RRFScore > fused[j].RRFScore
	})

	if topK < len(fused) {
		fused = fused[:topK]
	}
	return fused
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

	neighbors := s.graph.Search(queryVec, fetchK)
	vecResults := make([]SearchResult, len(neighbors))
	for i, n := range neighbors {
		vecResults[i] = SearchResult{RowID: n.ID, Score: float64(1 - n.Dist)}
	}

	bm25Query := "SELECT rowid, bm25(fts_content) AS score FROM fts_content WHERE fts_content MATCH ? ORDER BY score LIMIT ?"
	rows, err := s.db.Query(bm25Query, query, fetchK)
	if err != nil {
		return nil, fmt.Errorf("hybrid search bm25: %w", err)
	}
	defer rows.Close()

	var bm25Results []BM25Result
	for rows.Next() {
		var r BM25Result
		if err := rows.Scan(&r.RowID, &r.Score); err != nil {
			return nil, fmt.Errorf("hybrid search bm25 scan: %w", err)
		}
		bm25Results = append(bm25Results, r)
	}

	return rrfFuse(vecResults, bm25Results, opts.Alpha, opts.TopK), nil
}
