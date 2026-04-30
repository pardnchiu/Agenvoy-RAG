package openai

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

type embedRequest struct {
	Input      []string `json:"input"`
	Model      string   `json:"model"`
	Dimensions int      `json:"dimensions"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

type apiErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type Vector []float32

func (o *OpenAI) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	if o == nil || o.client == nil {
		return nil, fmt.Errorf("openai: not initialized")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embedRequest{Input: texts, Model: model, Dimensions: dim})
	if err != nil {
		return nil, fmt.Errorf("openai: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: NewRequest: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		var apiErr apiErrorBody
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("openai: %s [%s/%s] (status=%d)",
				apiErr.Error.Message, apiErr.Error.Type, apiErr.Error.Code, resp.StatusCode)
		}
		return nil, fmt.Errorf("openai: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var data embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("openai: decode: %w", err)
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
