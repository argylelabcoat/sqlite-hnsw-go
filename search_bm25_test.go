package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBM25Search_BasicQuery(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "Draw1Control creates a control"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "File Manager handles file operations"},
		{ID: 3, Vector: []float32{0, 0, 1, 0}, Content: "Draw1Control parameters and usage"},
	}))

	results, err := s.BM25Search("Draw1Control", 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "should find at least 2 docs mentioning Draw1Control")

	ids := make(map[int]bool)
	for _, r := range results {
		ids[r.RowID] = true
	}
	assert.True(t, ids[1])
	assert.True(t, ids[3])
}

func TestBM25Search_NoResults(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "hello world"},
	}))

	results, err := s.BM25Search("xyznonexistent", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestBM25Search_TopK(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	for i := 0; i < 20; i++ {
		require.NoError(t, s.Upsert([]Point{
			{ID: i + 1, Vector: []float32{float32(i), 0, 0, 0}, Content: "Draw1Control document"},
		}))
	}

	results, err := s.BM25Search("Draw1Control", 5)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestBM25SearchOpts_PhraseMatch(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "File Manager handles files"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "the file is managed by the system"},
	}))

	results, err := s.BM25SearchOpts("file manager", SearchOptions{
		TopK:   10,
		Phrase: true,
	})
	require.NoError(t, err)

	ids := make(map[int]bool)
	for _, r := range results {
		ids[r.RowID] = true
	}
	assert.True(t, ids[1], "exact phrase 'file manager' should match doc 1")
}

func TestBM25SearchOpts_Prefix(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "Draw1Control"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "DrawWindow"},
		{ID: 3, Vector: []float32{0, 0, 1, 0}, Content: "CreateWindow"},
	}))

	results, err := s.BM25SearchOpts("Draw", SearchOptions{
		TopK:   10,
		Prefix: true,
	})
	require.NoError(t, err)

	ids := make(map[int]bool)
	for _, r := range results {
		ids[r.RowID] = true
	}
	assert.True(t, ids[1], "Draw* should match Draw1Control")
	assert.True(t, ids[2], "Draw* should match DrawWindow")
}

func TestBM25SearchOpts_BookFilter(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "Draw1Control", BookID: "book1"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "Draw1Control", BookID: "book2"},
	}))

	results, err := s.BM25SearchOpts("Draw1Control", SearchOptions{
		TopK:   10,
		BookID: "book1",
	})
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].RowID)
}
