package segmenter

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-ego/gse"
)

var (
	once      sync.Once
	segmenter *gse.Segmenter
)

func New() {
	once.Do(func() {
		var seg gse.Segmenter
		if err := seg.LoadDictEmbed("zh_s"); err != nil {
			slog.Warn("seg.LoadDictEmbed",
				slog.String("error", err.Error()))
		}

		if err := seg.LoadStopEmbed(); err != nil {
			slog.Warn("seg.LoadStopEmbed",
				slog.String("error", err.Error()))
		}
		segmenter = &seg
	})
}

func Tokenize(text string) ([]string, error) {
	if segmenter == nil {
		return nil, errors.New("not initialized")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	raw := segmenter.Cut(text, true)
	words := segmenter.Trim(raw)
	seen := make(map[string]struct{}, len(words))
	out := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if word == "" {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		out = append(out, word)
	}
	return out, nil
}
