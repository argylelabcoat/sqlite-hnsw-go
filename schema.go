package sqlitehnsw

import "database/sql"

func initSchema(db *sql.DB) error {
	stmts := []string{
		// ── Content table ──────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS content (
			id           INTEGER PRIMARY KEY,
			book_id      TEXT,
			chapter_file TEXT NOT NULL UNIQUE,
			title        TEXT,
			text         TEXT NOT NULL,
			meta         JSON,
			embedded     INTEGER NOT NULL DEFAULT 0,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_content_book     ON content(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_content_embedded ON content(embedded) WHERE embedded = 0`,

		// FTS5 over full chapter text.
		`CREATE VIRTUAL TABLE IF NOT EXISTS fts_chapters USING fts5(
			text,
			content='content',
			content_rowid='id',
			tokenize='porter unicode61'
		)`,
		`CREATE TRIGGER IF NOT EXISTS content_ai AFTER INSERT ON content BEGIN
			INSERT INTO fts_chapters(rowid, text) VALUES (new.id, new.text);
		END`,
		`CREATE TRIGGER IF NOT EXISTS content_au AFTER UPDATE OF text ON content BEGIN
			INSERT INTO fts_chapters(fts_chapters, rowid, text) VALUES ('delete', old.id, old.text);
			INSERT INTO fts_chapters(rowid, text) VALUES (new.id, new.text);
		END`,
		`CREATE TRIGGER IF NOT EXISTS content_ad AFTER DELETE ON content BEGIN
			INSERT INTO fts_chapters(fts_chapters, rowid, text) VALUES ('delete', old.id, old.text);
		END`,

		// ── Books table ───────────────────────────────────────────────────────
		// One row per book; chapter counts and word counts are derived from content.
		`CREATE TABLE IF NOT EXISTS books (
			book_id    TEXT PRIMARY KEY,
			title      TEXT NOT NULL DEFAULT '',
			category   TEXT NOT NULL DEFAULT '',
			base_url   TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ── Vectors table ──────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS vectors (
			rowid        INTEGER PRIMARY KEY,
			vec          BLOB NOT NULL,
			content      TEXT,
			meta         JSON,
			content_id   INTEGER REFERENCES content(id),
			chunk_start  INTEGER,
			chunk_end    INTEGER,
			book_id      TEXT,
			chapter_file TEXT,
			title        TEXT,
			entity_name  TEXT,
			entity_kind  TEXT,
			manager      TEXT,
			return_type  TEXT,
			trap_id      TEXT,
			header_file  TEXT,
			availability TEXT,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_name  ON vectors(entity_name)  WHERE entity_name  IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_manager_kind ON vectors(manager, entity_kind)`,
		`CREATE INDEX IF NOT EXISTS idx_trap_id      ON vectors(trap_id)      WHERE trap_id      IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_return_type  ON vectors(return_type)  WHERE return_type  IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_book_chapter ON vectors(book_id, chapter_file)`,
		`CREATE INDEX IF NOT EXISTS idx_content_id   ON vectors(content_id)   WHERE content_id   IS NOT NULL`,

		// Chunk-level FTS5.
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

		// ── HNSW graph persistence ─────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS hnsw_graph (
			collection      TEXT    NOT NULL PRIMARY KEY,
			dim             INTEGER NOT NULL,
			metric          TEXT    NOT NULL,
			m               INTEGER NOT NULL,
			ef_construction INTEGER NOT NULL,
			entry_point     INTEGER,
			max_level       INTEGER DEFAULT 0,
			blob            BLOB,
			version         INTEGER DEFAULT 1
		)`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	// Migrations for existing databases: add columns silently if absent.
	migrations := []string{
		`ALTER TABLE vectors ADD COLUMN content_id  INTEGER`,
		`ALTER TABLE vectors ADD COLUMN chunk_start INTEGER`,
		`ALTER TABLE vectors ADD COLUMN chunk_end   INTEGER`,
		`ALTER TABLE content ADD COLUMN embedded INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		db.Exec(m) // ignore "duplicate column" errors
	}

	return nil
}
