package vector

import (
	"fmt"
	"sync"
)

var (
	once  sync.Once
	cache *Cache
)

type Cache struct {
	mu        sync.RWMutex
	dbBuckets map[string]*Bucket
}

type Bucket struct {
	mu            sync.RWMutex
	idVectors     map[int64][]float32
	idSource      map[int64]string
	sourceChunks  map[string][]int64
	sourceVectors map[string][]float32
}

func New() {
	once.Do(func() {
		cache = &Cache{dbBuckets: make(map[string]*Bucket)}
	})
}

func Check() bool {
	return cache != nil
}

func InitBucket(db string) error {
	if db == "" {
		return fmt.Errorf("db is required")
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, ok := cache.dbBuckets[db]; !ok {
		cache.dbBuckets[db] = &Bucket{
			idVectors:     make(map[int64][]float32),
			idSource:      make(map[int64]string),
			sourceChunks:  make(map[string][]int64),
			sourceVectors: make(map[string][]float32),
		}
	}
	return nil
}

func (c *Cache) bucket(db string) (*Bucket, error) {
	if db == "" {
		return nil, fmt.Errorf("db is required")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	bucket, ok := c.dbBuckets[db]
	if !ok {
		return nil, fmt.Errorf("bucket %q not initialized", db)
	}
	return bucket, nil
}
