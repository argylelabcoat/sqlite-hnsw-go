package sqlitehnsw

import (
	"encoding/json"
	"fmt"
)

func (s *Store) Upsert(points []Point) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("upsert: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO vectors
			(rowid, vec, content, meta, book_id, chapter_file, title,
			 entity_name, entity_kind, manager, return_type, trap_id,
			 header_file, availability)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("upsert: prepare: %w", err)
	}
	defer stmt.Close()

	for _, p := range points {
		if len(p.Vector) != s.cfg.Dimension {
			return fmt.Errorf("upsert point: %w: got %d, want %d",
				ErrDimensionMismatch, len(p.Vector), s.cfg.Dimension)
		}

		var metaJSON []byte
		if p.Meta != nil {
			metaJSON, _ = json.Marshal(p.Meta)
		}

		var id any
		if p.ID != 0 {
			id = p.ID
		}
		encoded := encodeVector(p.Vector)

		result, err := stmt.Exec(id, encoded, p.Content, metaJSON,
			p.BookID, p.ChapterFile, p.Title,
			p.EntityName, p.EntityKind, p.Manager, p.ReturnType, p.TrapID,
			p.HeaderFile, p.Availability)
		if err != nil {
			return fmt.Errorf("upsert: insert: %w", err)
		}

		if p.ID == 0 {
			rowid, _ := result.LastInsertId()
			p.ID = int(rowid)
		}

		s.graph.Insert(p.ID, p.Vector)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("upsert: commit: %w", err)
	}
	committed = true

	s.pendingFlushes += len(points)
	if s.pendingFlushes >= s.cfg.FlushThreshold {
		if err := s.flushGraphLocked(); err != nil {
			return fmt.Errorf("upsert: flush: %w", err)
		}
	}

	return nil
}
