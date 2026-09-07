package probe

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/ffprobe"
)

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c, err := NewCache(path)
	if err != nil {
		t.Fatal(err)
	}
	info := &ffprobe.Info{Format: map[string]interface{}{"format_name": "matroska"}}
	c.Set("movie.mkv", info)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := NewCache(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := loaded.Get("movie.mkv")
	if !ok {
		t.Fatal("expected cache hit")
	}
	got, ok := v.(*ffprobe.Info)
	if !ok || got.Format["format_name"] != "matroska" {
		t.Fatalf("value: %#v", v)
	}
}

func TestCacheMissingFileIsSoftEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "cache.json")
	c, err := NewCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get("nope"); ok {
		t.Fatal("expected miss")
	}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("cache file: %v", err)
	}
}
