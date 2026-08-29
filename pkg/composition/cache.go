package composition

import (
	"sync"

	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

// Cache stores immutable fragments by promoted component input hash.
type Cache interface {
	Get(key string) (Fragment, bool)
	Put(key string, fragment Fragment)
}

// MemoryCache is a race-safe content cache for long-lived composers and
// preview sessions.
type MemoryCache struct {
	mu        sync.RWMutex
	fragments map[string]Fragment
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{fragments: make(map[string]Fragment)}
}

func (cache *MemoryCache) Get(key string) (Fragment, bool) {
	if cache == nil {
		return Fragment{}, false
	}
	cache.mu.RLock()
	fragment, exists := cache.fragments[key]
	cache.mu.RUnlock()
	if !exists {
		return Fragment{}, false
	}
	return cloneFragment(fragment), true
}

func (cache *MemoryCache) Put(key string, fragment Fragment) {
	if cache == nil {
		return
	}
	cache.mu.Lock()
	cache.fragments[key] = cloneFragment(fragment)
	cache.mu.Unlock()
}

func cloneFragment(fragment Fragment) Fragment {
	fragment.Content = append([]byte(nil), fragment.Content...)
	fragment.Assets = append([]document.AssetRef(nil), fragment.Assets...)
	fragment.Children = append([]string(nil), fragment.Children...)
	fragment.Dependencies = append([]string(nil), fragment.Dependencies...)
	return fragment
}
