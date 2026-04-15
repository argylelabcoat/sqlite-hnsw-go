package sqlitehnsw

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Store) GetPayload(rowid int) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	var content, entityName, entityKind, manager string
	var returnType, trapID, headerFile, availability sql.NullString
	var bookID, chapterFile, title sql.NullString
	var metaJSON []byte

	err := s.db.QueryRow(`
		SELECT content, meta, entity_name, entity_kind, manager,
		       return_type, trap_id, header_file, availability,
		       book_id, chapter_file, title
		FROM vectors WHERE rowid = ?`, rowid,
	).Scan(&content, &metaJSON, &entityName, &entityKind, &manager,
		&returnType, &trapID, &headerFile, &availability,
		&bookID, &chapterFile, &title)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get payload: %w", err)
	}

	payload := map[string]any{
		"rowid":   rowid,
		"content": content,
	}
	if entityName != "" {
		payload["entity_name"] = entityName
	}
	if entityKind != "" {
		payload["entity_kind"] = entityKind
	}
	if manager != "" {
		payload["manager"] = manager
	}
	if returnType.Valid {
		payload["return_type"] = returnType.String
	}
	if trapID.Valid {
		payload["trap_id"] = trapID.String
	}
	if headerFile.Valid {
		payload["header_file"] = headerFile.String
	}
	if availability.Valid {
		payload["availability"] = availability.String
	}
	if bookID.Valid {
		payload["book_id"] = bookID.String
	}
	if chapterFile.Valid {
		payload["chapter_file"] = chapterFile.String
	}
	if title.Valid {
		payload["title"] = title.String
	}

	var meta map[string]any
	if len(metaJSON) > 0 && json.Unmarshal(metaJSON, &meta) == nil {
		for k, v := range meta {
			payload[k] = v
		}
	}

	return payload, nil
}

func (s *Store) GetPayloads(rowids []int) (map[int]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	if len(rowids) == 0 {
		return map[int]map[string]any{}, nil
	}

	placeholders := make([]string, len(rowids))
	args := make([]any, len(rowids))
	for i, id := range rowids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT rowid, content, entity_name, entity_kind, manager
		FROM vectors WHERE rowid IN (%s)`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get payloads: %w", err)
	}
	defer rows.Close()

	result := make(map[int]map[string]any)
	for rows.Next() {
		var id int
		var content, entityName, entityKind, manager string
		if err := rows.Scan(&id, &content, &entityName, &entityKind, &manager); err != nil {
			return nil, fmt.Errorf("get payloads: scan: %w", err)
		}
		payload := map[string]any{
			"rowid":   id,
			"content": content,
		}
		if entityName != "" {
			payload["entity_name"] = entityName
		}
		if entityKind != "" {
			payload["entity_kind"] = entityKind
		}
		if manager != "" {
			payload["manager"] = manager
		}
		result[id] = payload
	}
	return result, nil
}
