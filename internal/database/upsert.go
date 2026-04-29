package database

import (
	"context"
	"fmt"

	"github.com/pardnchiu/AgenvoyRAG/internal/filesystem/parser"
)

func (db *DB) Upsert(ctx context.Context, source string, files []parser.FileData) error {
	if db == nil || db.db == nil {
		return fmt.Errorf("database: not initialized")
	}
	if source == "" {
		return fmt.Errorf("database: source is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database: db.db.BeginTx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
UPDATE file_data
SET dismiss = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE source = ?
AND dismiss = FALSE;
`, source); err != nil {
		return fmt.Errorf("database: tx.ExecContext: %w", err)
	}

	for _, f := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO file_data (source, chunk, total, content, dismiss)
VALUES (?, ?, ?, ?, FALSE)
ON CONFLICT (source, chunk)
DO UPDATE SET
    total      = excluded.total,
    content    = excluded.content,
    dismiss    = FALSE,
    updated_at = CURRENT_TIMESTAMP;`,
			f.Source, f.Index, f.Total, f.Content,
		); err != nil {
			return fmt.Errorf("database: tx.ExecContext (chunk=%d): %w", f.Index, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: tx.Commit: %w", err)
	}
	return nil
}
