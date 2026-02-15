package aggregator

import (
	"fmt"
	"sync"
	"time"

	"flight-aggregator/models"
)

// Production-ready in-memory cache with expiration and size limit

type cacheEntry struct {
	value     models.SearchResponse
	expiresAt time.Time
}

type aggregatorCache struct {
	mu      sync.RWMutex
	store   map[string]cacheEntry
	maxSize int
	ttl     time.Duration
	order   []string // FIFO order for eviction
}

func newAggregatorCache(maxSize int, ttl time.Duration) *aggregatorCache {
	return &aggregatorCache{
		store:   make(map[string]cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		order:   make([]string, 0, maxSize),
	}
}

func (c *aggregatorCache) Get(key string) (models.SearchResponse, bool) {
	c.mu.RLock()
	entry, ok := c.store[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.store, key)
			c.removeOrder(key)
			c.mu.Unlock()
		}
		return models.SearchResponse{}, false
	}
	return entry.value, true
}

func (c *aggregatorCache) Set(key string, value models.SearchResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.store) >= c.maxSize {
		// Evict oldest
		oldest := c.order[0]
		delete(c.store, oldest)
		c.order = c.order[1:]
	}
	c.store[key] = cacheEntry{value: value, expiresAt: time.Now().Add(c.ttl)}
	c.order = append(c.order, key)
}

func (c *aggregatorCache) removeOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

var aggCache = newAggregatorCache(1000, 5*time.Minute) // 1000 entries, 5 min TTL

func cacheKey(req models.SearchRequest) string {
	// Use a simple key: all fields concatenated (for demo, not for production)
	return fmt.Sprintf("%s|%s|%s|%s|%s|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v",
		req.Origin, req.Destination, req.DepartureDate, req.Passengers, req.CabinClass,
		req.MinPrice, req.MaxPrice, req.MinStops, req.MaxStops,
		req.DepartureTimeStart, req.DepartureTimeEnd, req.ArrivalTimeStart, req.ArrivalTimeEnd,
		req.Airlines, req.MinDurationMinutes, req.MaxDurationMinutes, req.SortBy, req.ReturnDate)
}