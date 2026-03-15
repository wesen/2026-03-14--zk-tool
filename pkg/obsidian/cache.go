package obsidian

import "sync"

// Cache stores note contents by canonical path.
type Cache struct {
	mu      sync.RWMutex
	content map[string]string
}

// NewCache creates an empty note cache.
func NewCache() *Cache {
	return &Cache{
		content: map[string]string{},
	}
}

// Get returns one cached note body by path.
func (c *Cache) Get(path string) (string, bool) {
	if c == nil {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.content[path]
	return value, ok
}

// Set stores one note body in the cache.
func (c *Cache) Set(path string, value string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.content[path] = value
}

// Delete removes one cached note.
func (c *Cache) Delete(path string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.content, path)
}

// Clear removes all cached note bodies.
func (c *Cache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.content = map[string]string{}
}
