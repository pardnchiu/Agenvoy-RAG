package database

import (
	"context"
	"fmt"
)

func (db *DB) Dismiss(ctx context.Context, source string) error {
	if db == nil || db.db == nil {
		return fmt.Errorf("database: not initialized")
	}
	if source == "" {
		return fmt.Errorf("database: source is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := db.db.ExecContext(ctx, `
UPDATE file_data
SET dismiss = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE source = ?
AND dismiss = FALSE;`, source); err != nil {
		return fmt.Errorf("database: db.db.ExecContext: %w", err)
	}
	return nil
}
