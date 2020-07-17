package cache

import (
	"sync"
	"time"
)

type cacheItem struct {
	eol       time.Time
	storeTime time.Time
	data      interface{}
}

// Cache dns results
type Cache struct {
	mu     sync.RWMutex
	caches map[string]map[uint16]map[uint16]*cacheItem
	done   chan<- bool
}

// NewCache return new Cache obj
func NewCache() (cache *Cache) {
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool, 0)

	cache = &Cache{
		caches: make(map[string]map[uint16]map[uint16]*cacheItem),
		done:   done,
	}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cache.clean()
			case <-done:
				return
			}
		}
	}()

	return
}

// Get item from cache, got nil if no cache available
func (cache *Cache) Get(keyName string, keyType, keyClass uint16) (interface{}, time.Duration) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	now := time.Now()

	if nameDict, ok := cache.caches[keyName]; ok {
		if typeDict, ok := nameDict[keyType]; ok {
			if item, ok := typeDict[keyClass]; ok {
				if item.eol.After(now) {
					return item.data, now.Sub(item.storeTime)
				}
			}
		}
	}
	return nil, 0
}

// SetDataTTL set item into cache with ttl
func (cache *Cache) SetDataTTL(keyName string, keyType, keyClass uint16, data interface{}, ttl time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, ok := cache.caches[keyName]; !ok {
		cache.caches[keyName] = make(map[uint16]map[uint16]*cacheItem)
	}
	nameDict := cache.caches[keyName]
	if _, ok := nameDict[keyType]; !ok {
		nameDict[keyType] = make(map[uint16]*cacheItem)
	}
	typeDict := nameDict[keyType]

	now := time.Now()
	typeDict[keyClass] = &cacheItem{
		storeTime: now,
		eol:       now.Add(ttl),
		data:      data,
	}
}

func (cache *Cache) clean() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now()

	for keyName, nameDict := range cache.caches {
		for keyType, typeDict := range nameDict {
			for keyClass, item := range typeDict {
				if item.eol.Before(now) {
					delete(typeDict, keyClass)
				}
			}
			if len(typeDict) == 0 {
				delete(nameDict, keyType)
			}
		}
		if len(nameDict) == 0 {
			delete(cache.caches, keyName)
		}
	}
}

// Destroy caches, stop cleaning tick
func (cache *Cache) Destroy() {
	close(cache.done)
	for keyName := range cache.caches {
		delete(cache.caches, keyName)
	}
}
