package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlushGraph_PersistsToSQLite(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{Vector: []float32{1, 0, 0, 0}},
		{Vector: []float32{0, 1, 0, 0}},
	}))

	require.NoError(t, s.FlushGraph())

	assert.Equal(t, 2, s.Count())
}

func TestRebuildGraph_SearchStillWorks(t *testing.T) {
	s := testStore(t, 2)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0}},
		{ID: 2, Vector: []float32{0, 1}},
	}))

	require.NoError(t, s.RebuildGraph())

	results, err := s.Search([]float32{1, 0}, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0].RowID)
}

func TestStats_ReturnsCorrectCount(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{Vector: []float32{1, 0, 0, 0}},
	}))

	stats, err := s.Stats()
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalDocuments)
	assert.Equal(t, 1, stats.HNSWGraphEntries)
}

func TestOptimize_CompletesWithoutError(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{Vector: []float32{1, 0, 0, 0}, Content: "doc 1"},
		{Vector: []float32{0, 1, 0, 0}, Content: "doc 2"},
	}))

	err := s.Optimize()
	assert.NoError(t, err)
}
