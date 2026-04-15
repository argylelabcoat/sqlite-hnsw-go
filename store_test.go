package sqlitehnsw

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore_CreatesEmptyStore(t *testing.T) {
	path := testDBPath(t)
	s, err := NewStore(Config{DBPath: path, Dimension: 4})
	require.NoError(t, err)
	defer s.Close()

	assert.Equal(t, 0, s.Count())
}

func TestNewStore_ReopenPreservesCount(t *testing.T) {
	path := testDBPath(t)

	s, err := NewStore(Config{DBPath: path, Dimension: 4})
	require.NoError(t, err)
	require.NoError(t, s.Upsert([]Point{
		{Vector: []float32{1, 0, 0, 0}, Content: "hello"},
	}))
	require.NoError(t, s.Close())

	s2, err := NewStore(Config{DBPath: path, Dimension: 4})
	require.NoError(t, err)
	defer s2.Close()

	assert.Equal(t, 1, s2.Count())
}

func TestNewStore_InvalidDimension_ReturnsError(t *testing.T) {
	path := testDBPath(t)
	_, err := NewStore(Config{DBPath: path, Dimension: 0})
	assert.Error(t, err)
}

func testDBPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "sqlitehnsw-store-*.db")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })
	return path
}

func testStore(t *testing.T, dim int) *Store {
	t.Helper()
	path := testDBPath(t)
	s, err := NewStore(Config{DBPath: path, Dimension: dim})
	require.NoError(t, err)
	return s
}
