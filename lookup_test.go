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

func TestLookupEntity_FindsDocuments(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, EntityName: "Draw1Control"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, EntityName: "FileManager"},
		{ID: 3, Vector: []float32{0, 0, 1, 0}, EntityName: "Draw1Control"},
	}))

	ids, err := s.LookupEntity("Draw1Control")
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 3}, ids)
}

func TestLookupEntity_NotFound_ReturnsEmpty(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	ids, err := s.LookupEntity("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestLookupManager_FindsDocuments(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, Manager: "File Manager", EntityKind: "function"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, Manager: "Control Manager", EntityKind: "function"},
		{ID: 3, Vector: []float32{0, 0, 1, 0}, Manager: "File Manager", EntityKind: "type"},
	}))

	ids, err := s.LookupManager("File Manager")
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 3}, ids)
}

func TestLookupByFilter_SingleCondition(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, EntityKind: "function", Manager: "File Manager"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, EntityKind: "type", Manager: "File Manager"},
	}))

	ids, err := s.LookupByFilter("entity_kind = ?", "function")
	require.NoError(t, err)
	assert.Equal(t, []int{1}, ids)
}

func TestLookupByFilter_CombinedCondition(t *testing.T) {
	s := testStore(t, 4)
	defer s.Close()

	require.NoError(t, s.Upsert([]Point{
		{ID: 1, Vector: []float32{1, 0, 0, 0}, EntityKind: "function", Manager: "File Manager"},
		{ID: 2, Vector: []float32{0, 1, 0, 0}, EntityKind: "function", Manager: "Control Manager"},
	}))

	ids, err := s.LookupByFilter("entity_kind = ? AND manager = ?", "function", "File Manager")
	require.NoError(t, err)
	assert.Equal(t, []int{1}, ids)
}
