package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsert_SinglePoint(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	err := s.Upsert([]Point{{
		Vector:  []float32{1, 0, 0, 0},
		Content: "test document",
		Meta:    map[string]any{"key": "value"},
	}})
	require.NoError(t, err)
	assert.Equal(t, 1, s.Count())
}

func TestUpsert_BatchPoints(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	points := []Point{
		{Vector: []float32{1, 0, 0, 0}, Content: "doc 1"},
		{Vector: []float32{0, 1, 0, 0}, Content: "doc 2"},
		{Vector: []float32{0, 0, 1, 0}, Content: "doc 3"},
	}
	require.NoError(t, s.Upsert(points))
	assert.Equal(t, 3, s.Count())
}

func TestUpsert_WrongDimension_ReturnsError(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	err := s.Upsert([]Point{{Vector: []float32{1, 0}}})
	assert.ErrorIs(t, err, ErrDimensionMismatch)
}
