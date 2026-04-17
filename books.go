package sqlitehnsw

import (
	"database/sql"
	"fmt"
)

// BookSummary is a lightweight book record returned by ListBooks.
type BookSummary struct {
	BookID       string
	Title        string
	Category     string
	BaseURL      string
	ChapterCount int
	TotalChars   int
}

// UpsertBook inserts or updates a book metadata record.
func (s *Store) UpsertBook(bookID, title, category, baseURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	_, err := s.db.Exec(`
		INSERT INTO books (book_id, title, category, base_url)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(book_id) DO UPDATE SET
			title      = excluded.title,
			category   = excluded.category,
			base_url   = excluded.base_url,
			updated_at = CURRENT_TIMESTAMP`,
		bookID, title, category, baseURL)
	if err != nil {
		return fmt.Errorf("upsert book: %w", err)
	}
	return nil
}

// ListBooks returns summary information for every book, joined with chapter
// counts and total character lengths from the content table.
func (s *Store) ListBooks() ([]BookSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	rows, err := s.db.Query(`
		SELECT b.book_id, b.title, b.category, b.base_url,
		       COUNT(c.id)          AS chapter_count,
		       COALESCE(SUM(LENGTH(c.text)), 0) AS total_chars
		FROM books b
		LEFT JOIN content c ON c.book_id = b.book_id
		GROUP BY b.book_id
		ORDER BY b.title`)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []BookSummary
	for rows.Next() {
		var b BookSummary
		if err := rows.Scan(&b.BookID, &b.Title, &b.Category, &b.BaseURL, &b.ChapterCount, &b.TotalChars); err != nil {
			return nil, fmt.Errorf("list books: scan: %w", err)
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

// ListChaptersForBook returns all ContentEntry rows for a given book_id,
// ordered by chapter_file. Only metadata fields (no text) are returned to
// keep the response small — callers that need text use GetContent/GetContentByChapterFile.
func (s *Store) ListChaptersForBook(bookID string) ([]ContentEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	rows, err := s.db.Query(`
		SELECT id, book_id, chapter_file, title, '', NULL
		FROM content WHERE book_id = ? ORDER BY chapter_file`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list chapters for book %s: %w", bookID, err)
	}
	defer rows.Close()

	var entries []ContentEntry
	for rows.Next() {
		var e ContentEntry
		var bookIDNull, title sql.NullString
		var textPlaceholder string
		var metaPlaceholder []byte
		if err := rows.Scan(&e.ID, &bookIDNull, &e.ChapterFile, &title, &textPlaceholder, &metaPlaceholder); err != nil {
			return nil, fmt.Errorf("list chapters: scan: %w", err)
		}
		e.BookID = bookIDNull.String
		e.Title = title.String
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
