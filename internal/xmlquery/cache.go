package xmlquery

import (
	"sync"

	"github.com/beevik/etree"
)

// QueryCache provides caching for compiled XPath expressions
type QueryCache struct {
	cache   map[string]etree.Path
	mutex   sync.RWMutex
	maxSize int
}

// NewQueryCache creates a new query cache with the specified maximum size
func NewQueryCache(maxSize int) *QueryCache {
	if maxSize <= 0 {
		maxSize = 100 // default cache size
	}

	return &QueryCache{
		cache:   make(map[string]etree.Path),
		maxSize: maxSize,
	}
}

// Get retrieves a compiled path from cache
func (qc *QueryCache) Get(query string) (etree.Path, bool) {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	path, exists := qc.cache[query]
	return path, exists
}

// Put stores a compiled path in cache
func (qc *QueryCache) Put(query string, path etree.Path) {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	// Simple eviction: if cache is full, remove a random entry
	if len(qc.cache) >= qc.maxSize {
		// Remove first entry found (random eviction)
		for k := range qc.cache {
			delete(qc.cache, k)
			break
		}
	}

	qc.cache[query] = path
}

// Clear removes all cached queries
func (qc *QueryCache) Clear() {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	qc.cache = make(map[string]etree.Path)
}

// Size returns the current number of cached queries
func (qc *QueryCache) Size() int {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	return len(qc.cache)
}
