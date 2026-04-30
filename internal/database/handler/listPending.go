package databaseHandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pardnchiu/KuraDB/internal/database"
)

type PendingItem struct {
	ID      int64
	Content string
}

func ListPending(db *database.DB, ctx context.Context, limit int) ([]PendingItem, error) {
	if db == nil || db.DB == nil {
		return nil, fmt.Errorf("db is required")
	}
	if limit <= 0 {
		slog.Warn("ListPending: invalid limit, using 10 as default")
		limit = 10
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rows, err := db.DB.QueryContext(ctx, `
SELECT id, content
FROM file_data
WHERE is_embed = FALSE
AND dismiss = FALSE
ORDER BY id ASC
LIMIT ?;
`, limit)
	if err != nil {
		return nil, fmt.Errorf("db.DB.QueryContext: %w", err)
	}
	defer rows.Close()

	results := make([]PendingItem, 0)
	for rows.Next() {
		var p PendingItem
		if err := rows.Scan(&p.ID, &p.Content); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return results, nil
}
