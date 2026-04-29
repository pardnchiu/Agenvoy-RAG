package databaseHandler

import (
	"context"
	"fmt"

	"github.com/pardnchiu/AgenvoyRAG/internal/database"
)

func Dismiss(db *database.DB, ctx context.Context, source string) error {
	if db == nil || db.DB == nil {
		return fmt.Errorf("database: not initialized")
	}
	if source == "" {
		return fmt.Errorf("database: source is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := db.DB.ExecContext(ctx, `
UPDATE file_data
SET dismiss = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE source = ?
AND dismiss = FALSE;`, source); err != nil {
		return fmt.Errorf("database: db.db.ExecContext: %w", err)
	}
	return nil
}
