package sqlitehnsw

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestInitSchema_CreatesTables(t *testing.T) {
	db := testOpenDB(t)
	defer db.Close()

	err := initSchema(db)
	require.NoError(t, err)

	var name string
	tables := []string{"vectors", "hnsw_graph", "fts_content"}
	for _, tbl := range tables {
		err = db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl,
		).Scan(&name)
		assert.NoError(t, err, "table %s should exist", tbl)
	}
}

func TestInitSchema_Idempotent(t *testing.T) {
	db := testOpenDB(t)
	defer db.Close()

	require.NoError(t, initSchema(db))
	assert.NoError(t, initSchema(db), "running initSchema twice should not error")
}

func testOpenDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "sqlitehnsw-test-*.db")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })

	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	return db
}
