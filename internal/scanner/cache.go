package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vroom/internal/config"
)

type cacheEntry struct {
	Size      int64     `json:"size"`
	Count     int       `json:"count"`
	ModTime   time.Time `json:"mod_time"`
	InfoSize  int64     `json:"info_size"`
	ScannedAt time.Time `json:"scanned_at"`
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	path    string
	dirty   bool
}

func cachePath() string {
	return filepath.Join(config.DataDir(), "cache.json")
}

func DefaultCache() *Cache {
	c := &Cache{entries: map[string]cacheEntry{}, path: cachePath()}
	_ = c.Load()
	return c
}

func (c *Cache) Load() error {
	if c == nil {
		return nil
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, &c.entries)
}

func (c *Cache) Save() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.dirty {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return err
	}
	c.dirty = false
	return nil
}

func (c *Cache) Lookup(path string, mod time.Time, infoSize int64) (cacheEntry, bool) {
	if c == nil {
		return cacheEntry{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[path]
	if !ok {
		return cacheEntry{}, false
	}
	if !e.ModTime.Equal(mod) || e.InfoSize != infoSize {
		return cacheEntry{}, false
	}
	return e, true
}

func (c *Cache) Store(path string, mod time.Time, infoSize, size int64, count int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = cacheEntry{
		Size:      size,
		Count:     count,
		ModTime:   mod,
		InfoSize:  infoSize,
		ScannedAt: time.Now(),
	}
	c.dirty = true
}
