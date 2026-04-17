package sqlitehnsw

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
)

// ContentEntry represents a full source document (chapter, PDF page range, etc.).
// One ContentEntry is stored per source file; chunks reference it via ContentID.
type ContentEntry struct {
	ID          int
	BookID      string
	ChapterFile string // unique path relative to the organized dir
	Title       string
	Text        string         // complete organized text
	Meta        map[string]any // arbitrary metadata
}

// ChapterResult is returned by SearchChapters.
type ChapterResult struct {
	ContentID int
	Score     float64 // BM25 score (lower magnitude = better; SQLite BM25 is negative)
}

// UpsertContent inserts or replaces a ContentEntry.
// Returns the assigned content ID.
func (s *Store) UpsertContent(entry ContentEntry) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return 0, ErrStoreClosed
	}

	var metaJSON []byte
	if entry.Meta != nil {
		var err error
		metaJSON, err = json.Marshal(entry.Meta)
		if err != nil {
			return 0, fmt.Errorf("upsert content: marshal meta: %w", err)
		}
	}

	var result sql.Result
	var err error
	if entry.ID != 0 {
		result, err = s.db.Exec(`
			INSERT OR REPLACE INTO content (id, book_id, chapter_file, title, text, meta, embedded, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP)`,
			entry.ID, entry.BookID, entry.ChapterFile, entry.Title, entry.Text, metaJSON)
	} else {
		result, err = s.db.Exec(`
			INSERT INTO content (book_id, chapter_file, title, text, meta)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(chapter_file) DO UPDATE SET
				book_id    = excluded.book_id,
				title      = excluded.title,
				text       = excluded.text,
				meta       = excluded.meta,
				embedded   = CASE WHEN excluded.text != content.text THEN 0 ELSE content.embedded END,
				updated_at = CURRENT_TIMESTAMP`,
			entry.BookID, entry.ChapterFile, entry.Title, entry.Text, metaJSON)
	}
	if err != nil {
		return 0, fmt.Errorf("upsert content: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("upsert content: last insert id: %w", err)
	}
	if id == 0 {
		if err := s.db.QueryRow(
			"SELECT id FROM content WHERE chapter_file = ?", entry.ChapterFile,
		).Scan(&id); err != nil {
			return 0, fmt.Errorf("upsert content: lookup id: %w", err)
		}
	}
	return int(id), nil
}

// ListUnembedded returns all content entries whose embedded flag is 0,
// ordered by id ascending (stable processing order).
func (s *Store) ListUnembedded() ([]ContentEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	rows, err := s.db.Query(
		`SELECT id, book_id, chapter_file, title, text, meta FROM content WHERE embedded = 0 ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list unembedded: %w", err)
	}
	defer rows.Close()

	var entries []ContentEntry
	for rows.Next() {
		var e ContentEntry
		var bookID, title sql.NullString
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &bookID, &e.ChapterFile, &title, &e.Text, &metaJSON); err != nil {
			return nil, fmt.Errorf("list unembedded: scan: %w", err)
		}
		e.BookID = bookID.String
		e.Title = title.String
		if len(metaJSON) > 0 {
			e.Meta = make(map[string]any)
			json.Unmarshal(metaJSON, &e.Meta)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// MarkEmbedded sets embedded = 1 for the given content ID.
func (s *Store) MarkEmbedded(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	_, err := s.db.Exec(`UPDATE content SET embedded = 1 WHERE id = ?`, id)
	return err
}

// ResetEmbedded sets embedded = 0 for all content rows, triggering
// a full re-embedding on the next ingest run. Used by --force.
func (s *Store) ResetEmbedded() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	_, err := s.db.Exec(`UPDATE content SET embedded = 0`)
	return err
}

// GetContent retrieves a ContentEntry by its ID.
func (s *Store) GetContent(id int) (ContentEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return ContentEntry{}, ErrStoreClosed
	}

	return s.scanContentRow(
		s.db.QueryRow("SELECT id, book_id, chapter_file, title, text, meta FROM content WHERE id = ?", id),
	)
}

// GetContentByChapterFile retrieves a ContentEntry by chapter_file path.
func (s *Store) GetContentByChapterFile(chapterFile string) (ContentEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return ContentEntry{}, ErrStoreClosed
	}

	return s.scanContentRow(
		s.db.QueryRow("SELECT id, book_id, chapter_file, title, text, meta FROM content WHERE chapter_file = ?", chapterFile),
	)
}

// GetContentForChunk retrieves the ContentEntry linked to a vector rowID.
func (s *Store) GetContentForChunk(rowID int) (ContentEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return ContentEntry{}, ErrStoreClosed
	}

	return s.scanContentRow(s.db.QueryRow(`
		SELECT c.id, c.book_id, c.chapter_file, c.title, c.text, c.meta
		FROM content c
		JOIN vectors v ON v.content_id = c.id
		WHERE v.rowid = ?`, rowID))
}

// SearchChapters performs BM25 full-text search over the content table.
// bookID filters results to a single book when non-empty.
func (s *Store) SearchChapters(query string, topK int, bookID string) ([]ChapterResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	if topK <= 0 {
		topK = 10
	}

	var (
		sqlQuery string
		args     []any
	)
	if bookID != "" {
		sqlQuery = `
			SELECT f.rowid, bm25(fts_chapters) AS score
			FROM fts_chapters f
			JOIN content c ON f.rowid = c.id
			WHERE fts_chapters MATCH ? AND c.book_id = ?
			ORDER BY score LIMIT ?`
		args = []any{query, bookID, topK}
	} else {
		sqlQuery = `
			SELECT rowid, bm25(fts_chapters) AS score
			FROM fts_chapters
			WHERE fts_chapters MATCH ?
			ORDER BY score LIMIT ?`
		args = []any{query, topK}
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search chapters: %w", err)
	}
	defer rows.Close()

	var results []ChapterResult
	for rows.Next() {
		var r ChapterResult
		if err := rows.Scan(&r.ContentID, &r.Score); err != nil {
			return nil, fmt.Errorf("search chapters: scan: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}

// DeleteContent removes a content entry and all of its associated chunks.
func (s *Store) DeleteContent(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	// Find chunk rowIDs linked to this content entry.
	rows, err := s.db.Query("SELECT rowid FROM vectors WHERE content_id = ?", id)
	if err != nil {
		return fmt.Errorf("delete content: list chunks: %w", err)
	}
	var chunkIDs []int
	for rows.Next() {
		var rid int
		rows.Scan(&rid)
		chunkIDs = append(chunkIDs, rid)
	}
	rows.Close()

	// Delete chunks from the HNSW graph and vectors table.
	for _, rid := range chunkIDs {
		s.graph.Delete(rid)
	}
	if _, err := s.db.Exec("DELETE FROM vectors WHERE content_id = ?", id); err != nil {
		return fmt.Errorf("delete content: delete chunks: %w", err)
	}

	// Delete the content row (triggers FTS cleanup).
	if _, err := s.db.Exec("DELETE FROM content WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete content: %w", err)
	}
	return nil
}

// scanContentRow scans a *sql.Row into a ContentEntry.
func (s *Store) scanContentRow(row *sql.Row) (ContentEntry, error) {
	var e ContentEntry
	var bookID, title sql.NullString
	var metaJSON []byte

	if err := row.Scan(&e.ID, &bookID, &e.ChapterFile, &title, &e.Text, &metaJSON); err != nil {
		if err == sql.ErrNoRows {
			return ContentEntry{}, ErrNotFound
		}
		return ContentEntry{}, fmt.Errorf("scan content: %w", err)
	}
	e.BookID = bookID.String
	e.Title = title.String

	if len(metaJSON) > 0 {
		e.Meta = make(map[string]any)
		if err := json.Unmarshal(metaJSON, &e.Meta); err != nil {
			e.Meta = nil
		}
	}
	return e, nil
}

// ContentMeta returns the metadata map merged with denormalized fields,
// suitable for use as a GetPayload response for a chapter.
func (e ContentEntry) ContentMeta() map[string]any {
	out := map[string]any{
		"content_id":   e.ID,
		"book_id":      e.BookID,
		"chapter_file": e.ChapterFile,
		"title":        e.Title,
	}
	maps.Copy(out, e.Meta)
	return out
}
