package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullPipeline(t *testing.T) {
	s := testStore(t, 2)
	defer s.Close()

	points := []Point{
		{ID: 1, Vector: []float32{1, 0}, Content: "Draw1Control creates a control", EntityName: "Draw1Control", EntityKind: "function", Manager: "Control Manager"},
		{ID: 2, Vector: []float32{0, 1}, Content: "File Manager opens files", EntityName: "FileManager", EntityKind: "function", Manager: "File Manager"},
		{ID: 3, Vector: []float32{-1, 0}, Content: "Draw1Control parameters", EntityName: "Draw1Control", EntityKind: "function", Manager: "Control Manager"},
		{ID: 4, Vector: []float32{0, -1}, Content: "Window Manager handles windows", EntityName: "WindowManager", EntityKind: "type", Manager: "Window Manager"},
	}
	require.NoError(t, s.Upsert(points))
	assert.Equal(t, 4, s.Count())

	t.Run("vector search", func(t *testing.T) {
		results, err := s.Search([]float32{1, 0}, 2)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, 1, results[0].RowID)
	})

	t.Run("BM25 search", func(t *testing.T) {
		results, err := s.BM25Search("Draw1Control", 5)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)
	})

	t.Run("entity lookup", func(t *testing.T) {
		ids, err := s.LookupEntity("Draw1Control")
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{1, 3}, ids)
	})

	t.Run("manager lookup", func(t *testing.T) {
		ids, err := s.LookupManager("File Manager")
		require.NoError(t, err)
		assert.Equal(t, []int{2}, ids)
	})

	t.Run("filter lookup", func(t *testing.T) {
		ids, err := s.LookupByFilter("entity_kind = ? AND manager = ?", "function", "Control Manager")
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{1, 3}, ids)
	})

	t.Run("hybrid search", func(t *testing.T) {
		results, err := s.HybridSearch("Draw1Control", []float32{1, 0}, HybridOptions{TopK: 3, Alpha: 0.5})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)
	})

	t.Run("get payload", func(t *testing.T) {
		payload, err := s.GetPayload(1)
		require.NoError(t, err)
		assert.Equal(t, "Draw1Control creates a control", payload["content"])
		assert.Equal(t, "Draw1Control", payload["entity_name"])
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, s.Delete(1))
		assert.Equal(t, 3, s.Count())

		results, err := s.Search([]float32{1, 0}, 3)
		require.NoError(t, err)
		for _, r := range results {
			assert.NotEqual(t, 1, r.RowID)
		}
	})

	t.Run("rebuild graph", func(t *testing.T) {
		require.NoError(t, s.RebuildGraph())
		results, err := s.Search([]float32{1, 0}, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})
}

func TestIntegration_ReopenPreservesData(t *testing.T) {
	path := testDBPath(t)

	s, err := NewStore(Config{DBPath: path, Dimension: 2})
	require.NoError(t, err)
	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0}, Content: "hello"},
	}))
	require.NoError(t, s.Close())

	s2, err := NewStore(Config{DBPath: path, Dimension: 2})
	require.NoError(t, err)
	defer s2.Close()

	assert.Equal(t, 1, s2.Count())

	results, err := s2.Search([]float32{1, 0}, 5)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].RowID)
}

func TestIntegration_OptimizeAndFlush(t *testing.T) {
	s := testStore(t, 2)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0}, Content: "doc one"},
		{ID: 2, Vector: []float32{0, 1}, Content: "doc two"},
	}))

	require.NoError(t, s.Optimize())
	require.NoError(t, s.FlushGraph())

	results, err := s.Search([]float32{1, 0}, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
