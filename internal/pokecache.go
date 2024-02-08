package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time //holds the time of caching
	val       []byte    //holds the data we are caching
}

type Cache struct {
	mu sync.Mutex            //goroutine lock for map access
	m  map[string]cacheEntry //the cache itself, the key is the url, the value is the data and time
}

func NewCache(interval time.Duration) {
	cache := Cache{m: make(map[string]cacheEntry)}

	go cache.reapLoop(interval)
}

func (c *Cache) Add() {

}

func (c *Cache) Get() {

}

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
