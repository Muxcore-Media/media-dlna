package probe

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/anacrolix/ffprobe"
)

type persistedItem struct {
	Path string        `json:"path"`
	Info *ffprobe.Info `json:"info"`
}

// Cache persists ffprobe metadata for DLNA browsing.
type Cache struct {
	mu   sync.Mutex
	path string
	data map[string]*ffprobe.Info
}

// NewCache loads an on-disk ffprobe cache when path is set.
func NewCache(path string) (*Cache, error) {
	c := &Cache{
		path: path,
		data: make(map[string]*ffprobe.Info),
	}
	if path == "" {
		return c, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := c.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return c, nil
}

func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k, ok := cacheKey(key)
	if !ok {
		return nil, false
	}
	v, found := c.data[k]
	return v, found
}

func (c *Cache) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k, ok := cacheKey(key)
	if !ok {
		return
	}
	if info, ok := value.(*ffprobe.Info); ok {
		c.data[k] = info
	}
}

// Save writes the cache atomically to disk.
func (c *Cache) Save() error {
	if c.path == "" {
		return nil
	}
	c.mu.Lock()
	items := make([]persistedItem, 0, len(c.data))
	for path, info := range c.data {
		items = append(items, persistedItem{Path: path, Info: info})
	}
	c.mu.Unlock()

	tmp, err := os.CreateTemp(filepath.Dir(c.path), filepath.Base(c.path)+".*")
	if err != nil {
		return err
	}
	if err := json.NewEncoder(tmp).Encode(items); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), c.path)
}

func (c *Cache) load() error {
	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	var items []persistedItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return err
	}
	c.mu.Lock()
	for _, item := range items {
		if item.Info == nil || item.Path == "" {
			continue
		}
		c.data[item.Path] = item.Info
	}
	c.mu.Unlock()
	return nil
}

func cacheKey(key interface{}) (string, bool) {
	switch k := key.(type) {
	case string:
		return k, k != ""
	default:
		b, err := json.Marshal(k)
		if err != nil {
			return "", false
		}
		var tmp struct {
			Path string `json:"Path"`
		}
		if err := json.Unmarshal(b, &tmp); err != nil || tmp.Path == "" {
			return "", false
		}
		return tmp.Path, true
	}
}
