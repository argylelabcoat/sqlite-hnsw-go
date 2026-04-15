package sqlitehnsw

import "database/sql"

func initSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS vectors (
			rowid       INTEGER PRIMARY KEY,
			vec         BLOB NOT NULL,
			content     TEXT,
			meta        JSON,
			book_id     TEXT,
			chapter_file TEXT,
			title       TEXT,
			entity_name TEXT,
			entity_kind TEXT,
			manager     TEXT,
			return_type TEXT,
			trap_id     TEXT,
			header_file TEXT,
			availability TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_name ON vectors(entity_name) WHERE entity_name IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_manager_kind ON vectors(manager, entity_kind)`,
		`CREATE INDEX IF NOT EXISTS idx_trap_id ON vectors(trap_id) WHERE trap_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_return_type ON vectors(return_type) WHERE return_type IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_book_chapter ON vectors(book_id, chapter_file)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS fts_content USING fts5(
			content,
			content='vectors',
			content_rowid='rowid',
			tokenize='porter unicode61'
		)`,
		`CREATE TRIGGER IF NOT EXISTS vectors_ai AFTER INSERT ON vectors BEGIN
			INSERT INTO fts_content(rowid, content) VALUES (new.rowid, new.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS vectors_ad AFTER DELETE ON vectors BEGIN
			INSERT INTO fts_content(fts_content, rowid, content) VALUES('delete', old.rowid, old.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS vectors_au AFTER UPDATE ON vectors BEGIN
			INSERT INTO fts_content(fts_content, rowid, content) VALUES('delete', old.rowid, old.content);
			INSERT INTO fts_content(rowid, content) VALUES (new.rowid, new.content);
		END`,
		`CREATE TABLE IF NOT EXISTS hnsw_graph (
			collection      TEXT NOT NULL,
			dim             INTEGER NOT NULL,
			metric          TEXT NOT NULL,
			m               INTEGER NOT NULL,
			ef_construction INTEGER NOT NULL,
			entry_point     INTEGER,
			max_level       INTEGER DEFAULT 0,
			blob            BLOB,
			version         INTEGER DEFAULT 1,
			PRIMARY KEY (collection)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
