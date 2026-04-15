package hnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerialize_EmptyGraph(t *testing.T) {
	g := testNewGraph()
	data, err := g.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestSerialize_GraphWithNodes(t *testing.T) {
	g := testPopulatedGraph()
	data, err := g.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDeserialize_EmptyGraph(t *testing.T) {
	original := testNewGraph()
	data, err := original.Serialize()
	require.NoError(t, err)

	loaded := testNewGraph()
	err = loaded.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, original.Len(), loaded.Len())
}

func TestDeserialize_PreservesNodeCount(t *testing.T) {
	original := testPopulatedGraph()
	data, err := original.Serialize()
	require.NoError(t, err)

	loaded := testNewGraph()
	err = loaded.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, 20, loaded.Len())
}

func TestDeserialize_InvalidCBOR_ReturnsError(t *testing.T) {
	g := testNewGraph()
	err := g.Deserialize([]byte{0xff, 0xff, 0xff})
	assert.Error(t, err)
}

func TestRoundTrip_SearchResultsMatch(t *testing.T) {
	original := testPopulatedGraph()
	data, err := original.Serialize()
	require.NoError(t, err)

	loaded := testNewGraph()
	err = loaded.Deserialize(data)
	require.NoError(t, err)

	query := []float32{1, 0}
	originalResults := original.Search(query, 5)
	loadedResults := loaded.Search(query, 5)

	require.Len(t, loadedResults, len(originalResults))
	for i := range originalResults {
		assert.Equal(t, originalResults[i].ID, loadedResults[i].ID,
			"result rank %d: ID mismatch after round-trip", i)
	}
}

func TestRoundTrip_DeleteAfterDeserialize(t *testing.T) {
	original := testPopulatedGraph()
	data, err := original.Serialize()
	require.NoError(t, err)

	loaded := testNewGraph()
	err = loaded.Deserialize(data)
	require.NoError(t, err)

	loaded.Delete(0)
	assert.Equal(t, 19, loaded.Len())

	results := loaded.Search([]float32{1, 0}, 5)
	for _, r := range results {
		assert.NotEqual(t, 0, r.ID)
	}
}
