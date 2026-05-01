package openai

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

type Vector []float32

func (o *OpenAI) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	if o == nil || o.client == nil {
		return nil, fmt.Errorf("openai: not initialized")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body := map[string]any{
		"input":      texts,
		"model":      model,
		"dimensions": dim,
	}
	headers := map[string]string{"Authorization": "Bearer " + o.apiKey}

	data, _, err := go_pkg_http.POST[embedResponse](ctx, o.client, baseURL+"/embeddings", headers, body, "")
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if len(data.Data) != len(texts) {
		return nil, fmt.Errorf("openai: returned %d vectors, want %d", len(data.Data), len(texts))
	}

	out := make([]Vector, len(data.Data))
	for i, d := range data.Data {
		if len(d.Embedding) != dim {
			return nil, fmt.Errorf("openai: returned dim %d, want %d (index %d)", len(d.Embedding), dim, i)
		}
		v := make(Vector, len(d.Embedding))
		copy(v, d.Embedding)
		out[i] = v
	}
	return out, nil
}

func Encode(v Vector) []byte {
	b := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

func Decode(b []byte) (Vector, error) {
	if len(b)%4 != 0 {
		return nil, fmt.Errorf("vector: blob length %d not multiple of 4", len(b))
	}
	n := len(b) / 4
	v := make(Vector, n)
	for i := range n {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v, nil
}
