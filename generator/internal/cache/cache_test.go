package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCache_SaveAndLoad(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "scug_cache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cachePath := filepath.Join(tmpDir, "scug_cache.json")

	// 1. Create a cache, set values, save
	c1 := NewCache()
	c1.SetMeta("test_guid", ComponentInfo{ClassName: "MyClass", IsPublic: true}, "path/to/meta", 123456)
	c1.SetFileHash("my_prefab.prefab", "hash123")
	c1.MarkFileGenerated("path/to/out.cs")
	c1.Save(cachePath)

	// 2. Load the cache in a new instance and verify
	c2 := LoadCache(cachePath)

	info, exists := c2.GetMeta("test_guid")
	if !exists || info.ClassName != "MyClass" {
		t.Errorf("GetMeta failed: expected MyClass, got %+v", info)
	}

	if !c2.IsMetaUpToDate("path/to/meta", 123456) {
		t.Errorf("IsMetaUpToDate failed, expected true")
	}
	
	if c2.IsMetaUpToDate("path/to/meta", 999999) {
		t.Errorf("IsMetaUpToDate failed, expected false")
	}

	hash := c2.GetFileHash("my_prefab.prefab")
	if hash != "hash123" {
		t.Errorf("GetFileHash failed, got %s", hash)
	}

	// GeneratedFiles shouldn't be serialized in JSON, it should be empty
	if c2.IsFileGenerated("path/to/out.cs") {
		t.Errorf("GeneratedFiles should not persist between loads as per JSON tag")
	}

	// 3. Test Loading non-existent cache
	c3 := LoadCache(filepath.Join(tmpDir, "does_not_exist.json"))
	if c3.MetaMapping == nil {
		t.Errorf("LoadCache should initialize maps even if file doesn't exist")
	}
}
