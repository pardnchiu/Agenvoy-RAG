package vector

import "math"

func Rebuild(db, source string) error {
	if source == "" {
		return nil
	}

	bucket, err := cache.bucket(db)
	if err != nil {
		return err
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.rebuild(source)
	return nil
}

func RebuildAll(db string) error {
	bucket, err := cache.bucket(db)
	if err != nil {
		return err
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for source := range bucket.sourceChunks {
		bucket.rebuild(source)
	}
	return nil
}

func (b *Bucket) rebuild(source string) {
	ids := b.sourceChunks[source]
	if len(ids) == 0 {
		delete(b.sourceVectors, source)
		return
	}

	var dim int
	for _, id := range ids {
		if v, ok := b.idVectors[id]; ok && len(v) > 0 {
			dim = len(v)
			break
		}
	}
	if dim == 0 {
		delete(b.sourceVectors, source)
		return
	}

	sum := make([]float64, dim)
	num := 0
	for _, id := range ids {
		vector, ok := b.idVectors[id]
		if !ok || len(vector) != dim {
			continue
		}
		for i, x := range vector {
			sum[i] += float64(x)
		}
		num++
	}
	if num == 0 {
		delete(b.sourceVectors, source)
		return
	}

	var norm float64
	for _, s := range sum {
		norm += s * s
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		delete(b.sourceVectors, source)
		return
	}

	vectors := make([]float32, dim)
	for i, s := range sum {
		vectors[i] = float32(s / norm)
	}
	b.sourceVectors[source] = vectors
}
