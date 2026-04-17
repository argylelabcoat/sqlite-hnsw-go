package sqlitehnsw_test

import (
	"testing"

	sqlitehnsw "github.com/macintosh-codex/sqlite-hnsw"
)

func newTestStore(t *testing.T) *sqlitehnsw.Store {
	t.Helper()
	s, err := sqlitehnsw.NewStore(sqlitehnsw.Config{
		DBPath:    t.TempDir() + "/test.db",
		Dimension: 4,
	})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestUpsertContent_EmbeddedResetOnTextChange(t *testing.T) {
	s := newTestStore(t)

	id, err := s.UpsertContent(sqlitehnsw.ContentEntry{
		BookID: "book1", ChapterFile: "book1/ch1.md", Title: "Ch1", Text: "original",
	})
	if err != nil {
		t.Fatalf("UpsertContent: %v", err)
	}

	if err := s.MarkEmbedded(id); err != nil {
		t.Fatalf("MarkEmbedded: %v", err)
	}

	// Upsert with same text — embedded should stay 1.
	_, err = s.UpsertContent(sqlitehnsw.ContentEntry{
		BookID: "book1", ChapterFile: "book1/ch1.md", Title: "Ch1", Text: "original",
	})
	if err != nil {
		t.Fatalf("UpsertContent (same text): %v", err)
	}
	unembedded, err := s.ListUnembedded()
	if err != nil {
		t.Fatalf("ListUnembedded: %v", err)
	}
	if len(unembedded) != 0 {
		t.Errorf("expected 0 unembedded after same-text upsert, got %d", len(unembedded))
	}

	// Upsert with changed text — embedded should reset to 0.
	_, err = s.UpsertContent(sqlitehnsw.ContentEntry{
		BookID: "book1", ChapterFile: "book1/ch1.md", Title: "Ch1", Text: "changed",
	})
	if err != nil {
		t.Fatalf("UpsertContent (changed text): %v", err)
	}
	unembedded, err = s.ListUnembedded()
	if err != nil {
		t.Fatalf("ListUnembedded: %v", err)
	}
	if len(unembedded) != 1 {
		t.Errorf("expected 1 unembedded after text change, got %d", len(unembedded))
	}
}

func TestListUnembedded_OrderByID(t *testing.T) {
	s := newTestStore(t)

	for i, ch := range []string{"book1/a.md", "book1/b.md", "book1/c.md"} {
		id, err := s.UpsertContent(sqlitehnsw.ContentEntry{
			BookID: "book1", ChapterFile: ch, Title: ch, Text: "text",
		})
		if err != nil {
			t.Fatalf("UpsertContent %d: %v", i, err)
		}
		if i == 1 {
			s.MarkEmbedded(id) // mark middle one as embedded
		}
	}

	entries, err := s.ListUnembedded()
	if err != nil {
		t.Fatalf("ListUnembedded: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 unembedded, got %d", len(entries))
	}
	if entries[0].ChapterFile != "book1/a.md" || entries[1].ChapterFile != "book1/c.md" {
		t.Errorf("unexpected order: %v %v", entries[0].ChapterFile, entries[1].ChapterFile)
	}
}

func TestResetEmbedded(t *testing.T) {
	s := newTestStore(t)

	for _, ch := range []string{"book1/a.md", "book1/b.md"} {
		id, _ := s.UpsertContent(sqlitehnsw.ContentEntry{
			BookID: "book1", ChapterFile: ch, Title: ch, Text: "text",
		})
		s.MarkEmbedded(id)
	}

	if err := s.ResetEmbedded(); err != nil {
		t.Fatalf("ResetEmbedded: %v", err)
	}

	entries, err := s.ListUnembedded()
	if err != nil {
		t.Fatalf("ListUnembedded: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 unembedded after reset, got %d", len(entries))
	}
}

func TestUpsertBook_AndListBooks(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertBook("ctrl", "Control Manager", "Reference", ""); err != nil {
		t.Fatalf("UpsertBook: %v", err)
	}
	s.UpsertContent(sqlitehnsw.ContentEntry{BookID: "ctrl", ChapterFile: "ctrl/ch1.md", Title: "Ch1", Text: "hello world"})
	s.UpsertContent(sqlitehnsw.ContentEntry{BookID: "ctrl", ChapterFile: "ctrl/ch2.md", Title: "Ch2", Text: "foo bar baz"})

	books, err := s.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}
	b := books[0]
	if b.BookID != "ctrl" || b.Title != "Control Manager" || b.ChapterCount != 2 {
		t.Errorf("unexpected book: %+v", b)
	}
}

func TestListChaptersForBook(t *testing.T) {
	s := newTestStore(t)

	s.UpsertBook("files", "File Manager", "Reference", "")
	s.UpsertContent(sqlitehnsw.ContentEntry{BookID: "files", ChapterFile: "files/intro.md", Title: "Intro", Text: "text"})
	s.UpsertContent(sqlitehnsw.ContentEntry{BookID: "files", ChapterFile: "files/api.md", Title: "API", Text: "text"})

	chapters, err := s.ListChaptersForBook("files")
	if err != nil {
		t.Fatalf("ListChaptersForBook: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	// Ordered by chapter_file.
	if chapters[0].ChapterFile != "files/api.md" || chapters[1].ChapterFile != "files/intro.md" {
		t.Errorf("unexpected order: %v %v", chapters[0].ChapterFile, chapters[1].ChapterFile)
	}
}
