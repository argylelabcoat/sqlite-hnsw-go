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
