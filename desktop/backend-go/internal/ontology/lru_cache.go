package ontology

import (
	"container/list"
	"sync"
)

// LRUCache is a thread-safe Least Recently Used cache
type LRUCache struct {
	mu      sync.Mutex
	maxSize int
	cache   map[string]*list.Element
	order   *list.List // Doubly-linked list for LRU ordering
}

// CacheEntry holds a key-value pair
type CacheEntry struct {
	key   string
	value interface{}
}

// NewLRUCache creates a new LRU cache with the specified max size
func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &LRUCache{
		maxSize: maxSize,
		cache:   make(map[string]*list.Element, maxSize),
		order:   list.New(),
	}
}

// Get retrieves a value from the cache
// Returns the value and a boolean indicating if it was found
// Moves the accessed item to the front (most recently used)
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(elem)

	entry := elem.Value.(*CacheEntry)
	return entry.value, true
}

// Put adds or updates a key-value pair in the cache
// If the cache exceeds maxSize, the least recently used item is evicted
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update it
	if elem, exists := c.cache[key]; exists {
		c.order.MoveToFront(elem)
		elem.Value.(*CacheEntry).value = value
		return
	}

	// Add new entry to front
	entry := &CacheEntry{key: key, value: value}
	elem := c.order.PushFront(entry)
	c.cache[key] = elem

	// Evict least recently used if over capacity
	if c.order.Len() > c.maxSize {
		lastElem := c.order.Back()
		c.order.Remove(lastElem)
		delete(c.cache, lastElem.Value.(*CacheEntry).key)
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.order.Remove(elem)
		delete(c.cache, key)
	}
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element, c.maxSize)
	c.order = list.New()
}

// Size returns the current number of entries in the cache
func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cache)
}

// Capacity returns the maximum cache size
func (c *LRUCache) Capacity() int {
	return c.maxSize
}

// Keys returns all keys currently in the cache (in LRU order, most recent first)
func (c *LRUCache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.cache))
	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(*CacheEntry).key)
	}
	return keys
}

// Stats returns cache statistics
func (c *LRUCache) Stats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	return map[string]interface{}{
		"size":     len(c.cache),
		"capacity": c.maxSize,
		"usage":    float64(len(c.cache)) / float64(c.maxSize) * 100,
	}
}
