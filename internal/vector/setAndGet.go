package vector

import "slices"

func Set(db string, id int64, source string, vectors []float32) error {
	bucket, err := cache.bucket(db)
	if err != nil {
		return err
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if oldSource, ok := bucket.idSource[id]; ok && oldSource != source {
		bucket.remove(id, oldSource)
	}

	bucket.idVectors[id] = vectors
	bucket.idSource[id] = source

	if source == "" || slices.Contains(bucket.sourceChunks[source], id) {
		return nil
	}

	bucket.sourceChunks[source] = append(bucket.sourceChunks[source], id)
	return nil
}

func (b *Bucket) remove(id int64, source string) {
	ids := b.sourceChunks[source]
	if i := slices.Index(ids, id); i >= 0 {
		b.sourceChunks[source] = slices.Delete(ids, i, i+1)
	}

	if len(b.sourceChunks[source]) == 0 {
		delete(b.sourceChunks, source)
		delete(b.sourceVectors, source)
	}
}

func Get(db string, id int64) ([]float32, bool, error) {
	bucket, err := cache.bucket(db)
	if err != nil {
		return nil, false, err
	}

	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	vector, ok := bucket.idVectors[id]
	return vector, ok, nil
}
