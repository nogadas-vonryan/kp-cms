package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kpcms/server/core/document"

	_ "modernc.org/sqlite"
)

type MetadataStore struct {
	db *sql.DB
}

func NewMetadataStore(dbPath string) (*MetadataStore, error) {
	if dbPath == "" {
		return nil, errors.New("db path is required")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open metadata sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping metadata sqlite: %w", err)
	}

	if err := initMetadataSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &MetadataStore{db: db}, nil
}

func (m *MetadataStore) Close() error {
	if m == nil || m.db == nil {
		return nil
	}
	return m.db.Close()
}

func (m *MetadataStore) UpsertDocument(ctx context.Context, doc *document.Document) error {
	if m == nil || m.db == nil {
		return errors.New("metadata store not initialized")
	}
	if doc == nil {
		return errors.New("document is nil")
	}
	if doc.UUID == "" {
		return errors.New("document UUID is required")
	}

	fieldsJSON, err := json.Marshal(doc.Fields)
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	createdAt := formatTimeUTC(doc.CreatedAt)
	updatedAt := formatTimeUTC(doc.UpdatedAt)

	_, err = m.db.ExecContext(ctx, `
INSERT INTO documents (uuid, code, folder_name, title, fields_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uuid) DO UPDATE SET
	code = excluded.code,
	folder_name = excluded.folder_name,
	title = excluded.title,
	fields_json = excluded.fields_json,
	created_at = excluded.created_at,
	updated_at = excluded.updated_at;`,
		doc.UUID,
		doc.Code,
		doc.FolderName,
		doc.Title,
		string(fieldsJSON),
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}

	return nil
}

func (m *MetadataStore) GetDocumentByUUID(ctx context.Context, uuid string) (*document.Document, error) {
	if m == nil || m.db == nil {
		return nil, errors.New("metadata store not initialized")
	}
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	var doc document.Document
	var fieldsJSON string
	var createdAt string
	var updatedAt string

	err := m.db.QueryRowContext(ctx, `
SELECT uuid, code, folder_name, title, fields_json, created_at, updated_at
FROM documents
WHERE uuid = ?`, uuid).Scan(
		&doc.UUID,
		&doc.Code,
		&doc.FolderName,
		&doc.Title,
		&fieldsJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if fieldsJSON != "" {
		if err := json.Unmarshal([]byte(fieldsJSON), &doc.Fields); err != nil {
			return nil, fmt.Errorf("unmarshal fields: %w", err)
		}
	}

	parsedCreatedAt, err := parseTimeUTC(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	parsedUpdatedAt, err := parseTimeUTC(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	doc.CreatedAt = parsedCreatedAt
	doc.UpdatedAt = parsedUpdatedAt

	return &doc, nil
}

func (m *MetadataStore) DeleteDocument(ctx context.Context, uuid string) error {
	if m == nil || m.db == nil {
		return errors.New("metadata store not initialized")
	}
	if uuid == "" {
		return errors.New("uuid is required")
	}

	_, err := m.db.ExecContext(ctx, `DELETE FROM documents WHERE uuid = ?`, uuid)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

func (m *MetadataStore) ReplaceFiles(ctx context.Context, uuid string, files []document.File) error {
	if m == nil || m.db == nil {
		return errors.New("metadata store not initialized")
	}
	if uuid == "" {
		return errors.New("uuid is required")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE document_uuid = ?`, uuid); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clear files: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO files (document_uuid, filename, description, note, tags_json, type, size, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare insert file: %w", err)
	}
	defer stmt.Close()

	for _, f := range files {
		tagsJSON, err := json.Marshal(f.Tags)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("marshal tags: %w", err)
		}

		createdAt := formatTimeUTC(f.CreatedAt)

		if _, err := stmt.ExecContext(ctx,
			uuid,
			f.FileName,
			f.Description,
			f.Note,
			string(tagsJSON),
			f.Type,
			f.Size,
			createdAt,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert file: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit files: %w", err)
	}

	return nil
}

func (m *MetadataStore) GetFiles(ctx context.Context, uuid string) ([]document.File, error) {
	if m == nil || m.db == nil {
		return nil, errors.New("metadata store not initialized")
	}
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	rows, err := m.db.QueryContext(ctx, `
SELECT filename, description, note, tags_json, type, size, created_at
FROM files
WHERE document_uuid = ?
ORDER BY filename ASC`, uuid)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}
	defer rows.Close()

	files := []document.File{}
	for rows.Next() {
		var f document.File
		var tagsJSON string
		var createdAt string

		if err := rows.Scan(&f.FileName, &f.Description, &f.Note, &tagsJSON, &f.Type, &f.Size, &createdAt); err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}

		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &f.Tags); err != nil {
				return nil, fmt.Errorf("unmarshal tags: %w", err)
			}
		}

		parsedCreatedAt, err := parseTimeUTC(createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse file created_at: %w", err)
		}
		f.CreatedAt = parsedCreatedAt

		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return files, nil
}

func initMetadataSchema(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS documents (
	uuid TEXT PRIMARY KEY,
	code TEXT NOT NULL,
	folder_name TEXT NOT NULL,
	title TEXT NOT NULL,
	fields_json TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	document_uuid TEXT NOT NULL,
	filename TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	note TEXT NOT NULL DEFAULT '',
	tags_json TEXT NOT NULL DEFAULT '[]',
	type TEXT NOT NULL DEFAULT '',
	size INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	FOREIGN KEY(document_uuid) REFERENCES documents(uuid) ON DELETE CASCADE,
	UNIQUE(document_uuid, filename)
);

CREATE INDEX IF NOT EXISTS idx_documents_code ON documents(code);
CREATE INDEX IF NOT EXISTS idx_documents_folder_name ON documents(folder_name);
CREATE INDEX IF NOT EXISTS idx_files_document_uuid ON files(document_uuid);
`)
	if err != nil {
		return fmt.Errorf("init metadata schema: %w", err)
	}

	return nil
}

func formatTimeUTC(value time.Time) string {
	if value.IsZero() {
		return time.Time{}.UTC().Format(time.RFC3339Nano)
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTimeUTC(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}.UTC(), nil
	}
	return time.Parse(time.RFC3339Nano, value)
}
