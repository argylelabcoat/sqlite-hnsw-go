package sqlitehnsw

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_FindsClosest(t *testing.T) {
	s := testStore(t, 2)
	defer s.Close()

	points := []Point{
		{ID: 1, Vector: []float32{1, 0}, Content: "right"},
		{ID: 2, Vector: []float32{0, 1}, Content: "up"},
		{ID: 3, Vector: []float32{-1, 0}, Content: "left"},
	}
	require.NoError(t, s.Upsert(points))

	results, err := s.Search([]float32{0.9, 0.1}, 2)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 1, results[0].RowID, "node 1 ([1,0]) should be closest to [0.9,0.1]")
}

func TestSearch_EmptyStore_ReturnsEmpty(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	results, err := s.Search([]float32{1, 0, 0, 0}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_TopK(t *testing.T) {
	s := testStore(t, 2)
	defer s.Close()

	for i := 0; i < 20; i++ {
		angle := float32(i) * 2 * math.Pi / 20
		require.NoError(t, s.Upsert([]Point{{
			ID: i + 1, Vector: []float32{float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))},
		}}))
	}

	results, err := s.Search([]float32{1, 0}, 5)
	require.NoError(t, err)
	assert.Len(t, results, 5)
	assert.Equal(t, 1, results[0].RowID, "node 1 (angle=0) should be closest")
}
