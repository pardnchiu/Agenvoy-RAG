package parser

import (
	"context"
	"fmt"

	"github.com/ledongthuc/pdf"
)

type PDFFile struct {
	Source  string
	Page    int
	Total   int
	Content string
}

func PDF(ctx context.Context, path string) (docs []PDFFile, err error) {
	if path == "" {
		return nil, fmt.Errorf("pdf: path is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pdf: panic: %v", r)
		}
	}()

	file, reader, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("pdf: pdf.Open: %w", err)
	}
	defer file.Close()

	len := reader.NumPage()
	if len <= 0 {
		return nil, fmt.Errorf("pdf: empty")
	}

	docs = make([]PDFFile, 0, len)
	for i := 1; i <= len; i++ {
		if err := ctx.Err(); err != nil {
			return docs, err
		}

		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return docs, fmt.Errorf("pdf: page.GetPlainText (index: %d): %w", i, err)
		}

		docs = append(docs, PDFFile{
			Source:  path,
			Page:    i,
			Total:   len,
			Content: text,
		})
	}

	return docs, nil
}
