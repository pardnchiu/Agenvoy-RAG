package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pardnchiu/AgenvoyRAG/internal/filesystem"
)

const (
	binaryName   = "AgenRAG"
	pollInterval = 10 * time.Second
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	if err := ctx.Err(); err != nil {
		slog.Info("shutdown before start", "reason", err)
		os.Exit(1)
	}
	defer cancel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("os.UserHomeDir",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	if homeDir == "" {
		slog.Error("home directory is empty")
		os.Exit(1)
	}

	folderDir := filepath.Join(homeDir, binaryName)
	info, err := os.Stat(folderDir)
	switch {
	case err == nil:
		if !info.IsDir() {
			slog.Error("path exists and is not a directory")
			os.Exit(1)
		}

	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(folderDir, 0o755); err != nil {
			slog.Error("mkdir failed",
				slog.String("error", err.Error()))
			os.Exit(1)
		}

	default:
		slog.Error("os.Statd",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var prev *map[string]filesystem.FileData

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutdown",
				slog.String("reason", ctx.Err().Error()))
			return

		case <-ticker.C:
			prev = filesystem.WalkFiles(ctx, folderDir, folderDir, prev)
		}
	}
}
