package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPayload_ExistingDoc(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{{
		ID:      1,
		Vector:  []float32{1, 0, 0, 0},
		Content: "hello world",
		Meta:    map[string]any{"source": "test"},
	}}))

	payload, err := s.GetPayload(1)
	require.NoError(t, err)
	assert.Equal(t, "hello world", payload["content"])
}

func TestGetPayload_Nonexistent(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	_, err := s.GetPayload(999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetPayloads_Batch(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Content: "doc1"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Content: "doc2"},
	}))

	payloads, err := s.GetPayloads([]int{1, 2})
	require.NoError(t, err)
	assert.Len(t, payloads, 2)
	assert.Equal(t, "doc1", payloads[1]["content"])
	assert.Equal(t, "doc2", payloads[2]["content"])
}
