package filesystem

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileData struct {
	Size     int64
	ModTime  time.Time
	IsDir    bool
	Children *map[string]FileData
}

func WalkFiles(ctx context.Context, root, dir string, prev *map[string]FileData) *map[string]FileData {
	if err := ctx.Err(); err != nil {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Warn("os.ReadDir",
			slog.String("error", err.Error()))
		return nil
	}

	result := make(map[string]FileData, len(entries))

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return &result
		}

		path := filepath.Join(dir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			slog.Warn("entry.Info",
				slog.String("error", err.Error()))
			continue
		}

		data := FileData{
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		}

		unchanged := false
		var prevChildren *map[string]FileData
		if prev != nil {
			if p, ok := (*prev)[entry.Name()]; ok && p.IsDir == data.IsDir {
				prevChildren = p.Children
				if p.Size == data.Size && p.ModTime.Equal(data.ModTime) {
					unchanged = true
				}
			}
		}

		if !unchanged {
			slog.Info("changed",
				slog.String("path", path))
		}

		if entry.IsDir() {
			data.Children = WalkFiles(ctx, root, path, prevChildren)
		}

		result[entry.Name()] = data
	}
	return &result
}
