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
