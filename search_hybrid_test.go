package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRRFFuse_PureVector(t *testing.T) {
	vecResults := []SearchResult{
		{RowID: 1, Score: 0.95},
		{RowID: 2, Score: 0.87},
	}
	bm25Results := []BM25Result{}

	fused := rrfFuse(vecResults, bm25Results, 1.0, 10)
	assert.Len(t, fused, 2)
	assert.Equal(t, 1, fused[0].RowID, "highest vector score should rank first")
	assert.Equal(t, 1, fused[0].VectorRank)
	assert.Equal(t, 0, fused[0].BM25Rank)
}

func TestRRFFuse_PureBM25(t *testing.T) {
	vecResults := []SearchResult{}
	bm25Results := []BM25Result{
		{RowID: 3, Score: -1.2},
		{RowID: 4, Score: -0.8},
	}

	fused := rrfFuse(vecResults, bm25Results, 0.0, 10)
	assert.Len(t, fused, 2)
	assert.Equal(t, 3, fused[0].RowID, "best BM25 score should rank first")
	assert.Equal(t, 0, fused[0].VectorRank)
	assert.Equal(t, 1, fused[0].BM25Rank)
}

func TestRRFFuse_Blended(t *testing.T) {
	vecResults := []SearchResult{
		{RowID: 1, Score: 0.95},
		{RowID: 2, Score: 0.80},
	}
	bm25Results := []BM25Result{
		{RowID: 2, Score: -2.0},
		{RowID: 3, Score: -1.0},
	}

	fused := rrfFuse(vecResults, bm25Results, 0.5, 10)

	ids := make(map[int]bool)
	for _, f := range fused {
		ids[f.RowID] = true
	}
	assert.True(t, ids[1])
	assert.True(t, ids[2])
	assert.True(t, ids[3], "doc 3 should appear from BM25 results")

	for _, f := range fused {
		if f.RowID == 2 {
			assert.Greater(t, f.VectorRank, 0, "doc 2 should have a vector rank")
			assert.Greater(t, f.BM25Rank, 0, "doc 2 should have a BM25 rank")
		}
	}
}

func TestRRFFuse_TopK(t *testing.T) {
	vecResults := []SearchResult{
		{RowID: 1, Score: 0.9},
		{RowID: 2, Score: 0.8},
		{RowID: 3, Score: 0.7},
	}
	bm25Results := []BM25Result{
		{RowID: 4, Score: -1.0},
		{RowID: 5, Score: -0.5},
	}

	fused := rrfFuse(vecResults, bm25Results, 0.5, 3)
	assert.Len(t, fused, 3)
}

func TestHybridSearch_BlendsResults(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "Draw1Control creates controls"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "File Manager manages files"},
		{ID: 3, Vector: []float32{0.9, 0.1, 0, 0}, Content: "Draw1Control usage guide"},
	}))

	results, err := s.HybridSearch("Draw1Control", []float32{1, 0, 0, 0}, HybridOptions{
		TopK:  3,
		Alpha: 0.5,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	ids := make(map[int]bool)
	for _, r := range results {
		ids[r.RowID] = true
	}
	assert.True(t, ids[1], "doc 1 should appear (vector + BM25 match)")
}
