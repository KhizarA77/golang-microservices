package cache

import (
	"product-service/models"
	"sync"
	"time"
)

type CacheEntry struct {
	product  *models.Product
	cachedAt time.Time
}

type ProductCache struct {
	mu    sync.RWMutex
	cache map[int]CacheEntry
	ttl   time.Duration
}

func NewProductCache() *ProductCache {
	return &ProductCache{
		cache: make(map[int]CacheEntry),
		ttl:   30 * time.Second,
	}
}

func (c *ProductCache) Get(id int) (*models.Product, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache[id]
	if !ok || time.Since(entry.cachedAt) > c.ttl {
		return nil, false
	}
	return entry.product, true
}

func (c *ProductCache) Set(id int, p *models.Product) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = CacheEntry{product: p, cachedAt: time.Now()}
}

func (c *ProductCache) Invalidate(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, id)
}
