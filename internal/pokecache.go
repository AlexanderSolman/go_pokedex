package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time //holds the time of caching
	val       []byte    //holds the data we are caching
	next      string    //next url
	prev      string    //previous url
}

type Cache struct {
	mu sync.Mutex            //goroutine lock for map access
	m  map[string]cacheEntry //the cache itself, the key is the url, the value is the data and time
}

// Creates a new cache and returns a pointer of the cache to the caller
func NewCache(interval time.Duration) *Cache {
	cache := &Cache{m: make(map[string]cacheEntry)}
	go cache.reapLoop(interval)
	return cache
}

// Add data to the cache, takes a url as key and data as val
func (c *Cache) Add(key string, val []byte, next string, prev string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
		next:      next,
		prev:      prev,
	}
}

// Returns data if url as key exists and nil,false otherwise
func (c *Cache) Get(key string) ([]byte, bool, string, string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if i, ok := c.m[key]; !ok {
		return nil, false, i.next, i.prev
	} else {
		return i.val, true, i.next, i.prev
	}
}

// Checks if data gets too old and deletes it if thats the case
func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.mu.Lock()
		for key, t := range c.m {
			if time.Since(t.createdAt) >= interval {
				delete(c.m, key)
			}
		}
		c.mu.Unlock()
	}
}
