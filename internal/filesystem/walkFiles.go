package filesystem

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pardnchiu/AgenvoyRAG/internal/filesystem/parser"
)

type File struct {
	Size     int64
	ModTime  time.Time
	IsDir    bool
	Children *map[string]File
}

func WalkFiles(ctx context.Context, root, dir string, prev *map[string]File) *map[string]File {
	if err := ctx.Err(); err != nil {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Warn("os.ReadDir",
			slog.String("error", err.Error()))
		return nil
	}

	result := make(map[string]File, len(entries))

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

		data := File{
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		}

		unchanged := false
		var prevChildren *map[string]File
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

			if !data.IsDir {
				switch strings.ToLower(filepath.Ext(entry.Name())) {
				case ".pdf":
					docs, perr := parser.PDF(ctx, path)
					if perr != nil {
						slog.Warn("parsePDF",
							slog.String("error", perr.Error()))
						break
					}
					slog.Info("parsed pdf",
						slog.String("path", path),
						slog.Int("pages", len(docs)))
					for e := range docs {
						slog.Info("pdf page",
							slog.String("path", path),
							slog.String("content", docs[e].Content))
					}
				}
			}
		}

		if entry.IsDir() {
			data.Children = WalkFiles(ctx, root, path, prevChildren)
		}

		result[entry.Name()] = data
	}
	return &result
}
