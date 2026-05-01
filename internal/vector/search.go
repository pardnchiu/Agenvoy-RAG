package vector

import (
	"math"
	"runtime"
	"sort"
	"sync"
)

type Hit struct {
	ID    int64
	Score float64
}

type SourceHit struct {
	source string
	score  float64
}

const (
	minSourceCandidates  = 20
	sourceCandidateRatio = 20
	reservedCores        = 1
	minChunksPerWorker   = 200
)

func Search(db string, query []float32, topK int) ([]Hit, error) {
	if topK <= 0 || len(query) == 0 {
		return nil, nil
	}

	bucket, err := cache.bucket(db)
	if err != nil {
		return nil, err
	}

	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	if len(bucket.sourceVectors) == 0 {
		return bucket.searchLocked(query, topK), nil
	}

	source := min(max(len(bucket.sourceVectors)/sourceCandidateRatio, minSourceCandidates), len(bucket.sourceVectors))
	soruceHits := make([]SourceHit, 0, len(bucket.sourceVectors))
	for s, vectors := range bucket.sourceVectors {
		if len(vectors) != len(query) {
			continue
		}
		soruceHits = append(soruceHits, SourceHit{
			source: s,
			score:  cosine(query, vectors),
		})
	}
	sort.Slice(soruceHits, func(i, j int) bool {
		return soruceHits[i].score > soruceHits[j].score
	})

	source = min(source, len(soruceHits))
	soruceHits = soruceHits[:source]

	ids := make([]int64, 0, source*8)
	for _, sh := range soruceHits {
		ids = append(ids, bucket.sourceChunks[sh.source]...)
	}

	hits := bucket.getHits(query, ids)
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})
	topK = min(topK, len(hits))
	return hits[:topK], nil
}

func (b *Bucket) getHits(query []float32, ids []int64) []Hit {
	workers := max(runtime.NumCPU()-reservedCores, 1)
	chunkCap := max(len(ids)/minChunksPerWorker, 1)
	workers = min(workers, chunkCap, len(ids))
	chunkSize := (len(ids) + workers - 1) / workers

	partial := make([][]Hit, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := min(start+chunkSize, len(ids))
		if start >= end {
			continue
		}
		wg.Add(1)
		go func(idx int, slice []int64) {
			defer wg.Done()
			partial[idx] = b.getHit(query, slice)
		}(w, ids[start:end])
	}
	wg.Wait()

	total := 0
	for _, p := range partial {
		total += len(p)
	}
	hits := make([]Hit, 0, total)
	for _, p := range partial {
		hits = append(hits, p...)
	}
	return hits
}

func (b *Bucket) getHit(query []float32, ids []int64) []Hit {
	hits := make([]Hit, 0, len(ids))
	for _, id := range ids {
		vector, ok := b.idVectors[id]
		if !ok || len(vector) != len(query) {
			continue
		}
		hits = append(hits, Hit{
			ID:    id,
			Score: cosine(query, vector),
		})
	}
	return hits
}

func (b *Bucket) searchLocked(query []float32, topK int) []Hit {
	if len(b.idVectors) == 0 {
		return nil
	}

	hits := make([]Hit, 0, len(b.idVectors))
	for id, v := range b.idVectors {
		if len(v) != len(query) {
			continue
		}
		hits = append(hits, Hit{ID: id, Score: cosine(query, v)})
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})

	topK = min(topK, len(hits))
	return hits[:topK]
}

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i, x := range a {
		y := b[i]
		fx, fy := float64(x), float64(y)
		dot += fx * fy
		na += fx * fx
		nb += fy * fy
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
