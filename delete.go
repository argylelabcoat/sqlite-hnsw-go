package sqlitehnsw

import "fmt"

func (s *Store) Delete(rowid int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	_, err := s.db.Exec("DELETE FROM vectors WHERE rowid = ?", rowid)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	s.graph.Delete(rowid)
	return nil
}
