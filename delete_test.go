package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete_ExistingDoc(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "doc1"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "doc2"},
	}))

	require.NoError(t, s.Delete(1))
	assert.Equal(t, 1, s.Count())

	results, err := s.Search([]float32{1, 0, 0, 0}, 5)
	require.NoError(t, err)
	for _, r := range results {
		assert.NotEqual(t, 1, r.RowID, "deleted doc should not appear in results")
	}
}

func TestDelete_Nonexistent_NoError(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	err := s.Delete(999)
	assert.NoError(t, err)
}
